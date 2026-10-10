<template>
  <div class="space-y-10">
    <div class="flex items-start justify-between">
      <div>
        <h1 class="text-4xl font-bold text-white mb-2 font-outfit tracking-tight">Planos</h1>
        <p class="text-slate-400">Crie e edite os planos disponíveis na plataforma.</p>
      </div>
      <button
        @click="openCreate"
        class="h-11 px-6 bg-indigo-600 hover:bg-indigo-500 text-white font-bold rounded-xl transition-all shadow-lg shadow-indigo-900/20 active:scale-95 flex items-center gap-2"
      >
        <Plus class="w-4 h-4" />
        Novo plano
      </button>
    </div>

    <div v-if="plansLoading" class="text-slate-400 py-12 text-center">Carregando planos...</div>
    <div v-else-if="plansError" class="p-4 bg-red-500/10 border border-red-500/20 rounded-lg text-red-400 text-sm">
      Erro ao carregar planos.
    </div>

    <template v-else>
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div v-for="plan in plans" :key="plan.id" class="glass-card p-6 space-y-4">
          <div class="flex items-start justify-between gap-3">
            <div>
              <div class="flex items-center gap-2">
                <h2 class="text-xl font-bold text-white font-outfit">{{ plan.name }}</h2>
                <span
                  class="px-2 py-0.5 rounded-full text-[11px] font-bold uppercase tracking-wider"
                  :class="plan.type === 'paid' ? 'bg-amber-500/15 text-amber-300' : 'bg-emerald-500/15 text-emerald-300'"
                >
                  {{ plan.type === 'paid' ? 'Pago' : 'Grátis' }}
                </span>
                <span v-if="plan.is_default" class="px-2 py-0.5 rounded-full text-[11px] font-bold uppercase tracking-wider bg-indigo-500/20 text-indigo-300">
                  Padrão
                </span>
              </div>
              <div class="text-xs text-slate-500 mt-1 font-mono">{{ plan.slug }}</div>
            </div>
            <span
              class="px-2 py-0.5 rounded-full text-[11px] font-bold uppercase tracking-wider"
              :class="plan.active ? 'bg-emerald-500/15 text-emerald-300' : 'bg-slate-600/30 text-slate-400'"
            >
              {{ plan.active ? 'Ativo' : 'Inativo' }}
            </span>
          </div>

          <div class="text-2xl font-bold text-white">
            {{ plan.price_monthly === 0 ? 'Grátis' : `R$ ${formatPrice(plan.price_monthly)}` }}
            <span v-if="plan.price_monthly > 0" class="text-sm font-normal text-slate-400">/mês</span>
          </div>

          <ul class="space-y-1.5">
            <li v-for="perk in planPerks(plan)" :key="perk" class="text-sm text-slate-300 flex items-center gap-2">
              <Check class="w-4 h-4 text-emerald-400" />
              {{ perk }}
            </li>
          </ul>

          <div class="pt-4 border-t border-white/5 flex flex-wrap gap-2">
            <button @click="openEdit(plan)" class="px-3 py-2 bg-slate-700 hover:bg-slate-600 text-white text-sm font-medium rounded-lg transition-colors flex items-center gap-2">
              <Pencil class="w-3.5 h-3.5" />
              Editar
            </button>
            <button
              v-if="!plan.is_default"
              @click="makeDefault(plan)"
              class="px-3 py-2 bg-indigo-600/80 hover:bg-indigo-600 text-white text-sm font-medium rounded-lg transition-colors flex items-center gap-2"
            >
              <Star class="w-3.5 h-3.5" />
              Tornar padrão
            </button>
            <button
              @click="toggleActive(plan)"
              class="px-3 py-2 text-sm font-medium rounded-lg transition-colors flex items-center gap-2"
              :class="plan.active ? 'bg-slate-700 hover:bg-slate-600 text-white' : 'bg-emerald-600/80 hover:bg-emerald-600 text-white'"
            >
              <Power class="w-3.5 h-3.5" />
              {{ plan.active ? 'Desativar' : 'Ativar' }}
            </button>
          </div>
        </div>
      </div>
    </template>

    <!-- Formulário de criação/edição -->
    <div v-if="formOpen" ref="formRef" class="glass-card p-10 space-y-6">
      <div class="flex items-center justify-between">
        <h2 class="text-2xl font-bold text-white font-outfit">{{ isEditing ? 'Editar plano' : 'Novo plano' }}</h2>
        <button @click="closeForm" class="text-slate-400 hover:text-white">
          <X class="w-5 h-5" />
        </button>
      </div>

      <form @submit.prevent="submit" class="space-y-6">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-400 uppercase tracking-widest ml-1">Nome</label>
            <input v-model="form.name" type="text" class="input-field w-full h-12" placeholder="ex: Starter" />
          </div>
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-400 uppercase tracking-widest ml-1">Identificador (slug)</label>
            <input v-model="form.slug" type="text" :disabled="isEditing" class="input-field w-full h-12 disabled:opacity-60" placeholder="ex: starter" />
            <p v-if="isEditing" class="text-xs text-slate-500 ml-1">O identificador não pode ser alterado.</p>
          </div>
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-400 uppercase tracking-widest ml-1">Tipo</label>
            <select v-model="form.type" class="input-field w-full h-12">
              <option value="free">Gratuito</option>
              <option value="paid">Pago</option>
            </select>
          </div>
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-400 uppercase tracking-widest ml-1">Preço mensal (R$)</label>
            <input v-model="form.priceMonthly" type="number" min="0" step="0.01" class="input-field w-full h-12" />
          </div>
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-400 uppercase tracking-widest ml-1">Preço anual (R$)</label>
            <input v-model="form.priceYearly" type="number" min="0" step="0.01" class="input-field w-full h-12" />
          </div>
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-400 uppercase tracking-widest ml-1">Máx. vagas (-1 = ilimitado)</label>
            <input v-model="form.maxJobs" type="number" class="input-field w-full h-12" />
          </div>
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-400 uppercase tracking-widest ml-1">Análises de currículo / mês (-1 = ilimitado)</label>
            <input v-model="form.maxResumeAnalyses" type="number" class="input-field w-full h-12" />
          </div>
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-400 uppercase tracking-widest ml-1">Máx. fontes</label>
            <input v-model="form.maxSites" type="number" class="input-field w-full h-12" />
          </div>
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-400 uppercase tracking-widest ml-1">Crawls por dia</label>
            <input v-model="form.maxCrawls" type="number" class="input-field w-full h-12" />
          </div>
        </div>

        <div class="space-y-2">
          <label class="text-sm font-bold text-slate-400 uppercase tracking-widest ml-1">Recursos (um por linha)</label>
          <textarea v-model="form.features" rows="4" class="input-field w-full py-3" placeholder="500 vagas&#10;2 currículos"></textarea>
        </div>

        <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <label class="flex items-center gap-3 text-sm text-slate-300">
            <input v-model="form.aiMatching" type="checkbox" class="w-4 h-4" />
            Matching por IA
          </label>
          <label class="flex items-center gap-3 text-sm text-slate-300">
            <input v-model="form.interviewPrep" type="checkbox" class="w-4 h-4" />
            Preparação
          </label>
        </div>

        <label v-if="isEditing" class="flex items-center gap-3 text-sm text-slate-300">
          <input v-model="form.active" type="checkbox" class="w-4 h-4" />
          Plano ativo
        </label>

        <div v-if="formError" class="p-3 bg-red-500/10 border border-red-500/20 rounded-lg text-red-400 text-sm">
          {{ formError }}
        </div>
        <div class="flex items-center gap-3">
          <button type="submit" :disabled="submitting" class="h-11 px-6 bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 text-white font-bold rounded-xl transition-all flex items-center gap-2">
            <Loader2 v-if="submitting" class="w-4 h-4 animate-spin" />
            {{ isEditing ? 'Salvar alterações' : 'Criar plano' }}
          </button>
          <button type="button" @click="closeForm" class="h-11 px-6 bg-slate-700 hover:bg-slate-600 text-white font-medium rounded-xl transition-colors">
            Cancelar
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, nextTick } from 'vue'
import { useQueryClient } from '@tanstack/vue-query'
import { Plus, Pencil, Star, Check, Power, X } from 'lucide-vue-next'
import { useAdminPlans, adminCreatePlan, adminUpdatePlan } from '../services/api'

const queryClient = useQueryClient()
const { data: plansData, isLoading: plansLoading, isError: plansError } = useAdminPlans()
const plans = computed(() => plansData.value || [])

const formOpen = ref(false)
const editingId = ref(null)
const submitting = ref(false)
const formError = ref('')
const isEditing = computed(() => editingId.value !== null)

const form = ref(emptyForm())
const formRef = ref(null)

async function scrollToForm() {
  await nextTick()
  formRef.value?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

function emptyForm() {
  return {
    name: '', slug: '', type: 'free',
    priceMonthly: 0, priceYearly: 0,
    maxJobs: 1000, maxResumeAnalyses: 1, maxSites: 5, maxCrawls: 10,
    aiMatching: false, interviewPrep: false,
    features: '', active: true, isDefault: false
  }
}

function openCreate() {
  form.value = emptyForm()
  formError.value = ''
  formOpen.value = true
  scrollToForm()
}

function openEdit(plan) {
  form.value = {
    name: plan.name,
    slug: plan.slug,
    type: plan.type || 'free',
    priceMonthly: (plan.price_monthly || 0) / 100,
    priceYearly: (plan.price_yearly || 0) / 100,
    maxJobs: plan.max_jobs,
    maxResumeAnalyses: plan.max_resume_analyses,
    maxSites: plan.max_sites,
    maxCrawls: plan.max_crawls_per_day ?? 10,
    aiMatching: !!plan.ai_matching,
    interviewPrep: !!plan.interview_prep,
    features: parseFeatures(plan.features).join('\n'),
    active: plan.active,
    isDefault: plan.is_default
  }
  editingId.value = plan.id
  formError.value = ''
  formOpen.value = true
  scrollToForm()
}

function closeForm() {
  formOpen.value = false
  editingId.value = null
}

function toCents(value) {
  return Math.round((Number(value) || 0) * 100)
}

function validate() {
  if (!form.value.name.trim()) return 'Nome é obrigatório'
  if (!isEditing.value) {
    if (!form.value.slug.trim()) return 'Identificador é obrigatório'
    if (!/^[a-z0-9-]+$/.test(form.value.slug.trim())) return 'Identificador inválido: use letras minúsculas, números e hífen'
  }
  const monthly = toCents(form.value.priceMonthly)
  const yearly = toCents(form.value.priceYearly)
  if (form.value.type === 'free' && (monthly !== 0 || yearly !== 0)) return 'Plano gratuito deve ter preço zero'
  if (form.value.type === 'paid' && monthly <= 0 && yearly <= 0) return 'Plano pago exige preço maior que zero'
  return ''
}

async function submit() {
  formError.value = ''
  const error = validate()
  if (error) {
    formError.value = error
    return
  }
  const payload = {
    name: form.value.name.trim(),
    type: form.value.type,
    price_monthly: toCents(form.value.priceMonthly),
    price_yearly: toCents(form.value.priceYearly),
    max_jobs: Number(form.value.maxJobs) || 0,
    max_resume_analyses: Number(form.value.maxResumeAnalyses) || 0,
    max_sites: Number(form.value.maxSites) || 0,
    max_crawls_per_day: Number(form.value.maxCrawls) || 0,
    ai_matching: !!form.value.aiMatching,
    interview_prep: !!form.value.interviewPrep,
    features: form.value.features.split('\n').map((s) => s.trim()).filter(Boolean)
  }

  submitting.value = true
  try {
    if (isEditing.value) {
      payload.active = form.value.active
      payload.is_default = form.value.isDefault
      await adminUpdatePlan(editingId.value, payload)
    } else {
      payload.slug = form.value.slug.trim()
      await adminCreatePlan(payload)
    }
    await invalidate()
    closeForm()
  } catch (err) {
    formError.value = err.response?.data?.error || 'Erro ao salvar plano'
  } finally {
    submitting.value = false
  }
}

async function makeDefault(plan) {
  if (!confirm(`Tornar "${plan.name}" o plano padrão de cadastro?`)) return
  try {
    await adminUpdatePlan(plan.id, { is_default: true })
    await invalidate()
  } catch (err) {
    alert(err.response?.data?.error || 'Erro ao definir padrão')
  }
}

async function toggleActive(plan) {
  try {
    await adminUpdatePlan(plan.id, { active: !plan.active })
    await invalidate()
  } catch (err) {
    alert(err.response?.data?.error || 'Erro ao alterar estado')
  }
}

function invalidate() {
  return queryClient.invalidateQueries({ queryKey: ['plans'] })
}

function formatPrice(cents) {
  return (cents / 100).toFixed(2)
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

function planPerks(plan) {
  const perks = [
    limitLabel(plan.max_jobs, 'vaga', 'vagas', 'Vagas ilimitadas'),
    limitLabel(plan.max_resume_analyses, 'análise de currículo/mês', 'análises de currículo/mês', 'Análises de currículo ilimitadas'),
    limitLabel(plan.max_sites, 'fonte de busca', 'fontes de busca', 'Fontes de busca ilimitadas'),
  ]
  if (plan.ai_matching) 
    perks.push('Matching por IA')
  else
    perks.push('Matching Básico')
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
</script>
