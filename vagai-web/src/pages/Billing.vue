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
          <button v-if="planSlug === 'free'" @click="handleChangePlan('pro')" :disabled="changingPlan" class="btn-primary flex items-center gap-2">
            <Loader2 v-if="changingPlan" class="w-4 h-4 animate-spin" />
            Upgrade para Pro
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
            <div class="text-xs text-slate-500">Currículos</div>
            <div class="text-lg font-bold text-white mt-1">{{ usage.resumes }} / {{ maxResumes }}</div>
            <div class="w-full bg-slate-700 rounded-full h-2 mt-2">
              <div class="bg-indigo-500 h-2 rounded-full" :style="{ width: usageBar(usage.resumes, maxResumesRaw) + '%' }"></div>
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
            <li v-for="feature in parseFeatures(plan.features)" :key="feature" class="text-sm text-slate-300 flex items-center gap-2">
              <Check class="w-4 h-4 text-emerald-400" />
              {{ feature }}
            </li>
          </ul>
          <button v-if="planSlug !== plan.slug" @click="handleChangePlan(plan.slug)" :disabled="changingPlan" class="w-full mt-6 py-3 bg-slate-700 hover:bg-slate-600 disabled:opacity-50 text-white font-medium rounded-lg transition-colors flex items-center justify-center gap-2">
            <Loader2 v-if="changingPlan" class="w-4 h-4 animate-spin" />
            {{ planSlug === 'free' ? 'Começar' : 'Downgrade' }}
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

const planName = computed(() => meData.value?.plan?.name || meData.value?.organization?.plan || '-')
const planSlug = computed(() => meData.value?.plan?.slug || meData.value?.organization?.plan || 'free')
const planPrice = computed(() => meData.value?.plan?.price_monthly ?? 0)

const maxJobsRaw = computed(() => meData.value?.plan?.max_jobs ?? -1)
const maxResumesRaw = computed(() => meData.value?.plan?.max_resumes ?? -1)
const maxSitesRaw = computed(() => meData.value?.plan?.max_sites ?? -1)

const maxJobs = computed(() => maxJobsRaw.value === -1 ? '∞' : maxJobsRaw.value)
const maxResumes = computed(() => maxResumesRaw.value === -1 ? '∞' : maxResumesRaw.value)
const maxSites = computed(() => maxSitesRaw.value === -1 ? '∞' : maxSitesRaw.value)

const usage = computed(() => meData.value?.usage || { jobs: 0, resumes: 0, sites: 0 })

const changingPlan = ref(false)

async function handleChangePlan(slug) {
  const label = slug === 'free' ? 'Free' : 'Pro'
  if (!confirm(`Deseja alterar para o plano ${label}?`)) return
  changingPlan.value = true
  try {
    await changePlanApi(slug)
    window.location.reload()
  } catch (err) {
    alert(err.response?.data?.error || 'Erro ao alterar plano')
  } finally {
    changingPlan.value = false
  }
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
