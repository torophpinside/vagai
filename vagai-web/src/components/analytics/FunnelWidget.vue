<template>
  <Card title="Funil de Conversão" subtitle="Estágios do usuário e pontos de abandono">
    <template #icon><Filter class="w-6 h-6 text-rose-400" /></template>
    <EmptyState :data="data" :loading="loading" :error="error">
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
        <div class="h-[280px]">
          <h3 class="text-xs text-slate-400 uppercase tracking-widest mb-4">Estágios do Funil</h3>
          <Bar :data="funnelChart" :options="barOptions" />
        </div>
        <div class="space-y-6">
          <div class="bg-slate-900/60 border border-white/5 rounded-xl p-5">
            <div class="text-xs text-slate-500 uppercase tracking-widest mb-2">Total de Usuários</div>
            <div class="text-3xl font-bold text-white font-outfit">{{ data?.total_users }}</div>
          </div>
          <div>
            <h3 class="text-xs text-slate-400 uppercase tracking-widest mb-3">Pontos de Abandono</h3>
            <ul class="space-y-3">
              <li v-for="stage in data?.drop_off_points || []" :key="stage" class="flex items-center gap-3">
                <span class="w-1.5 h-1.5 rounded-full bg-rose-500"></span>
                <span class="text-sm text-slate-300">{{ stage }}</span>
              </li>
            </ul>
            <div v-if="!data?.drop_off_points?.length" class="text-slate-500 text-sm py-4 text-center">Sem pontos de abandono.</div>
          </div>
        </div>
      </div>
    </EmptyState>
  </Card>
</template>

<script setup>
import { computed } from 'vue'
import { Filter } from 'lucide-vue-next'
import { useAnalyticsFunnel } from '../../services/api'
import Card from './Card.vue'
import EmptyState from './EmptyState.vue'
import { Bar, barOptions, PALETTE } from './charts'

const { data, isLoading: loading, isError: error } = useAnalyticsFunnel()

const funnelChart = computed(() => ({
  labels: (data.value?.funnel_stages || []).map(s => s.stage),
  datasets: [{
    label: 'Usuários',
    data: (data.value?.funnel_stages || []).map(s => s.count),
    backgroundColor: PALETTE,
    borderRadius: 4
  }]
}))
</script>