<template>
  <form @submit.prevent="handleSubmit" class="space-y-8">
    <!-- Personal Info -->
    <div class="glass-card p-8">
      <div class="flex items-center gap-3 mb-6">
        <div class="w-10 h-10 bg-indigo-500/10 rounded-xl flex items-center justify-center text-indigo-400">
          <User class="w-5 h-5" />
        </div>
        <h3 class="text-lg font-bold text-white font-outfit">Informacoes Pessoais</h3>
      </div>
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div class="space-y-1 md:col-span-2">
          <label class="text-xs font-bold text-slate-500 uppercase tracking-widest">Nome completo *</label>
          <input v-model="form.personal_info.name" type="text" class="input-field w-full h-10 text-sm" placeholder="Seu nome" required />
        </div>
        <div class="space-y-1">
          <label class="text-xs font-bold text-slate-500 uppercase tracking-widest">Email</label>
          <input v-model="form.personal_info.email" type="email" class="input-field w-full h-10 text-sm" placeholder="email@exemplo.com" />
        </div>
        <div class="space-y-1">
          <label class="text-xs font-bold text-slate-500 uppercase tracking-widest">Telefone</label>
          <input v-model="form.personal_info.phone" type="text" class="input-field w-full h-10 text-sm" placeholder="(11) 99999-9999" />
        </div>
        <div class="space-y-1">
          <label class="text-xs font-bold text-slate-500 uppercase tracking-widest">Localizacao</label>
          <input v-model="form.personal_info.location" type="text" class="input-field w-full h-10 text-sm" placeholder="Sao Paulo, SP" />
        </div>
        <div class="space-y-1">
          <label class="text-xs font-bold text-slate-500 uppercase tracking-widest">LinkedIn</label>
          <input v-model="form.personal_info.linkedin" type="text" class="input-field w-full h-10 text-sm" placeholder="linkedin.com/in/seu-perfil" />
        </div>
        <div class="space-y-1 md:col-span-2">
          <label class="text-xs font-bold text-slate-500 uppercase tracking-widest">Website</label>
          <input v-model="form.personal_info.website" type="text" class="input-field w-full h-10 text-sm" placeholder="https://seusite.com" />
        </div>
      </div>
    </div>

    <!-- Summary -->
    <div class="glass-card p-8">
      <div class="flex items-center gap-3 mb-6">
        <div class="w-10 h-10 bg-emerald-500/10 rounded-xl flex items-center justify-center text-emerald-400">
          <FileText class="w-5 h-5" />
        </div>
        <h3 class="text-lg font-bold text-white font-outfit">Resumo / Objetivo</h3>
      </div>
      <textarea
        v-model="form.summary"
        class="input-field w-full text-sm min-h-[100px] resize-y"
        placeholder="Descreva seu perfil profissional, objetivos e diferencias..."
        rows="4"
      ></textarea>
    </div>

    <!-- Experience -->
    <div class="glass-card p-8">
      <div class="flex items-center justify-between mb-6">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 bg-blue-500/10 rounded-xl flex items-center justify-center text-blue-400">
            <Briefcase class="w-5 h-5" />
          </div>
          <h3 class="text-lg font-bold text-white font-outfit">Experiencia Profissional</h3>
        </div>
        <button
          type="button"
          @click="addExperience"
          class="px-4 py-2 bg-blue-500/10 hover:bg-blue-500/20 text-blue-400 rounded-xl text-sm font-bold transition-all flex items-center gap-2"
        >
          <Plus class="w-4 h-4" /> Adicionar
        </button>
      </div>
      <div v-if="form.experience.length === 0" class="text-center py-8 text-slate-500 text-sm">
        Nenhuma experiencia adicionada. Clique em "Adicionar" para comecar.
      </div>
      <div v-else class="space-y-4">
        <ExperienceEntry
          v-for="(exp, i) in form.experience"
          :key="i"
          :model-value="exp"
          @update:model-value="form.experience[i] = $event"
          @remove="removeExperience(i)"
        />
      </div>
    </div>

    <!-- Education -->
    <div class="glass-card p-8">
      <div class="flex items-center justify-between mb-6">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 bg-purple-500/10 rounded-xl flex items-center justify-center text-purple-400">
            <GraduationCap class="w-5 h-5" />
          </div>
          <h3 class="text-lg font-bold text-white font-outfit">Educacao</h3>
        </div>
        <button
          type="button"
          @click="addEducation"
          class="px-4 py-2 bg-purple-500/10 hover:bg-purple-500/20 text-purple-400 rounded-xl text-sm font-bold transition-all flex items-center gap-2"
        >
          <Plus class="w-4 h-4" /> Adicionar
        </button>
      </div>
      <div v-if="form.education.length === 0" class="text-center py-8 text-slate-500 text-sm">
        Nenhuma formacao adicionada.
      </div>
      <div v-else class="space-y-4">
        <EducationEntry
          v-for="(edu, i) in form.education"
          :key="i"
          :model-value="edu"
          @update:model-value="form.education[i] = $event"
          @remove="removeEducation(i)"
        />
      </div>
    </div>

    <!-- Skills -->
    <div class="glass-card p-8">
      <div class="flex items-center gap-3 mb-6">
        <div class="w-10 h-10 bg-amber-500/10 rounded-xl flex items-center justify-center text-amber-400">
          <Wrench class="w-5 h-5" />
        </div>
        <h3 class="text-lg font-bold text-white font-outfit">Habilidades</h3>
      </div>
      <TagInput v-model="form.skills" placeholder="Ex: JavaScript, Python, React..." />
    </div>

    <!-- Languages -->
    <div class="glass-card p-8">
      <div class="flex items-center gap-3 mb-6">
        <div class="w-10 h-10 bg-cyan-500/10 rounded-xl flex items-center justify-center text-cyan-400">
          <Globe class="w-5 h-5" />
        </div>
        <h3 class="text-lg font-bold text-white font-outfit">Idiomas</h3>
      </div>
      <TagInput v-model="form.languages" placeholder="Ex: Portugues, Ingles, Espanhol..." />
    </div>

    <!-- Certifications -->
    <div class="glass-card p-8">
      <div class="flex items-center gap-3 mb-6">
        <div class="w-10 h-10 bg-rose-500/10 rounded-xl flex items-center justify-center text-rose-400">
          <Award class="w-5 h-5" />
        </div>
        <h3 class="text-lg font-bold text-white font-outfit">Certificacoes</h3>
      </div>
      <TagInput v-model="form.certifications" placeholder="Ex: AWS Solutions Architect, PMP..." />
    </div>

    <!-- Actions -->
    <div class="flex gap-4">
      <button
        type="submit"
        class="flex-1 h-12 bg-emerald-600 hover:bg-emerald-500 text-white font-bold rounded-xl transition-all shadow-lg shadow-emerald-900/20 active:scale-95 flex items-center justify-center gap-2"
      >
        <Save class="w-5 h-5" /> Salvar Curriculo
      </button>
      <button
        type="button"
        @click="$emit('generate-pdf')"
        class="h-12 px-6 bg-indigo-600 hover:bg-indigo-500 text-white font-bold rounded-xl transition-all shadow-lg shadow-indigo-900/20 active:scale-95 flex items-center justify-center gap-2"
      >
        <Download class="w-5 h-5" /> Gerar PDF
      </button>
    </div>
  </form>
</template>

<script setup>
import { reactive, watch } from 'vue'
import { User, FileText, Briefcase, GraduationCap, Wrench, Globe, Award, Plus, Save, Download } from 'lucide-vue-next'
import ExperienceEntry from './ExperienceEntry.vue'
import EducationEntry from './EducationEntry.vue'
import TagInput from './TagInput.vue'

const props = defineProps({
  initialData: {
    type: Object,
    default: () => ({})
  }
})

const emit = defineEmits(['save', 'generate-pdf'])

const safeData = () => {
  const d = props.initialData || {}
  return {
    personal_info: d.personal_info || { name: '', email: '', phone: '', location: '', linkedin: '', website: '' },
    summary: d.summary || '',
    experience: Array.isArray(d.experience) ? d.experience : [],
    education: Array.isArray(d.education) ? d.education : [],
    skills: Array.isArray(d.skills) ? d.skills : [],
    languages: Array.isArray(d.languages) ? d.languages : [],
    certifications: Array.isArray(d.certifications) ? d.certifications : []
  }
}

const init = safeData()

const form = reactive({
  personal_info: { ...init.personal_info },
  summary: init.summary,
  experience: [...init.experience],
  education: [...init.education],
  skills: [...init.skills],
  languages: [...init.languages],
  certifications: [...init.certifications]
})

watch(() => props.initialData, (newData) => {
  if (newData) {
    Object.assign(form.personal_info, newData.personal_info || {})
    form.summary = newData.summary || ''
    form.experience.splice(0, form.experience.length, ...(newData.experience || []))
    form.education.splice(0, form.education.length, ...(newData.education || []))
    form.skills.splice(0, form.skills.length, ...(newData.skills || []))
    form.languages.splice(0, form.languages.length, ...(newData.languages || []))
    form.certifications.splice(0, form.certifications.length, ...(newData.certifications || []))
  }
}, { deep: true })

const addExperience = () => {
  form.experience.push({ company: '', role: '', start_date: '', end_date: '', description: '' })
}

const removeExperience = (index) => {
  form.experience.splice(index, 1)
}

const addEducation = () => {
  form.education.push({ institution: '', degree: '', field: '', start_date: '', end_date: '', notes: '' })
}

const removeEducation = (index) => {
  form.education.splice(index, 1)
}

const handleSubmit = () => {
  emit('save', { ...form })
}

const getFormData = () => ({ ...form })

defineExpose({ getFormData })
</script>
