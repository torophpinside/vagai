package analytics

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/anomalyco/vagai-api/internal/models"
	"gorm.io/gorm"
)

// Mensagens de estado vazio do interview-prep (T039).
const (
	EmptyInterviewPrepMessage  = "Nenhuma preparação de entrevista disponível"
	EmptyInterviewPrepGuidance = "Gere uma preparação de entrevista (POST /interview-prep) e conclua uma verificação (POST /interview-prep/:id/verify) para ver categorias, fontes e a evolução das notas."
)

// prepMaxTrendRows limita as linhas carregadas para classificar a tendência
// (o histórico por organização é pequeno; a agregação da média fica no SQL).
const prepMaxTrendRows = 500

// InterviewPrepAnalysis consolida a análise das preparações de entrevista:
// distribuição de categorias das perguntas (T035), distribuição de fontes das
// preparações — ai/template (T036) — nota média de verificação e tendência
// de evolução das notas (T037).
//
// Retorna nil quando a organização não possui preparações, perguntas nem
// histórico de verificação (estado vazio, T039).
func (s *Service) InterviewPrepAnalysis(ctx context.Context, userID uint) (*models.InterviewPrepAnalysis, error) {
	db, ctx, cancel := s.ctxDB(ctx)
	defer cancel()

	key := fmt.Sprintf("org:%d:interview-prep", userID)
	return cachedValue(ctx, s, key, func(ctx context.Context) (*models.InterviewPrepAnalysis, error) {
		return loadInterviewPrepAnalysis(db, userID)
	})
}

func loadInterviewPrepAnalysis(db *gorm.DB, userID uint) (*models.InterviewPrepAnalysis, error) {
	// T035 — contagem de interview_questions por categoria (escopo da org).
	type prepCategoryRow struct {
		Category models.InterviewCategory
		Cnt      int64
	}
	var catRows []prepCategoryRow
	if err := db.Raw(`
		SELECT category, COUNT(*) AS cnt
		FROM interview_questions
		WHERE user_id = ?
		GROUP BY category`, userID).Scan(&catRows).Error; err != nil {
		return nil, err
	}
	categories := make(map[models.InterviewCategory]int64, len(catRows))
	for _, r := range catRows {
		categories[r.Category] = r.Cnt
	}

	// T036 — contagem de interview_preparations por fonte (ai/template).
	type prepSourceRow struct {
		Source models.InterviewPreparationSource
		Cnt    int64
	}
	var srcRows []prepSourceRow
	if err := db.Raw(`
		SELECT source, COUNT(*) AS cnt
		FROM interview_preparations
		WHERE user_id = ?
		GROUP BY source`, userID).Scan(&srcRows).Error; err != nil {
		return nil, err
	}
	sources := make(map[models.InterviewPreparationSource]int64, len(srcRows))
	for _, r := range srcRows {
		sources[r.Source] = r.Cnt
	}

	// TotalPreparations — COUNT(*) de interview_preparations da org.
	var total int64
	if err := db.Table("interview_preparations").
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, err
	}

	// T037 — média agregada em SQL (sem linhas → NULL → 0) e notas ordenadas
	// por verified_at ASC para classificar a tendência em Go.
	var avgScore sql.NullFloat64
	if err := db.Raw(`
		SELECT AVG(score) AS avg_score
		FROM preparation_verification_histories
		WHERE user_id = ?`, userID).Scan(&avgScore).Error; err != nil {
		return nil, err
	}

	type prepScoreRow struct {
		Score float64
	}
	var scoreRows []prepScoreRow
	if err := db.Raw(`
		SELECT score
		FROM preparation_verification_histories
		WHERE user_id = ?
		ORDER BY verified_at ASC
		LIMIT ?`, userID, prepMaxTrendRows).Scan(&scoreRows).Error; err != nil {
		return nil, err
	}

	// T039 — vazio: sem preparações, sem perguntas e sem histórico.
	if total == 0 && len(catRows) == 0 && len(scoreRows) == 0 {
		return nil, nil
	}

	scores := make([]float64, 0, len(scoreRows))
	for _, r := range scoreRows {
		scores = append(scores, r.Score)
	}

	return &models.InterviewPrepAnalysis{
		TotalPreparations:        total,
		CategoryDistribution:     categories,
		SourceDistribution:       sources,
		AverageVerificationScore: prepRound1(avgScore.Float64),
		VerificationTrend:        prepTrendFromScores(scores),
	}, nil
}

// prepRound1 arredonda v para 1 casa decimal (nota média de verificação).
func prepRound1(v float64) float64 {
	if v < 0 {
		return float64(int64(v*10-0.5)) / 10
	}
	return float64(int64(v*10+0.5)) / 10
}

// prepTrendFromScores classifica a tendência das notas já ordenadas por
// verified_at ASC (T037): menos de 2 notas → "stable"; caso contrário a
// metade mais nova é comparada com a metade mais antiga (±5%). Metade mais
// antiga com média 0 → "stable".
func prepTrendFromScores(scores []float64) string {
	if len(scores) < 2 {
		return "stable"
	}
	mid := len(scores) / 2
	older := prepAvgFloat(scores[:mid])
	newer := prepAvgFloat(scores[mid:])
	if older == 0 {
		return "stable"
	}
	switch {
	case newer >= older*1.05:
		return "improving"
	case newer <= older*0.95:
		return "declining"
	default:
		return "stable"
	}
}

// prepAvgFloat devolve a média de values (0 quando vazia).
func prepAvgFloat(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}
