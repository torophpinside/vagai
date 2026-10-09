<template>
  <Card title="Tendências de Vagas" subtitle="Status, principais organizações e volume de postagens">
    <template #icon><Briefcase class="w-6 h-6 text-amber-400" /></template>
    <EmptyState :data="data" :loading="loading" :error="error">
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
        <div class="h-[260px]">
          <h3 class="text-xs text-slate-400 uppercase tracking-widest mb-4">Vagas por Status</h3>
          <Bar :data="statusChart" :options="barOptions" />
        </div>
        <div>
          <h3 class="text-xs text-slate-400 uppercase tracking-widest mb-4">Principais Organizações</h3>
          <FacetList :items="data?.top_organizations" color="bg-amber-500" />
        </div>
      </div>
    </EmptyState>
  </Card>
</template>

<script setup>
import { computed } from 'vue'
import { Briefcase } from 'lucide-vue-next'
import { useAnalyticsJobTrends } from '../../services/api'
import Card from './Card.vue'
import EmptyState from './EmptyState.vue'
import FacetList from './FacetList.vue'
import { Bar, barOptions, PALETTE } from './charts'

const { data, isLoading: loading, isError: error } = useAnalyticsJobTrends()

const statusChart = computed(() => ({
  labels: Object.keys(data.value?.jobs_by_status || {}),
  datasets: [{
    label: 'Vagas',
    data: Object.values(data.value?.jobs_by_status || {}),
    backgroundColor: PALETTE,
    borderRadius: 4
  }]
}))
</script>