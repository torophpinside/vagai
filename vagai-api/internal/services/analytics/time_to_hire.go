package analytics

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/anomalyco/vagai-api/internal/models"
	"gorm.io/gorm"
)

// Mensagens de estado vazio do time-to-hire (T051 / spec clarification:
// placeholder + guidance), espelhando EmptyResumeMessage/EmptyResumeGuidance.
const (
	EmptyTimeToHireMessage  = "Nenhum match disponível para calcular o tempo até a contratação"
	EmptyTimeToHireGuidance = "Execute o matching de vagas com seus currículos e registre candidaturas para medir o tempo entre coleta, análise e aplicação."
)

// TimeToHire calcula o tempo médio e mediano entre as etapas do pipeline:
// coleta→match (T047), match→análise IA (T048) e coleta→candidatura
// (T049), mais o vetor StageTiming com as três etapas na ordem contratada
// (FR-013, US7 — T047/T048/T049/T052).
//
// Retorna nil quando a organização (ou a visão global, orgID=0) não possui
// nenhum match (estado vazio, T051).
func (s *Service) TimeToHire(ctx context.Context, orgID uint) (*models.TimeToHireMetrics, error) {
	db, ctx, cancel := s.ctxDB(ctx)
	defer cancel()

	key := fmt.Sprintf("org:%d:time-to-hire", orgID)
	return cachedValue(ctx, s, key, func(ctx context.Context) (*models.TimeToHireMetrics, error) {
		return loadTimeToHire(db, orgID)
	})
}

func loadTimeToHire(db *gorm.DB, orgID uint) (*models.TimeToHireMetrics, error) {
	// T051 — estado vazio: nenhum match no escopo.
	total, err := tthMatchCount(db, orgID)
	if err != nil {
		return nil, err
	}
	if total == 0 {
		return nil, nil
	}

	// T047 — coleta → match (jobs.collected_at → matches.analyzed_at).
	matchAvg, matchMedian, matchSamples, err := tthStage(db,
		"GREATEST(0, DATEDIFF(matches.analyzed_at, jobs.collected_at))",
		"matches JOIN jobs ON jobs.id = matches.job_id",
		tthWhere(tthOrgCond(orgID, "matches")), tthArgs(orgID)...)
	if err != nil {
		return nil, err
	}

	// T048 — match → análise IA (m.analyzed_at → ai_analysis_results.created_at).
	analysisAvg, analysisMedian, analysisSamples, err := tthStage(db,
		"GREATEST(0, DATEDIFF(ai_analysis_results.created_at, m.analyzed_at))",
		"ai_analysis_results"+
			" JOIN interview_preparations ip ON ip.id = ai_analysis_results.preparation_id"+
			" JOIN matches m ON m.id = ip.match_id",
		tthWhere(tthOrgCond(orgID, "m")), tthArgs(orgID)...)
	if err != nil {
		return nil, err
	}

	// T049 — pipeline completo: coleta → candidatura (somente applied).
	applyAvg, applyMedian, applySamples, err := tthStage(db,
		"GREATEST(0, DATEDIFF(matches.applied_at, jobs.collected_at))",
		"matches JOIN jobs ON jobs.id = matches.job_id",
		tthWhere("matches.applied = TRUE", "matches.applied_at IS NOT NULL",
			tthOrgCond(orgID, "matches")), tthArgs(orgID)...)
	if err != nil {
		return nil, err
	}

	// T052 — StageTiming com exatamente 3 entradas, na ordem do contrato.
	return &models.TimeToHireMetrics{
		AverageTimeToMatchDays:    matchAvg,
		AverageTimeToAnalysisDays: analysisAvg,
		AverageTimeToApplyDays:    applyAvg,
		MedianTimeToMatchDays:     matchMedian,
		MedianTimeToAnalysisDays:  analysisMedian,
		MedianTimeToApplyDays:     applyMedian,
		StageTimings: []models.StageTiming{
			{Stage: "time_to_match", MedianDays: matchMedian, Samples: matchSamples},
			{Stage: "time_to_analysis", MedianDays: analysisMedian, Samples: analysisSamples},
			{Stage: "time_to_apply", MedianDays: applyMedian, Samples: applySamples},
		},
	}, nil
}

// tthMatchCount conta os matches do escopo (0 = global, orgID=0).
func tthMatchCount(db *gorm.DB, orgID uint) (int64, error) {
	var total int64
	q := db.Table("matches")
	if orgID > 0 {
		q = q.Where("organization_id = ?", orgID)
	}
	if err := q.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// tthStage calcula média, mediana e amostras (COUNT(*) do JOIN) de uma
// etapa do pipeline em dias. daysExpr é a expressão GREATEST(0, DATEDIFF(...)),
// fromSQL o FROM/JOIN e whereCond a cláusula WHERE completa (opcional).
// Retorna tudo arredondado em 2 casas (round2); sem linhas → 0.
func tthStage(db *gorm.DB, daysExpr, fromSQL, whereCond string, args ...any) (float64, float64, int64, error) {
	whereSQL := ""
	if whereCond != "" {
		whereSQL = " WHERE " + whereCond
	}

	// Média + amostras em uma única passada.
	type tthAggRow struct {
		AvgDays float64
		Samples int64
	}
	var agg tthAggRow
	aggSQL := fmt.Sprintf(`
		SELECT COALESCE(AVG(%s), 0) AS avg_days, COUNT(*) AS samples
		FROM %s%s`, daysExpr, fromSQL, whereSQL)
	if err := db.Raw(aggSQL, args...).Scan(&agg).Error; err != nil {
		return 0, 0, 0, err
	}

	// Mediana via window functions (MySQL 8) — mesmo padrão de matching.go.
	var med sql.NullFloat64
	medSQL := fmt.Sprintf(`
		SELECT COALESCE(AVG(v), 0) AS median
		FROM (
			SELECT %s AS v,
				ROW_NUMBER() OVER (ORDER BY %s) AS rn,
				COUNT(*) OVER () AS total
			FROM %s%s
		) t
		WHERE rn IN (FLOOR((total + 1) / 2), CEIL((total + 1) / 2))`,
		daysExpr, daysExpr, fromSQL, whereSQL)
	if err := db.Raw(medSQL, args...).Scan(&med).Error; err != nil {
		return 0, 0, 0, err
	}

	var median float64
	if med.Valid {
		median = med.Float64
	}
	return round2(agg.AvgDays), round2(median), agg.Samples, nil
}

// tthOrgCond devolve a condição de escopo por organização (vazio = global,
// orgID=0 — visão de super admin, padrão de job_trends.go). O prefixo do
// alias evita ambiguidade entre matches.organization_id e jobs.organization_id.
func tthOrgCond(orgID uint, alias string) string {
	if orgID == 0 {
		return ""
	}
	return fmt.Sprintf("%s.organization_id = ?", alias)
}

// tthArgs devolve os argumentos correspondentes a tthOrgCond.
func tthArgs(orgID uint) []any {
	if orgID == 0 {
		return nil
	}
	return []any{orgID}
}

// tthWhere junta condições não vazias com AND.
func tthWhere(conds ...string) string {
	parts := make([]string, 0, len(conds))
	for _, c := range conds {
		if c != "" {
			parts = append(parts, c)
		}
	}
	return strings.Join(parts, " AND ")
}
