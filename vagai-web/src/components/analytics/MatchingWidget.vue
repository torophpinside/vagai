<template>
  <Card title="Efetividade do Matching" subtitle="Distribuição de similaridade, média, mediana e conversão">
    <template #icon><Target class="w-6 h-6 text-indigo-400" /></template>
    <EmptyState :data="data" :loading="loading" :error="error">
      <div class="grid grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <Kpi :label="'Total de Matches'" :value="summary?.total_matches" />
        <Kpi :label="'Score Médio'" :value="pct(summary?.average_score)" :suffix="'%'" />
        <Kpi :label="'Score Mediano'" :value="pct(summary?.median_score)" :suffix="'%'" />
        <Kpi :label="'Taxa de Conversão'" :value="pct(summary?.conversion_rate)" :suffix="'%'" />
      </div>
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
        <div class="h-[260px]">
          <h3 class="text-xs text-slate-400 uppercase tracking-widest mb-4">Distribuição de Scores</h3>
          <Bar :data="bucketChart" :options="barOptions" />
        </div>
        <div>
          <h3 class="text-xs text-slate-400 uppercase tracking-widest mb-4">Melhores Títulos</h3>
          <div v-if="titles.length" class="space-y-3">
            <div v-for="t in titles" :key="t.title" class="flex items-center gap-4">
              <span class="flex-1 text-sm text-slate-300 truncate">{{ t.title }}</span>
              <div class="w-32 bg-slate-700 rounded-full h-1.5">
                <div class="bg-indigo-500 h-1.5 rounded-full" :style="{ width: t.success_rate + '%' }"></div>
              </div>
              <span class="text-xs text-slate-400 w-10 text-right">{{ t.success_rate.toFixed(1) }}%</span>
            </div>
          </div>
          <div v-else class="text-slate-500 text-sm py-8 text-center">Sem candidaturas registradas ainda.</div>
        </div>
      </div>
    </EmptyState>
  </Card>
</template>

<script setup>
import { computed } from 'vue'
import { Target } from 'lucide-vue-next'
import { useAnalyticsMatching } from '../../services/api'
import Card from './Card.vue'
import EmptyState from './EmptyState.vue'
import Kpi from './Kpi.vue'
import { Bar, barOptions, PALETTE } from './charts'

const { data, isLoading: loading, isError: error } = useAnalyticsMatching()

const summary = computed(() => data.value?.summary || {})
const titles = computed(() => data.value?.top_job_titles || [])

const bucketChart = computed(() => ({
  labels: (data.value?.buckets || []).map(b => b.bucket),
  datasets: [{
    label: 'Matches',
    data: (data.value?.buckets || []).map(b => b.count),
    backgroundColor: PALETTE,
    borderRadius: 4
  }]
}))

function pct(v) { return Math.round((Number(v) || 0) * 100) }
</script>