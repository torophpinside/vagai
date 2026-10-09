package models

// Tipos de resposta dos endpoints de analytics (data-model.md).
// São DTOs de leitura: não são migrados pelo AutoMigrate.

// ScoreBucket representa um intervalo de scores no histograma de matching.
type ScoreBucket struct {
	Bucket     string  `json:"bucket"`
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"`
}

// JobTitleSuccess mede a taxa de conversão (match → candidatura) por título.
type JobTitleSuccess struct {
	Title        string  `json:"title"`
	TotalMatches int64   `json:"total_matches"`
	Applied      int64   `json:"applied"`
	SuccessRate  float64 `json:"success_rate"`
}

// MatchingDistribution é o resultado do US1: distribuição de scores,
// média/mediana, taxa de conversão e títulos com maior sucesso.
type MatchingDistribution struct {
	Buckets        []ScoreBucket     `json:"buckets"`
	AverageScore   float64           `json:"average_score"`
	MedianScore    float64           `json:"median_score"`
	TotalMatches   int64             `json:"total_matches"`
	ConversionRate float64           `json:"conversion_rate"`
	TopJobTitles   []JobTitleSuccess `json:"top_job_titles"`
}

// SkillCount conta ocorrências de uma habilidade/força/fragilidade.
type SkillCount struct {
	Skill string `json:"skill"`
	Count int64  `json:"count"`
}

// SuggestionCount conta ocorrências de uma sugestão de melhoria.
type SuggestionCount struct {
	Suggestion string `json:"suggestion"`
	Count      int64  `json:"count"`
}

// ResumeQualityInsights é o resultado do US2.
type ResumeQualityInsights struct {
	TopStrengths   []SkillCount      `json:"top_strengths"`
	TopWeaknesses  []SkillCount      `json:"top_weaknesses"`
	TopSuggestions []SuggestionCount `json:"top_suggestions"`
	SkillFrequency map[string]int    `json:"skill_frequency"`
	TotalAnalyses  int64             `json:"total_analyses"`
}

// OrganizationSummary resume uma organização por volume de vagas.
type OrganizationSummary struct {
	OrganizationID uint   `json:"organization_id"`
	Name           string `json:"name"`
	JobCount       int64  `json:"job_count"`
}

// JobMarketTrends é o resultado do US3.
type JobMarketTrends struct {
	JobsByStatus        map[JobStatus]int64    `json:"jobs_by_status"`
	TopOrganizations    []OrganizationSummary  `json:"top_organizations"`
	PostingVolume       map[string]int64       `json:"posting_volume"`
	ActiveCompanies     int64                  `json:"active_companies"`
	TotalJobs           int64                  `json:"total_jobs"`
	PostingVolumeTotal  int64                  `json:"-"`
}

// FeatureUtilization indica quais features a organização efetivamente usa.
type FeatureUtilization struct {
	Matching      bool `json:"matching"`
	AIAnalysis    bool `json:"ai_analysis"`
	InterviewPrep bool `json:"interview_prep"`
	SkillGap      bool `json:"skill_gap"`
}

// OrgMetrics é o resultado do US4.
type OrgMetrics struct {
	OrganizationID      uint                 `json:"organization_id"`
	OrganizationName    string               `json:"organization_name"`
	PlanType            string               `json:"plan"`
	SubscriptionStatus  SubscriptionStatus   `json:"subscription_status"`
	TrialToPaid         bool                 `json:"trial_to_paid"`
	JobsPosted          int64                `json:"jobs_posted"`
	ResumesUploaded     int64                `json:"resumes_uploaded"`
	MatchesMade         int64                `json:"matches_made"`
	ApplicationsMade    int64                `json:"applications_made"`
	ConversionRate      float64              `json:"conversion_rate"`
	FeatureUtilization  FeatureUtilization   `json:"feature_utilization"`
	AverageTimeToMatchDays   float64         `json:"average_time_to_match_days"`
	AverageTimeToAnalysisDays float64        `json:"average_time_to_analysis_days"`
}

// InterviewPrepAnalysis é o resultado do US5.
type InterviewPrepAnalysis struct {
	TotalPreparations        int64                                `json:"total_preparations"`
	CategoryDistribution     map[InterviewCategory]int64          `json:"category_distribution"`
	SourceDistribution       map[InterviewPreparationSource]int64 `json:"source_distribution"`
	AverageVerificationScore float64                              `json:"average_verification_score"`
	VerificationTrend        string                               `json:"verification_trend"`
}

// ScoreBucketCount alias documental para buckets reutilizados fora do matching.

// SuccessRateBySource mede concluído/erro/pendente por fonte (ai/template).
type SuccessRateBySource struct {
	Source      VerificationSource `json:"source"`
	Completed   int64              `json:"completed"`
	Error       int64              `json:"error"`
	Pending     int64              `json:"pending"`
	SuccessRate float64            `json:"success_rate"`
}

// AIQualityInsights é o resultado do US6.
type AIQualityInsights struct {
	TotalAnalyses            int64                      `json:"total_analyses"`
	ScoreBuckets             []ScoreBucket              `json:"score_buckets"`
	AverageScore             float64                    `json:"average_score"`
	StatusCounts             map[AIAnalysisStatus]int64 `json:"status_counts"`
	SuccessRateBySource      []SuccessRateBySource      `json:"success_rate_by_source"`
	AverageProcessingSeconds float64                    `json:"average_processing_seconds"`
}

// StageTiming mede a duração mediana (em dias) entre dois estágios.
type StageTiming struct {
	Stage      string  `json:"stage"`
	MedianDays float64 `json:"median_days"`
	Samples    int64   `json:"samples"`
}

// TimeToHireMetrics é o resultado do US7.
type TimeToHireMetrics struct {
	AverageTimeToMatchDays    float64       `json:"average_time_to_match_days"`
	AverageTimeToAnalysisDays float64       `json:"average_time_to_analysis_days"`
	AverageTimeToApplyDays    float64       `json:"average_time_to_apply_days"`
	MedianTimeToMatchDays     float64       `json:"median_time_to_match_days"`
	MedianTimeToAnalysisDays  float64       `json:"median_time_to_analysis_days"`
	MedianTimeToApplyDays     float64       `json:"median_time_to_apply_days"`
	StageTimings              []StageTiming `json:"stage_timings"`
}

// FunnelStageMetric é um estágio do funil com contagem e conversão acumulada.
type FunnelStageMetric struct {
	Stage          string  `json:"stage"`
	Count          int64   `json:"count"`
	ConversionRate float64 `json:"conversion_rate"`
}

// FunnelMetrics é o resultado do US9.
type FunnelMetrics struct {
	FunnelStages       []FunnelStageMetric `json:"funnel_stages"`
	MedianTimePerStage []StageTiming       `json:"median_time_per_stage"`
	DropOffPoints      []string            `json:"drop_off_points"`
	TotalUsers         int64               `json:"total_users"`
}

// SkillGapAnalysis é o resultado do US8.
type SkillGapAnalysis struct {
	JobSkills           []string           `json:"job_skills"`
	CandidateSkills     []string           `json:"candidate_skills"`
	OverlappingSkills   []string           `json:"overlapping_skills"`
	InDemandSkills      []string           `json:"in_demand_skills"`
	MatchRateByCategory map[string]float64 `json:"match_rate_by_category"`
}

// CityCount conta registros por cidade.
type CityCount struct {
	City  string `json:"city"`
	Count int64  `json:"count"`
}

// StateCount conta registros por estado/UF.
type StateCount struct {
	State string `json:"state"`
	Count int64  `json:"count"`
}

// TimezoneCount conta registros por fuso horário.
type TimezoneCount struct {
	Timezone string `json:"timezone"`
	Count    int64  `json:"count"`
}

// GeographicInsights é o resultado do US10.
type GeographicInsights struct {
	UsersByCity         []CityCount     `json:"users_by_city"`
	UsersByTimezone     []TimezoneCount `json:"users_by_timezone"`
	JobsByCity          []CityCount     `json:"jobs_by_city"`
	JobsByState         []StateCount    `json:"jobs_by_state"`
	RemoteFriendlyRatio float64         `json:"remote_friendly_ratio"`
	RemoteToOnSiteRatio string          `json:"remote_to_on_site_ratio"`
}
