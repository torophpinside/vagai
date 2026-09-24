<template>
  <div class="space-y-8">
    <div>
      <router-link to="/interview-prep" class="inline-flex items-center gap-1.5 text-slate-400 hover:text-slate-200 text-sm mb-2 transition-colors">
        <ArrowLeft class="w-4 h-4" />
        Preparações
      </router-link>
      <h1 class="text-4xl font-bold text-white mb-2 font-outfit tracking-tight">Preparação Avulsa</h1>
      <p class="text-slate-400">Cole o conteúdo de uma vaga e gere uma preparação salva com 15 perguntas de entrevista — sem criar uma vaga candidatada.</p>
    </div>

    <div class="glass-card p-8">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
        <div>
          <label class="block text-xs uppercase tracking-widest font-bold text-slate-400 mb-2">Cargo</label>
          <input v-model="form.title" type="text" placeholder="Ex.: Backend Developer"
            class="w-full bg-slate-900/50 border border-white/10 rounded-xl px-4 py-2.5 text-slate-200 text-sm outline-none focus:border-indigo-500/50 transition-colors" />
        </div>
        <div>
          <label class="block text-xs uppercase tracking-widest font-bold text-slate-400 mb-2">Empresa</label>
          <input v-model="form.company" type="text" placeholder="Ex.: Tech Corp"
            class="w-full bg-slate-900/50 border border-white/10 rounded-xl px-4 py-2.5 text-slate-200 text-sm outline-none focus:border-indigo-500/50 transition-colors" />
        </div>
      </div>

      <div class="mb-6">
        <label class="block text-xs uppercase tracking-widest font-bold text-slate-400 mb-2">
          Conteúdo da vaga <span class="text-rose-400">*</span>
        </label>
        <textarea v-model="form.description" rows="6" placeholder="Cole aqui a descrição completa da vaga. Quanto mais conteúdo, melhores as perguntas."
          class="w-full bg-slate-900/50 border border-white/10 rounded-xl px-4 py-3 text-slate-200 text-sm resize-none outline-none focus:border-indigo-500/50 transition-colors"></textarea>
        <p v-if="error" class="mt-2 text-sm text-rose-400 font-semibold">{{ error }}</p>
      </div>

      <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div class="w-full sm:w-auto">
          <label class="block text-xs uppercase tracking-widest font-bold text-slate-400 mb-2">Semente (opcional)</label>
          <input v-model="form.seed" type="text" placeholder="Use um valor fixo para a mesma ordem de perguntas"
            class="w-full sm:w-96 bg-slate-900/50 border border-white/10 rounded-xl px-4 py-2.5 text-slate-200 text-sm outline-none focus:border-indigo-500/50 transition-colors" />
        </div>
        <button @click="generate" :disabled="isLoading" class="inline-flex items-center gap-2 px-6 py-2.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-sm font-bold shadow-lg shadow-indigo-500/20 transition-all active:scale-95 disabled:opacity-40 disabled:cursor-not-allowed">
          <span v-if="isLoading" class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></span>
          <Sparkles v-else class="w-4 h-4" />
          {{ isLoading ? 'Salvando preparação...' : 'Gerar e salvar' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { generateSpontaneousPrep } from '../services/api'
import { ArrowLeft, Sparkles } from 'lucide-vue-next'

const router = useRouter()

const form = ref({ title: '', company: '', description: '', seed: '' })
const isLoading = ref(false)
const error = ref('')

const generate = async () => {
  error.value = ''
  if (!form.value.description.trim()) {
    error.value = 'O conteúdo da vaga é obrigatório.'
    return
  }
  isLoading.value = true
  try {
    const data = await generateSpontaneousPrep({
      title: form.value.title.trim(),
      company: form.value.company.trim(),
      description: form.value.description,
      random: form.value.seed.trim()
    })
    const id = data?.preparation?.id
    if (!id) {
      throw new Error('Preparação criada sem identificador')
    }
    router.push('/interview-prep/' + id)
  } catch (e) {
    error.value = e.response?.data?.error || 'Falha ao gerar as perguntas. Tente novamente.'
    console.error('Falha ao gerar preparação avulsa', e)
  } finally {
    isLoading.value = false
  }
}
</script>