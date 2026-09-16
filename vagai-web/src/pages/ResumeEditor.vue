<template>
  <div class="space-y-8">
    <div class="flex items-end justify-between gap-4">
      <div>
        <h1 class="text-4xl font-bold text-white mb-2 font-outfit tracking-tight">Criar seu Curriculo</h1>
        <p class="text-slate-400">Importe um curriculo existente ou comece do zero. Edite todos os campos e gere um PDF profissional.</p>
        <span v-if="resumeId" class="inline-flex items-center gap-1.5 mt-3 px-2.5 py-1 bg-emerald-500/10 border border-emerald-500/20 rounded-lg text-xs text-emerald-300">
          <CircleCheck class="w-3.5 h-3.5" /> Salvo no servidor
        </span>
        <span v-else class="inline-flex items-center gap-1.5 mt-3 px-2.5 py-1 bg-amber-500/10 border border-amber-500/20 rounded-lg text-xs text-amber-300">
          <CircleAlert class="w-3.5 h-3.5" /> Não salvo no servidor
        </span>
      </div>
    </div>

    <!-- Error message -->
    <div v-if="errorMsg" class="glass-card p-6 border-red-500/30 bg-red-500/5">
      <p class="text-red-400 text-sm">{{ errorMsg }}</p>
      <button @click="errorMsg = ''" class="text-xs text-slate-500 mt-2 hover:text-white">Fechar</button>
    </div>

    <!-- Upload / Update Resume (always at top) -->
    <div class="glass-card p-10">
      <div class="flex items-center gap-3 mb-8">
        <div class="w-10 h-10 bg-indigo-500/10 rounded-xl flex items-center justify-center text-indigo-400">
          <Upload class="w-6 h-6" />
        </div>
        <div>
          <h2 class="text-2xl font-bold text-white font-outfit">Importar Curriculo</h2>
          <p class="text-sm text-slate-400">{{ resumeData ? 'Atualize o curriculo atual importando um novo arquivo' : 'Importe um curriculo existente para comecar' }}</p>
        </div>
      </div>

      <form @submit.prevent="handleParse" class="space-y-6">
        <div class="group relative h-48 border-2 border-dashed border-white/10 hover:border-indigo-500/50 rounded-2xl flex flex-col items-center justify-center transition-all duration-300 cursor-pointer overflow-hidden">
          <input type="file" @change="handleFileChange" class="absolute inset-0 opacity-0 cursor-pointer z-10" accept=".pdf,.txt,.docx" />
          <div class="flex flex-col items-center gap-3 group-hover:scale-110 transition-transform duration-500">
            <div class="w-12 h-12 bg-slate-900 rounded-full flex items-center justify-center text-slate-500">
              <Upload class="w-6 h-6" />
            </div>
            <div class="text-center">
              <p class="text-white font-bold">{{ parseFile ? parseFile.name : 'Clique ou arraste seu arquivo' }}</p>
              <p class="text-xs text-slate-500 mt-1">PDF, TXT ou DOCX (Max. 10MB)</p>
            </div>
          </div>
        </div>

        <button
          type="submit"
          class="w-full h-12 bg-indigo-600 hover:bg-indigo-500 text-white font-bold rounded-xl transition-all shadow-lg shadow-indigo-900/20 active:scale-95 flex items-center justify-center gap-2"
          :disabled="!parseFile || parsing"
        >
          <div v-if="parsing" class="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
          <FileSearch v-else class="w-5 h-5" />
          {{ parsing ? 'Analisando curriculo com IA...' : 'Parsear e Editar' }}
        </button>
      </form>

      <div v-if="!resumeData" class="mt-6 text-center">
        <button
          @click="startEmpty"
          class="text-sm text-slate-500 hover:text-indigo-400 transition-colors underline underline-offset-4"
        >
          Ou comece do zero
        </button>
      </div>
    </div>

    <!-- Resume Form -->
    <div v-if="resumeData" class="space-y-6">
      <ResumeForm
        ref="resumeFormRef"
        :initial-data="resumeData"
        @save="handleSave"
        @generate-pdf="handleGeneratePDF"
      />

      <!-- Status Messages -->
      <div v-if="saveMessage" class="fixed bottom-6 right-6 bg-emerald-600 text-white px-6 py-3 rounded-xl shadow-lg z-50">
        {{ saveMessage }}
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Upload, FileSearch, CircleCheck, CircleAlert } from 'lucide-vue-next'
import ResumeForm from '../components/ResumeForm.vue'
import { parseResume, getResumeData, updateResumeData, generateResumePDF, createResume } from '../services/api'

const route = useRoute()
const router = useRouter()
const resumeFormRef = ref(null)

const resumeId = ref(null)
const resumeData = ref(null)
const parseFile = ref(null)
const parsing = ref(false)
const saveMessage = ref('')
const errorMsg = ref('')

const STORAGE_KEY = 'vagai_resume_draft'

const loadFromStorage = () => {
  try {
    const saved = localStorage.getItem(STORAGE_KEY)
    if (saved) {
      const parsed = JSON.parse(saved)
      if (parsed && parsed.personal_info) {
        return parsed
      }
    }
  } catch (e) {
    // ignore
  }
  return null
}

const saveToStorage = (data) => {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(data))
  } catch (e) {
    // ignore
  }
}

onMounted(async () => {
  const id = route.params.id
  if (id && id !== 'new') {
    try {
      const result = await getResumeData(id)
      if (result && result.data) {
        resumeId.value = result.resume_id
        resumeData.value = result.data
      }
    } catch (e) {
      console.error('Erro ao carregar curriculo:', e)
      errorMsg.value = 'Erro ao carregar curriculo do servidor.'
    }
  } else {
    const draft = loadFromStorage()
    if (draft) {
      resumeData.value = draft
    }
  }
})

let autoSaveInterval = null

onMounted(() => {
  autoSaveInterval = setInterval(() => {
    if (resumeFormRef.value && typeof resumeFormRef.value.getFormData === 'function') {
      saveToStorage(resumeFormRef.value.getFormData())
    }
  }, 30000)
})

onUnmounted(() => {
  if (autoSaveInterval) {
    clearInterval(autoSaveInterval)
  }
})

const handleFileChange = (e) => {
  parseFile.value = e.target.files[0]
}

const handleParse = async () => {
  if (!parseFile.value) return
  parsing.value = true
  errorMsg.value = ''
  try {
    const formData = new FormData()
    formData.append('file', parseFile.value)
    if (resumeId.value) {
      formData.append('resume_id', resumeId.value)
    }
    const response = await parseResume(formData)

    const body = response.data || response

    if (body.error) {
      errorMsg.value = body.error
      return
    }

    if (body.data) {
      resumeData.value = body.data
      resumeId.value = body.resume?.id || null
    } else if (body.resume) {
      // Fallback: try to parse Data field from resume
      try {
        const parsed = JSON.parse(body.resume.data || '{}')
        resumeData.value = parsed
        resumeId.value = body.resume.id
      } catch (e) {
        resumeData.value = makeEmptyData()
        resumeId.value = body.resume.id
      }
    } else {
      errorMsg.value = 'Resposta inesperada do servidor.'
    }
  } catch (e) {
    console.error('Erro ao parsear curriculo:', e)
    const msg = e.response?.data?.error || 'Erro ao parsear curriculo. Verifique se o servidor esta rodando.'
    errorMsg.value = msg
  } finally {
    parsing.value = false
  }
}

const makeEmptyData = () => ({
  personal_info: { name: '', email: '', phone: '', location: '', linkedin: '', website: '' },
  summary: '',
  experience: [],
  education: [],
  skills: [],
  languages: [],
  certifications: []
})

const startEmpty = () => {
  resumeData.value = makeEmptyData()
}

const navigateToCreated = (id) => {
  if (route.params.id === 'new' || !route.params.id) {
    router.replace({ params: { id: String(id) } })
  }
}

const ensureResumeSaved = async (data) => {
  if (resumeId.value) {
    await updateResumeData(resumeId.value, data)
    return resumeId.value
  }
  const response = await createResume(data)
  const body = response.data || response
  resumeId.value = body.resume?.id || null
  if (!resumeId.value) {
    throw new Error('Falha ao criar curriculo no servidor')
  }
  navigateToCreated(resumeId.value)
  return resumeId.value
}

const handleSave = async (data) => {
  try {
    await ensureResumeSaved(data)
    saveToStorage(data)
    showSaveMessage('Curriculo salvo com sucesso!')
  } catch (e) {
    console.error('Erro ao salvar:', e)
    saveToStorage(data)
    showSaveMessage(e.response?.data?.error || 'Salvo localmente (erro ao sincronizar)')
  }
}

const handleGeneratePDF = async () => {
  try {
    let id = resumeId.value
    if (!id && resumeFormRef.value && typeof resumeFormRef.value.getFormData === 'function') {
      id = await ensureResumeSaved(resumeFormRef.value.getFormData())
    }
    if (!id) {
      errorMsg.value = 'Salve o curriculo antes de gerar o PDF.'
      return
    }
    const response = await generateResumePDF(id)
    const blob = new Blob([response.data], { type: 'application/pdf' })
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'curriculo.pdf'
    document.body.appendChild(a)
    a.click()
    window.URL.revokeObjectURL(url)
    document.body.removeChild(a)
  } catch (e) {
    console.error('Erro ao gerar PDF:', e)
    errorMsg.value = e.response?.data?.error || 'Erro ao gerar PDF. Tente novamente.'
  }
}

const showSaveMessage = (msg) => {
  saveMessage.value = msg
  setTimeout(() => { saveMessage.value = '' }, 3000)
}
</script>
