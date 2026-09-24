<template>
  <div class="space-y-8">
    <div class="flex items-center justify-between">
      <div>
        <router-link to="/applied" class="inline-flex items-center gap-1.5 text-slate-400 hover:text-slate-200 text-sm mb-2 transition-colors">
          <ArrowLeft class="w-4 h-4" />
          Candidatadas
        </router-link>
        <h1 class="text-4xl font-bold text-white font-outfit tracking-tight">Sabatina</h1>
        <p class="text-slate-400 mt-1">
          <span class="text-slate-300 font-medium">{{ prep?.title }}</span>
          <span v-if="prep?.company" class="text-slate-500"> · {{ prep.company }}</span>
        </p>
      </div>
      <button
        v-if="prep?.match_id"
        @click="regenerate" :disabled="regenerating" class="inline-flex items-center gap-2 px-4 py-2 rounded-xl bg-slate-800 border border-white/10 text-slate-300 text-sm font-semibold hover:bg-slate-700 transition-all disabled:opacity-40">
        <span v-if="regenerating" class="w-4 h-4 border-2 border-slate-400/30 border-t-slate-300 rounded-full animate-spin"></span>
        <RefreshCw v-else class="w-4 h-4" />
        Regenerar
      </button>
    </div>

    <div v-if="isLoading" class="flex items-center justify-center h-64">
      <div class="w-12 h-12 border-4 border-indigo-500/30 border-t-indigo-500 rounded-full animate-spin"></div>
    </div>

    <div v-else-if="!prep || questions.length === 0" class="flex flex-col items-center justify-center h-64 text-center">
      <div class="w-20 h-20 bg-slate-800 rounded-full flex items-center justify-center mb-4">
        <GraduationCap class="w-10 h-10 text-slate-600" />
      </div>
      <p class="text-slate-500 max-w-md">Nenhuma pergunta disponível.</p>
    </div>

    <template v-else>
      <div class="glass-card p-6">
        <div class="flex items-center justify-between mb-3">
          <span class="text-xs uppercase tracking-widest font-bold text-slate-400">Progresso</span>
          <span class="text-sm font-mono text-slate-300">{{ answeredCount }}/{{ questions.length }} respondidas · {{ masteredCount }} dominadas · {{ practicedCount }} praticadas</span>
        </div>
        <div class="w-full h-2.5 bg-slate-800 rounded-full overflow-hidden">
          <div class="h-full bg-gradient-to-r from-indigo-500 to-emerald-500 transition-all duration-500" :style="{ width: progressPercent + '%' }"></div>
        </div>
      </div>

      <div v-if="reviewQueue.length > 0" class="glass-card p-6 border-amber-500/30">
        <div class="flex items-center justify-between mb-3">
          <span class="text-xs uppercase tracking-widest font-bold text-amber-300">Fila de revisão ({{ reviewQueue.length }})</span>
          <span class="text-[10px] text-slate-500">Marcadas como difíceis há 3+ dias</span>
        </div>
        <div class="flex flex-wrap gap-2">
          <button v-for="item in reviewQueue" :key="item.q.id" @click="goto(item.index)" class="inline-flex items-center gap-2 px-3 py-1.5 rounded-lg bg-amber-500/10 border border-amber-500/20 text-amber-200 text-xs font-semibold hover:bg-amber-500/20 transition-all">
            <span class="font-mono">#{{ item.index + 1 }}</span>
            <span class="max-w-56 truncate">{{ item.q.topic || item.q.text }}</span>
            <span class="font-mono text-amber-400/80">{{ item.days }}d</span>
          </button>
        </div>
      </div>

      <div class="glass-card p-6 border-indigo-500/30">
        <div class="flex items-center justify-between mb-3">
          <span class="text-xs uppercase tracking-widest font-bold text-slate-400">Verificação por IA</span>
          <span v-if="verification" class="px-2.5 py-1 rounded-lg text-[10px] uppercase tracking-widest font-bold" :class="verificationStatusClass">
            {{ verificationStatusLabel }}
          </span>
        </div>

        <div v-if="verification?.status === 'verified'" class="flex flex-col sm:flex-row gap-6 items-start sm:items-center">
          <div class="text-center shrink-0">
            <div class="text-5xl font-bold font-outfit text-indigo-400">{{ verification.score.toFixed(1) }}</div>
            <div class="text-[10px] uppercase tracking-widest font-bold text-slate-500 mt-1">Nota geral</div>
          </div>
          <div class="flex-1">
            <p class="text-slate-300 text-sm leading-relaxed whitespace-pre-line">{{ verification.feedback }}</p>
            <p class="text-[10px] text-slate-500 mt-2 font-mono">
              {{ verification.source === 'ai' ? 'Avaliação por IA' : 'Avaliação automática' }}
              <span v-if="verification.verified_at"> · {{ new Date(verification.verified_at).toLocaleString('pt-BR') }}</span>
            </p>
            <div v-if="historyAsc.length > 1" class="flex items-center gap-2 mt-3 flex-wrap">
              <span class="text-[10px] uppercase tracking-widest font-bold text-slate-500">Evolução:</span>
              <span v-for="(h, i) in historyAsc" :key="i" class="inline-flex items-center gap-1.5">
                <span class="px-2 py-0.5 rounded-lg bg-white/5 border border-white/10 text-xs font-mono text-slate-300" :title="new Date(h.verified_at).toLocaleString('pt-BR')">
                  {{ Number(h.score).toFixed(1) }}
                </span>
                <span v-if="i < historyAsc.length - 1" class="text-slate-600 text-xs">→</span>
              </span>
            </div>
          </div>
        </div>

        <div v-else-if="verification?.status === 'running' || verification?.status === 'pending'" class="flex items-center gap-4 py-2">
          <div class="w-8 h-8 border-2 border-indigo-500/30 border-t-indigo-500 rounded-full animate-spin shrink-0"></div>
          <div>
            <p class="text-slate-300 text-sm font-semibold">Verificando suas respostas...</p>
            <p class="text-slate-500 text-xs mt-0.5">A nota e o feedback aparecerão aqui quando prontos. Você pode fechar esta página.</p>
          </div>
        </div>

        <div v-else-if="verification?.status === 'outdated'" class="flex flex-col sm:flex-row gap-4 items-start sm:items-center justify-between">
          <p class="text-amber-300 text-sm font-semibold">Suas respostas mudaram após a verificação. Gere uma nova para atualizar a nota e o feedback.</p>
          <button @click="verify" :disabled="verifying" class="inline-flex items-center gap-2 px-4 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-sm font-bold disabled:opacity-40 transition-all shrink-0">
            <span v-if="verifying" class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></span>
            <RefreshCw v-else class="w-4 h-4" />
            Gerar nova verificação
          </button>
        </div>

        <div v-else-if="answeredCount === 0" class="text-slate-500 text-sm">
          Responda ao menos uma questão para pedir a correção por IA ({{ progress.answered || 0 }}/{{ progress.total || 0 }} respondidas).
        </div>

        <div v-else class="flex flex-col sm:flex-row gap-4 items-start sm:items-center justify-between">
          <p class="text-slate-400 text-sm">
            <span v-if="remainingCount > 0">Correção parcial: o que já foi respondido será corrigido — faltam <span class="text-slate-300 font-semibold">{{ remainingCount }}</span>.</span>
            <span v-else>Todas as questões respondidas. Gere a verificação para receber sua nota e feedback.</span>
          </p>
          <button @click="verify" :disabled="verifying" class="inline-flex items-center gap-2 px-4 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-sm font-bold disabled:opacity-40 transition-all shrink-0">
            <span v-if="verifying" class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></span>
            <Sparkles v-else class="w-4 h-4" />
            Verificar minhas respostas
          </button>
        </div>
      </div>

      <div class="flex items-center justify-center gap-1.5 flex-wrap">
        <button v-for="(q, i) in questions" :key="q.id" @click="goto(i)"
          class="w-8 h-8 rounded-lg text-xs font-bold transition-all"
          :class="dotClass(q, i)" :title="'Pergunta ' + (i + 1)">
          {{ i + 1 }}
        </button>
      </div>

      <div class="glass-card p-8 min-h-[24rem] flex flex-col">
        <div class="flex items-center gap-3 mb-6">
          <span class="px-2.5 py-1 rounded-lg text-[10px] uppercase tracking-widest font-bold" :class="categoryClass(current.category)">
            {{ categoryLabel(current.category) }}
          </span>
          <span v-if="current.topic" class="px-2.5 py-1 rounded-lg bg-white/5 text-slate-400 text-[10px] uppercase tracking-widest font-bold">{{ current.topic }}</span>
          <span class="ml-auto text-xs text-slate-500 font-mono">{{ currentIndex + 1 }} / {{ questions.length }}</span>
          <select v-model.number="simMinutes" title="Modo simulado: tempo por pergunta" class="bg-slate-900/60 border border-white/10 rounded-lg px-2 py-1 text-[10px] uppercase tracking-widest font-bold text-slate-400 focus:outline-none focus:border-indigo-500/50">
            <option :value="0">Simulado off</option>
            <option :value="2">Simulado 2min</option>
            <option :value="3">Simulado 3min</option>
            <option :value="5">Simulado 5min</option>
          </select>
          <span v-if="simMinutes > 0" class="px-2.5 py-1 rounded-lg text-[10px] font-mono font-bold" :class="timeLeft <= 30 ? 'bg-rose-500/15 text-rose-300' : 'bg-white/5 text-slate-300'">{{ fmtClock(timeLeft) }}</span>
          <span v-if="entrySpent > 0" class="text-[10px] text-slate-500 font-mono" :title="'Tempo total nesta pergunta'">{{ fmtClock(entrySpent) }}</span>
        </div>

        <h2 class="text-2xl font-bold text-white font-outfit leading-relaxed mb-6">{{ current.text }}</h2>

        <div v-if="revealed" class="bg-slate-900/70 border border-white/5 rounded-2xl p-6 mb-6">
          <div class="flex items-center gap-2 mb-3 text-indigo-400 font-bold text-xs uppercase tracking-widest">
            <Lightbulb class="w-4 h-4" />
            Guia de resposta
          </div>
          <p class="text-slate-300 text-sm leading-relaxed whitespace-pre-line">{{ current.answer_guide || 'Sem guia disponível.' }}</p>
        </div>

        <div class="mb-6">
          <textarea v-model="draftAnswer" rows="4" placeholder="Opcional: escreva sua resposta..."
            class="w-full bg-slate-900/60 border border-white/10 rounded-2xl p-4 text-slate-200 text-sm resize-none focus:outline-none focus:border-indigo-500/50 transition-colors"></textarea>
        </div>

        <div v-if="currentAnalysis" class="bg-emerald-500/5 border border-emerald-500/20 rounded-2xl p-6 mb-6">
          <div class="flex items-center gap-2 mb-3 text-emerald-300 font-bold text-xs uppercase tracking-widest">
            <Sparkles class="w-4 h-4" />
            Feedback da correção por IA
          </div>
          <p class="text-slate-300 text-sm leading-relaxed whitespace-pre-line">{{ currentAnalysis.text }}</p>
        </div>

        <div class="mt-auto flex flex-col sm:flex-row items-center justify-between gap-4">
          <button @click="prev" :disabled="currentIndex === 0" class="inline-flex items-center gap-2 px-4 py-2.5 rounded-xl bg-slate-800 text-slate-300 text-sm font-semibold hover:bg-slate-700 disabled:opacity-30 disabled:cursor-not-allowed transition-all">
            <ChevronLeft class="w-4 h-4" />
            Anterior
          </button>

          <div v-if="!revealed" class="flex-1 flex justify-center">
            <button @click="revealed = true" class="inline-flex items-center gap-2 px-6 py-2.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-sm font-bold shadow-lg shadow-indigo-500/20 transition-all active:scale-95">
              <Eye class="w-4 h-4" />
              Revelar resposta
            </button>
          </div>
          <div v-else class="flex-1 flex items-center justify-center gap-2">
            <button @click="assess('low')" class="px-4 py-2.5 rounded-xl bg-rose-500/15 border border-rose-500/30 text-rose-300 text-sm font-semibold hover:bg-rose-500/25 transition-all">Difícil</button>
            <button @click="assess('medium')" class="px-4 py-2.5 rounded-xl bg-amber-500/15 border border-amber-500/30 text-amber-300 text-sm font-semibold hover:bg-amber-500/25 transition-all">Ok</button>
            <button @click="assess('high')" class="px-4 py-2.5 rounded-xl bg-emerald-500/15 border border-emerald-500/30 text-emerald-300 text-sm font-semibold hover:bg-emerald-500/25 transition-all">Dominada</button>
          </div>

          <button @click="next" :disabled="currentIndex === questions.length - 1" class="inline-flex items-center gap-2 px-4 py-2.5 rounded-xl bg-slate-800 text-slate-300 text-sm font-semibold hover:bg-slate-700 disabled:opacity-30 disabled:cursor-not-allowed transition-all">
            Próxima
            <ChevronRight class="w-4 h-4" />
          </button>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
import { computed, ref, watch, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { getInterviewPrep, updateQuestionStatus, saveQuestionAnswer, generateInterviewPrep, verifyInterviewPrep } from '../services/api'
import {
  ArrowLeft,
  ChevronLeft,
  ChevronRight,
  Eye,
  GraduationCap,
  Lightbulb,
  RefreshCw,
  Sparkles
} from 'lucide-vue-next'

const route = useRoute()

const prep = ref(null)
const questions = ref([])
const isLoading = ref(true)
const regenerating = ref(false)
const verifying = ref(false)
const currentIndex = ref(0)
const revealed = ref(false)
const draftAnswer = ref('')

// Modo simulado: cronômetro por pergunta com avanço automático e registro do
// tempo gasto (acumulado no backend via time_spent_seconds).
const simMinutes = ref(0)
const timeLeft = ref(0)
const questionStartAt = ref(Date.now())
let simTimer = null

const fmtClock = (s) => `${Math.floor(Math.max(0, s) / 60)}:${String(Math.max(0, s) % 60).padStart(2, '0')}`
const questionElapsed = () => Math.max(0, Math.floor((Date.now() - questionStartAt.value) / 1000))

const stopSimTimer = () => {
  if (simTimer) {
    clearInterval(simTimer)
    simTimer = null
  }
}
const startSimTimer = () => {
  stopSimTimer()
  questionStartAt.value = Date.now()
  if (simMinutes.value <= 0 || questions.value.length === 0) return
  timeLeft.value = simMinutes.value * 60
  simTimer = setInterval(() => {
    timeLeft.value--
    if (timeLeft.value <= 0) {
      stopSimTimer()
      if (currentIndex.value < questions.value.length - 1) next()
    }
  }, 1000)
}

const current = computed(() => questions.value[currentIndex.value] || {})

const verification = computed(() => prep.value?.verification || null)
// Histórico de tentativas em ordem cronológica (a API retorna desc).
const historyAsc = computed(() => [...(prep.value?.history || [])].reverse())
// Feedback individual da correção por IA da pergunta atual. Só existe quando
// a correção foi pedida e concluída; sem feedback o bloco fica oculto.
const currentAnalysis = computed(() => {
  const a = prep.value?.analyses?.[current.value.id]
  if (!a || a.status !== 'completed' || !a.text?.trim()) return null
  return a
})
const progress = computed(() => prep.value?.progress || {})
const remainingCount = computed(() => progress.value.remaining ?? 0)
const answeredCount = computed(() => progress.value.answered ?? 0)

const verificationStatusLabel = computed(() => ({
  verified: 'Verificada',
  running: 'Verificando...',
  pending: 'Na fila',
  outdated: 'Desatualizada'
}[verification.value?.status] || verification.value?.status || ''))

const verificationStatusClass = computed(() => ({
  verified: 'bg-emerald-500/15 text-emerald-300',
  running: 'bg-indigo-500/15 text-indigo-300',
  pending: 'bg-slate-500/15 text-slate-300',
  outdated: 'bg-amber-500/15 text-amber-300'
}[verification.value?.status] || 'bg-white/5 text-slate-400'))

const masteredCount = computed(() => questions.value.filter(q => q.status === 'mastered').length)
const practicedCount = computed(() => questions.value.filter(q => q.status === 'practiced').length)

// Tempo acumulado na pergunta atual (modo simulado).
const entrySpent = computed(() => prep.value?.entries?.[current.value.id]?.time_spent_seconds || 0)

// Fila de revisão espaçada: perguntas marcadas como "Difícil" há 3+ dias.
const REVIEW_DAYS = 3
const reviewQueue = computed(() => {
  const out = []
  const now = Date.now()
  questions.value.forEach((q, i) => {
    const e = prep.value?.entries?.[q.id]
    if (e?.self_assessment === 'low' && e?.updated_at) {
      const days = Math.floor((now - new Date(e.updated_at).getTime()) / 86400000)
      if (days >= REVIEW_DAYS) out.push({ index: i, q, days })
    }
  })
  return out
})

const progressPercent = computed(() => {
  if (questions.value.length === 0) return 0
  return Math.round(((masteredCount.value + practicedCount.value) / questions.value.length) * 100)
})

const applyPrep = (data) => {
   prep.value = data.preparation
   questions.value = data.preparation?.questions || []
   
   // Recupera a primeira pergunta que não está masterizada
   const firstPending = questions.value.findIndex(q => q.status !== 'mastered')
   currentIndex.value = firstPending === -1 ? 0 : firstPending
   
   revealed.value = false
   
   // Recupera a resposta salva para a pergunta selecionada
   const currentQ = questions.value[currentIndex.value]
   if (currentQ && prep.value?.entries) {
     draftAnswer.value = prep.value.entries[currentQ.id]?.answer || ''
    } else {
      draftAnswer.value = ''
    }

    startPolling()
    startSimTimer()
  }

let pollTimer = null

const stopPolling = () => {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

const startPolling = () => {
  const st = verification.value?.status
  if (st === 'pending' || st === 'running') {
    stopPolling()
    pollTimer = setInterval(async () => {
      await refreshMeta()
      const st2 = verification.value?.status
      if (!st2 || st2 === 'verified' || st2 === 'outdated') stopPolling()
    }, 2500)
  } else {
    stopPolling()
  }
}

const refreshMeta = async () => {
   try {
     const data = await getInterviewPrep(route.params.id)
     prep.value = data.preparation
     questions.value = data.preparation?.questions || []
   } catch (e) {
     console.error('Falha ao atualizar preparação', e)
   }
 }

onUnmounted(() => {
  stopPolling()
  stopSimTimer()
})

const load = async () => {
  isLoading.value = true
  try {
    const data = await getInterviewPrep(route.params.id)
    applyPrep(data)
  } catch (e) {
    console.error('Falha ao carregar preparação', e)
  } finally {
    isLoading.value = false
  }
}
load()

watch(() => route.params.id, load)

watch(currentIndex, () => {
  revealed.value = false

  // Recupera a resposta salva para a nova pergunta ao navegar
  const currentQ = questions.value[currentIndex.value]
  if (currentQ && prep.value?.entries) {
    draftAnswer.value = prep.value.entries[currentQ.id]?.answer || ''
  } else {
    draftAnswer.value = ''
  }

  startSimTimer()
})

watch(simMinutes, () => {
  startSimTimer()
})

const goto = async (i) => {
   if (i !== currentIndex.value) {
     await saveDraft()
     currentIndex.value = i
   }
 }
 const next = async () => {
   if (currentIndex.value < questions.value.length - 1) {
     await saveDraft()
     currentIndex.value++
   }
 }
const prev = async () => {
   if (currentIndex.value > 0) {
     await saveDraft()
     currentIndex.value--
   }
 }

const saveDraft = async () => {
   const q = current.value
   if (!q || !prep.value) return
   try {
     const currentSelfAssessment = prep.value.entries?.[q.id]?.self_assessment || ''
     
      // Atualização Otimista: atualiza localmente antes da API para evitar races na navegação
      if (!prep.value.entries) prep.value.entries = {}
     const wasEmpty = !(prep.value.entries[q.id]?.answer || '').trim()
     const isNowEmpty = !draftAnswer.value.trim()
     
     prep.value.entries[q.id] = { 
       id: prep.value.entries[q.id]?.id, 
       question_id: q.id, 
       answer: draftAnswer.value, 
       self_assessment: currentSelfAssessment 
     }
     
     if (prep.value.progress) {
       if (wasEmpty && !isNowEmpty) prep.value.progress.answered++
       else if (!wasEmpty && isNowEmpty) prep.value.progress.answered--
     }

      await saveQuestionAnswer(prep.value.id, q.id, draftAnswer.value, currentSelfAssessment, questionElapsed())
    } catch (e) {
      console.error('Falha ao salvar rascunho', e)
      // Opcional: Reverter estado local em caso de erro grave
    }
  }

const assess = async (level) => {
  const q = current.value
  if (!q) return
  const status = level === 'high' ? 'mastered' : 'practiced'
  try {
    await saveQuestionAnswer(prep.value.id, q.id, draftAnswer.value, level, questionElapsed())
    const res = await updateQuestionStatus(prep.value.id, q.id, status)
    if (res.question) {
      q.status = res.question.status
    }
    await refreshMeta()
    startPolling()
    if (currentIndex.value < questions.value.length - 1) next()
    else revealed.value = false
  } catch (e) {
    console.error('Falha ao salvar avaliação', e)
  }
}

const regenerate = async () => {
  if (!prep.value) return
  regenerating.value = true
  try {
    const data = await generateInterviewPrep(prep.value.match_id)
    applyPrep(data)
  } catch (e) {
    console.error('Falha ao regenerar preparação', e)
  } finally {
    regenerating.value = false
  }
}

const verify = async () => {
  if (!prep.value) return
  verifying.value = true
  try {
    await verifyInterviewPrep(prep.value.id)
    await refreshMeta()
    startPolling()
  } catch (e) {
    console.error('Falha ao verificar respostas', e)
  } finally {
    verifying.value = false
  }
}

const categoryLabel = (c) => ({ technology: 'Tecnologia', foundations: 'Fundamentos', architecture: 'Arquitetura' }[c] || c)
const categoryClass = (c) => ({
  technology: 'bg-indigo-500/15 text-indigo-300 border border-indigo-500/30',
  foundations: 'bg-amber-500/15 text-amber-300 border border-amber-500/30',
  architecture: 'bg-emerald-500/15 text-emerald-300 border border-emerald-500/30'
}[c] || 'bg-white/5 text-slate-400')

const dotClass = (q, i) => {
  if (i === currentIndex.value) return 'bg-indigo-600 text-white ring-2 ring-indigo-400/40'
  if (q.status === 'mastered') return 'bg-emerald-500/20 text-emerald-300'
  if (q.status === 'practiced') return 'bg-amber-500/20 text-amber-300'
  return 'bg-slate-800 text-slate-400 hover:bg-slate-700'
}
</script>
