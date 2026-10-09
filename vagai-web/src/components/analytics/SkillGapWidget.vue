<template>
  <Card title="Gap de Skills" subtitle="Sobreposição entre vagas e candidatos por categoria">
    <template #icon><Puzzle class="w-6 h-6 text-lime-400" /></template>
    <EmptyState :data="data" :loading="loading" :error="error">
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
        <div class="h-[280px]">
          <h3 class="text-xs text-slate-400 uppercase tracking-widest mb-4">Match Rate por Categoria</h3>
          <Bar :data="categoryChart" :options="barOptions" />
        </div>
        <div class="space-y-6">
          <ChipList title="Habilidades em Demanda" :items="data?.in_demand_skills" color="bg-lime-500" />
          <ChipList title="Habilidades Sobrepostas" :items="data?.overlapping_skills" color="bg-indigo-500" />
        </div>
      </div>
    </EmptyState>
  </Card>
</template>

<script setup>
import { computed } from 'vue'
import { Puzzle } from 'lucide-vue-next'
import { useAnalyticsSkillsGap } from '../../services/api'
import Card from './Card.vue'
import EmptyState from './EmptyState.vue'
import ChipList from './ChipList.vue'
import { Bar, barOptions, PALETTE } from './charts'

const { data, isLoading: loading, isError: error } = useAnalyticsSkillsGap()

const categoryChart = computed(() => ({
  labels: Object.keys(data.value?.match_rate_by_category || {}),
  datasets: [{
    label: '%',
    data: Object.values(data.value?.match_rate_by_category || {}),
    backgroundColor: PALETTE,
    borderRadius: 4
  }]
}))
</script>