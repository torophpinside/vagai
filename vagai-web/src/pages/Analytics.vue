<template>
  <div class="space-y-12">
    <div class="flex items-end justify-between">
      <div>
        <h1 class="text-4xl font-bold text-white mb-2 font-outfit tracking-tight">Analytics</h1>
        <p class="text-slate-400">Indicadores de matching, currículos, vagas e organização.</p>
      </div>
    </div>

    <MatchingWidget />
    <ResumeWidget />
    <JobTrendsWidget />
    <OrgMetricsWidget />
    <InterviewPrepWidget />
    <AIQualityWidget />
    <TimeToHireWidget />
    <SkillGapWidget />
    <FunnelWidget />
    <GeographicWidget />
  </div>
</template>

<script setup>
import MatchingWidget from '../components/analytics/MatchingWidget.vue'
import ResumeWidget from '../components/analytics/ResumeWidget.vue'
import JobTrendsWidget from '../components/analytics/JobTrendsWidget.vue'
import OrgMetricsWidget from '../components/analytics/OrgMetricsWidget.vue'
import InterviewPrepWidget from '../components/analytics/InterviewPrepWidget.vue'
import AIQualityWidget from '../components/analytics/AIQualityWidget.vue'
import TimeToHireWidget from '../components/analytics/TimeToHireWidget.vue'
import SkillGapWidget from '../components/analytics/SkillGapWidget.vue'
import FunnelWidget from '../components/analytics/FunnelWidget.vue'
import GeographicWidget from '../components/analytics/GeographicWidget.vue'
</script>