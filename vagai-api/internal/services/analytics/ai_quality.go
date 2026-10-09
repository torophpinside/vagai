package analytics

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/anomalyco/vagai-api/internal/models"
	"gorm.io/gorm"
)

const (
	// EmptyAIQualityMessage e EmptyAIQualityGuidance são mensagens de estado vazio para AIQualityInsights.
	EmptyAIQualityMessage  = "Nenhuma análise de IA disponível"
	EmptyAIQualityGuidance = "Execute análises de IA (ex.: via correção de respostas) para gerar insights de qualidade."
)

// AIQuality retorna os insights de qualidade das análises de IA (US6).
// Retorna (nil, nil) quando não existem análises para o escopo informado (T045).
func (s *Service) AIQuality(ctx context.Context, orgID uint) (*models.AIQualityInsights, error) {
	db, ctx, cancel := s.ctxDB(ctx)
	defer cancel()

	key := fmt.Sprintf("org:%d:ai-quality", orgID)
	return cachedValue(ctx, s, key, func(ctx context.Context) (*models.AIQualityInsights, error) {
		return aiqLoadInsights(db, orgID)
	})
}

func aiqLoadInsights(db *gorm.DB, orgID uint) (*models.AIQualityInsights, error) {
	// Total de análises
	var total int64
	qTotal := db.Model(&models.AIAnalysisResult{})
	if orgID > 0 {
		qTotal = qTotal.Where("organization_id = ?", orgID)
	}
	if err := qTotal.Count(&total).Error; err != nil {
		return nil, err
	}
	if total == 0 {
		return nil, nil
	}

	// T041 - ScoreBuckets: 5 buckets de largura 2 (0-10)
	type aiqBucketRow struct {
		Bucket string
		Cnt    int64
	}
	var bucketRows []aiqBucketRow
	qb := db.Table("ai_analysis_results").Select(`
		CASE
			WHEN analysis_score < 2.0 THEN '0-2'
			WHEN analysis_score < 4.0 THEN '2-4'
			WHEN analysis_score < 6.0 THEN '4-6'
			WHEN analysis_score < 8.0 THEN '6-8'
			ELSE '8-10'
		END AS bucket,
		COUNT(*) AS cnt
	`)
	if orgID > 0 {
		qb = qb.Where("organization_id = ?", orgID)
	}
	qb = qb.Group("bucket")
	if err := qb.Scan(&bucketRows).Error; err != nil {
		return nil, err
	}

	bucketOrder := []string{"0-2", "2-4", "4-6", "6-8", "8-10"}
	byBucket := make(map[string]int64, len(bucketRows))
	for _, r := range bucketRows {
		byBucket[r.Bucket] = r.Cnt
	}
	buckets := make([]models.ScoreBucket, 0, len(bucketOrder))
	for _, b := range bucketOrder {
		cnt := byBucket[b]
		buckets = append(buckets, models.ScoreBucket{
			Bucket:     b,
			Count:      cnt,
			Percentage: pct(cnt, total),
		})
	}

	// AverageScore
	var avg sql.NullFloat64
	qa := db.Table("ai_analysis_results").Select("AVG(analysis_score) AS avg_score")
	if orgID > 0 {
		qa = qa.Where("organization_id = ?", orgID)
	}
	if err := qa.Scan(&avg).Error; err != nil {
		return nil, err
	}
	avgScore := 0.0
	if avg.Valid {
		avgScore = round2(avg.Float64)
	}
	if avgScore > 10 {
		avgScore = round2(minFloat(avgScore, 10))
	}
	if avgScore < 0 {
		avgScore = 0
	}

	// StatusCounts
	type aiqStatusRow struct {
		Status models.AIAnalysisStatus
		Cnt    int64
	}
	var statusRows []aiqStatusRow
	qs := db.Table("ai_analysis_results").Select("status, COUNT(*) AS cnt")
	if orgID > 0 {
		qs = qs.Where("organization_id = ?", orgID)
	}
	qs = qs.Group("status")
	if err := qs.Scan(&statusRows).Error; err != nil {
		return nil, err
	}
	statusCounts := make(map[models.AIAnalysisStatus]int64, len(statusRows))
	for _, r := range statusRows {
		statusCounts[r.Status] = r.Cnt
	}

	// T042 SuccessRateBySource
	type aiqSrcRow struct {
		Source    models.VerificationSource
		Completed int64
		Error     int64
		Pending   int64
	}
	var srcRows []aiqSrcRow
	qsrc := db.Table("ai_analysis_results").Select(`
		source,
		SUM(CASE WHEN status='completed' THEN 1 ELSE 0 END) AS completed,
		SUM(CASE WHEN status='error' THEN 1 ELSE 0 END) AS error,
		SUM(CASE WHEN status='pending' THEN 1 ELSE 0 END) AS pending
	`)
	if orgID > 0 {
		qsrc = qsrc.Where("organization_id = ?", orgID)
	}
	qsrc = qsrc.Group("source")
	if err := qsrc.Scan(&srcRows).Error; err != nil {
		return nil, err
	}
	succ := make([]models.SuccessRateBySource, 0, len(srcRows))
	for _, r := range srcRows {
		sr := 0.0
		den := r.Completed + r.Error
		if den > 0 {
			sr = float64(r.Completed) / float64(den) * 100
		}
		succ = append(succ, models.SuccessRateBySource{
			Source:      r.Source,
			Completed:   r.Completed,
			Error:       r.Error,
			Pending:     r.Pending,
			SuccessRate: round2(sr),
		})
	}

	// T043 AverageProcessingSeconds: AVG(updated_at - created_at) where completed/error
	var proc sql.NullFloat64
	qp := db.Table("ai_analysis_results").Select("AVG(UNIX_TIMESTAMP(updated_at) - UNIX_TIMESTAMP(created_at)) AS proc_sec")
	if orgID > 0 {
		qp = qp.Where("organization_id = ?", orgID)
	}
	qp = qp.Where("status IN ('completed','error')")
	if err := qp.Scan(&proc).Error; err != nil {
		return nil, err
	}
	procSec := 0.0
	if proc.Valid {
		procSec = round2(proc.Float64)
	}

	return &models.AIQualityInsights{
		TotalAnalyses:            total,
		ScoreBuckets:             buckets,
		AverageScore:             avgScore,
		StatusCounts:             statusCounts,
		SuccessRateBySource:      succ,
		AverageProcessingSeconds: procSec,
	}, nil
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
