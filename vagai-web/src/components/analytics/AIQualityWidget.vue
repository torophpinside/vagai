<template>
  <Card title="Qualidade de Análises de IA" subtitle="Distribuição de score, status e sucesso por fonte">
    <template #icon><Sparkles class="w-6 h-6 text-cyan-400" /></template>
    <EmptyState :data="data" :loading="loading" :error="error">
      <div class="grid grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <Kpi :label="'Análises'" :value="data?.total_analyses" />
        <Kpi :label="'Score Médio'" :value="data?.average_score?.toFixed?.(1)" />
        <Kpi :label="'Tempo Médio'" :value="data?.average_processing_seconds" :suffix="'s'" />
        <Kpi :label="'Tendência'" :value="trendLabel(data?.verification_trend)" />
      </div>
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
        <div class="h-[260px]">
          <h3 class="text-xs text-slate-400 uppercase tracking-widest mb-4">Score por Bucket</h3>
          <Bar :data="bucketChart" :options="barOptions" />
        </div>
        <div>
          <h3 class="text-xs text-slate-400 uppercase tracking-widest mb-4">Sucesso por Fonte</h3>
          <FacetList :items="sourceList" color="bg-cyan-500" label-key="source" />
        </div>
      </div>
    </EmptyState>
  </Card>
</template>

<script setup>
import { computed } from 'vue'
import { Sparkles } from 'lucide-vue-next'
import { useAnalyticsAIQuality } from '../../services/api'
import Card from './Card.vue'
import EmptyState from './EmptyState.vue'
import Kpi from './Kpi.vue'
import FacetList from './FacetList.vue'
import { Bar, barOptions, PALETTE } from './charts'

const { data, isLoading: loading, isError: error } = useAnalyticsAIQuality()

const bucketChart = computed(() => ({
  labels: (data.value?.score_buckets || []).map(b => b.bucket),
  datasets: [{
    label: 'Análises',
    data: (data.value?.score_buckets || []).map(b => b.count),
    backgroundColor: PALETTE,
    borderRadius: 4
  }]
}))

const sourceList = computed(() =>
  (data.value?.success_rate_by_source || []).map(s => ({ source: s.source, count: s.completed, name: `${s.success_rate?.toFixed?.(1)}%` }))
)

function trendLabel(t) {
  return ({ improving: 'Melhorando', stable: 'Estável', declining: 'Caindo' })[t] || t || '—'
}
</script>