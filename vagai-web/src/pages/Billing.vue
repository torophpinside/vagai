<template>
  <div>
    <div class="mb-8">
      <h1 class="text-2xl font-bold font-outfit text-white">Billing</h1>
      <p class="text-slate-400 mt-1">Gerencie seu plano e pagamentos</p>
    </div>

    <!-- Current Plan -->
    <div class="glass-card p-6 mb-8">
      <div class="flex items-center justify-between">
        <div>
          <div class="text-sm text-slate-400">Plano atual</div>
          <div class="text-2xl font-bold text-white mt-1">{{ currentPlan.name }}</div>
          <div class="text-sm text-slate-400 mt-1">
            {{ currentPlan.price === 0 ? 'Grátis' : `R$ ${currentPlan.price / 100}/mês` }}
          </div>
        </div>
        <div v-if="trialEndsAt" class="text-right">
          <div class="text-sm text-indigo-400 font-medium">Período de teste</div>
          <div class="text-xs text-slate-500">Expira em {{ trialDaysLeft }} dias</div>
        </div>
        <button v-if="currentPlan.slug === 'free'" class="btn-primary">
          Upgrade para Pro
        </button>
      </div>

      <!-- Usage -->
      <div class="mt-6 pt-6 border-t border-white/5 grid grid-cols-3 gap-6">
        <div>
          <div class="text-xs text-slate-500">Vagas</div>
          <div class="text-lg font-bold text-white mt-1">{{ usage.jobs }} / {{ currentPlan.maxJobs === -1 ? '∞' : currentPlan.maxJobs }}</div>
          <div class="w-full bg-slate-700 rounded-full h-2 mt-2">
            <div class="bg-indigo-500 h-2 rounded-full" :style="{ width: usagePercent('jobs') + '%' }"></div>
          </div>
        </div>
        <div>
          <div class="text-xs text-slate-500">Currículos</div>
          <div class="text-lg font-bold text-white mt-1">{{ usage.resumes }} / {{ currentPlan.maxResumes === -1 ? '∞' : currentPlan.maxResumes }}</div>
          <div class="w-full bg-slate-700 rounded-full h-2 mt-2">
            <div class="bg-indigo-500 h-2 rounded-full" :style="{ width: usagePercent('resumes') + '%' }"></div>
          </div>
        </div>
        <div>
          <div class="text-xs text-slate-500">Sites</div>
          <div class="text-lg font-bold text-white mt-1">{{ usage.sites }} / {{ currentPlan.maxSites === -1 ? '∞' : currentPlan.maxSites }}</div>
          <div class="w-full bg-slate-700 rounded-full h-2 mt-2">
            <div class="bg-indigo-500 h-2 rounded-full" :style="{ width: usagePercent('sites') + '%' }"></div>
          </div>
        </div>
      </div>
    </div>

    <!-- Plans -->
    <h2 class="text-lg font-semibold text-white mb-4">Planos disponíveis</h2>
    <div class="grid grid-cols-3 gap-6">
      <div v-for="plan in plans" :key="plan.slug" :class="['glass-card p-6', currentPlan.slug === plan.slug ? 'ring-2 ring-indigo-500' : '']">
        <div class="text-sm text-slate-400 mb-2">{{ plan.name }}</div>
        <div class="text-3xl font-bold text-white">
          {{ plan.price === 0 ? 'Grátis' : `R$ ${plan.price / 100}` }}
          <span v-if="plan.price > 0" class="text-sm font-normal text-slate-400">/mês</span>
        </div>
        <ul class="mt-4 space-y-2">
          <li v-for="feature in plan.featuresList" :key="feature" class="text-sm text-slate-300 flex items-center gap-2">
            <Check class="w-4 h-4 text-emerald-400" />
            {{ feature }}
          </li>
        </ul>
        <button v-if="currentPlan.slug !== plan.slug" class="w-full mt-6 py-3 bg-slate-700 hover:bg-slate-600 text-white font-medium rounded-lg transition-colors">
          {{ currentPlan.slug === 'free' ? 'Começar' : 'Downgrade' }}
        </button>
        <button v-else class="w-full mt-6 py-3 bg-indigo-600 text-white font-medium rounded-lg cursor-default">
          Plano atual
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { Check } from 'lucide-vue-next'

const currentPlan = ref({
  name: 'Free',
  slug: 'free',
  price: 0,
  maxJobs: 100,
  maxResumes: 3,
  maxSites: 5,
  featuresList: ['100 vagas', '3 currículos', '5 sites', 'Matching básico']
})

const trialEndsAt = ref(new Date(Date.now() + 14 * 24 * 60 * 60 * 1000))
const trialDaysLeft = computed(() => {
  if (!trialEndsAt.value) return 0
  return Math.max(0, Math.ceil((trialEndsAt.value - Date.now()) / (24 * 60 * 60 * 1000)))
})

const usage = ref({ jobs: 42, resumes: 2, sites: 3 })

function usagePercent(resource) {
  const max = currentPlan.value['max' + resource.charAt(0).toUpperCase() + resource.slice(1)]
  if (max === -1) return 0
  return Math.min(100, (usage.value[resource] / max) * 100)
}

const plans = ref([
  {
    name: 'Free', slug: 'free', price: 0,
    featuresList: ['100 vagas', '3 currículos', '5 sites', 'Matching básico']
  },
  {
    name: 'Pro', slug: 'pro', price: 4900,
    featuresList: ['1000 vagas', '10 currículos', '20 sites', 'Matching AI avançado', 'Alertas por email', 'Análise de currículo']
  },
  {
    name: 'Enterprise', slug: 'enterprise', price: 14900,
    featuresList: ['Vagas ilimitadas', 'Currículos ilimitados', 'Sites ilimitados', 'AI avançado', 'Alertas por email', 'Análise de currículo', 'API access', 'Webhooks', 'Suporte prioritário']
  }
])
</script>
