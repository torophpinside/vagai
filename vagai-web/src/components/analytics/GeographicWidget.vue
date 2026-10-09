<template>
  <Card title="Insights Geográficos" subtitle="Distribuição de usuários e vagas, e proporção remoto/presencial">
    <template #icon><MapPin class="w-6 h-6 text-teal-400" /></template>
    <EmptyState :data="data" :loading="loading" :error="error">
      <div class="grid grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <Kpi :label="'Cidades'" :value="data?.users_by_city?.length" />
        <Kpi :label="'Fusos'" :value="data?.users_by_timezone?.length" />
        <Kpi :label="'Cidades de Vagas'" :value="data?.jobs_by_city?.length" />
        <Kpi :label="'Proporção'" :value="data?.remote_to_on_site_ratio" />
      </div>
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
        <div class="h-[260px]">
          <h3 class="text-xs text-slate-400 uppercase tracking-widest mb-4">Usuários por Cidade</h3>
          <Bar :data="usersChart" :options="barOptions" />
        </div>
        <div class="h-[260px]">
          <h3 class="text-xs text-slate-400 uppercase tracking-widest mb-4">Vagas por Cidade</h3>
          <Bar :data="jobsChart" :options="barOptions" />
        </div>
      </div>
    </EmptyState>
  </Card>
</template>

<script setup>
import { computed } from 'vue'
import { MapPin } from 'lucide-vue-next'
import { useAnalyticsGeographic } from '../../services/api'
import Card from './Card.vue'
import EmptyState from './EmptyState.vue'
import Kpi from './Kpi.vue'
import { Bar, barOptions, PALETTE } from './charts'

const { data, isLoading: loading, isError: error } = useAnalyticsGeographic()

const usersChart = computed(() => ({
  labels: (data.value?.users_by_city || []).map(c => c.city),
  datasets: [{ label: 'Usuários', data: (data.value?.users_by_city || []).map(c => c.count), backgroundColor: PALETTE, borderRadius: 4 }]
}))

const jobsChart = computed(() => ({
  labels: (data.value?.jobs_by_city || []).map(c => c.city),
  datasets: [{ label: 'Vagas', data: (data.value?.jobs_by_city || []).map(c => c.count), backgroundColor: PALETTE, borderRadius: 4 }]
}))
</script>