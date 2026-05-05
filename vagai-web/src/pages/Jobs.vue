<template>
  <div class="space-y-8">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-4xl font-bold text-white mb-2 font-outfit tracking-tight">Vagas</h1>
        <p class="text-slate-400">Gerencie as oportunidades encontradas pelo scanner.</p>
      </div>
      <div class="flex gap-4 items-center">
        <div class="relative">
          <select 
            v-model="filters.site" 
            class="bg-slate-800/50 border border-white/10 text-white text-sm rounded-xl px-4 py-2.5 pr-10 focus:outline-none focus:ring-2 focus:ring-indigo-500/50 appearance-none min-w-[200px]"
          >
            <option value="">Todos os Sites</option>
            <option v-for="site in sites" :key="site.id" :value="site.id">
              {{ site.name }}
            </option>
          </select>
          <div class="absolute right-4 top-1/2 -translate-y-1/2 pointer-events-none text-slate-400">
            <Filter class="w-4 h-4" />
          </div>
        </div>
        <button class="btn-primary flex items-center gap-2">
          <RefreshCcw class="w-4 h-4" />
          Sincronizar
        </button>
      </div>
    </div>

    <div v-if="isLoading" class="flex items-center justify-center h-64">
      <div class="w-12 h-12 border-4 border-indigo-500/30 border-t-indigo-500 rounded-full animate-spin"></div>
    </div>

    <div v-else class="glass-card overflow-hidden">
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-white/5">
          <thead class="bg-white/5">
            <tr>
              <th class="px-8 py-5 text-left text-xs font-bold text-slate-400 uppercase tracking-widest">Título</th>
              <th class="px-8 py-5 text-left text-xs font-bold text-slate-400 uppercase tracking-widest">Empresa</th>
              <th class="px-8 py-5 text-left text-xs font-bold text-slate-400 uppercase tracking-widest">Status</th>
              <th class="px-8 py-5 text-left text-xs font-bold text-slate-400 uppercase tracking-widest">Coletado em</th>
              <th class="px-8 py-5 text-right text-xs font-bold text-slate-400 uppercase tracking-widest">Ações</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-white/5">
            <tr v-for="job in jobs" :key="job.id" class="hover:bg-white/5 transition-colors group">
              <td class="px-8 py-6">
                <div class="flex flex-col">
                  <span class="text-white font-bold group-hover:text-indigo-400 transition-colors">{{ job.title }}</span>
                  <span class="text-xs text-slate-500 font-mono mt-1">{{ job.url.substring(0, 50) }}...</span>
                </div>
              </td>
              <td class="px-8 py-6">
                <div class="flex items-center gap-2">
                  <div class="w-8 h-8 rounded-lg bg-slate-700/50 flex items-center justify-center text-xs font-bold text-slate-400 uppercase">
                    {{ job.company?.charAt(0) || '?' }}
                  </div>
                  <span class="text-slate-300">{{ job.company || 'Não identificada' }}</span>
                </div>
              </td>
              <td class="px-8 py-6">
                <span :class="statusClass(job.status)" class="px-3 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider border">
                  {{ job.status }}
                </span>
              </td>
              <td class="px-8 py-6 text-slate-400 text-sm">
                {{ formatDate(job.collected_at) }}
              </td>
              <td class="px-8 py-6 text-right">
                <div class="flex items-center justify-end gap-3">
                  <a :href="job.url" target="_blank" class="p-2 bg-indigo-500/10 text-indigo-400 rounded-lg hover:bg-indigo-500 hover:text-white transition-all">
                    <ExternalLink class="w-4 h-4" />
                  </a>
                  <button @click="hideJob(job.id)" class="p-2 bg-slate-700/50 text-slate-400 rounded-lg hover:bg-red-500/20 hover:text-red-400 transition-all">
                    <Trash2 class="w-4 h-4" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useQueryClient } from '@tanstack/vue-query'
import { useJobs, useSites, updateJobStatus } from '../services/api'
import { ExternalLink, Trash2, RefreshCcw, Filter } from 'lucide-vue-next'

const filters = reactive({
  site: ''
})

const queryClient = useQueryClient()
const { data: jobs, isLoading } = useJobs(filters)
const { data: sites } = useSites()

const hideJob = async (jobId) => {
  await updateJobStatus(jobId, 'ignored')
  queryClient.invalidateQueries({ queryKey: ['jobs'] })
}

const statusClass = (status) => {
  const classes = {
    new: 'bg-indigo-500/10 text-indigo-400 border-indigo-500/20',
    matched: 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20',
    ignored: 'bg-slate-700/50 text-slate-400 border-white/5'
  }
  return classes[status] || classes.new
}

const formatDate = (date) => {
  return new Intl.DateTimeFormat('pt-BR', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  }).format(new Date(date))
}
</script>
