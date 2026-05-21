<template>
  <div class="space-y-8">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-4xl font-bold text-white mb-2 font-outfit tracking-tight">Vagas</h1>
        <p class="text-slate-400">Gerencie as oportunidades encontradas pelo scanner.</p>
      </div>
      <div class="flex gap-4 items-center">
        <div class="relative flex-1 max-w-xs">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500" />
          <input v-model="searchQuery" type="text" placeholder="Buscar por título..." class="input-field pl-10 h-12 w-full" />
        </div>
        <DropdownMenu v-model:open="siteOpen">
          <template #trigger="{ open, toggle }">
            <button @click="toggle" class="flex items-center gap-2 px-4 py-2 bg-slate-900/50 border border-white/10 rounded-xl hover:bg-slate-800/50 transition-all whitespace-nowrap">
              <Filter class="w-4 h-4 text-slate-500 shrink-0" />
              <span class="text-sm text-slate-300">
                <template v-if="filters.site.length === 0">Todos os Sites</template>
                <template v-else>{{ filters.site.length }} site{{ filters.site.length > 1 ? 's' : '' }}</template>
              </span>
              <ChevronDown class="w-3.5 h-3.5 text-slate-500" :class="open ? 'rotate-180' : ''" />
            </button>
          </template>
          <template #default="{ close }">
            <button v-for="site in sites" :key="site.id"
              @click="toggleSite(site.id)"
              class="flex items-center gap-3 w-full px-3 py-2 rounded-lg text-sm transition-all"
              :class="filters.site.includes(site.id) ? 'bg-indigo-500/20 text-indigo-300' : 'text-slate-400 hover:text-slate-300 hover:bg-slate-800/50'"
            >
              <div class="w-4 h-4 rounded border flex items-center justify-center transition-all" :class="filters.site.includes(site.id) ? 'bg-indigo-500 border-indigo-500' : 'border-slate-600'">
                <svg v-if="filters.site.includes(site.id)" class="w-3 h-3 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" /></svg>
              </div>
              {{ site.name }}
            </button>
            <div v-if="filters.site.length > 0" class="border-t border-white/5 p-2">
              <button @click="filters.site = []; close()" class="flex items-center gap-2 w-full px-3 py-2 rounded-lg text-sm text-slate-500 hover:text-red-400 hover:bg-red-500/10 transition-all">
                <X class="w-4 h-4" />
                Limpar filtros
              </button>
            </div>
          </template>
        </DropdownMenu>
        <DropdownMenu v-model:open="statusOpen" position="bottom-right">
          <template #trigger="{ open, toggle }">
            <button @click="toggle" class="flex items-center gap-2 px-4 py-2 bg-slate-900/50 border border-white/10 rounded-xl hover:bg-slate-800/50 transition-all whitespace-nowrap">
              <Filter class="w-4 h-4 text-slate-500 shrink-0" />
              <span class="text-sm text-slate-300">
                <template v-if="filters.status.length === 0">Status</template>
                <template v-else>{{ filters.status.length }} status</template>
              </span>
              <ChevronDown class="w-3.5 h-3.5 text-slate-500" :class="open ? 'rotate-180' : ''" />
            </button>
          </template>
          <template #default="{ close }">
            <button v-for="opt in statusOptions" :key="opt.value"
              @click="toggleStatus(opt.value)"
              class="flex items-center gap-3 w-full px-3 py-2 rounded-lg text-sm transition-all"
              :class="filters.status.includes(opt.value) ? 'bg-indigo-500/20 text-indigo-300' : 'text-slate-400 hover:text-slate-300 hover:bg-slate-800/50'"
            >
              <div class="w-4 h-4 rounded border flex items-center justify-center transition-all" :class="filters.status.includes(opt.value) ? 'bg-indigo-500 border-indigo-500' : 'border-slate-600'">
                <svg v-if="filters.status.includes(opt.value)" class="w-3 h-3 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" /></svg>
              </div>
              {{ opt.label }}
            </button>
            <div v-if="filters.status.length > 0" class="border-t border-white/5 p-2">
              <button @click="filters.status = []; close()" class="flex items-center gap-2 w-full px-3 py-2 rounded-lg text-sm text-slate-500 hover:text-red-400 hover:bg-red-500/10 transition-all">
                <X class="w-4 h-4" />
                Limpar filtros
              </button>
            </div>
          </template>
        </DropdownMenu>
        <router-link to="/jobs/new" class="btn-primary flex items-center gap-2">
          <Plus class="w-4 h-4" />
          Adicionar Vaga
        </router-link>
        <button class="btn-primary flex items-center gap-2">
          <RefreshCcw class="w-4 h-4" />
          Sincronizar
        </button>
      </div>
    </div>

    <div v-if="isLoading" class="flex items-center justify-center h-64">
      <div class="w-12 h-12 border-4 border-indigo-500/30 border-t-indigo-500 rounded-full animate-spin"></div>
    </div>

    <div v-else-if="jobs.length === 0" class="flex flex-col items-center justify-center h-64 text-center">
      <div class="w-20 h-20 bg-slate-800 rounded-full flex items-center justify-center mb-4">
        <X class="w-10 h-10 text-slate-600" />
      </div>
      <p class="text-slate-500 max-w-md">Nenhum resultado</p>
    </div>

    <div v-else class="glass-card overflow-hidden">
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-white/5">
          <thead class="bg-white/5">
            <tr>
              <th class="px-8 py-5 text-left text-xs font-bold text-slate-400 uppercase tracking-widest">Título</th>
              <th class="px-8 py-5 text-left text-xs font-bold text-slate-400 uppercase tracking-widest">Empresa</th>
              <th class="px-8 py-5 text-left text-xs font-bold text-slate-400 uppercase tracking-widest">Status</th>
              <th class="px-8 py-5 text-left text-xs font-bold text-slate-400 uppercase tracking-widest">Score</th>
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
                  {{ statusLabelMap[job.status] || job.status }}
                </span>
              </td>
              <td class="px-8 py-6">
                <div v-if="job.max_score > 0" :class="scoreClass(job.max_score)" class="px-3 py-1 rounded-full text-xs font-bold text-center min-w-[3rem]">
                  {{ Math.round(job.max_score) }}%
                </div>
                <span v-else class="text-slate-600 text-sm">—</span>
              </td>
              <td class="px-8 py-6 text-slate-400 text-sm">
                {{ formatDate(job.collected_at) }}
              </td>
              <td class="px-8 py-6 text-right">
                <div class="flex items-center justify-end gap-3">
                  <button v-if="job.status !== 'matched' && job.status !== 'ignored'" @click="markAsMatched(job.id)" class="p-2 bg-emerald-500/10 text-emerald-400 rounded-lg hover:bg-emerald-500 hover:text-white transition-all" title="Marcar como Match">
                    <CheckCircle2 class="w-4 h-4" />
                  </button>
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
      <div v-if="totalPages > 0" class="flex items-center justify-between px-8 py-4 border-t border-white/5">
        <span class="text-sm text-slate-400">
          {{ showingInfo }}
        </span>
        <div class="flex items-center gap-2">
          <button @click="filters.page--" :disabled="filters.page <= 1" class="px-3 py-1.5 text-sm rounded-lg bg-slate-700/50 text-slate-300 hover:bg-slate-600/50 disabled:opacity-30 disabled:cursor-not-allowed transition-all">
            <ChevronLeft class="w-4 h-4 inline-block -ml-0.5" />
            Anterior
          </button>
          <span class="text-sm text-slate-400 px-2 font-mono">{{ filters.page }} / {{ totalPages }}</span>
          <button @click="filters.page++" :disabled="filters.page >= totalPages" class="px-3 py-1.5 text-sm rounded-lg bg-slate-700/50 text-slate-300 hover:bg-slate-600/50 disabled:opacity-30 disabled:cursor-not-allowed transition-all">
            Próximo
            <ChevronRight class="w-4 h-4 inline-block -mr-0.5" />
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref, computed, watch } from 'vue'
import { useQueryClient } from '@tanstack/vue-query'
import { useJobs, useSites, updateJobStatus } from '../services/api'
import { ExternalLink, Trash2, RefreshCcw, Filter, ChevronDown, ChevronLeft, ChevronRight, X, CheckCircle2, Search, Plus } from 'lucide-vue-next'
import DropdownMenu from '../components/DropdownMenu.vue'

const queryClient = useQueryClient()
const { data: sites } = useSites()

const filters = reactive({
  site: [],
  status: [],
  search: '',
  page: 1,
  limit: 20
})

const searchQuery = ref('')

let searchTimer
watch(searchQuery, (val) => {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    filters.search = val
  }, 300)
})

watch([() => filters.site, () => filters.status, () => filters.search], () => {
  filters.page = 1
}, { deep: true })

const { data: jobsResponse, isLoading } = useJobs(filters)

const jobs = computed(() => jobsResponse.value?.data || [])
const total = computed(() => jobsResponse.value?.total || 0)
const totalPages = computed(() => jobsResponse.value?.totalPages || 0)

const showingInfo = computed(() => {
  if (total.value === 0) return 'Nenhuma vaga encontrada'
  const from = (filters.page - 1) * filters.limit + 1
  const to = Math.min(filters.page * filters.limit, total.value)
  return `Mostrando ${from} a ${to} de ${total.value} vagas`
})

const statusOptions = [
  { value: 'new', label: 'Novas' },
  { value: 'matched', label: 'Match' },
  { value: 'analyzed', label: 'Analisadas' },
  { value: 'unmatched', label: 'Sem Match' },
  { value: 'ignored', label: 'Ignoradas' }
]

const statusLabelMap = {
  new: 'Nova',
  matched: 'Match',
  analyzed: 'Analisada',
  unmatched: 'Sem Match',
  ignored: 'Ignorada'
}

const siteOpen = ref(false)

const toggleSite = (id) => {
  const idx = filters.site.indexOf(id)
  if (idx >= 0) {
    filters.site.splice(idx, 1)
  } else {
    filters.site.push(id)
  }
}

const statusOpen = ref(false)

const toggleStatus = (value) => {
  const idx = filters.status.indexOf(value)
  if (idx >= 0) {
    filters.status.splice(idx, 1)
  } else {
    filters.status.push(value)
  }
}

const hideJob = async (jobId) => {
  await updateJobStatus(jobId, 'ignored')
  queryClient.invalidateQueries({ queryKey: ['jobs'] })
}

const markAsMatched = async (jobId) => {
  await updateJobStatus(jobId, 'matched')
  queryClient.invalidateQueries({ queryKey: ['jobs'] })
}

const statusClass = (status) => {
  const classes = {
    new: 'bg-indigo-500/10 text-indigo-400 border-indigo-500/20',
    matched: 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20',
    analyzed: 'bg-amber-500/10 text-amber-400 border-amber-500/20',
    unmatched: 'bg-orange-500/10 text-orange-400 border-orange-500/20',
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

const scoreClass = (score) => {
  if (score >= 0.8) return 'bg-emerald-500/10 text-emerald-400'
  if (score >= 0.6) return 'bg-indigo-500/10 text-indigo-400'
  if (score >= 0.4) return 'bg-amber-500/10 text-amber-400'
  return 'bg-slate-500/10 text-slate-400'
}
</script>

<style scoped>
</style>
