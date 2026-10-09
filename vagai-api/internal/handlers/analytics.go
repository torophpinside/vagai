package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/anomalyco/vagai-api/internal/middleware"
	"github.com/anomalyco/vagai-api/internal/models"
	"github.com/anomalyco/vagai-api/internal/services/analytics"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterAnalyticsRoutes registra o grupo /api/v1/analytics/* (contracts/
// api_contract.md) com autenticação JWT + controle de acesso FR-020
// (org admins e super admins). As rotas individuais são registradas por
// user story (US1–US10).
func RegisterAnalyticsRoutes(rg *gin.RouterGroup) {
	rg.Use(middleware.RequireAnalyticsAccess())

	// US1 — matching effectiveness
	rg.GET("/matching-distribution", GetMatchingDistribution)

	// US2 — resume quality
	rg.GET("/resume-insights", GetResumeInsights)

	// US3 — job market trends
	rg.GET("/job-trends", GetJobTrends)

	// US4 — organization metrics
	rg.GET("/org/:organization_id", GetOrgMetrics)

	// US5 — interview preparation analysis
	rg.GET("/interview-prep", GetInterviewPrepAnalytics)

	// US6 — AI analysis quality
	rg.GET("/ai-quality", GetAIQuality)

	// US7 — time-to-hire
	rg.GET("/time-to-hire", GetTimeToHire)

	// US8 — skill gap
	rg.GET("/skills-gap", GetSkillsGap)

	// US9 — funnel
	rg.GET("/funnel", GetFunnel)

	// US10 — geographic
	rg.GET("/geographic", GetGeographic)
}

// analyticsService cria o serviço com o DB já escopado pelo request.
func analyticsService(c *gin.Context) *analytics.Service {
	return analytics.New(getDB(c))
}

// parsePage extrai page/page_size da query string (422 em valor inválido,
// conforme contracts/api_contract.md).
func parsePage(c *gin.Context) (analytics.PageOptions, bool) {
	opts, err := analytics.ParsePageOptions(c.Query("page"), c.Query("page_size"))
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": "Parâmetros de paginação inválidos (page ≥ 1, page_size entre 1 e 100)",
		})
		return opts, false
	}
	return opts, true
}

// queryOrgID lê o parâmetro opcional organization_id da query string.
// Retorna 0 quando ausente (o middleware resolve para a organização do
// próprio usuário; super admin com 0 consulta todas). OK=false indica
// parâmetro inválido — resposta 422 já emitida.
func queryOrgID(c *gin.Context) (uint, bool) {
	raw := c.Query("organization_id")
	if raw == "" {
		return 0, true
	}
	id, err := strconv.ParseUint(raw, 10, 32)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": "organization_id inválido",
		})
		return 0, false
	}
	return uint(id), true
}

// abortAnalyticsError padroniza erros de serviço → 500 ( contrato: "Analytics
// service unavailable").
func abortAnalyticsError(c *gin.Context, err error) {
	if errors.Is(err, analytics.ErrInvalidPage) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": "Parâmetros de paginação inválidos",
		})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": "Analytics service unavailable",
	})
}

// ---------------------------------------------------------------------------
// Handlers — um por user story (US1–US10).
// ---------------------------------------------------------------------------

// US1 — GET /api/v1/analytics/matching-distribution (FR-010).
func GetMatchingDistribution(c *gin.Context) {
	opts, ok := parsePage(c)
	if !ok {
		return
	}

	requested, ok := queryOrgID(c)
	if !ok {
		return
	}
	orgID, ok := middleware.ResolveAnalyticsOrgID(c, requested)
	if !ok {
		return
	}

	svc := analyticsService(c)
	result, err := svc.MatchingDistribution(c.Request.Context(), orgID, opts)
	if err != nil {
		abortAnalyticsError(c, err)
		return
	}

	if result == nil {
		c.JSON(http.StatusOK, analytics.EmptyEnvelope(map[string]any{
			"buckets": []gin.H{},
			"summary": gin.H{
				"average_score":   0,
				"median_score":    0,
				"total_matches":   0,
				"conversion_rate": 0,
			},
			"top_job_titles": []gin.H{},
		}, analytics.EmptyMatchingMessage, analytics.EmptyMatchingGuidance))
		return
	}

	page := analytics.NewPagination(opts, result.TotalMatches)
	c.JSON(http.StatusOK, analytics.OKEnvelope(map[string]any{
		"buckets": result.Buckets,
		"summary": gin.H{
			"average_score":   result.AverageScore,
			"median_score":    result.MedianScore,
			"total_matches":   result.TotalMatches,
			"conversion_rate": result.ConversionRate,
		},
		"top_job_titles": result.TopJobTitles,
	}, &page))
}

// US2 — GET /api/v1/analytics/resume-insights (FR-011).
func GetResumeInsights(c *gin.Context) {
	requested, ok := queryOrgID(c)
	if !ok {
		return
	}
	orgID, ok := middleware.ResolveAnalyticsOrgID(c, requested)
	if !ok {
		return
	}

	svc := analyticsService(c)
	result, err := svc.ResumeInsights(c.Request.Context(), orgID)
	if err != nil {
		abortAnalyticsError(c, err)
		return
	}

	if result == nil {
		c.JSON(http.StatusOK, analytics.EmptyEnvelope(map[string]any{
			"top_strengths":   []gin.H{},
			"top_weaknesses":  []gin.H{},
			"top_suggestions": []gin.H{},
			"skill_frequency": map[string]int{},
		}, analytics.EmptyResumeMessage, analytics.EmptyResumeGuidance))
		return
	}

	c.JSON(http.StatusOK, analytics.OKEnvelope(map[string]any{
		"top_strengths":   result.TopStrengths,
		"top_weaknesses":  result.TopWeaknesses,
		"top_suggestions": result.TopSuggestions,
		"skill_frequency": result.SkillFrequency,
		"total_analyses":  result.TotalAnalyses,
	}, nil))
}

// US3 — GET /api/v1/analytics/job-trends (FR-012).
func GetJobTrends(c *gin.Context) {
	opts, ok := parsePage(c)
	if !ok {
		return
	}

	period := c.Query("period")
	if period != "" && !analytics.ValidPeriods[period] {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": "period inválido: use week, month, quarter ou year",
		})
		return
	}

	requested, ok := queryOrgID(c)
	if !ok {
		return
	}
	orgID, ok := middleware.ResolveAnalyticsOrgID(c, requested)
	if !ok {
		return
	}

	svc := analyticsService(c)
	result, err := svc.JobTrends(c.Request.Context(), orgID, period, opts)
	if err != nil {
		abortAnalyticsError(c, err)
		return
	}

	if result == nil {
		c.JSON(http.StatusOK, analytics.EmptyEnvelope(map[string]any{
			"jobs_by_status":    map[string]int64{},
			"top_organizations": []gin.H{},
			"posting_volume":    map[string]int64{},
			"active_companies":  0,
			"total_jobs":        0,
		}, analytics.EmptyJobTrendsMessage, analytics.EmptyJobTrendsGuidance))
		return
	}

	page := analytics.NewPagination(opts, result.PostingVolumeTotal)
	c.JSON(http.StatusOK, analytics.OKEnvelope(map[string]any{
		"jobs_by_status":    result.JobsByStatus,
		"top_organizations": result.TopOrganizations,
		"posting_volume":    result.PostingVolume,
		"active_companies":  result.ActiveCompanies,
		"total_jobs":        result.TotalJobs,
	}, &page))
}

// US4 — GET /api/v1/analytics/org/:organization_id (FR-013).
// Contrato: 404 organização inexistente, 403 acesso negado, 500 serviço.
func GetOrgMetrics(c *gin.Context) {
	requested, err := strconv.ParseUint(c.Param("organization_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Organização não encontrada"})
		return
	}
	orgID, ok := middleware.ResolveAnalyticsOrgID(c, uint(requested))
	if !ok {
		return
	}
	if orgID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Organização não encontrada"})
		return
	}

	var org models.Organization
	if err := getDB(c).First(&org, orgID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Organização não encontrada"})
			return
		}
		abortAnalyticsError(c, err)
		return
	}

	svc := analyticsService(c)
	result, err := svc.OrgMetrics(c.Request.Context(), orgID)
	if err != nil {
		abortAnalyticsError(c, err)
		return
	}

	if result == nil {
		c.JSON(http.StatusOK, analytics.EmptyEnvelope(map[string]any{
			"organization": gin.H{
				"id":                  orgID,
				"name":                org.Name,
				"plan":                org.Plan,
				"subscription_status": "",
				"trial_to_paid":       false,
			},
			"metrics": gin.H{
				"jobs_posted":         0,
				"resumes_uploaded":    0,
				"matches_made":        0,
				"applications_made":   0,
				"conversion_rate":     0,
				"feature_utilization": models.FeatureUtilization{},
			},
			"timing": gin.H{
				"average_time_to_match_days":    0,
				"average_time_to_analysis_days": 0,
			},
		}, analytics.EmptyOrgMetricsMessage, analytics.EmptyOrgMetricsGuidance))
		return
	}

	c.JSON(http.StatusOK, analytics.OKEnvelope(map[string]any{
		"organization": gin.H{
			"id":                  result.OrganizationID,
			"name":                result.OrganizationName,
			"plan":                result.PlanType,
			"subscription_status": result.SubscriptionStatus,
			"trial_to_paid":       result.TrialToPaid,
		},
		"metrics": gin.H{
			"jobs_posted":         result.JobsPosted,
			"resumes_uploaded":    result.ResumesUploaded,
			"matches_made":        result.MatchesMade,
			"applications_made":   result.ApplicationsMade,
			"conversion_rate":     result.ConversionRate,
			"feature_utilization": result.FeatureUtilization,
		},
		"timing": gin.H{
			"average_time_to_match_days":    result.AverageTimeToMatchDays,
			"average_time_to_analysis_days": result.AverageTimeToAnalysisDays,
		},
	}, nil))
}

// US5 — GET /api/v1/analytics/interview-prep (FR-014).
func GetInterviewPrepAnalytics(c *gin.Context) {
	requested, ok := queryOrgID(c)
	if !ok {
		return
	}
	orgID, ok := middleware.ResolveAnalyticsOrgID(c, requested)
	if !ok {
		return
	}

	svc := analyticsService(c)
	result, err := svc.InterviewPrepAnalysis(c.Request.Context(), orgID)
	if err != nil {
		abortAnalyticsError(c, err)
		return
	}

	if result == nil {
		c.JSON(http.StatusOK, analytics.EmptyEnvelope(map[string]any{
			"total_preparations":         int64(0),
			"category_distribution":      map[string]int64{},
			"source_distribution":        map[string]int64{},
			"average_verification_score": 0,
			"verification_trend":         "stable",
		}, analytics.EmptyInterviewPrepMessage, analytics.EmptyInterviewPrepGuidance))
		return
	}

	c.JSON(http.StatusOK, analytics.OKEnvelope(map[string]any{
		"total_preparations":         result.TotalPreparations,
		"category_distribution":      result.CategoryDistribution,
		"source_distribution":        result.SourceDistribution,
		"average_verification_score": result.AverageVerificationScore,
		"verification_trend":         result.VerificationTrend,
	}, nil))
}

// US6 — GET /api/v1/analytics/ai-quality (FR-015).
func GetAIQuality(c *gin.Context) {
	requested, ok := queryOrgID(c)
	if !ok {
		return
	}
	orgID, ok := middleware.ResolveAnalyticsOrgID(c, requested)
	if !ok {
		return
	}

	svc := analyticsService(c)
	result, err := svc.AIQuality(c.Request.Context(), orgID)
	if err != nil {
		abortAnalyticsError(c, err)
		return
	}

	if result == nil {
		c.JSON(http.StatusOK, analytics.EmptyEnvelope(map[string]any{
			"total_analyses":             int64(0),
			"score_buckets":              []gin.H{},
			"average_score":              0,
			"status_counts":              map[string]int64{},
			"success_rate_by_source":     []gin.H{},
			"average_processing_seconds": 0,
		}, analytics.EmptyAIQualityMessage, analytics.EmptyAIQualityGuidance))
		return
	}

	c.JSON(http.StatusOK, analytics.OKEnvelope(map[string]any{
		"total_analyses":             result.TotalAnalyses,
		"score_buckets":              result.ScoreBuckets,
		"average_score":              result.AverageScore,
		"status_counts":              result.StatusCounts,
		"success_rate_by_source":     result.SuccessRateBySource,
		"average_processing_seconds": result.AverageProcessingSeconds,
	}, nil))
}

// US7 — GET /api/v1/analytics/time-to-hire (FR-016).
func GetTimeToHire(c *gin.Context) {
	requested, ok := queryOrgID(c)
	if !ok {
		return
	}
	orgID, ok := middleware.ResolveAnalyticsOrgID(c, requested)
	if !ok {
		return
	}

	svc := analyticsService(c)
	result, err := svc.TimeToHire(c.Request.Context(), orgID)
	if err != nil {
		abortAnalyticsError(c, err)
		return
	}

	if result == nil {
		c.JSON(http.StatusOK, analytics.EmptyEnvelope(map[string]any{
			"average_time_to_match_days":    0,
			"average_time_to_analysis_days": 0,
			"average_time_to_apply_days":    0,
			"median_time_to_match_days":     0,
			"median_time_to_analysis_days":  0,
			"median_time_to_apply_days":     0,
			"stage_timings":                 []gin.H{},
		}, analytics.EmptyTimeToHireMessage, analytics.EmptyTimeToHireGuidance))
		return
	}

	c.JSON(http.StatusOK, analytics.OKEnvelope(map[string]any{
		"average_time_to_match_days":    result.AverageTimeToMatchDays,
		"average_time_to_analysis_days": result.AverageTimeToAnalysisDays,
		"average_time_to_apply_days":    result.AverageTimeToApplyDays,
		"median_time_to_match_days":     result.MedianTimeToMatchDays,
		"median_time_to_analysis_days":  result.MedianTimeToAnalysisDays,
		"median_time_to_apply_days":     result.MedianTimeToApplyDays,
		"stage_timings":                 result.StageTimings,
	}, nil))
}

// US8 — GET /api/v1/analytics/skills-gap (FR-017).
func GetSkillsGap(c *gin.Context) {
	requested, ok := queryOrgID(c)
	if !ok {
		return
	}
	orgID, ok := middleware.ResolveAnalyticsOrgID(c, requested)
	if !ok {
		return
	}

	minFreq := 1
	if raw := c.Query("min_frequency"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"error": "min_frequency inválido (inteiro ≥ 1)",
			})
			return
		}
		minFreq = v
	}

	svc := analyticsService(c)
	result, err := svc.SkillGap(c.Request.Context(), orgID, minFreq)
	if err != nil {
		abortAnalyticsError(c, err)
		return
	}

	if result == nil {
		c.JSON(http.StatusOK, analytics.EmptyEnvelope(map[string]any{
			"job_skills":             []string{},
			"candidate_skills":       []string{},
			"overlapping_skills":     []string{},
			"in_demand_skills":       []string{},
			"match_rate_by_category": map[string]float64{},
		}, analytics.EmptySkillGapMessage, analytics.EmptySkillGapGuidance))
		return
	}

	c.JSON(http.StatusOK, analytics.OKEnvelope(map[string]any{
		"job_skills":             result.JobSkills,
		"candidate_skills":       result.CandidateSkills,
		"overlapping_skills":     result.OverlappingSkills,
		"in_demand_skills":       result.InDemandSkills,
		"match_rate_by_category": result.MatchRateByCategory,
	}, nil))
}

// US9 — GET /api/v1/analytics/funnel (FR-018).
func GetFunnel(c *gin.Context) {
	requested, ok := queryOrgID(c)
	if !ok {
		return
	}
	orgID, ok := middleware.ResolveAnalyticsOrgID(c, requested)
	if !ok {
		return
	}

	svc := analyticsService(c)
	result, err := svc.Funnel(c.Request.Context(), orgID)
	if err != nil {
		abortAnalyticsError(c, err)
		return
	}

	if result == nil {
		c.JSON(http.StatusOK, analytics.EmptyEnvelope(map[string]any{
			"funnel_stages":         []gin.H{},
			"median_time_per_stage": []gin.H{},
			"drop_off_points":       []string{},
			"total_users":           int64(0),
		}, analytics.EmptyFunnelMessage, analytics.EmptyFunnelGuidance))
		return
	}

	c.JSON(http.StatusOK, analytics.OKEnvelope(map[string]any{
		"funnel_stages":         result.FunnelStages,
		"median_time_per_stage": result.MedianTimePerStage,
		"drop_off_points":       result.DropOffPoints,
		"total_users":           result.TotalUsers,
	}, nil))
}

// US10 — GET /api/v1/analytics/geographic (FR-019).
func GetGeographic(c *gin.Context) {
	requested, ok := queryOrgID(c)
	if !ok {
		return
	}
	orgID, ok := middleware.ResolveAnalyticsOrgID(c, requested)
	if !ok {
		return
	}

	svc := analyticsService(c)
	result, err := svc.Geographic(c.Request.Context(), orgID)
	if err != nil {
		abortAnalyticsError(c, err)
		return
	}

	if result == nil {
		c.JSON(http.StatusOK, analytics.EmptyEnvelope(map[string]any{
			"users_by_city":           []gin.H{},
			"users_by_timezone":       []gin.H{},
			"jobs_by_city":            []gin.H{},
			"jobs_by_state":           []gin.H{},
			"remote_friendly_ratio":   0,
			"remote_to_on_site_ratio": "0:0",
		}, analytics.EmptyGeographicMessage, analytics.EmptyGeographicGuidance))
		return
	}

	c.JSON(http.StatusOK, analytics.OKEnvelope(map[string]any{
		"users_by_city":           result.UsersByCity,
		"users_by_timezone":       result.UsersByTimezone,
		"jobs_by_city":            result.JobsByCity,
		"jobs_by_state":           result.JobsByState,
		"remote_friendly_ratio":   result.RemoteFriendlyRatio,
		"remote_to_on_site_ratio": result.RemoteToOnSiteRatio,
	}, nil))
}
