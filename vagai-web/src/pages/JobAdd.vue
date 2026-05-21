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
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { extractJob, createJob } from '../services/api'
import { ArrowLeft, Search, Loader2, CheckCircle2 } from 'lucide-vue-next'

const router = useRouter()
const url = ref('')
const extracting = ref(false)
const saving = ref(false)
const error = ref('')
const extracted = ref(false)

const form = reactive({
  title: '',
  company: '',
  url: '',
  description: ''
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
      description: form.description
    })
    router.push('/jobs')
  } catch (e) {
    error.value = e.response?.data?.error || 'Erro ao salvar vaga'
  } finally {
    saving.value = false
  }
}
</script>
