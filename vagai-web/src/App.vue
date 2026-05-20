<template>
  <div v-if="!isAuthenticated" class="min-h-screen bg-[#0f172a]">
    <router-view />
  </div>
  <div v-else class="flex min-h-screen bg-[#0f172a] text-slate-200 overflow-hidden">
    <!-- Sidebar -->
    <aside class="w-72 bg-slate-900/50 border-r border-white/5 backdrop-blur-xl flex flex-col">
      <div class="p-8">
        <div class="flex items-center gap-3 mb-12">
          <div class="w-10 h-10 bg-indigo-600 rounded-xl flex items-center justify-center shadow-lg shadow-indigo-500/20">
            <Target class="text-white w-6 h-6" />
          </div>
          <span class="text-2xl font-bold font-outfit text-white tracking-tight">VagAI</span>
        </div>

        <nav class="space-y-2">
          <router-link to="/" class="nav-link" active-class="active">
            <LayoutDashboard class="w-5 h-5" />
            <span>Dashboard</span>
          </router-link>
          <router-link to="/jobs" class="nav-link" active-class="active">
            <Briefcase class="w-5 h-5" />
            <span>Vagas</span>
          </router-link>
          <router-link to="/matches" class="nav-link" active-class="active">
            <CheckCircle2 class="w-5 h-5" />
            <span>Matches</span>
          </router-link>
          <router-link to="/applied" class="nav-link" active-class="active">
            <Send class="w-5 h-5" />
            <span>Candidatadas</span>
          </router-link>
          <router-link to="/analysis" class="nav-link" active-class="active">
            <FileText class="w-5 h-5" />
            <span>Análise</span>
          </router-link>
          <router-link to="/settings" class="nav-link" active-class="active">
            <SettingsIcon class="w-5 h-5" />
            <span>Configurações</span>
          </router-link>
        </nav>
      </div>

      <div class="mt-auto p-8">
        <DropdownMenu v-model:open="showUserMenu" position="top-left" width="" panel-class="right-0">
          <template #trigger="{ open, toggle }">
            <div class="glass-card p-4 flex items-center gap-4" @click="toggle">
              <div class="w-10 h-10 rounded-full bg-indigo-500/20 flex items-center justify-center">
                <User class="w-6 h-6 text-indigo-400" />
              </div>
              <div class="flex-1 min-w-0">
                <div class="text-sm font-bold text-white truncate">{{ userName }}</div>
                <div class="text-xs text-slate-500 truncate">{{ orgName }}</div>
              </div>
              <button class="text-slate-400 hover:text-white">
                <ChevronUp v-if="open" class="w-4 h-4" />
                <ChevronDown v-else class="w-4 h-4" />
              </button>
            </div>
          </template>
          <template #default="{ close }">
            <router-link @click="close" to="/settings" class="block px-3 py-2 text-sm text-slate-300 hover:bg-white/5 rounded-lg">
              <SettingsIcon class="w-4 h-4 inline mr-2" />
              Configurações
            </router-link>
            <router-link @click="close" to="/settings/billing" class="block px-3 py-2 text-sm text-slate-300 hover:bg-white/5 rounded-lg">
              <CreditCard class="w-4 h-4 inline mr-2" />
              Billing
            </router-link>
            <button @click="handleLogout" class="w-full text-left px-3 py-2 text-sm text-red-400 hover:bg-white/5 rounded-lg">
              <LogOut class="w-4 h-4 inline mr-2" />
              Sair
            </button>
          </template>
        </DropdownMenu>
      </div>
    </aside>

    <!-- Main Content -->
    <main class="flex-1 overflow-y-auto bg-gradient-to-br from-slate-950 via-slate-900 to-indigo-950/20">
      <header class="h-20 border-b border-white/5 flex items-center justify-between px-12 backdrop-blur-md sticky top-0 z-10">
        <div class="text-slate-400 text-sm">
          Bem-vindo de volta, <span class="text-white font-bold">{{ firstName }}</span>
        </div>
        <div class="flex items-center gap-6">
          <div class="w-[1px] h-6 bg-white/10"></div>
          <div class="text-right">
            <div class="text-xs text-slate-500">Status do Scanner</div>
            <div class="text-xs text-emerald-400 font-bold flex items-center gap-1">
              <span class="w-1.5 h-1.5 bg-emerald-400 rounded-full animate-pulse"></span>
              Ativo
            </div>
          </div>
        </div>
      </header>

      <div class="p-12 animate-in">
        <router-view />
      </div>
    </main>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import {
  Target,
  LayoutDashboard,
  Briefcase,
  CheckCircle2,
  Send,
  FileText,
  Settings as SettingsIcon,
  User,
  ChevronDown,
  ChevronUp,
  CreditCard,
  LogOut
} from 'lucide-vue-next'
import { useAuth } from './composables/auth'
import DropdownMenu from './components/DropdownMenu.vue'

const router = useRouter()
const { user, isAuthenticated, initAuth, logout } = useAuth()

const showUserMenu = ref(false)

onMounted(() => {
  initAuth()
})

const userName = computed(() => user.value?.name || 'Usuário')
const firstName = computed(() => user.value?.name?.split(' ')[0] || 'Usuário')
const orgName = computed(() => user.value?.organization?.name || 'VagAI')

function handleLogout() {
  logout()
  router.push('/login')
}
</script>
