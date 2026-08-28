<template>
  <div class="space-y-2">
    <div class="flex flex-wrap gap-2">
      <span
        v-for="(tag, i) in modelValue"
        :key="i"
        class="inline-flex items-center gap-1.5 px-3 py-1 bg-indigo-500/10 border border-indigo-500/20 rounded-full text-sm text-indigo-300"
      >
        {{ tag }}
        <button
          type="button"
          @click="removeTag(i)"
          class="w-4 h-4 flex items-center justify-center rounded-full bg-indigo-500/20 hover:bg-red-500/30 hover:text-red-400 transition-colors text-xs"
        >
          &times;
        </button>
      </span>
    </div>
    <div class="flex gap-2">
      <input
        v-model="newTag"
        @keydown.enter.prevent="addTag"
        type="text"
        class="input-field flex-1 h-10 text-sm"
        :placeholder="placeholder"
      />
      <button
        type="button"
        @click="addTag"
        class="px-4 h-10 bg-slate-700/50 hover:bg-indigo-500/20 text-slate-300 hover:text-indigo-400 rounded-xl text-sm font-bold transition-all"
      >
        Adicionar
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  placeholder: { type: String, default: 'Digite e pressione Enter' }
})

const emit = defineEmits(['update:modelValue'])

const newTag = ref('')

const addTag = () => {
  const val = newTag.value.trim()
  if (val && !props.modelValue.includes(val)) {
    emit('update:modelValue', [...props.modelValue, val])
  }
  newTag.value = ''
}

const removeTag = (index) => {
  const updated = [...props.modelValue]
  updated.splice(index, 1)
  emit('update:modelValue', updated)
}
</script>
