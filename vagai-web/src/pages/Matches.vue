<template>
  <div class="space-y-8">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-4xl font-bold text-white mb-2 font-outfit tracking-tight">Matches</h1>
        <p class="text-slate-400">Correspondências inteligentes baseadas no seu currículo.</p>
      </div>
      <div class="flex items-center gap-4">
        <!-- Site Filter -->
        <div class="flex items-center gap-2 px-4 py-2 bg-slate-900/50 border border-white/10 rounded-xl">
          <Filter class="w-4 h-4 text-slate-500" />
          <select v-model="filters.site" class="bg-transparent text-sm text-slate-300 outline-none min-w-[150px]">
            <option value="" class="bg-slate-900">Todos os Sites</option>
            <option v-for="site in sites" :key="site.id" :value="site.id" class="bg-slate-900">
              {{ site.name }}
            </option>
          </select>
        </div>

        <!-- Sort -->
        <div class="flex items-center gap-2 px-4 py-2 bg-slate-900/50 border border-white/10 rounded-xl">
          <ArrowUpDown class="w-4 h-4 text-slate-500" />
          <select v-model="filters.sort" class="bg-transparent text-sm text-slate-300 outline-none">
            <option value="desc" class="bg-slate-900">Maior similaridade</option>
            <option value="asc" class="bg-slate-900">Menor similaridade</option>
          </select>
        </div>
      </div>
    </div>

    <div v-if="isLoading" class="flex items-center justify-center h-64">
      <div class="w-12 h-12 border-4 border-indigo-500/30 border-t-indigo-500 rounded-full animate-spin"></div>
    </div>

    <div v-else class="grid grid-cols-1 gap-6">
      <div v-for="match in matches" :key="match.id" class="glass-card p-8 group transition-all duration-500 hover:bg-slate-800/60 border-l-4" :class="scoreColor(match.similarity_score)">
        <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-6">
          <div class="flex-1">
            <div class="flex items-center gap-3 mb-2">
              <h3 class="text-xl font-bold text-white font-outfit">{{ match.job?.title }}</h3>
              <span class="px-2 py-0.5 bg-white/5 rounded text-[10px] text-slate-500 font-mono">{{ match.job?.id }}</span>
            </div>
            <div class="flex items-center gap-4 text-slate-400 text-sm">
              <span class="flex items-center gap-1.5"><Building2 class="w-4 h-4" /> {{ match.job?.company }}</span>
              <span class="w-1 h-1 bg-slate-700 rounded-full"></span>
              <span class="flex items-center gap-1.5"><Calendar class="w-4 h-4" /> {{ formatDate(match.analyzed_at) }}</span>
            </div>
          </div>
          
          <div class="flex items-center gap-8">
            <div class="text-right">
              <div class="text-4xl font-bold font-outfit" :class="scoreTextColor(match.similarity_score)">
                {{ Math.round(match.similarity_score) }}%
              </div>
              <div class="text-[10px] uppercase tracking-widest font-bold text-slate-500">Similaridade</div>
            </div>
            <button @click="hideMatch(match.id, match.job.id)" class="w-12 h-12 rounded-2xl bg-slate-700 flex items-center justify-center text-slate-400 hover:bg-red-500/20 hover:text-red-400 transition-all">
              <EyeOff class="w-6 h-6" />
            </button>
            <button @click="toggleApplied(match.id, match.applied)" :class="match.applied ? 'bg-emerald-600 hover:bg-emerald-500 shadow-emerald-900/20' : 'bg-slate-700 hover:bg-indigo-600'" class="w-12 h-12 rounded-2xl flex items-center justify-center text-white shadow-lg transition-all active:scale-95">
              <CheckCircle v-if="match.applied" class="w-6 h-6" />
              <Paperclip v-else class="w-6 h-6" />
            </button>
            <a :href="match.job?.url" target="_blank" class="w-12 h-12 rounded-2xl bg-indigo-600 flex items-center justify-center text-white shadow-lg shadow-indigo-500/20 hover:scale-110 transition-transform active:scale-95">
              <ExternalLink class="w-6 h-6" />
            </a>
          </div>
        </div>

        <div class="mt-8 space-y-6">
          <!-- Description Preview -->
          <div v-if="match.job?.description" class="bg-slate-900/50 border border-white/5 rounded-2xl p-4">
            <div class="text-xs text-slate-500 font-bold uppercase tracking-widest mb-2">Descrição da Vaga</div>
            <p class="text-slate-400 text-sm line-clamp-4">{{ match.job.description }}</p>
          </div>

          <!-- Keywords -->
          <div class="flex flex-wrap gap-2">
            <div v-for="kw in parseKeywords(match.keywords_matched)" :key="kw" class="px-3 py-1 bg-indigo-500/10 border border-indigo-500/20 rounded-lg text-xs text-indigo-300 font-medium">
              {{ kw }}
            </div>
          </div>

          <!-- AI Reason -->
          <div v-if="match.ai_reason" class="bg-slate-900/80 border border-white/5 rounded-2xl p-6 relative overflow-hidden">
            <div class="absolute top-0 right-0 p-4 opacity-10">
              <Sparkles class="w-12 h-12 text-indigo-400" />
            </div>
            <div class="flex items-center gap-2 mb-3 text-indigo-400 font-bold text-xs uppercase tracking-widest">
              <BrainCircuit class="w-4 h-4" />
              Análise de Inteligência Artificial
            </div>
            <p class="text-slate-300 text-sm leading-relaxed relative z-10">{{ match.ai_reason }}</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive } from 'vue'
import { useQueryClient } from '@tanstack/vue-query'
import { useMatches, useSites, updateJobStatus, updateMatch } from '../services/api'
import { 
  Building2, 
  Calendar, 
  ExternalLink, 
  ArrowUpDown, 
  Sparkles, 
  BrainCircuit,
  Filter,
  EyeOff,
  CheckCircle,
  Paperclip
} from 'lucide-vue-next'

const filters = reactive({
  sort: 'desc',
  site: '',
  threshold: 1,
  applied: 'false'
})

const queryClient = useQueryClient()
const { data: matches, isLoading } = useMatches(filters)
const { data: sites } = useSites()

const hideMatch = async (matchId, jobId) => {
  await updateJobStatus(jobId, 'ignored')
  queryClient.invalidateQueries({ queryKey: ['matches'] })
}

const toggleApplied = async (matchId, currentApplied) => {
  await updateMatch(matchId, !currentApplied)
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

const formatDate = (date) => {
  return new Intl.DateTimeFormat('pt-BR', {
    day: '2-digit',
    month: 'short'
  }).format(new Date(date))
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
