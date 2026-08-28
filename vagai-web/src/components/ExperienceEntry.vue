<template>
  <div class="bg-slate-950/50 border border-white/5 rounded-2xl p-5 space-y-4 relative group">
    <div class="flex items-center justify-between">
      <h4 class="text-sm font-bold text-slate-400 uppercase tracking-widest">Experiencia</h4>
      <button
        type="button"
        @click="$emit('remove')"
        class="p-1.5 text-slate-500 hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-all opacity-0 group-hover:opacity-100"
      >
        <Trash2 class="w-4 h-4" />
      </button>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div class="space-y-1">
        <label class="text-xs font-bold text-slate-500 uppercase tracking-widest">Empresa</label>
        <input
          :value="modelValue.company"
          @input="update('company', $event.target.value)"
          type="text"
          class="input-field w-full h-10 text-sm"
          placeholder="Nome da empresa"
        />
      </div>
      <div class="space-y-1">
        <label class="text-xs font-bold text-slate-500 uppercase tracking-widest">Cargo</label>
        <input
          :value="modelValue.role"
          @input="update('role', $event.target.value)"
          type="text"
          class="input-field w-full h-10 text-sm"
          placeholder="Seu cargo"
        />
      </div>
      <div class="space-y-1">
        <label class="text-xs font-bold text-slate-500 uppercase tracking-widest">Data Inicio</label>
        <input
          :value="modelValue.start_date"
          @input="update('start_date', $event.target.value)"
          type="text"
          class="input-field w-full h-10 text-sm"
          placeholder="MM/AAAA"
        />
      </div>
      <div class="space-y-1">
        <label class="text-xs font-bold text-slate-500 uppercase tracking-widest">Data Fim</label>
        <input
          :value="modelValue.end_date"
          @input="update('end_date', $event.target.value)"
          type="text"
          class="input-field w-full h-10 text-sm"
          placeholder="MM/AAAA ou Presente"
        />
      </div>
    </div>

    <div class="space-y-1">
      <label class="text-xs font-bold text-slate-500 uppercase tracking-widest">Descricao da atividade do cargo</label>
      <textarea
        :value="modelValue.description"
        @input="update('description', $event.target.value)"
        class="input-field w-full text-sm min-h-[80px] resize-y"
        placeholder="Descreva suas atividades e responsabilidades neste cargo..."
        rows="3"
      ></textarea>
    </div>
  </div>
</template>

<script setup>
import { Trash2 } from 'lucide-vue-next'

const props = defineProps({
  modelValue: {
    type: Object,
    default: () => ({ company: '', role: '', start_date: '', end_date: '', description: '' })
  }
})

const emit = defineEmits(['update:modelValue', 'remove'])

const update = (field, value) => {
  emit('update:modelValue', { ...props.modelValue, [field]: value })
}
</script>
