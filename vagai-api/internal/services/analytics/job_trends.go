package analytics

import (
	"context"
	"fmt"

	"github.com/anomalyco/vagai-api/internal/models"
	"gorm.io/gorm"
)

// Mensagens de estado vazio do job-trends (T027).
const (
	EmptyJobTrendsMessage  = "Nenhuma vaga registrada ainda"
	EmptyJobTrendsGuidance = "Cadastre fontes de vagas e execute o crawler para começar a ver tendências de mercado."
)

// ValidPeriods aceita o parâmetro period do job-trends (contrato: week,
// month — default — quarter, year).
var ValidPeriods = map[string]bool{
	"week": true, "month": true, "quarter": true, "year": true,
}

// JobTrends consolida as tendências de vagas: distribuição por status, top 5
// organizações por volume, volume de postagens por período e empresas ativas
// (FR-012, US3 — T022–T025).
//
// Retorna nil quando a organização (ou a visão global) não possui vagas.
func (s *Service) JobTrends(ctx context.Context, orgID uint, period string, opts PageOptions) (*models.JobMarketTrends, error) {
	period = periodOrDefault(period)
	db, ctx, cancel := s.ctxDB(ctx)
	defer cancel()

	key := fmt.Sprintf("org:%d:job-trends:%s:page:%d:size:%d", orgID, period, opts.Page, opts.PageSize)
	return cachedValue(ctx, s, key, func(ctx context.Context) (*models.JobMarketTrends, error) {
		return loadJobTrends(db, orgID, period, opts)
	})
}

func loadJobTrends(db *gorm.DB, orgID uint, period string, opts PageOptions) (*models.JobMarketTrends, error) {
	// T022 — distribuição de status por organização.
	type statusRow struct {
		Status models.JobStatus
		Cnt    int64
	}
	var statusRows []statusRow
	q := db.Table("jobs").Select("status, COUNT(*) AS cnt").Group("status")
	if orgID > 0 {
		q = q.Where("organization_id = ?", orgID)
	}
	if err := q.Scan(&statusRows).Error; err != nil {
		return nil, err
	}

	byStatus := make(map[models.JobStatus]int64, len(statusRows))
	var total int64
	for _, r := range statusRows {
		byStatus[r.Status] = r.Cnt
		total += r.Cnt
	}
	if total == 0 {
		return nil, nil
	}

	// T023 — top 5 organizações ativas por volume de vagas.
	type orgRow struct {
		OrganizationID uint
		Name           string
		JobCount       int64
	}
	var orgRows []orgRow
	orgQ := db.Table("jobs").
		Select("jobs.organization_id, organizations.name, COUNT(*) AS job_count").
		Joins("JOIN organizations ON organizations.id = jobs.organization_id").
		Group("jobs.organization_id, organizations.name").
		Order("job_count DESC").
		Limit(5)
	if orgID > 0 {
		orgQ = orgQ.Where("jobs.organization_id = ?", orgID)
	}
	if err := orgQ.Scan(&orgRows).Error; err != nil {
		return nil, err
	}
	topOrgs := make([]models.OrganizationSummary, 0, len(orgRows))
	for _, r := range orgRows {
		topOrgs = append(topOrgs, models.OrganizationSummary{
			OrganizationID: r.OrganizationID,
			Name:           r.Name,
			JobCount:       r.JobCount,
		})
	}

	// T024 — volume de postagens por período (posted_date com fallback em
	// collected_at), mais recente primeiro; lista paginada.
	bucketExpr := periodBucketSQL(period)
	type volRow struct {
		Period string
		Cnt    int64
	}
	var volTotal int64
	countQ := db.Table("jobs").Select(fmt.Sprintf("COUNT(DISTINCT %s)", bucketExpr)).
		Where("COALESCE(posted_date, collected_at) IS NOT NULL")
	if orgID > 0 {
		countQ = countQ.Where("organization_id = ?", orgID)
	}
	if err := countQ.Scan(&volTotal).Error; err != nil {
		return nil, err
	}

	var volRows []volRow
	volQ := db.Table("jobs").
		Select(fmt.Sprintf("%s AS period, COUNT(*) AS cnt", bucketExpr)).
		Where("COALESCE(posted_date, collected_at) IS NOT NULL")
	if orgID > 0 {
		volQ = volQ.Where("organization_id = ?", orgID)
	}
	if err := volQ.Group("period").Order("period DESC").
		Offset(opts.Offset()).Limit(opts.Limit()).
		Scan(&volRows).Error; err != nil {
		return nil, err
	}
	volume := make(map[string]int64, len(volRows))
	for _, r := range volRows {
		volume[r.Period] = r.Cnt
	}

	// T025 — empresas ativas: organizações com >0 vagas.
	var active int64
	activeQ := db.Table("jobs").Select("COUNT(DISTINCT organization_id)")
	if orgID > 0 {
		activeQ = activeQ.Where("organization_id = ?", orgID)
	}
	if err := activeQ.Scan(&active).Error; err != nil {
		return nil, err
	}

	return &models.JobMarketTrends{
		JobsByStatus:     byStatus,
		TopOrganizations: topOrgs,
		PostingVolume:    volume,
		ActiveCompanies:  active,
		TotalJobs:        total,
		PostingVolumeTotal: volTotal,
	}, nil
}

// periodBucketSQL devolve a expressão MySQL de agrupamento por período.
func periodBucketSQL(period string) string {
	// COALESCE(posted_date, collected_at) — vagas sem data de publicação caem
	// na data de coleta.
	src := "COALESCE(posted_date, collected_at)"
	switch period {
	case "week":
		return fmt.Sprintf("DATE_FORMAT(%s, '%%x-W%%v')", src)
	case "quarter":
		return fmt.Sprintf("CONCAT(YEAR(%s), '-Q', QUARTER(%s))", src, src)
	case "year":
		return fmt.Sprintf("DATE_FORMAT(%s, '%%Y')", src)
	default: // month
		return fmt.Sprintf("DATE_FORMAT(%s, '%%Y-%%m')", src)
	}
}

// periodOrDefault normaliza o parâmetro period (default: month).
func periodOrDefault(period string) string {
	if ValidPeriods[period] {
		return period
	}
	return "month"
}
