package analytics

import (
	"context"
	"errors"
	"fmt"

	"github.com/anomalyco/vagai-api/internal/models"
	"gorm.io/gorm"
)

// Mensagens de estado vazio do user-metrics (T033).
const (
	EmptyUserMetricsMessage  = "Nenhum indicador do usuário disponível"
	EmptyUserMetricsGuidance = "Cadastre vagas, envie currículos e execute o matching para começar a acompanhar os seus indicadores."
)

// UserMetrics consolida os indicadores do usuário: volumes de vagas,
// currículos, matches e candidaturas, taxa de conversão, utilização de
// features, status da assinatura e tempos médios até match/análise
// (FR-013, US4 — T029/T030/T031/T033/T034).
//
// userID=0 agrega os dados de todos os usuários (visão global do admin).
// Retorna nil quando o usuário não existe (soft delete) ou quando vagas,
// currículos e matches são simultaneamente zero (T033).
func (s *Service) UserMetrics(ctx context.Context, userID uint) (*models.UserMetrics, error) {
	db, ctx, cancel := s.ctxDB(ctx)
	defer cancel()

	key := fmt.Sprintf("user:%d:user-metrics", userID)
	return cachedValue(ctx, s, key, func(ctx context.Context) (*models.UserMetrics, error) {
		return userMetricsLoad(db, userID)
	})
}

// userMetricsLoad agrega tudo em SQL (COUNT/AVG/joins) — nunca carrega tabelas
// inteiras em Go. userID=0 executa as mesmas consultas sem o filtro de
// user_id (agregação global).
func userMetricsLoad(db *gorm.DB, userID uint) (*models.UserMetrics, error) {
	// Usuário: query por modelo aplica o soft delete (deleted_at IS NULL).
	var user models.User
	if userID != 0 {
		if err := db.First(&user, userID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, nil
			}
			return nil, err
		}
	}

	// Fragmentos de WHERE condicionais: sem filtro quando userID=0 (globais).
	userWhere := ""
	appliedWhere := " WHERE applied = ?"
	if userID != 0 {
		userWhere = " WHERE user_id = ?"
		appliedWhere = " WHERE user_id = ? AND applied = ?"
	}

	// T029 — contagens em subqueries escalares (uma única query).
	type userMetricsCounts struct {
		JobsPosted       int64
		ResumesUploaded  int64
		MatchesMade      int64
		ApplicationsMade int64
		AIAnalyses       int64
		InterviewPreps   int64
	}
	var counts userMetricsCounts
	var countArgs []interface{}
	if userID != 0 {
		countArgs = []interface{}{userID, userID, userID, userID, true, userID, userID}
	} else {
		countArgs = []interface{}{true}
	}
	if err := db.Raw(`
		SELECT
			(SELECT COUNT(*) FROM jobs`+userWhere+`) AS jobs_posted,
			(SELECT COUNT(*) FROM resumes`+userWhere+`) AS resumes_uploaded,
			(SELECT COUNT(*) FROM matches`+userWhere+`) AS matches_made,
			(SELECT COUNT(*) FROM matches`+appliedWhere+`) AS applications_made,
			(SELECT COUNT(*) FROM ai_analysis_results`+userWhere+`) AS ai_analyses,
			(SELECT COUNT(*) FROM interview_preparations`+userWhere+`) AS interview_preps`,
		countArgs...).
		Scan(&counts).Error; err != nil {
		return nil, err
	}

	// T033 — estado vazio: sem vagas, currículos e matches.
	if counts.JobsPosted == 0 && counts.ResumesUploaded == 0 && counts.MatchesMade == 0 {
		return nil, nil
	}

	subStatus := user.SubscriptionStatus

	// T030 — tempo médio até match: matches → jobs (GREATEST(0, ...) evita
	// valores negativos; COALESCE zera o AVG vazio).
	var timeToMatch float64
	matchWhere := ""
	if userID != 0 {
		matchWhere = " WHERE m.user_id = ?"
	}
	matchQuery := `
		SELECT COALESCE(AVG(GREATEST(0, DATEDIFF(m.analyzed_at, j.collected_at))), 0)
		FROM matches m
		JOIN jobs j ON j.id = m.job_id` + matchWhere
	var matchArgs []interface{}
	if userID != 0 {
		matchArgs = []interface{}{userID}
	}
	if err := db.Raw(matchQuery, matchArgs...).Scan(&timeToMatch).Error; err != nil {
		return nil, err
	}

	// T030 — tempo médio até análise: matches → interview_preparations →
	// ai_analysis_results.
	var timeToAnalysis float64
	analysisQuery := `
		SELECT COALESCE(AVG(GREATEST(0, DATEDIFF(ai.created_at, m.analyzed_at))), 0)
		FROM matches m
		JOIN interview_preparations ip ON ip.match_id = m.id
		JOIN ai_analysis_results ai ON ai.preparation_id = ip.id` + matchWhere
	if err := db.Raw(analysisQuery, matchArgs...).Scan(&timeToAnalysis).Error; err != nil {
		return nil, err
	}

	userName := user.Name
	if userID == 0 {
		userName = "Todos os usuários"
	}

	return &models.UserMetrics{
		UserID:             userID,
		UserName:           userName,
		PlanType:           user.PlanSlug,
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
