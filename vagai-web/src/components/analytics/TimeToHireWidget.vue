<template>
  <Card title="Tempo até a Contratação" subtitle="Duração média e mediana por estágio do pipeline">
    <template #icon><Timer class="w-6 h-6 text-orange-400" /></template>
    <EmptyState :data="data" :loading="loading" :error="error">
      <div class="grid grid-cols-2 lg:grid-cols-6 gap-4 mb-8">
        <Kpi :label="'Match (média)'" :value="days(data?.average_time_to_match_days)" :suffix="'d'" />
        <Kpi :label="'Análise (média)'" :value="days(data?.average_time_to_analysis_days)" :suffix="'d'" />
        <Kpi :label="'Apply (média)'" :value="days(data?.average_time_to_apply_days)" :suffix="'d'" />
        <Kpi :label="'Match (mediana)'" :value="days(data?.median_time_to_match_days)" :suffix="'d'" />
        <Kpi :label="'Análise (mediana)'" :value="days(data?.median_time_to_analysis_days)" :suffix="'d'" />
        <Kpi :label="'Apply (mediana)'" :value="days(data?.median_time_to_apply_days)" :suffix="'d'" />
      </div>
      <div class="h-[280px]">
        <h3 class="text-xs text-slate-400 uppercase tracking-widest mb-4">Tempo Mediano por Estágio</h3>
        <Bar :data="stageChart" :options="barOptions" />
      </div>
    </EmptyState>
  </Card>
</template>

<script setup>
import { computed } from 'vue'
import { Timer } from 'lucide-vue-next'
import { useAnalyticsTimeToHire } from '../../services/api'
import Card from './Card.vue'
import EmptyState from './EmptyState.vue'
import Kpi from './Kpi.vue'
import { Bar, barOptions, PALETTE } from './charts'

const { data, isLoading: loading, isError: error } = useAnalyticsTimeToHire()

const stageChart = computed(() => ({
  labels: (data.value?.stage_timings || []).map(s => s.stage),
  datasets: [{
    label: 'Dias',
    data: (data.value?.stage_timings || []).map(s => s.median_days),
    backgroundColor: PALETTE,
    borderRadius: 4
  }]
}))

function days(v) { return (Number(v) || 0).toFixed(1) }
</script>