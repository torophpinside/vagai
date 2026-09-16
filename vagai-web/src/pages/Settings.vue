<template>
  <div class="space-y-12">
    <div>
      <h1 class="text-4xl font-bold text-white mb-2 font-outfit tracking-tight">Configurações</h1>
      <p class="text-slate-400">Personalize seu scanner, gerencie sua equipe e seu plano.</p>
    </div>

    <!-- Perfil do Usuário -->
    <div class="glass-card p-10">
      <div class="flex items-center gap-3 mb-8">
        <div class="w-10 h-10 bg-indigo-500/10 rounded-xl flex items-center justify-center text-indigo-400">
          <MapPin class="w-6 h-6" />
        </div>
        <div>
          <h2 class="text-2xl font-bold text-white font-outfit">Seu Perfil</h2>
          <p class="text-sm text-slate-400">Configure sua localização para melhor matching</p>
        </div>
      </div>

      <form @submit.prevent="handleUpdateProfile" class="space-y-6">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-400 uppercase tracking-widest ml-1">Nome</label>
            <input v-model="profileForm.name" type="text" class="input-field w-full h-12" placeholder="Seu nome" />
          </div>
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-400 uppercase tracking-widest ml-1">Cidade</label>
            <input v-model="profileForm.city" type="text" class="input-field w-full h-12" placeholder="ex: São Paulo, Rio de Janeiro" />
            <p class="text-xs text-slate-500 ml-1">Usada para filtrar vagas presenciais e híbridas</p>
          </div>
        </div>
        <button type="submit" class="h-10 px-6 bg-indigo-600 hover:bg-indigo-500 text-white font-bold rounded-xl transition-all shadow-lg shadow-indigo-900/20 active:scale-95 flex items-center justify-center gap-2" :disabled="profileMutation.isPending.value">
          <CheckCircle2 v-if="!profileMutation.isPending.value" class="w-4 h-4" />
          <div v-else class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
          {{ profileMutation.isPending.value ? 'Salvando...' : 'Salvar Perfil' }}
        </button>
      </form>
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
        <div class="w-10 h-10 bg-red-500/10 rounded-xl flex items-center justify-center text-red-400">
          <Ban class="w-6 h-6" />
        </div>
        <div>
          <h2 class="text-2xl font-bold text-white font-outfit">Palavras-chave de bloqueio</h2>
          <p class="text-sm text-slate-400">Vagas que citarem essas palavras perdem 3 pontos no score de match.</p>
        </div>
      </div>

      <div class="space-y-6">
        <div class="flex gap-3">
          <input v-model="newKeyword" @keyup.enter="addKeyword" type="text" class="input-field flex-1 h-12" placeholder="ex: freela, CLT, bilingue..." />
          <button type="button" @click="addKeyword" class="h-12 px-5 bg-red-600/80 hover:bg-red-500 text-white font-bold rounded-xl transition-all flex items-center gap-2">
            <Plus class="w-5 h-5" /> Adicionar
          </button>
        </div>

        <div v-if="keywordsLoading" class="flex justify-center py-8">
          <div class="w-8 h-8 border-4 border-red-500/30 border-t-red-500 rounded-full animate-spin"></div>
        </div>
        <div v-else-if="!keywords.length" class="text-center py-10 bg-slate-950/30 rounded-2xl border border-white/5">
          <p class="text-slate-400 font-medium">Nenhuma palavra de bloqueio configurada.</p>
        </div>
        <div v-else class="flex flex-wrap gap-3">
          <span v-for="(kw, i) in keywords" :key="`${kw}-${i}`" class="flex items-center gap-2 px-4 py-2 bg-red-500/10 border border-red-500/20 rounded-xl text-sm text-red-300 font-medium">
            {{ kw }}
            <button @click="removeKeyword(i)" class="text-red-400/60 hover:text-red-300 transition-colors" :title="`Remover ${kw}`">
              <X class="w-4 h-4" />
            </button>
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { useMutation, useQueryClient } from '@tanstack/vue-query'
import { addSite, updateSite, deleteSite, useSites, useMe, updateProfile, useNegativeKeywords, updateNegativeKeywords } from '../services/api'
import { useAuth } from '../composables/auth'
import {
  Globe,
  Plus,
  Trash2,
  ExternalLink,
  Users,
  CreditCard,
  Power,
  MapPin,
  Ban,
  X
} from 'lucide-vue-next'

const queryClient = useQueryClient()
const { updateUser } = useAuth()
const newSite = ref({ name: '', url: '' })
const newKeyword = ref('')
const keywords = ref([])

const sitesQuery = useSites()
const meQuery = useMe()
const negativeKeywordsQuery = useNegativeKeywords()

const sites = sitesQuery.data
const sitesLoading = sitesQuery.isLoading
const keywordsLoading = negativeKeywordsQuery.isLoading

watch(() => negativeKeywordsQuery.data?.value, (list) => {
  if (Array.isArray(list)) {
    keywords.value = [...list]
  }
}, { immediate: true })

const profileForm = ref({ name: '', city: '' })

watch(() => meQuery.data?.value, (resp) => {
  if (resp?.user) {
    profileForm.value = {
      name: resp.user.name || '',
      city: resp.user.city || ''
    }
  }
}, { immediate: true })

const profileMutation = useMutation({
  mutationFn: updateProfile,
  onSuccess: (response) => {
    queryClient.invalidateQueries({ queryKey: ['me'] })
    if (response?.data?.user) {
      updateUser(response.data.user)
    }
  }
})

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

const handleAddSite = () => siteMutation.mutate(newSite.value)
const handleDeleteSite = (id) => {
  if (confirm('Remover esta fonte?')) {
    deleteSiteMutation.mutate(id)
  }
}
const toggleSiteActive = (site) => {
  toggleSiteMutation.mutate({ id: site.id, data: { active: !site.active } })
}

const handleUpdateProfile = () => {
  profileMutation.mutate(profileForm.value)
}

const keywordMutation = useMutation({
  mutationFn: updateNegativeKeywords,
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['negative-keywords'] })
  }
})

const addKeyword = () => {
  const trimmed = newKeyword.value.trim()
  if (!trimmed) return
  if (!keywords.value.some(k => k.toLowerCase() === trimmed.toLowerCase())) {
    keywords.value.push(trimmed)
    keywordMutation.mutate(keywords.value)
  }
  newKeyword.value = ''
}

const removeKeyword = (index) => {
  if (index < 0 || index >= keywords.value.length) return
  keywords.value.splice(index, 1)
  keywordMutation.mutate(keywords.value)
}
</script>
