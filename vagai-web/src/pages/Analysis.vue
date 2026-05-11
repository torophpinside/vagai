<template>
  <div class="space-y-8">
    <div>
      <h1 class="text-4xl font-bold text-white mb-2 font-outfit tracking-tight">Análise de Currículo</h1>
      <p class="text-slate-400"> Faça upload do seu currículo e receba feedback de um Analista de RH Sênior.</p>
    </div>

    <!-- Upload Area -->
    <div class="glass-card p-8">
      <div v-if="!resumeFile && !analysis" class="border-2 border-dashed border-white/10 rounded-2xl p-12 text-center hover:border-indigo-500/50 transition-colors">
        <div class="flex flex-col items-center gap-4">
          <div class="w-16 h-16 bg-indigo-500/10 rounded-full flex items-center justify-center">
            <Upload class="w-8 h-8 text-indigo-400" />
          </div>
          <div>
            <label class="cursor-pointer">
              <span class="text-indigo-400 font-medium hover:text-indigo-300">Clique para fazer upload</span>
              <input type="file" accept=".pdf,.doc,.docx,.txt" class="hidden" @change="handleFileSelect" />
            </label>
            <span class="text-slate-500"> ou arraste o arquivo</span>
          </div>
          <div class="text-xs text-slate-500">PDF, DOC, DOCX ou TXT (max 5MB)</div>
        </div>
      </div>

      <!-- File Selected -->
      <div v-if="resumeFile && !analysis" class="space-y-4">
        <div class="flex items-center justify-between p-4 bg-slate-800/50 rounded-xl">
          <div class="flex items-center gap-3">
            <FileText class="w-8 h-8 text-indigo-400" />
            <div>
              <div class="text-white font-medium">{{ resumeFile.name }}</div>
              <div class="text-xs text-slate-500">{{ formatFileSize(resumeFile.size) }}</div>
            </div>
          </div>
          <button @click="analyzeResume" :disabled="isAnalyzing" class="px-6 py-2 bg-indigo-600 hover:bg-indigo-500 disabled:bg-slate-600 rounded-xl text-white font-medium transition-colors flex items-center gap-2">
            <Loader v-if="isAnalyzing" class="w-5 h-5 animate-spin" />
            <span>{{ isAnalyzing ? 'Analisando...' : 'Analisar' }}</span>
          </button>
        </div>
      </div>

      <!-- Loading -->
      <div v-if="isAnalyzing" class="flex flex-col items-center justify-center py-12">
        <Loader class="w-12 h-12 text-indigo-400 animate-spin mb-4" />
        <div class="text-slate-400">Analisando seu currículo...</div>
        <div class="text-xs text-slate-500 mt-2">Isso pode levar alguns segundos</div>
      </div>

      <!-- Analysis Result -->
      <div v-if="analysis" class="space-y-6">
        <div class="flex items-center justify-between">
          <button @click="reset" class="text-slate-400 hover:text-white text-sm flex items-center gap-2">
            <ArrowLeft class="w-4 h-4" />
            <span>Nova análise</span>
          </button>
        </div>

        <!-- Score Card -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div class="bg-emerald-500/10 border border-emerald-500/20 rounded-2xl p-6">
            <div class="text-emerald-400 font-bold text-xs uppercase tracking-widest mb-2">Pontos Fortes</div>
            <div class="text-3xl font-bold text-white">{{ analysis.strengths?.length || 0 }}</div>
            <div class="text-xs text-slate-500">itens identificados</div>
          </div>
          <div class="bg-amber-500/10 border border-amber-500/20 rounded-2xl p-6">
            <div class="text-amber-400 font-bold text-xs uppercase tracking-widest mb-2">Pontos de Atenção</div>
            <div class="text-3xl font-bold text-white">{{ analysis.weaknesses?.length || 0 }}</div>
            <div class="text-xs text-slate-500">itens identificados</div>
          </div>
          <div class="bg-indigo-500/10 border border-indigo-500/20 rounded-2xl p-6">
            <div class="text-indigo-400 font-bold text-xs uppercase tracking-widest mb-2">Sugestões</div>
            <div class="text-3xl font-bold text-white">{{ analysis.suggestions?.length || 0 }}</div>
            <div class="text-xs text-slate-500">recomendações</div>
          </div>
        </div>

        <!-- Strengths -->
        <div v-if="analysis.strengths?.length" class="bg-emerald-500/5 border border-emerald-500/20 rounded-2xl p-6">
          <div class="flex items-center gap-2 mb-4 text-emerald-400 font-bold text-sm uppercase tracking-widest">
            <ThumbsUp class="w-5 h-5" />
            Pontos Fortes
          </div>
          <ul class="space-y-3">
            <li v-for="(strength, idx) in analysis.strengths" :key="idx" class="flex gap-3 text-slate-300">
              <CheckCircle class="w-5 h-5 text-emerald-400 flex-shrink-0 mt-0.5" />
              <span>{{ strength }}</span>
            </li>
          </ul>
        </div>

        <!-- Weaknesses -->
        <div v-if="analysis.weaknesses?.length" class="bg-amber-500/5 border border-amber-500/20 rounded-2xl p-6">
          <div class="flex items-center gap-2 mb-4 text-amber-400 font-bold text-sm uppercase tracking-widest">
            <AlertTriangle class="w-5 h-5" />
            Pontos de Atenção
          </div>
          <ul class="space-y-3">
            <li v-for="(weakness, idx) in analysis.weaknesses" :key="idx" class="flex gap-3 text-slate-300">
              <AlertCircle class="w-5 h-5 text-amber-400 flex-shrink-0 mt-0.5" />
              <span>{{ weakness }}</span>
            </li>
          </ul>
        </div>

        <!-- Suggestions -->
        <div v-if="analysis.suggestions?.length" class="bg-indigo-500/5 border border-indigo-500/20 rounded-2xl p-6">
          <div class="flex items-center gap-2 mb-4 text-indigo-400 font-bold text-sm uppercase tracking-widest">
            <Lightbulb class="w-5 h-5" />
            Sugestões de Melhoria
          </div>
          <ul class="space-y-3">
            <li v-for="(suggestion, idx) in analysis.suggestions" :key="idx" class="flex gap-3 text-slate-300">
              <ChevronRight class="w-5 h-5 text-indigo-400 flex-shrink-0 mt-0.5" />
              <span>{{ suggestion }}</span>
            </li>
          </ul>
        </div>

        <!-- Full Analysis -->
        <div v-if="analysis.fullAnalysis" class="bg-slate-900/50 border border-white/5 rounded-2xl p-6">
          <div class="flex items-center gap-2 mb-4 text-indigo-400 font-bold text-sm uppercase tracking-widest">
            <BrainCircuit class="w-5 h-5" />
            Análise Completa
          </div>
          <div class="text-slate-300 text-sm leading-relaxed whitespace-pre-line">{{ analysis.fullAnalysis }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { api } from '../services/api'
import { 
  Upload,
  FileText,
  ArrowLeft,
  Loader,
  ThumbsUp,
  AlertTriangle,
  AlertCircle,
  Lightbulb,
  ChevronRight,
  CheckCircle,
  BrainCircuit
} from 'lucide-vue-next'

const resumeFile = ref(null)
const analysis = ref(null)
const isAnalyzing = ref(false)

const handleFileSelect = (event) => {
  const file = event.target.files[0]
  if (file) {
    if (file.size > 5 * 1024 * 1024) {
      alert('Arquivo muito grande. Máximo 5MB.')
      return
    }
    resumeFile.value = file
    analysis.value = null
  }
}

const analyzeResume = async () => {
  if (!resumeFile.value) return

  isAnalyzing.value = true
  const formData = new FormData()
  formData.append('file', resumeFile.value)
  formData.append('type', 'analysis')

  try {
    const response = await api.post('/resumes/analyze', formData, {
      timeout: 180000
    })
    analysis.value = response.data
  } catch (error) {
    console.error('Erro ao analisar:', error)
    alert('Erro ao analisar currículo. Tente novamente.')
  } finally {
    isAnalyzing.value = false
  }
}

const reset = () => {
  resumeFile.value = null
  analysis.value = null
}

const formatFileSize = (bytes) => {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}
</script>