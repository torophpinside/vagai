<template>
  <div>
    <h3 class="text-xs text-slate-400 uppercase tracking-widest mb-3">{{ title }}</h3>
    <div class="flex flex-wrap gap-2">
      <span v-for="item in items" :key="item" class="px-3 py-1 bg-slate-900/60 border border-white/5 rounded-full text-xs text-slate-300">{{ item }}</span>
      <span v-if="!items?.length" class="text-slate-500 text-sm">Sem dados.</span>
    </div>
  </div>
</template>

<script setup>
defineProps({
  title: { type: String, required: true },
  items: { type: Array, default: () => [] },
  color: { type: String, default: 'bg-slate-500' }
})
</script>