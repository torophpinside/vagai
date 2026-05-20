<template>
  <div class="space-y-12">
    <div>
      <h1 class="text-4xl font-bold text-white mb-2 font-outfit tracking-tight">Configurações</h1>
      <p class="text-slate-400">Personalize seu scanner, gerencie sua equipe e seu plano.</p>
    </div>

    <!-- Quick Links -->
    <div class="grid grid-cols-2 gap-6">
      <router-link to="/settings/team" class="glass-card p-6 hover:border-indigo-500/30 transition-all cursor-pointer group">
        <div class="flex items-center gap-4">
          <div class="w-12 h-12 bg-indigo-500/10 rounded-xl flex items-center justify-center text-indigo-400 group-hover:bg-indigo-500/20 transition-colors">
            <Users class="w-6 h-6" />
          </div>
          <div>
            <h3 class="font-bold text-white">Equipe</h3>
            <p class="text-sm text-slate-400">Gerencie membros da organização</p>
          </div>
        </div>
      </router-link>
      <router-link to="/settings/billing" class="glass-card p-6 hover:border-indigo-500/30 transition-all cursor-pointer group">
        <div class="flex items-center gap-4">
          <div class="w-12 h-12 bg-emerald-500/10 rounded-xl flex items-center justify-center text-emerald-400 group-hover:bg-emerald-500/20 transition-colors">
            <CreditCard class="w-6 h-6" />
          </div>
          <div>
            <h3 class="font-bold text-white">Billing</h3>
            <p class="text-sm text-slate-400">Plano, pagamentos e limites</p>
          </div>
        </div>
      </router-link>
    </div>
    
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-12">
      <div class="glass-card p-10">
        <div class="flex items-center gap-3 mb-8">
          <div class="w-10 h-10 bg-indigo-500/10 rounded-xl flex items-center justify-center text-indigo-400">
            <Globe class="w-6 h-6" />
          </div>
          <h2 class="text-2xl font-bold text-white font-outfit">Adicionar Fonte</h2>
        </div>

        <form @submit.prevent="handleAddSite" class="space-y-6">
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-400 uppercase tracking-widest ml-1">Nome da Plataforma</label>
            <input v-model="newSite.name" type="text" class="input-field w-full h-12" placeholder="ex: LinkedIn, RemoteOK" required />
          </div>
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-400 uppercase tracking-widest ml-1">URL de Busca</label>
            <input v-model="newSite.url" type="url" class="input-field w-full h-12" placeholder="https://..." required />
          </div>
          <button type="submit" class="btn-primary w-full h-12 flex items-center justify-center gap-2" :disabled="siteMutation.isPending.value">
            <Plus v-if="!siteMutation.isPending.value" class="w-5 h-5" />
            <div v-else class="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
            {{ siteMutation.isPending.value ? 'Analisando site com IA...' : 'Adicionar Fonte' }}
          </button>
        </form>
      </div>

      <div class="glass-card p-10">
        <div class="flex items-center justify-between mb-10">
          <h2 class="text-2xl font-bold text-white font-outfit">Fontes Cadastradas</h2>
          <Globe class="text-slate-500 w-6 h-6" />
        </div>

        <div v-if="sitesLoading" class="flex justify-center py-12">
          <div class="w-10 h-10 border-4 border-indigo-500/30 border-t-indigo-500 rounded-full animate-spin"></div>
        </div>
        <div v-else-if="!sites?.length" class="text-center py-20 bg-slate-950/30 rounded-3xl border border-white/5">
          <div class="text-slate-500 mb-4 flex justify-center">
            <Globe class="w-12 h-12 opacity-20" />
          </div>
          <p class="text-slate-400 font-medium">Nenhuma fonte cadastrada.</p>
        </div>
        <div v-else class="space-y-4">
          <div v-for="site in sites" :key="site.id" class="bg-slate-950/50 border border-white/5 rounded-2xl p-5 flex items-center justify-between hover:border-indigo-500/30 transition-all">
            <div class="flex items-center gap-4">
              <div class="w-10 h-10 bg-indigo-500/10 rounded-xl flex items-center justify-center text-indigo-400">
                <Globe class="w-5 h-5" />
              </div>
              <div>
                <h3 class="font-bold text-white">{{ site.name }}</h3>
                <p class="text-xs text-slate-500 mt-1 flex items-center gap-2">
                  <a :href="site.url" target="_blank" class="hover:text-indigo-400 transition-colors flex items-center gap-1">
                    {{ site.url?.substring(0, 40) }}... <ExternalLink class="w-3 h-3" />
                  </a>
                </p>
              </div>
            </div>
            <div class="flex items-center gap-3">
              <button @click="toggleSiteActive(site)" :class="site.active ? 'bg-emerald-500/10 text-emerald-400 hover:bg-emerald-500/20' : 'bg-slate-700/50 text-slate-400 hover:bg-slate-600/50'" class="p-2 rounded-lg transition-all" :title="site.active ? 'Desativar' : 'Ativar'">
                <Power class="w-4 h-4" />
              </button>
              <button @click="handleDeleteSite(site.id)" class="p-2 bg-slate-700/50 text-slate-400 rounded-lg hover:bg-red-500/20 hover:text-red-400 transition-all">
                <Trash2 class="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="glass-card p-10">
      <div class="flex items-center gap-3 mb-8">
        <div class="w-10 h-10 bg-emerald-500/10 rounded-xl flex items-center justify-center text-emerald-400">
          <FileText class="w-6 h-6" />
        </div>
        <h2 class="text-2xl font-bold text-white font-outfit">Seu Currículo</h2>
      </div>

      <form @submit.prevent="handleUploadResume" class="space-y-6">
        <div class="group relative h-48 border-2 border-dashed border-white/10 hover:border-indigo-500/50 rounded-2xl flex flex-col items-center justify-center transition-all duration-300 cursor-pointer overflow-hidden">
          <input type="file" @change="handleFileChange" class="absolute inset-0 opacity-0 cursor-pointer z-10" accept=".pdf,.txt,.docx" />
          <div class="flex flex-col items-center gap-3 group-hover:scale-110 transition-transform duration-500">
            <div class="w-12 h-12 bg-slate-900 rounded-full flex items-center justify-center text-slate-500">
              <Upload class="w-6 h-6" />
            </div>
            <div class="text-center">
              <p class="text-white font-bold">{{ resumeFile ? resumeFile.name : 'Clique ou arraste seu arquivo' }}</p>
              <p class="text-xs text-slate-500 mt-1">PDF, TXT ou DOCX (Max. 10MB)</p>
            </div>
          </div>
        </div>
        
        <button type="submit" class="w-full h-12 bg-emerald-600 hover:bg-emerald-500 text-white font-bold rounded-xl transition-all shadow-lg shadow-emerald-900/20 active:scale-95 flex items-center justify-center gap-2" :disabled="!resumeFile || resumeMutation.isPending.value">
          <CheckCircle2 v-if="!resumeMutation.isPending.value" class="w-5 h-5" />
          <div v-else class="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
          {{ resumeMutation.isPending.value ? 'Enviando...' : 'Fazer Upload' }}
        </button>
      </form>
    </div>

    <div class="glass-card p-10">
      <div class="flex items-center justify-between mb-10">
        <h2 class="text-2xl font-bold text-white font-outfit">Currículos Ativos</h2>
        <History class="text-slate-500 w-6 h-6" />
      </div>

  <div v-if="resumesLoading" class="flex justify-center py-12">
    <div class="w-10 h-10 border-4 border-indigo-500/30 border-t-indigo-500 rounded-full animate-spin"></div>
  </div>
  <div v-else-if="!resumes || !resumes.length" class="text-center py-20 bg-slate-950/30 rounded-3xl border border-white/5">
    <div class="text-slate-500 mb-4 flex justify-center">
      <FileX class="w-12 h-12 opacity-20" />
    </div>
    <p class="text-slate-400 font-medium">Nenhum currículo cadastrado no sistema.</p>
  </div>
  <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-6">
    <div v-for="resume in resumes" :key="resume.id" class="bg-slate-950/50 border border-white/5 rounded-3xl p-6 hover:border-indigo-500/30 transition-all duration-300 group">
          <div class="flex justify-between items-start mb-6">
            <div class="flex gap-4">
              <div class="w-12 h-12 bg-indigo-500/10 rounded-2xl flex items-center justify-center text-indigo-400">
                <FileText class="w-6 h-6" />
              </div>
              <div>
                <h3 class="font-bold text-white text-lg group-hover:text-indigo-400 transition-colors">{{ resume.name }}</h3>
                <p class="text-xs text-slate-500 mt-1">ID #{{ resume.id }} • {{ formatDate(resume.uploaded_at) }}</p>
              </div>
            </div>
            <span class="px-3 py-1 rounded-full text-[10px] font-bold uppercase tracking-widest border border-emerald-500/20 bg-emerald-500/10 text-emerald-400">Ativo</span>
          </div>
          
          <div v-if="resume.content" class="space-y-3">
            <div class="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Conteúdo Extraído</div>
            <div class="bg-slate-900/50 border border-white/5 p-4 rounded-2xl text-xs text-slate-400 leading-relaxed max-h-40 overflow-y-auto custom-scrollbar">
              {{ resume.content }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useMutation, useQueryClient } from '@tanstack/vue-query'
import { addSite, updateSite, deleteSite, uploadResume, useSites, useResumes } from '../services/api'
import { 
  Globe, 
  Plus, 
  FileText, 
  Upload, 
  CheckCircle2, 
  History, 
  FileX,
  Trash2,
  ExternalLink,
  Users,
  CreditCard,
  Power
} from 'lucide-vue-next'

const queryClient = useQueryClient()
const newSite = ref({ name: '', url: '' })
const resumeFile = ref(null)

const resumesQuery = useResumes()
const sitesQuery = useSites()

const resumes = resumesQuery.data
const sites = sitesQuery.data
const resumesLoading = resumesQuery.isLoading
const sitesLoading = sitesQuery.isLoading

const siteMutation = useMutation({
  mutationFn: addSite,
  onSuccess: () => {
    newSite.value = { name: '', url: '' }
    queryClient.invalidateQueries({ queryKey: ['sites'] })
  }
})

const deleteSiteMutation = useMutation({
  mutationFn: deleteSite,
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['sites'] })
  }
})

const toggleSiteMutation = useMutation({
  mutationFn: ({ id, data }) => updateSite(id, data),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['sites'] })
  }
})

const resumeMutation = useMutation({
  mutationFn: uploadResume,
  onSuccess: () => {
    resumeFile.value = null
    queryClient.invalidateQueries({ queryKey: ['resumes'] })
  }
})

const handleAddSite = () => siteMutation.mutate(newSite.value)
const handleDeleteSite = (id) => {
  if (confirm('Remover esta fonte?')) {
    deleteSiteMutation.mutate(id)
  }
}
const toggleSiteActive = (site) => {
  toggleSiteMutation.mutate({ id: site.id, data: { active: !site.active } })
}
const handleFileChange = (e) => { resumeFile.value = e.target.files[0] }
const handleUploadResume = () => {
  if (resumeFile.value) {
    const formData = new FormData()
    formData.append('file', resumeFile.value)
    resumeMutation.mutate(formData)
  }
}

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  return new Intl.DateTimeFormat('pt-BR', {
    day: '2-digit',
    month: 'short',
    year: 'numeric'
  }).format(new Date(dateStr))
}
</script>
