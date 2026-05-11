<template>
  <div class="space-y-12">
    <div>
      <h1 class="text-4xl font-bold text-white mb-2 font-outfit tracking-tight">Dashboard</h1>
      <p class="text-slate-400">Visão geral de suas vagas.</p>
    </div>

    <div v-if="isLoading" class="flex items-center justify-center h-64">
      <div class="w-12 h-12 border-4 border-indigo-500/30 border-t-indigo-500 rounded-full animate-spin"></div>
    </div>

    <div v-else class="space-y-12">
      <!-- Stats Grid -->
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8">
        <div class="glass-card p-8 group hover:bg-slate-800/60 transition-all duration-500 cursor-default">
          <div class="flex items-start justify-between">
            <div>
              <div class="text-slate-400 font-medium mb-1 uppercase tracking-widest text-xs">Total de Vagas</div>
              <div class="text-5xl font-bold text-white font-outfit">{{ stats?.total_jobs || 0 }}</div>
            </div>
            <div class="w-12 h-12 bg-blue-500/10 rounded-2xl flex items-center justify-center text-blue-400 group-hover:scale-110 transition-transform duration-500">
              <Briefcase class="w-6 h-6" />
            </div>
          </div>
          <div class="mt-6 flex items-center gap-2 text-xs text-emerald-400 font-bold uppercase">
            <TrendingUp class="w-4 h-4" />
            <span>+12% este mês</span>
          </div>
        </div>

        <div class="glass-card p-8 group hover:bg-slate-800/60 transition-all duration-500 cursor-default">
          <div class="flex items-start justify-between">
            <div>
              <div class="text-slate-400 font-medium mb-1 uppercase tracking-widest text-xs">Matches Ideais</div>
              <div class="text-5xl font-bold text-white font-outfit">{{ stats?.total_matches || 0 }}</div>
            </div>
            <div class="w-12 h-12 bg-indigo-500/10 rounded-2xl flex items-center justify-center text-indigo-400 group-hover:scale-110 transition-transform duration-500">
              <CheckCircle2 class="w-6 h-6" />
            </div>
          </div>
          <div class="mt-6 flex items-center gap-2 text-xs text-amber-400 font-bold uppercase">
            <Target class="w-4 h-4" />
            <span>Alta Similaridade</span>
          </div>
        </div>

        <div class="glass-card p-8 group hover:bg-slate-800/60 transition-all duration-500 cursor-default">
          <div class="flex items-start justify-between">
            <div>
              <div class="text-slate-400 font-medium mb-1 uppercase tracking-widest text-xs">Candidatadas</div>
              <div class="text-5xl font-bold text-white font-outfit">{{ stats?.total_applied || 0 }}</div>
            </div>
            <div class="w-12 h-12 bg-emerald-500/10 rounded-2xl flex items-center justify-center text-emerald-400 group-hover:scale-110 transition-transform duration-500">
              <Send class="w-6 h-6" />
            </div>
          </div>
          <div class="mt-6 flex items-center gap-2 text-xs text-emerald-400 font-bold uppercase">
            <Rocket class="w-4 h-4" />
            <span>BOA sorte!</span>
          </div>
        </div>

        <div class="glass-card p-8 group hover:bg-slate-800/60 transition-all duration-500 cursor-default">
          <div class="flex items-start justify-between">
            <div>
              <div class="text-slate-400 font-medium mb-1 uppercase tracking-widest text-xs">Fontes Ativas</div>
              <div class="text-5xl font-bold text-white font-outfit">{{ stats?.active_sites || 0 }}</div>
            </div>
            <div class="w-12 h-12 bg-purple-500/10 rounded-2xl flex items-center justify-center text-purple-400 group-hover:scale-110 transition-transform duration-500">
              <Globe class="w-6 h-6" />
            </div>
          </div>
          <div class="mt-6 flex items-center gap-2 text-xs text-slate-500 font-bold uppercase">
            <Activity class="w-4 h-4" />
            <span>Scanner Saudável</span>
          </div>
        </div>
      </div>

      <!-- Charts Section -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-12">
        <div class="glass-card p-10">
          <div class="flex items-center justify-between mb-10">
            <h2 class="text-2xl font-bold text-white font-outfit">Linguagens</h2>
            <Code2 class="text-slate-500 w-6 h-6" />
          </div>
          <div class="h-[400px]">
            <Bar :data="langChartData" :options="chartOptions" />
          </div>
        </div>

        <div class="glass-card p-10">
          <div class="flex items-center justify-between mb-10">
            <h2 class="text-2xl font-bold text-white font-outfit">Tecnologias</h2>
            <Cpu class="text-slate-500 w-6 h-6" />
          </div>
          <div class="h-[400px]">
            <Bar :data="kwChartData" :options="chartOptions" />
          </div>
        </div>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-2 gap-12">
        <div class="glass-card p-10">
          <div class="flex items-center justify-between mb-10">
            <h2 class="text-2xl font-bold text-white font-outfit">Market Share (Linguagens)</h2>
            <PieChart class="text-slate-500 w-6 h-6" />
          </div>
          <div class="h-[300px]">
            <Doughnut :data="langPieData" :options="pieOptions" />
          </div>
        </div>

        <div class="glass-card p-10">
          <div class="flex items-center justify-between mb-10">
            <h2 class="text-2xl font-bold text-white font-outfit">Market Share (Tech)</h2>
            <Layers class="text-slate-500 w-6 h-6" />
          </div>
          <div class="h-[300px]">
            <Doughnut :data="kwPieData" :options="pieOptions" />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useStats } from '../services/api'
import { Bar, Doughnut } from 'vue-chartjs'
import { 
  Briefcase, 
  CheckCircle2, 
  Globe, 
  TrendingUp, 
  Target, 
  Activity,
  Code2,
  Cpu,
  PieChart,
  Layers,
  Send,
  Rocket
} from 'lucide-vue-next'
import { Chart as ChartJS, Title, Tooltip, Legend, BarElement, CategoryScale, LinearScale, ArcElement } from 'chart.js'

ChartJS.register(Title, Tooltip, Legend, BarElement, CategoryScale, LinearScale, ArcElement)

const { data: stats, isLoading } = useStats()

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false }
  },
  scales: {
    y: { beginAtZero: true }
  }
}

const pieOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { position: 'right' }
  }
}

const langColors = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#06b6d4', '#84cc16', '#f97316', '#6366f1']

const langChartData = computed(() => ({
  labels: (stats.value?.languages || []).map(l => l.name),
  datasets: [{
    label: 'Vagas',
    data: (stats.value?.languages || []).map(l => l.count),
    backgroundColor: langColors,
    borderRadius: 4
  }]
}))

const kwChartData = computed(() => ({
  labels: (stats.value?.keywords || []).map(k => k.name),
  datasets: [{
    label: 'Vagas',
    data: (stats.value?.keywords || []).map(k => k.count),
    backgroundColor: langColors,
    borderRadius: 4
  }]
}))

const langPieData = computed(() => {
  const langs = (stats.value?.languages || []).slice(0, 5)
  return {
    labels: langs.map(l => l.name),
    datasets: [{
      data: langs.map(l => l.count),
      backgroundColor: langColors.slice(0, 5)
    }]
  }
})

const kwPieData = computed(() => {
  const kws = (stats.value?.keywords || []).slice(0, 5)
  return {
    labels: kws.map(k => k.name),
    datasets: [{
      data: kws.map(k => k.count),
      backgroundColor: langColors.slice(0, 5)
    }]
  }
})
</script>
