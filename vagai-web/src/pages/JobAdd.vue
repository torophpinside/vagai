<template>
  <div class="space-y-8">
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-4">
        <router-link to="/jobs" class="p-2 bg-slate-700/50 text-slate-400 rounded-lg hover:bg-slate-600/50 transition-all">
          <ArrowLeft class="w-5 h-5" />
        </router-link>
        <div>
          <h1 class="text-4xl font-bold text-white mb-2 font-outfit tracking-tight">Adicionar Vaga</h1>
          <p class="text-slate-400">Cole o link da vaga para extrair os dados automaticamente.</p>
        </div>
      </div>
    </div>

    <div class="glass-card p-10">
      <form @submit.prevent="handleExtract" class="space-y-6">
        <div class="space-y-2">
          <label class="text-sm font-bold text-slate-400 uppercase tracking-widest ml-1">Fonte</label>
          <DropdownMenu v-model:open="siteOpen" width="w-64">
            <template #trigger="{ open, toggle }">
              <button @click="toggle" type="button" class="flex items-center gap-2 px-4 py-2.5 bg-slate-900/50 border border-white/10 rounded-xl hover:bg-slate-800/50 transition-all w-full">
                <span class="text-sm flex-1 text-left" :class="selectedSite ? 'text-slate-300' : 'text-slate-500'">
                  {{ selectedSite?.name || 'Nenhuma' }}
                </span>
                <ChevronDown class="w-3.5 h-3.5 text-slate-500 shrink-0" :class="open ? 'rotate-180' : ''" />
              </button>
            </template>
            <template #default="{ close }">
              <button @click="form.site_id = null; close()" type="button"
                class="flex items-center gap-3 w-full px-3 py-2 rounded-lg text-sm transition-all"
                :class="form.site_id === null ? 'bg-indigo-500/20 text-indigo-300' : 'text-slate-400 hover:text-slate-300 hover:bg-slate-800/50'"
              >
                <div class="w-4 h-4 rounded-full border flex items-center justify-center transition-all" :class="form.site_id === null ? 'bg-indigo-500 border-indigo-500' : 'border-slate-600'">
                  <div v-if="form.site_id === null" class="w-1.5 h-1.5 rounded-full bg-white"></div>
                </div>
                Nenhuma
              </button>
              <button v-for="site in sites" :key="site.id" @click="form.site_id = site.id; close()" type="button"
                class="flex items-center gap-3 w-full px-3 py-2 rounded-lg text-sm transition-all"
                :class="form.site_id === site.id ? 'bg-indigo-500/20 text-indigo-300' : 'text-slate-400 hover:text-slate-300 hover:bg-slate-800/50'"
              >
                <div class="w-4 h-4 rounded-full border flex items-center justify-center transition-all" :class="form.site_id === site.id ? 'bg-indigo-500 border-indigo-500' : 'border-slate-600'">
                  <div v-if="form.site_id === site.id" class="w-1.5 h-1.5 rounded-full bg-white"></div>
                </div>
                {{ site.name }}
              </button>
            </template>
          </DropdownMenu>
        </div>
        <div class="space-y-2">
          <label class="text-sm font-bold text-slate-400 uppercase tracking-widest ml-1">URL da Vaga</label>
          <div class="flex gap-3">
            <input v-model="url" type="url" class="input-field flex-1 h-12" placeholder="https://..." required />
            <button type="submit" class="btn-primary h-12 px-8 flex items-center gap-2 whitespace-nowrap" :disabled="extracting">
              <Loader2 v-if="extracting" class="w-5 h-5 animate-spin" />
              <Search v-else class="w-5 h-5" />
              {{ extracting ? 'Extraindo...' : 'Extrair' }}
            </button>
          </div>
        </div>
      </form>

      <div v-if="error" class="mt-6 p-4 bg-red-500/10 border border-red-500/20 rounded-xl text-red-400 text-sm">
        {{ error }}
      </div>

      <div v-if="extracted" class="mt-10 space-y-8">
        <div class="border-t border-white/5 pt-8">
          <h2 class="text-xl font-bold text-white font-outfit mb-6">Dados Extraídos</h2>

          <div class="space-y-6">
            <div class="space-y-2">
              <label class="text-sm font-bold text-slate-400 uppercase tracking-widest ml-1">Título</label>
              <input v-model="form.title" type="text" class="input-field w-full h-12" required />
            </div>
            <div class="space-y-2">
              <label class="text-sm font-bold text-slate-400 uppercase tracking-widest ml-1">Empresa</label>
              <input v-model="form.company" type="text" class="input-field w-full h-12" />
            </div>
            <div class="space-y-2">
              <label class="text-sm font-bold text-slate-400 uppercase tracking-widest ml-1">URL</label>
              <input v-model="form.url" type="url" class="input-field w-full h-12 bg-slate-800/50 text-slate-500 cursor-not-allowed" readonly />
            </div>
            <div class="space-y-2">
              <label class="text-sm font-bold text-slate-400 uppercase tracking-widest ml-1">Descrição</label>
              <textarea v-model="form.description" class="input-field w-full h-48 resize-y" />
            </div>
          </div>

          <div class="flex gap-3 mt-8">
            <button @click="handleSave" class="btn-primary h-12 px-8 flex items-center gap-2" :disabled="saving">
              <Loader2 v-if="saving" class="w-5 h-5 animate-spin" />
              <CheckCircle2 v-else class="w-5 h-5" />
              {{ saving ? 'Salvando...' : 'Salvar Vaga' }}
            </button>
            <router-link to="/jobs" class="h-12 px-8 flex items-center gap-2 bg-slate-700/50 text-slate-300 rounded-xl hover:bg-slate-600/50 transition-all">
              Cancelar
            </router-link>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { extractJob, createJob, listSites } from '../services/api'
import DropdownMenu from '../components/DropdownMenu.vue'
import { ArrowLeft, Search, Loader2, CheckCircle2, ChevronDown } from 'lucide-vue-next'

const router = useRouter()
const url = ref('')
const extracting = ref(false)
const saving = ref(false)
const error = ref('')
const extracted = ref(false)
const sites = ref([])
const siteOpen = ref(false)

const selectedSite = computed(() => {
  return sites.value.find(s => s.id === form.site_id) || null
})

const form = reactive({
  title: '',
  company: '',
  url: '',
  description: '',
  site_id: null
})

onMounted(async () => {
  try {
    const data = await listSites()
    sites.value = data || []
  } catch {
    sites.value = []
  }
})

const handleExtract = async () => {
  extracting.value = true
  error.value = ''
  try {
    const data = await extractJob(url.value)
    form.title = data.title || ''
    form.company = data.company || ''
    form.url = data.url || url.value
    form.description = data.description || ''
    extracted.value = true
  } catch (e) {
    error.value = e.response?.data?.error || 'Erro ao extrair dados da vaga'
  } finally {
    extracting.value = false
  }
}

const handleSave = async () => {
  if (!form.title) return
  saving.value = true
  try {
    await createJob({
      title: form.title,
      company: form.company,
      url: form.url,
      description: form.description,
      site_id: form.site_id || null
    })
    router.push('/jobs')
  } catch (e) {
    error.value = e.response?.data?.error || 'Erro ao salvar vaga'
  } finally {
    saving.value = false
  }
}
</script>
