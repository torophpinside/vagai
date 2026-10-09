<template>
  <Card title="Qualidade de Currículos" subtitle="Pontos fortes, fracos e melhorias mais frequentes">
    <template #icon><FileText class="w-6 h-6 text-emerald-400" /></template>
    <EmptyState :data="data" :loading="loading" :error="error">
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
        <FacetList title="Pontos Fortes" :items="data?.top_strengths" color="bg-emerald-500" />
        <FacetList title="Pontos Fracos" :items="data?.top_weaknesses" color="bg-red-500" />
        <FacetList title="Sugestões" :items="data?.top_suggestions" color="bg-amber-500" label-key="suggestion" />
      </div>
    </EmptyState>
  </Card>
</template>

<script setup>
import { FileText } from 'lucide-vue-next'
import { useAnalyticsResume } from '../../services/api'
import Card from './Card.vue'
import EmptyState from './EmptyState.vue'
import FacetList from './FacetList.vue'

const { data, isLoading: loading, isError: error } = useAnalyticsResume()
</script>