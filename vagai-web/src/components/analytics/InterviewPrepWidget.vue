<template>
  <Card title="Preparação para Entrevistas" subtitle="Categorias de perguntas, fontes e score de verificação">
    <template #icon><GraduationCap class="w-6 h-6 text-pink-400" /></template>
    <EmptyState :data="data" :loading="loading" :error="error">
      <div class="grid grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <Kpi :label="'Preparações'" :value="data?.total_preparations" />
        <Kpi :label="'Score Médio'" :value="data?.average_verification_score?.toFixed?.(1)" />
        <Kpi :label="'Tendência'" :value="trendLabel(data?.verification_trend)" />
        <Kpi :label="'Categorias'" :value="categoryCount" />
      </div>
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
        <div class="h-[260px]">
          <h3 class="text-xs text-slate-400 uppercase tracking-widest mb-4">Por Categoria</h3>
          <Doughnut :data="categoryChart" :options="pieOptions" />
        </div>
        <div class="h-[260px]">
          <h3 class="text-xs text-slate-400 uppercase tracking-widest mb-4">Por Fonte</h3>
          <Bar :data="sourceChart" :options="barOptions" />
        </div>
      </div>
    </EmptyState>
  </Card>
</template>

<script setup>
import { computed } from 'vue'
import { GraduationCap } from 'lucide-vue-next'
import { useAnalyticsInterviewPrep } from '../../services/api'
import Card from './Card.vue'
import EmptyState from './EmptyState.vue'
import Kpi from './Kpi.vue'
import { Bar, Doughnut, barOptions, pieOptions, PALETTE } from './charts'

const { data, isLoading: loading, isError: error } = useAnalyticsInterviewPrep()

const categoryChart = computed(() => ({
  labels: Object.keys(data.value?.category_distribution || {}),
  datasets: [{ data: Object.values(data.value?.category_distribution || {}), backgroundColor: PALETTE }]
}))

const sourceChart = computed(() => ({
  labels: Object.keys(data.value?.source_distribution || {}),
  datasets: [{ label: 'Preparações', data: Object.values(data.value?.source_distribution || {}), backgroundColor: PALETTE, borderRadius: 4 }]
}))

const categoryCount = computed(() => Object.keys(data.value?.category_distribution || {}).length)

function trendLabel(t) {
  return ({ improving: 'Melhorando', stable: 'Estável', declining: 'Caindo' })[t] || t || '—'
}
</script>