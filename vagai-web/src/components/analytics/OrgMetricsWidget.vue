<template>
  <Card title="Métricas da Organização" subtitle="Plano, assinatura e conversão">
    <template #icon><Building2 class="w-6 h-6 text-purple-400" /></template>
    <EmptyState :data="data" :loading="loading" :error="error">
      <div class="grid grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <Kpi :label="'Vagas'" :value="metrics?.jobs_posted" />
        <Kpi :label="'Currículos'" :value="metrics?.resumes_uploaded" />
        <Kpi :label="'Matches'" :value="metrics?.matches_made" />
        <Kpi :label="'Conversão'" :value="pct(metrics?.conversion_rate)" :suffix="'%'" />
      </div>
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4 text-center">
        <div v-for="(active, key) in metrics?.feature_utilization || {}" :key="key" class="bg-slate-900/60 border border-white/5 rounded-xl p-4">
          <div class="text-xs text-slate-500 uppercase tracking-widest mb-2">{{ featureLabel(key) }}</div>
          <div :class="active ? 'text-emerald-400' : 'text-slate-600'" class="font-bold">{{ active ? 'Ativo' : 'Inativo' }}</div>
        </div>
      </div>
    </EmptyState>
  </Card>
</template>

<script setup>
import { computed } from 'vue'
import { Building2 } from 'lucide-vue-next'
import { useAnalyticsOrgMetrics, useMe } from '../../services/api'
import Card from './Card.vue'
import EmptyState from './EmptyState.vue'
import Kpi from './Kpi.vue'

const { data: me, isLoading: meLoading } = useMe()
const orgId = computed(() => me.value?.organization?.id)
const { data, isLoading: loading, isError: error } = useAnalyticsOrgMetrics(orgId)

const metrics = computed(() => data.value?.metrics)

function pct(v) { return Math.round((Number(v) || 0) * 100) }

const FEATURES = { matching: 'Matching', ai_analysis: 'Análise de IA', interview_prep: 'Entrevista', skill_gap: 'Skill Gap' }
function featureLabel(key) { return FEATURES[key] || key }
</script>