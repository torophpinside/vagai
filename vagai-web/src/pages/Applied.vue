<template>
  <div class="space-y-8">
    <div>
      <h1 class="text-4xl font-bold text-white mb-2 font-outfit tracking-tight">Candidatadas</h1>
      <p class="text-slate-400">Vagas onde você já se candidatou.</p>
    </div>

    <div v-if="isLoading" class="flex items-center justify-center h-64">
      <div class="w-12 h-12 border-4 border-indigo-500/30 border-t-indigo-500 rounded-full animate-spin"></div>
    </div>

    <div v-else-if="appliedJobs.length === 0" class="flex flex-col items-center justify-center h-64 text-center">
      <div class="w-20 h-20 bg-slate-800 rounded-full flex items-center justify-center mb-4">
        <CheckCircle class="w-10 h-10 text-slate-600" />
      </div>
      <h3 class="text-xl font-bold text-white mb-2">Nenhuma candidatura ainda</h3>
      <p class="text-slate-500 max-w-md">Você não se candidatou a nenhuma vaga. Acesse os matches e clique no botão de candidatar.</p>
    </div>

    <div v-else class="grid grid-cols-1 gap-6">
      <div v-for="match in appliedJobs" :key="match.id" class="glass-card p-8 group transition-all duration-500 hover:bg-slate-800/60 border-l-4" :class="scoreColor(match.similarity_score)">
        <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-6">
          <div class="flex-1">
            <div class="flex items-center gap-3 mb-2">
              <h3 class="text-xl font-bold text-white font-outfit">{{ match.job?.title }}</h3>
              <span class="px-2 py-0.5 bg-white/5 rounded text-[10px] text-slate-500 font-mono">{{ match.job?.id }}</span>
            </div>
            <div class="flex items-center gap-4 text-slate-400 text-sm">
              <span class="flex items-center gap-1.5"><Building2 class="w-4 h-4" /> {{ match.job?.company }}</span>
              <span class="w-1 h-1 bg-slate-700 rounded-full"></span>
              <span class="flex items-center gap-1.5"><MapPin class="w-4 h-4" /> {{ match.job?.location }}</span>
            </div>
          </div>
          
          <div class="flex items-center gap-8">
            <div class="text-right">
              <div class="text-4xl font-bold font-outfit" :class="scoreTextColor(match.similarity_score)">
                {{ Math.round(match.similarity_score) }}%
              </div>
              <div class="text-[10px] uppercase tracking-widest font-bold text-slate-500">Similaridade</div>
            </div>
            <button @click="removeMatch(match.id)" class="w-12 h-12 rounded-2xl bg-red-500/20 hover:bg-red-500 flex items-center justify-center text-red-400 hover:text-white shadow-lg transition-all active:scale-95">
              <Trash2 class="w-6 h-6" />
            </button>
            <a :href="match.job?.url" target="_blank" class="w-12 h-12 rounded-2xl bg-indigo-600 flex items-center justify-center text-white shadow-lg shadow-indigo-500/20 hover:scale-110 transition-transform active:scale-95">
              <ExternalLink class="w-6 h-6" />
            </a>
          </div>
        </div>

        <div class="mt-8 space-y-6">
          <div class="flex flex-wrap gap-2">
            <div v-for="kw in parseKeywords(match.keywords_matched)" :key="kw" class="px-3 py-1 bg-indigo-500/10 border border-indigo-500/20 rounded-lg text-xs text-indigo-300 font-medium">
              {{ kw }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useQueryClient } from '@tanstack/vue-query'
import { useMatches, updateMatch, deleteMatch } from '../services/api'
import { 
  Building2, 
  MapPin,
  ExternalLink, 
  Trash2
} from 'lucide-vue-next'

const queryClient = useQueryClient()
const { data: matches, isLoading } = useMatches({ applied: 'true' })

const appliedJobs = computed(() => {
  if (!matches.value) return []
  return matches.value.filter(m => m.applied)
})

const removeMatch = async (matchId) => {
  await deleteMatch(matchId)
  queryClient.invalidateQueries({ queryKey: ['matches'] })
}

const scoreColor = (score) => {
  if (score >= 80) return 'border-emerald-500'
  if (score >= 60) return 'border-indigo-500'
  return 'border-slate-700'
}

const scoreTextColor = (score) => {
  if (score >= 80) return 'text-emerald-400'
  if (score >= 60) return 'text-indigo-400'
  return 'text-slate-400'
}

const parseKeywords = (kws) => {
  if (!kws) return []
  try {
    const parsed = JSON.parse(kws)
    return Array.isArray(parsed) ? parsed : []
  } catch (e) {
    return kws.split(',').map(s => s.trim())
  }
}
</script>