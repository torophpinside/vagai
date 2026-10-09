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
// api_contract.md) com autenticação JWT. O escopo por papel é resolvido em
// middleware.ResolveAnalyticsUserID: admin consulta dados globais (0) ou de
// qualquer usuário; usuário comum consulta apenas os próprios. As rotas
// individuais são registradas por user story (US1–US10).
func RegisterAnalyticsRoutes(rg *gin.RouterGroup) {
	// US1 — matching effectiveness
	rg.GET("/matching-distribution", GetMatchingDistribution)

	// US2 — resume quality
	rg.GET("/resume-insights", GetResumeInsights)

	// US3 — job market trends
	rg.GET("/job-trends", GetJobTrends)

	// US4 — organization metrics
	rg.GET("/user/:user_id", GetUserMetrics)

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

// queryUserID lê o parâmetro opcional user_id da query string.
// Retorna 0 quando ausente (o middleware resolve para a organização do
// próprio usuário; super admin com 0 consulta todas). OK=false indica
// parâmetro inválido — resposta 422 já emitida.
func queryUserID(c *gin.Context) (uint, bool) {
	raw := c.Query("user_id")
	if raw == "" {
		return 0, true
	}
	id, err := strconv.ParseUint(raw, 10, 32)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": "user_id inválido",
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

	requested, ok := queryUserID(c)
	if !ok {
		return
	}
	userID, ok := middleware.ResolveAnalyticsUserID(c, requested)
	if !ok {
		return
	}

	svc := analyticsService(c)
	result, err := svc.MatchingDistribution(c.Request.Context(), userID, opts)
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
	requested, ok := queryUserID(c)
	if !ok {
		return
	}
	userID, ok := middleware.ResolveAnalyticsUserID(c, requested)
	if !ok {
		return
	}

	svc := analyticsService(c)
	result, err := svc.ResumeInsights(c.Request.Context(), userID)
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

	requested, ok := queryUserID(c)
	if !ok {
		return
	}
	userID, ok := middleware.ResolveAnalyticsUserID(c, requested)
	if !ok {
		return
	}

	svc := analyticsService(c)
	result, err := svc.JobTrends(c.Request.Context(), userID, period, opts)
	if err != nil {
		abortAnalyticsError(c, err)
		return
	}

	if result == nil {
		c.JSON(http.StatusOK, analytics.EmptyEnvelope(map[string]any{
			"jobs_by_status":    map[string]int64{},
			"top_users": []gin.H{},
			"posting_volume":    map[string]int64{},
			"active_companies":  0,
			"total_jobs":        0,
		}, analytics.EmptyJobTrendsMessage, analytics.EmptyJobTrendsGuidance))
		return
	}

	page := analytics.NewPagination(opts, result.PostingVolumeTotal)
	c.JSON(http.StatusOK, analytics.OKEnvelope(map[string]any{
		"jobs_by_status":    result.JobsByStatus,
		"top_users": result.TopUsers,
		"posting_volume":    result.PostingVolume,
		"active_companies":  result.ActiveCompanies,
		"total_jobs":        result.TotalJobs,
	}, &page))
}

// US4 — GET /api/v1/analytics/user/:user_id (FR-013).
// user_id=0 agrega os dados de todos os usuários (visão global do admin);
// para usuário comum o ResolveAnalyticsUserID força os próprios dados.
// Contrato: 404 usuário inexistente, 500 serviço.
func GetUserMetrics(c *gin.Context) {
	requested, err := strconv.ParseUint(c.Param("user_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
		return
	}
	userID, ok := middleware.ResolveAnalyticsUserID(c, uint(requested))
	if !ok {
		return
	}

	var user models.User
	if userID != 0 {
		if err := getDB(c).First(&user, userID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
				return
			}
			abortAnalyticsError(c, err)
			return
		}
	}

	svc := analyticsService(c)
	result, err := svc.UserMetrics(c.Request.Context(), userID)
	if err != nil {
		abortAnalyticsError(c, err)
		return
	}

	if result == nil {
		userName := user.Name
		if userID == 0 {
			userName = "Todos os usuários"
		}
		c.JSON(http.StatusOK, analytics.EmptyEnvelope(map[string]any{
			"user": gin.H{
				"id":                  userID,
				"name":                userName,
				"plan":                user.PlanSlug,
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
		}, analytics.EmptyUserMetricsMessage, analytics.EmptyUserMetricsGuidance))
		return
	}

	c.JSON(http.StatusOK, analytics.OKEnvelope(map[string]any{
		"user": gin.H{
			"id":                  result.UserID,
			"name":                result.UserName,
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
	requested, ok := queryUserID(c)
	if !ok {
		return
	}
	userID, ok := middleware.ResolveAnalyticsUserID(c, requested)
	if !ok {
		return
	}

	svc := analyticsService(c)
	result, err := svc.InterviewPrepAnalysis(c.Request.Context(), userID)
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
	requested, ok := queryUserID(c)
	if !ok {
		return
	}
	userID, ok := middleware.ResolveAnalyticsUserID(c, requested)
	if !ok {
		return
	}

	svc := analyticsService(c)
	result, err := svc.AIQuality(c.Request.Context(), userID)
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
	requested, ok := queryUserID(c)
	if !ok {
		return
	}
	userID, ok := middleware.ResolveAnalyticsUserID(c, requested)
	if !ok {
		return
	}

	svc := analyticsService(c)
	result, err := svc.TimeToHire(c.Request.Context(), userID)
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
	requested, ok := queryUserID(c)
	if !ok {
		return
	}
	userID, ok := middleware.ResolveAnalyticsUserID(c, requested)
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
	result, err := svc.SkillGap(c.Request.Context(), userID, minFreq)
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
	requested, ok := queryUserID(c)
	if !ok {
		return
	}
	userID, ok := middleware.ResolveAnalyticsUserID(c, requested)
	if !ok {
		return
	}

	svc := analyticsService(c)
	result, err := svc.Funnel(c.Request.Context(), userID)
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
	requested, ok := queryUserID(c)
	if !ok {
		return
	}
	userID, ok := middleware.ResolveAnalyticsUserID(c, requested)
	if !ok {
		return
	}

	svc := analyticsService(c)
	result, err := svc.Geographic(c.Request.Context(), userID)
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
