<template>
  <div class="min-h-screen bg-[#0f172a] text-slate-200">
    <!-- Navbar -->
    <header class="sticky top-0 z-20 backdrop-blur-md bg-slate-950/60 border-b border-white/5">
      <div class="max-w-6xl mx-auto px-6 h-16 flex items-center justify-between">
        <a href="#top" class="flex items-center gap-2">
          <div class="w-9 h-9 bg-indigo-600 rounded-xl flex items-center justify-center shadow-lg shadow-indigo-500/20">
            <Target class="text-white w-5 h-5" />
          </div>
          <span class="text-xl font-bold font-outfit text-white tracking-tight">VagAI</span>
        </a>
        <nav class="flex items-center gap-8 text-sm text-slate-400">
          <a href="#funcionalidades" class="hidden md:inline hover:text-white transition-colors">Funcionalidades</a>
          <a href="#como-funciona" class="hidden md:inline hover:text-white transition-colors">Como funciona</a>
          <a href="#planos" class="hover:text-white transition-colors">Preços</a>
        </nav>
        <div class="flex items-center gap-3">
          <template v-if="isLoggedIn">
            <router-link to="/dashboard" class="btn-primary">Ir para o painel</router-link>
          </template>
          <template v-else>
            <router-link to="/login" class="text-sm text-slate-300 hover:text-white px-3 py-2">Entrar</router-link>
            <router-link to="/register" class="btn-primary text-sm">Criar conta grátis</router-link>
          </template>
        </div>
      </div>
    </header>

    <!-- Hero -->
    <section id="top" class="relative overflow-hidden">
      <div class="absolute inset-0 bg-gradient-to-b from-indigo-950/40 via-transparent to-transparent pointer-events-none"></div>
      <div class="absolute -top-40 left-1/2 -translate-x-1/2 w-[800px] h-[800px] bg-indigo-600/10 rounded-full blur-3xl pointer-events-none"></div>
      <div class="relative max-w-4xl mx-auto px-6 pt-24 pb-28 text-center">
        <span class="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-indigo-500/10 border border-indigo-500/20 text-indigo-300 text-xs font-medium mb-6">
          <Sparkles class="w-3.5 h-3.5" />
          Assistente de vagas com inteligência artificial
        </span>
        <h1 class="text-4xl md:text-6xl font-bold font-outfit text-white tracking-tight leading-tight">
          Encontre as vagas certas.<br />
          <span class="bg-gradient-to-r from-indigo-400 to-violet-400 bg-clip-text text-transparent">Sem perder horas procurando.</span>
        </h1>
        <p class="mt-6 text-lg text-slate-400 max-w-2xl mx-auto">
          O VagAI monitora sites de emprego, compara cada vaga com o seu currículo e entrega um
          ranking por compatibilidade — além de te preparar para a entrevista.
        </p>
        <div class="mt-10 flex flex-col sm:flex-row items-center justify-center gap-4">
          <router-link :to="isLoggedIn ? '/dashboard' : '/register'" class="btn-primary text-base px-8 py-3 flex items-center gap-2">
            {{ isLoggedIn ? 'Ir para o painel' : 'Começar grátis' }}
            <ArrowRight class="w-5 h-5" />
          </router-link>
          <a href="#funcionalidades" class="px-8 py-3 rounded-xl border border-white/10 text-slate-300 hover:bg-white/5 transition-colors">
            Ver funcionalidades
          </a>
        </div>
      </div>
    </section>

    <!-- Funcionalidades -->
    <section id="funcionalidades" class="max-w-6xl mx-auto px-6 py-24">
      <div class="text-center mb-14">
        <h2 class="text-3xl md:text-4xl font-bold font-outfit text-white">Tudo para a sua busca em um só lugar</h2>
        <p class="mt-3 text-slate-400">Do primeiro filtro à entrevista final.</p>
      </div>
      <div class="grid sm:grid-cols-2 lg:grid-cols-3 gap-6">
        <div v-for="f in features" :key="f.title" class="glass-card p-6 hover:border-indigo-500/30 transition-colors">
          <div class="w-11 h-11 rounded-xl bg-indigo-500/15 flex items-center justify-center mb-4">
            <component :is="f.icon" class="w-5 h-5 text-indigo-400" />
          </div>
          <h3 class="text-lg font-semibold text-white mb-2">{{ f.title }}</h3>
          <p class="text-sm text-slate-400 leading-relaxed">{{ f.text }}</p>
        </div>
      </div>
    </section>

    <!-- Como funciona -->
    <section id="como-funciona" class="bg-slate-900/40 border-y border-white/5">
      <div class="max-w-6xl mx-auto px-6 py-24">
        <div class="text-center mb-14">
          <h2 class="text-3xl md:text-4xl font-bold font-outfit text-white">Como funciona</h2>
        </div>
        <div class="grid md:grid-cols-4 gap-8">
          <div v-for="(s, i) in steps" :key="s.title" class="text-center">
            <div class="w-12 h-12 mx-auto rounded-full bg-indigo-600 text-white font-bold flex items-center justify-center mb-4 shadow-lg shadow-indigo-500/20">
              {{ i + 1 }}
            </div>
            <h3 class="font-semibold text-white mb-2">{{ s.title }}</h3>
            <p class="text-sm text-slate-400">{{ s.text }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- Planos -->
    <section id="planos" class="max-w-6xl mx-auto px-6 py-24">
      <div class="text-center mb-14">
        <h2 class="text-3xl md:text-4xl font-bold font-outfit text-white">Preços</h2>
        <p class="mt-3 text-slate-400">Comece grátis e evolua quando precisar.</p>
      </div>
      <div v-if="plansLoading" class="text-center text-slate-500 text-sm">Carregando planos...</div>
      <div v-else-if="!plans.length" class="text-center text-slate-500 text-sm">
        Crie sua conta para conhecer os planos disponíveis.
      </div>
      <div v-else class="grid gap-6 mx-auto" :class="plans.length > 1 ? 'md:grid-cols-2 max-w-3xl' : 'max-w-sm'">
        <div v-for="plan in plans" :key="plan.slug" class="glass-card p-8 flex flex-col">
          <div class="text-sm text-slate-400 mb-2">{{ plan.name }}</div>
          <div class="text-4xl font-bold text-white">
            {{ plan.price_monthly === 0 ? 'Grátis' : formatPrice(plan.price_monthly) }}
            <span v-if="plan.price_monthly > 0" class="text-sm font-normal text-slate-400">/mês</span>
          </div>
          <ul class="mt-6 space-y-3 flex-1">
            <li v-for="perk in planPerks(plan)" :key="perk" class="text-sm text-slate-300 flex items-center gap-2">
              <Check class="w-4 h-4 text-emerald-400 shrink-0" />
              {{ perk }}
            </li>
          </ul>
          <router-link :to="isLoggedIn ? '/settings/billing' : '/register'" class="btn-primary mt-8 text-center">
            {{ isLoggedIn ? 'Gerenciar plano' : 'Começar agora' }}
          </router-link>
        </div>
      </div>
    </section>

    <!-- CTA final -->
    <section class="max-w-4xl mx-auto px-6 pb-24">
      <div class="glass-card p-12 text-center bg-gradient-to-br from-indigo-600/20 to-violet-600/10">
        <h2 class="text-3xl font-bold font-outfit text-white">Pronto para achar sua próxima vaga?</h2>
        <p class="mt-3 text-slate-400">Crie sua conta em menos de um minuto.</p>
        <router-link :to="isLoggedIn ? '/dashboard' : '/register'" class="btn-primary inline-block mt-8 px-8 py-3">
          {{ isLoggedIn ? 'Ir para o painel' : 'Criar conta grátis' }}
        </router-link>
      </div>
    </section>

    <footer class="border-t border-white/5 py-8 text-center text-sm text-slate-500">
      © {{ year }} VagAI. Todos os direitos reservados.
    </footer>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import {
  Target, Sparkles, ArrowRight, Check, Radar, Gauge, FileEdit, GraduationCap,
  BarChart3, Send, FileSearch, ShieldCheck
} from 'lucide-vue-next'
import { useAuth } from '../composables/auth'
import { usePublicPlans } from '../services/api'

const { isAuthenticated } = useAuth()
// isAuthenticated só é preenchido após initAuth(); o token no storage cobre o primeiro render.
const isLoggedIn = computed(() => {
  if (isAuthenticated.value) return true
  try { return !!localStorage.getItem('vagai_token') } catch { return false }
})

const year = new Date().getFullYear()

const features = [
  { icon: Radar, title: 'Busca automática de vagas', text: 'Cadastre os sites de emprego e o VagAI descobre os seletores com IA e rastreia novas vagas continuamente.' },
  { icon: Gauge, title: 'Match por compatibilidade', text: 'Cada vaga é comparada ao seu currículo e ranqueada por score, com os pontos fortes e as lacunas explicados.' },
  { icon: FileSearch, title: 'Análise de currículo', text: 'Envie PDF, DOCX ou TXT e receba uma análise de como seu perfil se encaixa em cada oportunidade.' },
  { icon: FileEdit, title: 'Editor de currículo', text: 'Monte e ajuste currículos direcionados por vaga e exporte em PDF pronto para enviar.' },
  { icon: GraduationCap, title: 'Preparação para entrevistas', text: 'Perguntas técnicas e comportamentais geradas para a vaga, incluindo preparação avulsa e questões situacionais.' },
  { icon: Send, title: 'Controle de candidaturas', text: 'Acompanhe onde você já se candidatou e mantenha o histórico da sua busca organizado.' },
  { icon: BarChart3, title: 'Analytics da busca', text: 'Funil, tendências de vagas, lacunas de habilidades e tempo de resposta em gráficos claros.' },
  { icon: ShieldCheck, title: 'Palavras-chave negativas', text: 'Filtre o que não interessa e veja só vagas relevantes para o seu momento.' },
  { icon: Sparkles, title: 'Dados só seus', text: 'Cada conta é isolada: suas vagas, currículos e métricas pertencem somente a você.' },
]

const steps = [
  { title: 'Crie sua conta', text: 'Cadastro grátis, sem cartão de crédito.' },
  { title: 'Envie seu currículo', text: 'Suba um arquivo ou monte no editor.' },
  { title: 'Adicione os sites', text: 'Informe onde buscar. A IA faz o resto.' },
  { title: 'Receba os matches', text: 'Veja as melhores vagas e prepare-se.' },
]

const { data: plansData, isLoading: plansLoading } = usePublicPlans()
const plans = computed(() => Array.isArray(plansData.value) ? plansData.value : [])

function formatPrice(cents) {
  return (cents / 100).toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })
}

function limitLabel(value, singular, plural, unlimited) {
  if (value === -1) return unlimited
  return `${value} ${value === 1 ? singular : plural}`
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
  if (plan.ai_matching) perks.push('Matching por IA')
  if (plan.interview_prep) perks.push('Preparação para entrevistas')
  const seen = new Set()
  return perks.concat(parseFeatures(plan.features)).filter((perk) => {
    const key = perk.toLowerCase()
    if (seen.has(key)) return false
    seen.add(key)
    return true
  })
}
</script>
