package analytics

import (
	"context"
	"errors"
	"fmt"

	"github.com/anomalyco/vagai-api/internal/models"
	"gorm.io/gorm"
)

// Mensagens de estado vazio do org-metrics (T033).
const (
	EmptyOrgMetricsMessage  = "Nenhum indicador da organização disponível"
	EmptyOrgMetricsGuidance = "Cadastre vagas, envie currículos e execute o matching para começar a acompanhar os indicadores da organização."
)

// OrgMetrics consolida os indicadores da organização: volumes de vagas,
// currículos, matches e candidaturas, taxa de conversão, utilização de
// features, assinatura vigente e tempos médios até match/análise
// (FR-013, US4 — T029/T030/T031/T033/T034).
//
// Retorna nil quando orgID é 0, quando a organização não existe (soft delete)
// ou quando vagas, currículos e matches são simultaneamente zero (T033).
func (s *Service) OrgMetrics(ctx context.Context, orgID uint) (*models.OrgMetrics, error) {
	if orgID == 0 {
		return nil, nil
	}

	db, ctx, cancel := s.ctxDB(ctx)
	defer cancel()

	key := fmt.Sprintf("org:%d:org-metrics", orgID)
	return cachedValue(ctx, s, key, func(ctx context.Context) (*models.OrgMetrics, error) {
		return orgMetricsLoad(db, orgID)
	})
}

// orgMetricsLoad agrega tudo em SQL (COUNT/AVG/joins) — nunca carrega tabelas
// inteiras em Go.
func orgMetricsLoad(db *gorm.DB, orgID uint) (*models.OrgMetrics, error) {
	// Organização: query por modelo aplica o soft delete (deleted_at IS NULL).
	var org models.Organization
	if err := db.First(&org, orgID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// T029 — contagens em subqueries escalares (uma única query).
	type orgMetricsCounts struct {
		JobsPosted       int64
		ResumesUploaded  int64
		MatchesMade      int64
		ApplicationsMade int64
		AIAnalyses       int64
		InterviewPreps   int64
	}
	var counts orgMetricsCounts
	if err := db.Raw(`
		SELECT
			(SELECT COUNT(*) FROM jobs WHERE organization_id = ?) AS jobs_posted,
			(SELECT COUNT(*) FROM resumes WHERE organization_id = ?) AS resumes_uploaded,
			(SELECT COUNT(*) FROM matches WHERE organization_id = ?) AS matches_made,
			(SELECT COUNT(*) FROM matches WHERE organization_id = ? AND applied = ?) AS applications_made,
			(SELECT COUNT(*) FROM ai_analysis_results WHERE organization_id = ?) AS ai_analyses,
			(SELECT COUNT(*) FROM interview_preparations WHERE organization_id = ?) AS interview_preps`,
		orgID, orgID, orgID, orgID, true, orgID, orgID).
		Scan(&counts).Error; err != nil {
		return nil, err
	}

	// T033 — estado vazio: sem vagas, currículos e matches.
	if counts.JobsPosted == 0 && counts.ResumesUploaded == 0 && counts.MatchesMade == 0 {
		return nil, nil
	}

	// Assinatura mais recente da org (model aplica soft delete; created_at DESC).
	var sub models.Subscription
	var subStatus models.SubscriptionStatus
	err := db.Where("organization_id = ?", orgID).
		Order("created_at DESC").
		First(&sub).Error
	switch {
	case err == nil:
		subStatus = sub.Status
	case !errors.Is(err, gorm.ErrRecordNotFound):
		return nil, err
	}

	// T030 — tempo médio até match: matches → jobs (GREATEST(0, ...) evita
	// valores negativos; COALESCE zera o AVG vazio).
	var timeToMatch float64
	if err := db.Raw(`
		SELECT COALESCE(AVG(GREATEST(0, DATEDIFF(m.analyzed_at, j.collected_at))), 0)
		FROM matches m
		JOIN jobs j ON j.id = m.job_id
		WHERE m.organization_id = ?`, orgID).Scan(&timeToMatch).Error; err != nil {
		return nil, err
	}

	// T030 — tempo médio até análise: matches → interview_preparations →
	// ai_analysis_results.
	var timeToAnalysis float64
	if err := db.Raw(`
		SELECT COALESCE(AVG(GREATEST(0, DATEDIFF(ai.created_at, m.analyzed_at))), 0)
		FROM matches m
		JOIN interview_preparations ip ON ip.match_id = m.id
		JOIN ai_analysis_results ai ON ai.preparation_id = ip.id
		WHERE m.organization_id = ?`, orgID).Scan(&timeToAnalysis).Error; err != nil {
		return nil, err
	}

	return &models.OrgMetrics{
		OrganizationID:     orgID,
		OrganizationName:   org.Name,
		PlanType:           org.Plan,
		SubscriptionStatus: subStatus,
		// T031 — trial→paid: apenas assinatura ativa (paga) converte.
		TrialToPaid:      subStatus == models.SubStatusActive,
		JobsPosted:       counts.JobsPosted,
		ResumesUploaded:  counts.ResumesUploaded,
		MatchesMade:      counts.MatchesMade,
		ApplicationsMade: counts.ApplicationsMade,
		ConversionRate:   round2(pct(counts.ApplicationsMade, counts.MatchesMade)),
		FeatureUtilization: models.FeatureUtilization{
			Matching:      counts.MatchesMade > 0,
			AIAnalysis:    counts.AIAnalyses > 0,
			InterviewPrep: counts.InterviewPreps > 0,
			SkillGap:      counts.ResumesUploaded > 0 && counts.MatchesMade > 0,
		},
		AverageTimeToMatchDays:    round2(timeToMatch),
		AverageTimeToAnalysisDays: round2(timeToAnalysis),
	}, nil
}
