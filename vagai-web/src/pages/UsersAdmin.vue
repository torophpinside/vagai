<template>
  <div class="space-y-8">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-4xl font-bold text-white mb-2 font-outfit tracking-tight">Usuários</h1>
        <p class="text-slate-400">Gerencie as contas do tipo usuário da plataforma.</p>
      </div>
      <div class="flex gap-4 items-center">
        <div class="relative flex-1 max-w-xs">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500" />
          <input v-model="searchQuery" type="text" placeholder="Buscar por nome ou e-mail..." class="input-field pl-10 h-12 w-full" />
        </div>
        <select v-model="filters.status" class="input-field h-12 w-44">
          <option value="active">Ativos</option>
          <option value="suspended">Suspensos</option>
          <option value="deleted">Excluídos</option>
        </select>
      </div>
    </div>

    <div v-if="isLoading" class="flex items-center justify-center h-64">
      <div class="w-12 h-12 border-4 border-indigo-500/30 border-t-indigo-500 rounded-full animate-spin"></div>
    </div>

    <div v-else-if="isError" class="bg-red-500/10 border border-red-500/20 rounded-xl p-4 text-red-400 text-sm">
      Erro ao carregar usuários. Tente novamente.
    </div>

    <div v-else-if="users.length === 0" class="flex flex-col items-center justify-center h-64 text-center">
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
              <th class="px-8 py-5 text-left text-xs font-bold text-slate-400 uppercase tracking-widest">Usuário</th>
              <th class="px-8 py-5 text-left text-xs font-bold text-slate-400 uppercase tracking-widest">Plano</th>
              <th class="px-8 py-5 text-left text-xs font-bold text-slate-400 uppercase tracking-widest">Assinatura</th>
              <th class="px-8 py-5 text-left text-xs font-bold text-slate-400 uppercase tracking-widest">Situação</th>
              <th class="px-8 py-5 text-left text-xs font-bold text-slate-400 uppercase tracking-widest">Cadastro</th>
              <th class="px-8 py-5 text-right text-xs font-bold text-slate-400 uppercase tracking-widest">Ações</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-white/5">
            <tr v-for="u in users" :key="u.id" class="hover:bg-white/5 transition-colors group">
              <td class="px-8 py-6">
                <div class="flex items-center gap-3">
                  <div class="w-9 h-9 rounded-lg bg-slate-700/50 flex items-center justify-center text-xs font-bold text-slate-400 uppercase">
                    {{ u.name?.charAt(0) || '?' }}
                  </div>
                  <div class="flex flex-col">
                    <span class="text-white font-bold group-hover:text-indigo-400 transition-colors">{{ u.name }}</span>
                    <span class="text-xs text-slate-500 font-mono">{{ u.email }}</span>
                  </div>
                </div>
              </td>
              <td class="px-8 py-6">
                <span class="px-3 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider border border-indigo-500/30 text-indigo-300 bg-indigo-500/10">
                  {{ u.plan_slug }}
                </span>
              </td>
              <td class="px-8 py-6 text-slate-400 text-sm">{{ subStatusLabels[u.subscription_status] || u.subscription_status }}</td>
              <td class="px-8 py-6">
                <span :class="situationClass(u)" class="px-3 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider border">
                  {{ situationLabel(u) }}
                </span>
                <div v-if="u.purge_after" class="text-[10px] text-slate-500 mt-1 font-mono">
                  expurgo em {{ formatDate(u.purge_after) }}
                </div>
              </td>
              <td class="px-8 py-6 text-slate-400 text-sm">{{ formatDate(u.created_at) }}</td>
              <td class="px-8 py-6 text-right">
                <div class="flex items-center justify-end gap-2">
                  <button @click="openProfile(u)" class="p-2 bg-indigo-500/10 text-indigo-400 rounded-lg hover:bg-indigo-500 hover:text-white transition-all" title="Ver perfil">
                    <Eye class="w-4 h-4" />
                  </button>
                  <template v-if="!u.deleted_at">
                    <button v-if="!u.suspended" @click="suspendUser(u)" class="p-2 bg-amber-500/10 text-amber-400 rounded-lg hover:bg-amber-500 hover:text-white transition-all" title="Suspender">
                      <Ban class="w-4 h-4" />
                    </button>
                    <button v-else @click="restoreUser(u)" class="p-2 bg-emerald-500/10 text-emerald-400 rounded-lg hover:bg-emerald-500 hover:text-white transition-all" title="Restaurar">
                      <RotateCcw class="w-4 h-4" />
                    </button>
                    <button @click="deleteUser(u)" class="p-2 bg-slate-700/50 text-slate-400 rounded-lg hover:bg-red-500/20 hover:text-red-400 transition-all" title="Excluir">
                      <Trash2 class="w-4 h-4" />
                    </button>
                  </template>
                  <button v-else @click="reactivateUser(u)" class="p-2 bg-emerald-500/10 text-emerald-400 rounded-lg hover:bg-emerald-500 hover:text-white transition-all" title="Reativar">
                    <Undo2 class="w-4 h-4" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div v-if="totalPages > 0" class="flex items-center justify-between px-8 py-4 border-t border-white/5">
        <span class="text-sm text-slate-400">{{ showingInfo }}</span>
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

    <!-- Painel de perfil (edição e atividade) -->
    <div v-if="selected" ref="profilePanel" class="glass-card p-8 space-y-6">
      <div class="flex items-center justify-between">
        <div>
          <h2 class="text-2xl font-bold text-white font-outfit">{{ profile?.name || selected.name }}</h2>
          <p class="text-slate-400 text-sm font-mono">{{ profile?.email || selected.email }}</p>
        </div>
        <button @click="closeProfile" class="p-2 bg-slate-700/50 text-slate-400 rounded-lg hover:bg-slate-600/50 transition-all">
          <X class="w-5 h-5" />
        </button>
      </div>

      <div v-if="profile?.activity" class="grid grid-cols-4 gap-4">
        <div class="bg-slate-800/40 rounded-xl p-4 text-center">
          <div class="text-3xl font-bold text-white">{{ profile.activity.jobs }}</div>
          <div class="text-xs text-slate-400 uppercase tracking-wider mt-1">Vagas</div>
        </div>
        <div class="bg-slate-800/40 rounded-xl p-4 text-center">
          <div class="text-3xl font-bold text-white">{{ profile.activity.resumes }}</div>
          <div class="text-xs text-slate-400 uppercase tracking-wider mt-1">Currículos analisados (mês)</div>
        </div>
        <div class="bg-slate-800/40 rounded-xl p-4 text-center">
          <div class="text-3xl font-bold text-white">{{ profile.activity.matches }}</div>
          <div class="text-xs text-slate-400 uppercase tracking-wider mt-1">Matches</div>
        </div>
        <div class="bg-slate-800/40 rounded-xl p-4 text-center">
          <div class="text-3xl font-bold text-white">{{ profile.activity.applied }}</div>
          <div class="text-xs text-slate-400 uppercase tracking-wider mt-1">Candidatadas</div>
        </div>
      </div>

      <div v-if="formError" class="bg-red-500/10 border border-red-500/20 rounded-xl p-3 text-red-400 text-sm">
        {{ formError }}
      </div>

      <form @submit.prevent="saveProfile" class="grid grid-cols-2 gap-4">
        <div class="space-y-1">
          <label class="text-xs font-bold text-slate-400 uppercase tracking-wider">Nome</label>
          <input v-model="form.name" type="text" class="input-field w-full" />
        </div>
        <div class="space-y-1">
          <label class="text-xs font-bold text-slate-400 uppercase tracking-wider">E-mail</label>
          <input v-model="form.email" type="email" class="input-field w-full" />
        </div>
        <div class="space-y-1">
          <label class="text-xs font-bold text-slate-400 uppercase tracking-wider">Cidade</label>
          <input v-model="form.city" type="text" class="input-field w-full" />
        </div>
        <div class="flex items-end gap-3">
          <button type="submit" :disabled="saving" class="btn-primary flex items-center gap-2">
            <Save class="w-4 h-4" />
            {{ saving ? 'Salvando...' : 'Salvar' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref, computed, watch, nextTick } from 'vue'
import { useQueryClient } from '@tanstack/vue-query'
import {
  useAdminUsers, useAdminUser,
  adminUpdateUser, adminSuspendUser, adminRestoreUser,
  adminDeleteUser, adminReactivateUser
} from '../services/api'
import {
  Eye, Ban, RotateCcw, Trash2, Undo2, X, Save, Search,
  ChevronLeft, ChevronRight
} from 'lucide-vue-next'

const queryClient = useQueryClient()

const filters = reactive({
  search: '',
  status: 'active',
  page: 1,
  limit: 20
})

const searchQuery = ref('')
let searchTimer
watch(searchQuery, (val) => {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => { filters.search = val }, 300)
})
watch([() => filters.search, () => filters.status], () => { filters.page = 1 })

const { data: usersData, isLoading, isError } = useAdminUsers(filters)
const users = computed(() => usersData.value?.data || [])
const total = computed(() => usersData.value?.total || 0)
const totalPages = computed(() => usersData.value?.totalPages || 0)
const showingInfo = computed(() => {
  if (total.value === 0) return 'Nenhum usuário encontrado'
  const from = (filters.page - 1) * filters.limit + 1
  const to = Math.min(filters.page * filters.limit, total.value)
  return `Mostrando ${from} a ${to} de ${total.value} usuários`
})

const subStatusLabels = {
  trial: 'Trial', active: 'Ativa', past_due: 'Pendente',
  canceled: 'Cancelada', expired: 'Expirada'
}

const situationLabel = (u) => {
  if (u.deleted_at) return 'Excluída'
  if (u.suspended) return 'Suspenso'
  return 'Ativo'
}
const situationClass = (u) => {
  if (u.deleted_at) return 'bg-slate-500/10 text-slate-400 border-slate-500/30'
  if (u.suspended) return 'bg-amber-500/10 text-amber-400 border-amber-500/30'
  return 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30'
}

const formatDate = (raw) => {
  if (!raw) return '—'
  return new Date(raw).toLocaleDateString('pt-BR')
}

const invalidate = () => {
  queryClient.invalidateQueries({ queryKey: ['admin'] })
}

// --- Perfil (US2) ---
const selected = ref(null)
const profilePanel = ref(null)
const form = reactive({ name: '', email: '', city: '' })
const formError = ref('')
const saving = ref(false)

const { data: profile } = useAdminUser(computed(() => selected.value?.id))

const openProfile = async (u) => {
  selected.value = u
  formError.value = ''
  await nextTick()
  profilePanel.value?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

watch(profile, (p) => {
  if (!p) return
  form.name = p.name || ''
  form.email = p.email || ''
  form.city = p.city || ''
})

const closeProfile = () => {
  selected.value = null
  formError.value = ''
}

const saveProfile = async () => {
  formError.value = ''
  if (!form.name.trim()) {
    formError.value = 'Nome é obrigatório'
    return
  }
  saving.value = true
  try {
    const id = selected.value.id
    await adminUpdateUser(id, {
      name: form.name.trim(),
      email: form.email.trim(),
      city: form.city.trim()
    })
    await invalidate()
  } catch (err) {
    formError.value = err.response?.data?.error || 'Erro ao salvar usuário'
  } finally {
    saving.value = false
  }
}

// --- Moderação (US3) ---
const suspendUser = async (u) => {
  if (!confirm(`Suspender a conta de ${u.name}? O acesso será bloqueado imediatamente.`)) return
  try {
    await adminSuspendUser(u.id)
    await invalidate()
  } catch (err) {
    alert(err.response?.data?.error || 'Erro ao suspender usuário')
  }
}

const restoreUser = async (u) => {
  try {
    await adminRestoreUser(u.id)
    await invalidate()
  } catch (err) {
    alert(err.response?.data?.error || 'Erro ao restaurar usuário')
  }
}

const deleteUser = async (u) => {
  if (!confirm(`Excluir a conta de ${u.name}? O acesso será bloqueado na hora; os dados ficarão guardados por 30 dias para reativação e depois serão removidos.`)) return
  try {
    await adminDeleteUser(u.id)
    if (selected.value?.id === u.id) closeProfile()
    await invalidate()
  } catch (err) {
    alert(err.response?.data?.error || 'Erro ao excluir usuário')
  }
}

const reactivateUser = async (u) => {
  if (!confirm(`Reativar a conta de ${u.name}?`)) return
  try {
    await adminReactivateUser(u.id)
    await invalidate()
  } catch (err) {
    alert(err.response?.data?.error || 'Erro ao reativar usuário')
  }
}
</script>
