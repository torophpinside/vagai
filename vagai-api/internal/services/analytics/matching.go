package analytics

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/anomalyco/vagai-api/internal/models"
	"gorm.io/gorm"
)

// analyticsOrgWhere monta o filtro de tenant "WHERE organization_id = ?"
// (ou vazio, para super admin com orgID == 0 = todas as organizações).
func analyticsOrgWhere(orgID uint) (string, []any) {
	if orgID == 0 {
		return "", nil
	}
	return " WHERE organization_id = ?", []any{orgID}
}

// matchingJoinOrgWhere/Args montam o filtro de tenant no join matches→jobs
// (alias m), respeitando o agregado global do super admin.
func matchingJoinOrgWhere(orgID uint) string {
	if orgID == 0 {
		return ""
	}
	return " WHERE m.organization_id = ?"
}

func matchingJoinOrgArgs(orgID uint) []any {
	if orgID == 0 {
		return nil
	}
	return []any{orgID}
}

// Mensagens de estado vazio do matching-distribution (T015 / spec
// clarification: placeholder + guidance).
const (
	EmptyMatchingMessage  = "Nenhum match registrado ainda"
	EmptyMatchingGuidance = "Execute o matching de vagas com seus currículos para começar a ver a distribuição de scores e taxas de conversão."
)

// scoreBucketOrder garante os 5 buckets do histograma, mesmo vazios.
var scoreBucketOrder = []string{"0.0-0.2", "0.2-0.4", "0.4-0.6", "0.6-0.8", "0.8-1.0"}

// MatchingDistribution calcula a distribuição de scores de similaridade,
// média/mediana, taxa de conversão match→candidatura e os títulos de vaga
// com maior taxa de sucesso (FR-010, US1 — T011/T012/T013).
//
// Retorna nil quando a organização não possui nenhum match (estado vazio,
// T015). A escala armazenada em matches.similarity_score é 0–100; aqui é
// normalizada para 0–1 conforme contracts/api_contract.md.
func (s *Service) MatchingDistribution(ctx context.Context, orgID uint, opts PageOptions) (*models.MatchingDistribution, error) {
	db, ctx, cancel := s.ctxDB(ctx)
	defer cancel()

	key := fmt.Sprintf("org:%d:matching:page:%d:size:%d", orgID, opts.Page, opts.PageSize)
	return cachedValue(ctx, s, key, func(ctx context.Context) (*models.MatchingDistribution, error) {
		return loadMatchingDistribution(db, orgID, opts)
	})
}

func loadMatchingDistribution(db *gorm.DB, orgID uint, opts PageOptions) (*models.MatchingDistribution, error) {
	// T011 — histograma de scores normalizado para 0–1. orgID == 0 (super
	// admin) agrega todas as organizações; orgID > 0 escopa pelo tenant.
	type bucketRow struct {
		Bucket   string
		Cnt      int64
		AvgScore float64
	}

	matchWhere, matchArgs := analyticsOrgWhere(orgID)

	var rows []bucketRow
	err := db.Raw(`
		SELECT
			CASE
				WHEN similarity_score < 20  THEN '0.0-0.2'
				WHEN similarity_score < 40  THEN '0.2-0.4'
				WHEN similarity_score < 60  THEN '0.4-0.6'
				WHEN similarity_score < 80  THEN '0.6-0.8'
				ELSE '0.8-1.0'
			END AS bucket,
			COUNT(*) AS cnt,
			AVG(similarity_score) AS avg_score
		FROM matches`+matchWhere+`
		GROUP BY bucket`, matchArgs...).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	var total int64
	qTotal := db.Model(&models.Match{})
	if orgID > 0 {
		qTotal = qTotal.Where("organization_id = ?", orgID)
	}
	if err := qTotal.Count(&total).Error; err != nil {
		return nil, err
	}
	if total == 0 {
		return nil, nil
	}

	byBucket := make(map[string]bucketRow, len(rows))
	for _, r := range rows {
		byBucket[r.Bucket] = r
	}
	buckets := make([]models.ScoreBucket, 0, len(scoreBucketOrder))
	var weightedAvg float64
	for _, name := range scoreBucketOrder {
		row, ok := byBucket[name]
		if !ok {
			row = bucketRow{Bucket: name}
		}
		weightedAvg += float64(row.Cnt) * row.AvgScore
		buckets = append(buckets, models.ScoreBucket{
			Bucket:     name,
			Count:      row.Cnt,
			Percentage: pct(row.Cnt, total),
		})
	}
	average := weightedAvg / float64(total)

	// Mediana via window functions (MySQL 8) — sem transferir todas as linhas.
	var med sql.NullFloat64
	if err := db.Raw(`
		SELECT AVG(similarity_score) / 100.0 AS median
		FROM (
			SELECT similarity_score,
				ROW_NUMBER() OVER (ORDER BY similarity_score) AS rn,
				COUNT(*) OVER () AS total
			FROM matches`+matchWhere+`
		) ranked
		WHERE rn IN (FLOOR((total + 1) / 2), CEIL((total + 1) / 2))`,
		matchArgs...).Scan(&med).Error; err != nil {
		return nil, err
	}
	median01 := med.Float64
	if !med.Valid {
		median01 = average / 100
	}

	// T012 — taxa de conversão match → candidatura.
	qApplied := db.Model(&models.Match{}).Where("applied = ?", true)
	if orgID > 0 {
		qApplied = qApplied.Where("organization_id = ?", orgID)
	}
	var applied int64
	if err := qApplied.Count(&applied).Error; err != nil {
		return nil, err
	}

	// T013 — top 5 títulos por taxa de sucesso.
	type titleRow struct {
		Title   string
		Total   int64
		Applied int64
	}
	var titles []titleRow
	if err := db.Raw(`
		SELECT j.title AS title, COUNT(*) AS total, COALESCE(SUM(m.applied), 0) AS applied
		FROM matches m
		JOIN jobs j ON j.id = m.job_id`+matchingJoinOrgWhere(orgID)+`
		GROUP BY j.title
		ORDER BY COALESCE(SUM(m.applied), 0) / COUNT(*) DESC, COUNT(*) DESC
		LIMIT 5`, matchingJoinOrgArgs(orgID)...).Scan(&titles).Error; err != nil {
		return nil, err
	}
	top := make([]models.JobTitleSuccess, 0, len(titles))
	for _, t := range titles {
		top = append(top, models.JobTitleSuccess{
			Title:        t.Title,
			TotalMatches: t.Total,
			Applied:      t.Applied,
			SuccessRate:  pct(t.Applied, t.Total),
		})
	}

	return &models.MatchingDistribution{
		Buckets:        buckets,
		AverageScore:   round2(average / 100),
		MedianScore:    round2(median01),
		TotalMatches:   total,
		ConversionRate: round2(pct(applied, total)),
		TopJobTitles:   top,
	}, nil
}
