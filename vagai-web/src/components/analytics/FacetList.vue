<template>
  <div>
    <h3 class="text-xs text-slate-400 uppercase tracking-widest mb-3">{{ title }}</h3>
    <ul class="space-y-3">
      <li v-for="(item, i) in items" :key="i" class="flex items-center gap-3">
        <span class="w-1.5 h-1.5 rounded-full" :class="color"></span>
        <span class="flex-1 text-sm text-slate-300 truncate">{{ item[labelKey] || item.skill || item.name }}</span>
        <span class="text-xs font-bold text-slate-400">{{ item.count }}</span>
      </li>
    </ul>
    <div v-if="!items?.length" class="text-slate-500 text-sm py-6 text-center">Sem dados.</div>
  </div>
</template>

<script setup>
defineProps({
  title: { type: String, required: true },
  items: { type: Array, default: () => [] },
  color: { type: String, default: 'bg-slate-500' },
  labelKey: { type: String, default: 'skill' }
})
</script>