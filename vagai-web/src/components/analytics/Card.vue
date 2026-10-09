<template>
  <div class="glass-card p-8">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h2 class="text-xl font-bold text-white font-outfit">{{ title }}</h2>
        <p v-if="subtitle" class="text-sm text-slate-500 mt-1">{{ subtitle }}</p>
      </div>
      <slot name="icon" />
    </div>
    <slot />
  </div>
</template>

<script setup>
defineProps({
  title: { type: String, required: true },
  subtitle: { type: String, default: '' }
})
</script>