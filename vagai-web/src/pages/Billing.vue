<template>
  <div>
    <div class="mb-8">
      <h1 class="text-2xl font-bold font-outfit text-white">Billing</h1>
      <p class="text-slate-400 mt-1">Gerencie seu plano e pagamentos</p>
    </div>

    <div v-if="meLoading" class="text-slate-400 text-center py-12">Carregando...</div>
    <div v-else-if="meError" class="p-4 bg-red-500/10 border border-red-500/20 rounded-lg text-red-400 text-sm">
      Erro ao carregar dados do plano.
    </div>

    <template v-else-if="meData">
      <div class="glass-card p-6 mb-8">
        <div class="flex items-center justify-between">
          <div>
            <div class="text-sm text-slate-400">Plano atual</div>
            <div class="text-2xl font-bold text-white mt-1">{{ planName }}</div>
            <div class="text-sm text-slate-400 mt-1">
              {{ planPrice === 0 ? 'Grátis' : `R$ ${planPrice / 100}/mês` }}
            </div>
          </div>
          <button v-if="upgradePlan" @click="handleChangePlan(upgradePlan)" :disabled="changingPlan" class="btn-primary flex items-center gap-2">
            <Loader2 v-if="changingPlan" class="w-4 h-4 animate-spin" />
            Upgrade para {{ upgradePlan.name }}
          </button>
        </div>

        <div class="mt-6 pt-6 border-t border-white/5 grid grid-cols-3 gap-6">
          <div>
            <div class="text-xs text-slate-500">Vagas</div>
            <div class="text-lg font-bold text-white mt-1">{{ usage.jobs }} / {{ maxJobs }}</div>
            <div class="w-full bg-slate-700 rounded-full h-2 mt-2">
              <div class="bg-indigo-500 h-2 rounded-full" :style="{ width: usageBar(usage.jobs, maxJobsRaw) + '%' }"></div>
            </div>
          </div>
          <div>
            <div class="text-xs text-slate-500">Análises de currículo / mês</div>
            <div class="text-lg font-bold text-white mt-1">{{ usage.resume_analyses ?? 0 }} / {{ maxResumeAnalyses }}</div>
            <div class="w-full bg-slate-700 rounded-full h-2 mt-2">
              <div class="bg-indigo-500 h-2 rounded-full" :style="{ width: usageBar(usage.resume_analyses ?? 0, maxResumeAnalysesRaw) + '%' }"></div>
            </div>
          </div>
          <div>
            <div class="text-xs text-slate-500">Sites</div>
            <div class="text-lg font-bold text-white mt-1">{{ usage.sites }} / {{ maxSites }}</div>
            <div class="w-full bg-slate-700 rounded-full h-2 mt-2">
              <div class="bg-indigo-500 h-2 rounded-full" :style="{ width: usageBar(usage.sites, maxSitesRaw) + '%' }"></div>
            </div>
          </div>
        </div>
      </div>

      <h2 class="text-lg font-semibold text-white mb-4">Planos disponíveis</h2>
      <div v-if="plansLoading" class="text-slate-500 text-sm">Carregando planos...</div>
      <div v-else-if="plansError" class="text-slate-500 text-sm">Não foi possível carregar planos.</div>
      <div v-else class="grid grid-cols-2 gap-6">
        <div v-for="plan in plansData" :key="plan.slug" :class="['glass-card p-6', planSlug === plan.slug ? 'ring-2 ring-indigo-500' : '']">
          <div class="text-sm text-slate-400 mb-2">{{ plan.name }}</div>
          <div class="text-3xl font-bold text-white">
            {{ plan.price_monthly === 0 ? 'Grátis' : `R$ ${plan.price_monthly / 100}` }}
            <span v-if="plan.price_monthly > 0" class="text-sm font-normal text-slate-400">/mês</span>
          </div>
          <ul class="mt-4 space-y-2">
            <li v-for="perk in planPerks(plan)" :key="perk" class="text-sm text-slate-300 flex items-center gap-2">
              <Check class="w-4 h-4 text-emerald-400" />
              {{ perk }}
            </li>
          </ul>
          <button v-if="planSlug !== plan.slug" @click="handleChangePlan(plan)" :disabled="changingPlan" class="w-full mt-6 py-3 bg-slate-700 hover:bg-slate-600 disabled:opacity-50 text-white font-medium rounded-lg transition-colors flex items-center justify-center gap-2">
            <Loader2 v-if="changingPlan" class="w-4 h-4 animate-spin" />
            {{ planActionLabel(plan) }}
          </button>
          <button v-else class="w-full mt-6 py-3 bg-indigo-600 text-white font-medium rounded-lg cursor-default">
            Plano atual
          </button>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { Check, Loader2 } from 'lucide-vue-next'
import { useMe, usePlans, changePlan as changePlanApi } from '../services/api'

const { data: meData, isLoading: meLoading, isError: meError } = useMe()
const { data: plansData, isLoading: plansLoading, isError: plansError } = usePlans()

const plans = computed(() => plansData.value || [])

const planName = computed(() => meData.value?.plan?.name || '-')
const planSlug = computed(() => meData.value?.plan?.slug || meData.value?.user?.plan || 'free')
const planPrice = computed(() => meData.value?.plan?.price_monthly ?? 0)

const upgradePlan = computed(() => {
  return plans.value
    .filter((p) => p.slug !== planSlug.value && p.price_monthly > planPrice.value)
    .sort((a, b) => a.price_monthly - b.price_monthly)[0] || null
})

const maxJobsRaw = computed(() => meData.value?.plan?.max_jobs ?? -1)
const maxResumeAnalysesRaw = computed(() => meData.value?.plan?.max_resume_analyses ?? -1)
const maxSitesRaw = computed(() => meData.value?.plan?.max_sites ?? -1)

const maxJobs = computed(() => maxJobsRaw.value === -1 ? '∞' : maxJobsRaw.value)
const maxResumeAnalyses = computed(() => maxResumeAnalysesRaw.value === -1 ? '∞' : maxResumeAnalysesRaw.value)
const maxSites = computed(() => maxSitesRaw.value === -1 ? '∞' : maxSitesRaw.value)

const usage = computed(() => meData.value?.usage || { jobs: 0, resumes: 0, sites: 0 })

const changingPlan = ref(false)

async function handleChangePlan(plan) {
  const target = typeof plan === 'string' ? { slug: plan, name: plan } : plan
  if (!confirm(`Deseja alterar para o plano ${target.name}?`)) return
  changingPlan.value = true
  try {
    await changePlanApi(target.slug)
    window.location.reload()
  } catch (err) {
    alert(err.response?.data?.error || 'Erro ao alterar plano')
  } finally {
    changingPlan.value = false
  }
}

function planActionLabel(plan) {
  if (plan.price_monthly > planPrice.value) return 'Fazer upgrade'
  if (plan.price_monthly === 0) return 'Mudar para grátis'
  if (plan.price_monthly < planPrice.value) return 'Fazer downgrade'
  return 'Mudar para este plano'
}

function planPerks(plan) {
  const perks = [
    limitLabel(plan.max_jobs, 'vaga', 'vagas', 'Vagas ilimitadas'),
    limitLabel(plan.max_resume_analyses, 'análise de currículo/mês', 'análises de currículo/mês', 'Análises de currículo ilimitadas'),
    limitLabel(plan.max_sites, 'fonte de busca', 'fontes de busca', 'Fontes de busca ilimitadas'),
  ]
  if (plan.ai_matching) perks.push('Matching por IA')
  if (plan.interview_prep) perks.push('Preparação')
  const seen = new Set()
  return perks.concat(parseFeatures(plan.features)).filter((perk) => {
    const key = perk.toLowerCase()
    if (seen.has(key)) return false
    seen.add(key)
    return true
  })
}

function limitLabel(value, singular, plural, unlimited) {
  if (value === -1) return unlimited
  return `${value} ${value === 1 ? singular : plural}`
}

function usageBar(current, max) {
  if (max === -1) return 0
  return Math.min(100, (current / max) * 100)
}

function parseFeatures(features) {
  if (!features) return []
  try {
    const parsed = JSON.parse(features)
    return Array.isArray(parsed) ? parsed : []
  } catch {
    return []
  }
}
</script>
