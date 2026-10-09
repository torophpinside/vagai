<template>
  <div v-if="empty" class="text-center py-12">
    <div class="w-12 h-12 mx-auto mb-4 bg-slate-700/40 rounded-2xl flex items-center justify-center text-slate-400">
      <Inbox class="w-6 h-6" />
    </div>
    <div class="text-slate-300 font-semibold">{{ state?.message }}</div>
    <div class="text-sm text-slate-500 mt-1 max-w-sm mx-auto">{{ state?.guidance }}</div>
  </div>
  <div v-else-if="loading" class="flex items-center justify-center h-48">
    <div class="w-10 h-10 border-4 border-indigo-500/30 border-t-indigo-500 rounded-full animate-spin"></div>
  </div>
  <div v-else-if="error" class="text-center py-12 text-red-400 text-sm">
    Não foi possível carregar os dados de analytics.
  </div>
  <slot v-else />
</template>

<script setup>
import { computed } from 'vue'
import { Inbox } from 'lucide-vue-next'

const props = defineProps({
  data: { type: [Object, Array, null], default: null },
  loading: { type: Boolean, default: false },
  error: { type: Boolean, default: false }
})

const state = computed(() => props.data?.empty_state || null)
const empty = computed(() => !!props.data?.empty_state?.empty)
</script>