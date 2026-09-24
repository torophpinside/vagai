<template>
  <div class="space-y-8">
    <div>
      <h1 class="text-4xl font-bold text-white mb-2 font-outfit tracking-tight">Preparação para Entrevista</h1>
      <p class="text-slate-400">Suas preparações de entrevista por vaga candidatada, com acompanhamento de progresso.</p>
    </div>

    <div v-if="isLoading" class="flex items-center justify-center h-64">
      <div class="w-12 h-12 border-4 border-indigo-500/30 border-t-indigo-500 rounded-full animate-spin"></div>
    </div>

    <div v-else-if="preps.length === 0" class="flex flex-col items-center justify-center h-64 text-center">
      <div class="w-20 h-20 bg-slate-800 rounded-full flex items-center justify-center mb-4">
        <GraduationCap class="w-10 h-10 text-slate-600" />
      </div>
      <p class="text-slate-500 max-w-md mb-6">Nenhuma preparação criada ainda. Gere uma a partir de uma vaga candidatada.</p>
      <router-link to="/applied" class="inline-flex items-center gap-2 px-5 py-2.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-sm font-bold shadow-lg shadow-indigo-500/20 transition-all active:scale-95">
        <GraduationCap class="w-4 h-4" />
        Ver candidatadas
      </router-link>
    </div>

    <div v-else class="grid grid-cols-1 gap-6">
      <div v-for="prep in preps" :key="prep.id" class="glass-card p-8 border-l-4 border-indigo-500">
        <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-6">
          <div class="flex-1">
            <div class="flex items-center gap-3 mb-2">
              <h3 class="text-xl font-bold text-white font-outfit">{{ prep.title }}</h3>
              <span class="px-2 py-0.5 bg-white/5 rounded text-[10px] text-slate-500 font-mono">{{ prep.id }}</span>
            </div>
            <div class="flex items-center gap-4 text-slate-400 text-sm">
              <span class="flex items-center gap-1.5"><Building2 class="w-4 h-4" /> {{ prep.company || 'Sem empresa' }}</span>
              <span class="w-1 h-1 bg-slate-700 rounded-full"></span>
              <span class="px-2 py-0.5 rounded bg-white/5 text-slate-500 text-[10px] font-mono">
                {{ prep.source === 'ai' ? 'IA' : 'Modelo' }}
              </span>
            </div>
          </div>
          <div class="text-right">
            <div class="flex items-center justify-end gap-2">
              <span v-if="isCompleted(prep)" class="px-2.5 py-1 rounded-lg text-xs font-bold bg-emerald-500/15 border border-emerald-500/30 text-emerald-300">
                Concluída
              </span>
              <span v-else class="px-2.5 py-1 rounded-lg text-xs font-bold bg-amber-500/15 border border-amber-500/30 text-amber-300">
                Em andamento
              </span>
              <span v-if="verification(prep)?.status === 'verified'" class="px-2.5 py-1 rounded-lg text-xs font-bold bg-emerald-500/15 border border-emerald-500/30 text-emerald-300">
                {{ verification(prep).score.toFixed(1) }}/10
              </span>
              <span v-else-if="verification(prep)?.status === 'outdated'" class="px-2.5 py-1 rounded-lg text-xs font-bold bg-amber-500/15 border border-amber-500/30 text-amber-300">
                Desatualizada
              </span>
              <span v-else-if="['running', 'pending'].includes(verification(prep)?.status)" class="px-2.5 py-1 rounded-lg text-xs font-bold bg-indigo-500/15 border border-indigo-500/30 text-indigo-300">
                Verificando...
              </span>
            </div>
            <div class="text-3xl font-bold font-outfit text-indigo-400 mt-1">{{ progressPercent(prep) }}%</div>
            <div class="text-[10px] uppercase tracking-widest font-bold text-slate-500">Respondidas</div>
          </div>
        </div>

        <div class="mt-6 space-y-3">
          <div class="w-full h-2.5 bg-slate-800 rounded-full overflow-hidden">
            <div class="h-full bg-gradient-to-r from-indigo-500 to-emerald-500 transition-all duration-500" :style="{ width: progressPercent(prep) + '%' }"></div>
          </div>
          <div class="flex items-center justify-between text-sm">
            <span class="text-slate-400 font-mono">
              {{ answeredCount(prep) }}/{{ prep.progress?.total || 0 }} respondidas
            </span>
            <div class="flex gap-4 text-xs">
              <span class="text-amber-300">{{ prep.progress?.practiced || 0 }} praticadas</span>
              <span class="text-emerald-300">{{ prep.progress?.mastered || 0 }} dominadas</span>
            </div>
          </div>
        </div>

        <div class="mt-6 flex flex-wrap gap-2">
          <div v-for="tech in prep.tech_stack || []" :key="tech" class="px-3 py-1 bg-indigo-500/10 border border-indigo-500/20 rounded-lg text-xs text-indigo-300 font-medium">
            {{ tech }}
          </div>
        </div>

        <div class="mt-8 flex items-center gap-3 flex-wrap">
          <router-link :to="`/interview-prep/${prep.id}`" class="inline-flex items-center gap-2 px-5 py-2.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-sm font-bold shadow-lg shadow-indigo-500/20 transition-all active:scale-95">
            <Play class="w-4 h-4" />
            Retomar Preparação
          </router-link>
          <button v-if="canVerify(prep)" @click="verifyPrep(prep)" :disabled="isVerifying(prep)" class="inline-flex items-center gap-2 px-5 py-2.5 rounded-xl bg-slate-800 border border-white/10 text-slate-200 text-sm font-semibold hover:bg-slate-700 transition-all disabled:opacity-40 disabled:cursor-wait">
            <span v-if="isVerifying(prep)" class="w-4 h-4 border-2 border-slate-400/30 border-t-slate-200 rounded-full animate-spin"></span>
            <RefreshCw v-else-if="verification(prep)" class="w-4 h-4" />
            <Sparkles v-else class="w-4 h-4" />
            {{ isVerifying(prep) ? 'Corrigindo respostas...' : verifyLabel(prep) }}
          </button>
          <router-link to="/applied" class="inline-flex items-center gap-2 px-4 py-2.5 rounded-xl bg-slate-800 text-slate-300 text-sm font-semibold hover:bg-slate-700 transition-all">
            <Briefcase class="w-4 h-4" />
            Ver candidatadas
          </router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onUnmounted, ref, watch } from 'vue'
import { useQueryClient } from '@tanstack/vue-query'
import { useInterviewPreps, verifyInterviewPrep } from '../services/api'
import { Building2, GraduationCap, Play, Briefcase, Sparkles, RefreshCw } from 'lucide-vue-next'

const queryClient = useQueryClient()
const { data: prepsResponse, isLoading } = useInterviewPreps()

const preps = computed(() => prepsResponse.value?.data || [])
const verifyingId = ref(null)

// Revalida a lista enquanto houver verificação ativa, para o selo e o botão
// refletirem a conclusão sem refresh manual.
let pollTimer = null
const stopPolling = () => {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}
const hasActiveVerification = computed(() => preps.value.some((p) => ['running', 'pending'].includes(p.verification?.status)))
watch(hasActiveVerification, (active) => {
  if (active) {
    stopPolling()
    pollTimer = setInterval(() => queryClient.invalidateQueries({ queryKey: ['interview-preps'] }), 3000)
  } else {
    stopPolling()
  }
}, { immediate: true })
onUnmounted(stopPolling)

const verifyPrep = async (prep) => {
  if (!prep || verifyingId.value) return
  verifyingId.value = prep.id
  try {
    await verifyInterviewPrep(prep.id)
    await queryClient.invalidateQueries({ queryKey: ['interview-preps'] })
  } catch (e) {
    console.error('Falha ao verificar respostas', e)
  } finally {
    verifyingId.value = null
  }
}

const answeredCount = (prep) => {
  const p = prep.progress
  return p ? (p.answered || 0) : 0
}

const progressPercent = (prep) => {
  const total = prep.progress?.total || 0
  if (total === 0) return 0
  return Math.round((answeredCount(prep) / total) * 100)
}

const isCompleted = (prep) => {
  const p = prep.progress
  return !!(p && p.total > 0 && (p.answered || 0) >= p.total)
}

const verification = (prep) => prep.verification || null

// Botão sempre visível em prep concluída; durante a correção fica inativo
// (desabilitado com spinner) em vez de sumir. Nada a verificar enquanto
// faltam respostas; concluída e verificada permite reverificar (ex.:
// reavaliar por IA uma nota gerada pelo fallback).
const canVerify = (prep) => isCompleted(prep)

const isVerifying = (prep) => {
  if (verifyingId.value === prep.id) return true
  const st = verification(prep)?.status
  return st === 'running' || st === 'pending'
}

const verifyLabel = () => 'Verificar Respostas'
</script>