<template>
  <div class="bg-slate-900/60 border border-white/5 rounded-xl p-5">
    <div class="text-xs text-slate-500 uppercase tracking-widest mb-2">{{ label }}</div>
    <div class="text-3xl font-bold text-white font-outfit">{{ value }}{{ suffix }}</div>
  </div>
</template>

<script setup>
defineProps({
  label: { type: String, required: true },
  value: { type: [Number, String], default: 0 },
  suffix: { type: String, default: '' }
})
</script>