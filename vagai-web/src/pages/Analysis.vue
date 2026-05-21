<template>
  <div class="space-y-8">
    <div>
      <h1 class="text-4xl font-bold text-white mb-2 font-outfit tracking-tight">Análise de Currículo</h1>
      <p class="text-slate-400">Faça upload do seu currículo e receba feedback de melhorias e sugestões.</p>
    </div>

    <div class="glass-card p-8">
      <div v-if="!resumeFile && !currentAnalysis" class="border-2 border-dashed border-white/10 rounded-2xl p-12 text-center hover:border-indigo-500/50 transition-colors">
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

      <div v-if="resumeFile && !currentAnalysis" class="space-y-4">
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

      <div v-if="isAnalyzing" class="flex flex-col items-center justify-center py-12">
        <Loader class="w-12 h-12 text-indigo-400 animate-spin mb-4" />
        <div class="text-slate-400">Analisando seu currículo...</div>
        <div class="text-xs text-slate-500 mt-2">Isso pode levar alguns segundos</div>
      </div>

      <div v-if="currentAnalysis" class="space-y-6">
        <div class="flex items-center justify-between">
          <button @click="resetCurrent" class="text-slate-400 hover:text-white text-sm flex items-center gap-2">
            <ArrowLeft class="w-4 h-4" />
            <span>Nova análise</span>
          </button>
          <div class="text-xs text-slate-500">{{ currentAnalysis.file_name }}</div>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div class="bg-emerald-500/10 border border-emerald-500/20 rounded-2xl p-6">
            <div class="text-emerald-400 font-bold text-xs uppercase tracking-widest mb-2">Pontos Fortes</div>
            <div class="text-3xl font-bold text-white">{{ currentAnalysis.strengths?.length || 0 }}</div>
            <div class="text-xs text-slate-500">itens identificados</div>
          </div>
          <div class="bg-amber-500/10 border border-amber-500/20 rounded-2xl p-6">
            <div class="text-amber-400 font-bold text-xs uppercase tracking-widest mb-2">Pontos de Atenção</div>
            <div class="text-3xl font-bold text-white">{{ currentAnalysis.weaknesses?.length || 0 }}</div>
            <div class="text-xs text-slate-500">itens identificados</div>
          </div>
          <div class="bg-indigo-500/10 border border-indigo-500/20 rounded-2xl p-6">
            <div class="text-indigo-400 font-bold text-xs uppercase tracking-widest mb-2">Sugestões</div>
            <div class="text-3xl font-bold text-white">{{ currentAnalysis.suggestions?.length || 0 }}</div>
            <div class="text-xs text-slate-500">recomendações</div>
          </div>
        </div>

        <div v-if="currentAnalysis.strengths?.length" class="bg-emerald-500/5 border border-emerald-500/20 rounded-2xl p-6">
          <div class="flex items-center gap-2 mb-4 text-emerald-400 font-bold text-sm uppercase tracking-widest">
            <ThumbsUp class="w-5 h-5" />
            Pontos Fortes
          </div>
          <ul class="space-y-3">
            <li v-for="(strength, idx) in currentAnalysis.strengths" :key="idx" class="flex gap-3 text-slate-300">
              <CheckCircle class="w-5 h-5 text-emerald-400 flex-shrink-0 mt-0.5" />
              <span>{{ strength }}</span>
            </li>
          </ul>
        </div>

        <div v-if="currentAnalysis.weaknesses?.length" class="bg-amber-500/5 border border-amber-500/20 rounded-2xl p-6">
          <div class="flex items-center gap-2 mb-4 text-amber-400 font-bold text-sm uppercase tracking-widest">
            <AlertTriangle class="w-5 h-5" />
            Pontos de Atenção
          </div>
          <ul class="space-y-3">
            <li v-for="(weakness, idx) in currentAnalysis.weaknesses" :key="idx" class="flex gap-3 text-slate-300">
              <AlertCircle class="w-5 h-5 text-amber-400 flex-shrink-0 mt-0.5" />
              <span>{{ weakness }}</span>
            </li>
          </ul>
        </div>

        <div v-if="currentAnalysis.suggestions?.length" class="bg-indigo-500/5 border border-indigo-500/20 rounded-2xl p-6">
          <div class="flex items-center gap-2 mb-4 text-indigo-400 font-bold text-sm uppercase tracking-widest">
            <Lightbulb class="w-5 h-5" />
            Sugestões de Melhoria
          </div>
          <ul class="space-y-3">
            <li v-for="(suggestion, idx) in currentAnalysis.suggestions" :key="idx" class="flex gap-3 text-slate-300">
              <ChevronRight class="w-5 h-5 text-indigo-400 flex-shrink-0 mt-0.5" />
              <span>{{ suggestion }}</span>
            </li>
          </ul>
        </div>

        <div v-if="currentAnalysis.fullAnalysis" class="bg-slate-900/50 border border-white/5 rounded-2xl p-6">
          <div class="flex items-center gap-2 mb-4 text-indigo-400 font-bold text-sm uppercase tracking-widest">
            <BrainCircuit class="w-5 h-5" />
            Análise Completa
          </div>
          <div class="text-slate-300 text-sm leading-relaxed whitespace-pre-line">{{ currentAnalysis.fullAnalysis }}</div>
        </div>
      </div>
    </div>

    <div class="glass-card p-8">
      <div class="flex items-center gap-3 mb-8">
        <div class="w-10 h-10 bg-indigo-500/10 rounded-xl flex items-center justify-center text-indigo-400">
          <History class="w-6 h-6" />
        </div>
        <h2 class="text-2xl font-bold text-white font-outfit">Histórico de Análises</h2>
      </div>

      <div v-if="analysesLoading" class="flex justify-center py-12">
        <div class="w-10 h-10 border-4 border-indigo-500/30 border-t-indigo-500 rounded-full animate-spin"></div>
      </div>
      <div v-else-if="!analyses?.length" class="text-center py-12 bg-slate-950/30 rounded-3xl border border-white/5">
        <FileX class="w-12 h-12 text-slate-500 mx-auto mb-4 opacity-20" />
        <p class="text-slate-400 font-medium">Nenhuma análise encontrada.</p>
        <p class="text-xs text-slate-500 mt-2">Faça o upload de um currículo acima para começar.</p>
      </div>
      <div v-else class="space-y-3">
        <div v-for="item in analyses" :key="item.id"
          :class="[
            'flex items-center justify-between p-4 rounded-xl border transition-all group',
            selectedId === item.id
              ? 'bg-indigo-500/10 border-indigo-500/30 cursor-default'
              : 'bg-slate-950/50 border-white/5 cursor-pointer hover:border-indigo-500/20'
          ]">
          <div class="flex items-center gap-4 min-w-0 flex-1" @click="viewAnalysis(item)">
            <div class="w-10 h-10 bg-indigo-500/10 rounded-xl flex items-center justify-center text-indigo-400 flex-shrink-0">
              <FileText class="w-5 h-5" />
            </div>
            <div class="min-w-0">
              <div class="text-white font-medium truncate">{{ item.file_name }}</div>
              <div class="text-xs text-slate-500 mt-1">{{ formatDate(item.created_at) }}</div>
            </div>
          </div>
          <div class="flex items-center gap-4 flex-shrink-0 ml-4">
            <div class="flex items-center gap-3 text-xs text-slate-500">
              <span class="text-emerald-400">{{ item.strengths?.length || 0 }} fortes</span>
              <span class="text-amber-400">{{ item.weaknesses?.length || 0 }} atencao</span>
              <span class="text-indigo-400">{{ item.suggestions?.length || 0 }} sugestoes</span>
            </div>
            <button @click.stop="handleDelete(item.id)" :disabled="deletingId === item.id"
              class="p-2 rounded-lg text-slate-500 hover:bg-red-500/20 hover:text-red-400 transition-all opacity-0 group-hover:opacity-100">
              <Trash2 v-if="deletingId !== item.id" class="w-4 h-4" />
              <div v-else class="w-4 h-4 border-2 border-red-400/30 border-t-red-400 rounded-full animate-spin"></div>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { api, useResumeAnalyses, deleteResumeAnalysis } from '../services/api'
import { useQueryClient } from '@tanstack/vue-query'
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
  BrainCircuit,
  History,
  FileX,
  Trash2
} from 'lucide-vue-next'

const queryClient = useQueryClient()
const resumeFile = ref(null)
const currentAnalysis = ref(null)
const isAnalyzing = ref(false)
const selectedId = ref(null)
const deletingId = ref(null)

const { data: analyses, isLoading: analysesLoading } = useResumeAnalyses()

const handleFileSelect = (event) => {
  const file = event.target.files[0]
  if (file) {
    if (file.size > 5 * 1024 * 1024) {
      alert('Arquivo muito grande. Maximo 5MB.')
      return
    }
    resumeFile.value = file
    currentAnalysis.value = null
    selectedId.value = null
  }
}

const analyzeResume = async () => {
  if (!resumeFile.value) return
  isAnalyzing.value = true
  const formData = new FormData()
  formData.append('file', resumeFile.value)
  formData.append('type', 'analysis')
  try {
    const response = await api.post('/resumes/analyze', formData, { timeout: 240000 })
    currentAnalysis.value = response.data
    selectedId.value = response.data.id
    resumeFile.value = null
  } catch (error) {
    console.error('Erro ao analisar:', error)
    alert('Erro ao analisar curriculo. Tente novamente.')
  } finally {
    isAnalyzing.value = false
  }
}

const viewAnalysis = (item) => {
  currentAnalysis.value = item
  selectedId.value = item.id
  resumeFile.value = null
}

const handleDelete = async (id) => {
  if (!confirm('Excluir esta analise?')) return
  deletingId.value = id
  try {
    await deleteResumeAnalysis(id)
    queryClient.invalidateQueries({ queryKey: ['resume-analyses'] })
    if (selectedId.value === id) {
      currentAnalysis.value = null
      selectedId.value = null
    }
  } catch {
    alert('Erro ao excluir analise.')
  } finally {
    deletingId.value = null
  }
}

const resetCurrent = () => {
  currentAnalysis.value = null
  selectedId.value = null
  resumeFile.value = null
}

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  return new Intl.DateTimeFormat('pt-BR', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  }).format(new Date(dateStr))
}

const formatFileSize = (bytes) => {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}
</script>
