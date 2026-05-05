<template>
  <div>
    <div class="flex items-center justify-between mb-8">
      <div>
        <h1 class="text-2xl font-bold font-outfit text-white">Equipe</h1>
        <p class="text-slate-400 mt-1">Gerencie os membros da sua organização</p>
      </div>
      <button @click="showInviteModal = true" class="btn-primary flex items-center gap-2">
        <UserPlus class="w-5 h-5" />
        Convidar membro
      </button>
    </div>

    <div class="glass-card">
      <div class="p-6 border-b border-white/5">
        <h2 class="text-lg font-semibold text-white">Membros</h2>
      </div>

      <div class="divide-y divide-white/5">
        <div v-for="member in members" :key="member.id" class="p-6 flex items-center justify-between">
          <div class="flex items-center gap-4">
            <div class="w-10 h-10 rounded-full bg-indigo-500/20 flex items-center justify-center">
              <User class="w-5 h-5 text-indigo-400" />
            </div>
            <div>
              <div class="text-sm font-medium text-white">{{ member.name }}</div>
              <div class="text-xs text-slate-500">{{ member.email }}</div>
            </div>
          </div>
          <div class="flex items-center gap-3">
            <span :class="['px-2 py-1 rounded text-xs font-medium', roleBadgeClass(member.role)]">
              {{ member.role }}
            </span>
            <button v-if="member.role !== 'owner'" @click="removeMember(member.id)" class="text-slate-500 hover:text-red-400">
              <X class="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>

      <div v-if="members.length === 0" class="p-12 text-center text-slate-500">
        Nenhum membro encontrado. Convide sua equipe para colaborar.
      </div>
    </div>

    <!-- Invite Modal -->
    <div v-if="showInviteModal" class="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50" @click.self="showInviteModal = false">
      <div class="glass-card p-6 w-full max-w-md">
        <h3 class="text-lg font-semibold text-white mb-4">Convidar membro</h3>
        <form @submit.prevent="handleInvite">
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-slate-300 mb-2">Email</label>
              <input v-model="inviteEmail" type="email" required class="w-full px-4 py-3 bg-slate-800/50 border border-white/10 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-indigo-500" placeholder="email@exemplo.com" />
            </div>
            <div>
              <label class="block text-sm font-medium text-slate-300 mb-2">Função</label>
              <select v-model="inviteRole" class="w-full px-4 py-3 bg-slate-800/50 border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-indigo-500">
                <option value="member">Membro</option>
                <option value="admin">Administrador</option>
              </select>
            </div>
          </div>
          <div class="flex gap-3 mt-6">
            <button type="button" @click="showInviteModal = false" class="flex-1 py-3 bg-slate-700 hover:bg-slate-600 text-white font-medium rounded-lg">Cancelar</button>
            <button type="submit" class="flex-1 py-3 bg-indigo-600 hover:bg-indigo-700 text-white font-medium rounded-lg">Enviar convite</button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { User, UserPlus, X } from 'lucide-vue-next'

const members = ref([
  { id: 1, name: 'Você', email: 'seu@email.com', role: 'owner' }
])
const showInviteModal = ref(false)
const inviteEmail = ref('')
const inviteRole = ref('member')

function roleBadgeClass(role) {
  const classes = {
    owner: 'bg-indigo-500/20 text-indigo-400',
    admin: 'bg-emerald-500/20 text-emerald-400',
    member: 'bg-slate-500/20 text-slate-400'
  }
  return classes[role] || classes.member
}

function removeMember(id) {
  members.value = members.value.filter(m => m.id !== id)
}

async function handleInvite() {
  members.value.push({
    id: Date.now(),
    name: inviteEmail.value.split('@')[0],
    email: inviteEmail.value,
    role: inviteRole.value
  })
  showInviteModal.value = false
  inviteEmail.value = ''
}
</script>
