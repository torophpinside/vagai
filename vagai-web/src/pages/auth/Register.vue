<template>
  <div class="min-h-screen bg-[#0f172a] flex items-center justify-center p-4">
    <div class="w-full max-w-md">
      <div class="text-center mb-8">
        <div class="flex items-center justify-center gap-3 mb-4">
          <div class="w-12 h-12 bg-indigo-600 rounded-xl flex items-center justify-center shadow-lg shadow-indigo-500/20">
            <Target class="text-white w-7 h-7" />
          </div>
          <span class="text-3xl font-bold font-outfit text-white tracking-tight">VagAI</span>
        </div>
        <p class="text-slate-400">Crie sua conta e comece a usar IA para encontrar vagas</p>
      </div>

      <div class="glass-card p-8">
        <form @submit.prevent="handleRegister" class="space-y-5">
          <div>
            <label class="block text-sm font-medium text-slate-300 mb-2">Nome</label>
            <input
              v-model="name"
              type="text"
              required
              minlength="2"
              class="w-full px-4 py-3 bg-slate-800/50 border border-white/10 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
              placeholder="Seu nome"
            />
          </div>

          <div>
            <label class="block text-sm font-medium text-slate-300 mb-2">Email</label>
            <input
              v-model="email"
              type="email"
              required
              class="w-full px-4 py-3 bg-slate-800/50 border border-white/10 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
              placeholder="seu@email.com"
            />
          </div>

          <div>
            <label class="block text-sm font-medium text-slate-300 mb-2">Senha</label>
            <input
              v-model="password"
              type="password"
              required
              minlength="8"
              class="w-full px-4 py-3 bg-slate-800/50 border border-white/10 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
              placeholder="Mínimo 8 caracteres"
            />
          </div>

          <div>
            <label class="block text-sm font-medium text-slate-300 mb-2">Nome da Empresa</label>
            <input
              v-model="organization"
              type="text"
              required
              minlength="2"
              class="w-full px-4 py-3 bg-slate-800/50 border border-white/10 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
              placeholder="Sua empresa ou nome pessoal"
            />
          </div>

          <div v-if="error" class="p-3 bg-red-500/10 border border-red-500/20 rounded-lg text-red-400 text-sm">
            {{ error }}
          </div>

          <button
            type="submit"
            :disabled="loading"
            class="w-full py-3 bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white font-semibold rounded-lg transition-colors flex items-center justify-center gap-2"
          >
            <Loader2 v-if="loading" class="w-5 h-5 animate-spin" />
            {{ loading ? 'Criando conta...' : 'Criar conta grátis' }}
          </button>
        </form>

        <div class="mt-6 text-center">
          <p class="text-slate-400 text-sm">
            Já tem conta?
            <router-link to="/login" class="text-indigo-400 hover:text-indigo-300 font-medium">
              Fazer login
            </router-link>
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { Target, Loader2 } from 'lucide-vue-next'
import { useAuth } from '../../composables/auth'

const router = useRouter()
const { register } = useAuth()

const name = ref('')
const email = ref('')
const password = ref('')
const organization = ref('')
const error = ref('')
const loading = ref(false)

async function handleRegister() {
  loading.value = true
  error.value = ''

  try {
    await register(name.value, email.value, password.value, organization.value)
    router.push('/')
  } catch (err) {
    error.value = err.response?.data?.error || 'Erro ao criar conta'
  } finally {
    loading.value = false
  }
}
</script>
