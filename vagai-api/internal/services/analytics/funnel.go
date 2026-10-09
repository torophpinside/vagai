package analytics

import (
	"context"
	"database/sql"
	"fmt"
	"sort"

	"github.com/anomalyco/vagai-api/internal/models"
	"gorm.io/gorm"
)

// Mensagens de estado vazio do funil (T063) — espelham o padrão
// placeholder + guidance de EmptyResumeMessage/EmptyResumeGuidance.
const (
	EmptyFunnelMessage  = "Nenhum usuário no funil ainda"
	EmptyFunnelGuidance = "Cadastre usuários na organização, envie currículos e execute o matching para começar a acompanhar o funil de conversão."
)

// fnFunnelStages fixa a ordem dos 4 estágios do funil (T059): signups →
// currículos enviados → matches → candidaturas. A ordem é estável porque a
// lista é percorrida na ordem declarada — nunca deriva de iteração de mapa.
var fnFunnelStages = []string{
	"signups",
	"uploaded_resumes",
	"matched_with_jobs",
	"applied_to_jobs",
}

// Funnel calcula o funil de conversão em 4 estágios (T059), o tempo mediano
// entre estágios (T061) e os pontos de maior abandono (T060) — US9.
//
// Retorna nil quando não há nenhum usuário (estado vazio, T063). Com orgID
// == 0 agrega todas as organizações; orgID > 0 escopa cada tabela pelo tenant
// (Constitution III).
func (s *Service) Funnel(ctx context.Context, orgID uint) (*models.FunnelMetrics, error) {
	db, ctx, cancel := s.ctxDB(ctx)
	defer cancel()

	key := fmt.Sprintf("org:%d:funnel", orgID)
	return cachedValue(ctx, s, key, func(ctx context.Context) (*models.FunnelMetrics, error) {
		return fnFunnelLoad(db, orgID)
	})
}

func fnFunnelLoad(db *gorm.DB, orgID uint) (*models.FunnelMetrics, error) {
	// T059 — contagens agregadas em SQL, na ordem fixa dos estágios.
	signups, err := fnFunnelCountSignups(db, orgID)
	if err != nil {
		return nil, err
	}
	resumes, err := fnFunnelCount(db, "resumes", orgID, "", nil)
	if err != nil {
		return nil, err
	}
	matches, err := fnFunnelCount(db, "matches", orgID, "", nil)
	if err != nil {
		return nil, err
	}
	applied, err := fnFunnelCount(db, "matches", orgID, "applied = TRUE", nil)
	if err != nil {
		return nil, err
	}

	// T063 — sem usuários não há funil a reportar (estado vazio).
	if signups == 0 {
		return nil, nil
	}

	counts := []int64{signups, resumes, matches, applied}
	stages := make([]models.FunnelStageMetric, 0, len(fnFunnelStages))
	for i, name := range fnFunnelStages {
		// Conversão acumulada sobre o estágio 1 (signups = 100 quando há
		// contagem), arredondada em 1 casa decimal (T059).
		stages = append(stages, models.FunnelStageMetric{
			Stage:          name,
			Count:          counts[i],
			ConversionRate: fnFunnelRound1(pct(counts[i], signups)),
		})
	}

	// T061 — tempo mediano por estágio.
	timings, err := fnFunnelStageTimings(db, orgID, counts)
	if err != nil {
		return nil, err
	}

	return &models.FunnelMetrics{
		FunnelStages:       stages,
		MedianTimePerStage: timings,
		DropOffPoints:      fnFunnelDropOffs(stages),
		TotalUsers:         signups,
	}, nil
}

// fnFunnelCountSignups conta usuários distintos que pertencem a pelo menos
// uma membership (T059, estágio 1): COUNT(DISTINCT users.id) com JOIN em
// memberships, excluindo usuários soft-deletados. Org 0 agrega tudo.
func fnFunnelCountSignups(db *gorm.DB, orgID uint) (int64, error) {
	q := db.Table("users").
		Select("COUNT(DISTINCT users.id)").
		Joins("JOIN memberships ms ON ms.user_id = users.id").
		Where("users.deleted_at IS NULL")
	if orgID > 0 {
		q = q.Where("ms.organization_id = ?", orgID)
	}
	var n int64
	if err := q.Scan(&n).Error; err != nil {
		return 0, err
	}
	return n, nil
}

// fnFunnelCount conta linhas de <table>, escopadas por organization_id quando
// orgID > 0, com <extra> opcional no WHERE (ex.: "applied = TRUE").
func fnFunnelCount(db *gorm.DB, table string, orgID uint, extra string, extraArgs []any) (int64, error) {
	q := db.Table(table)
	if orgID > 0 {
		q = q.Where(table+".organization_id = ?", orgID)
	}
	if extra != "" {
		q = q.Where(extra, extraArgs...)
	}
	var n int64
	if err := q.Count(&n).Error; err != nil {
		return 0, err
	}
	return n, nil
}

// fnFunnelStageTimings calcula o tempo mediano (em dias) de cada transição do
// funil (T061). Cada valor é a DIFERENÇA entre duas MEDIANAS de timestamps
// (e não a mediana das diferenças): a mediana é obtida com o padrão de window
// functions de matching.go aplicado a UNIX_TIMESTAMP da coluna —
//
//	AVG(v) WHERE rn IN (FLOOR((total+1)/2), CEIL((total+1)/2))
//
// Samples é a contagem de linhas do estágio de destino da transição.
func fnFunnelStageTimings(db *gorm.DB, orgID uint, counts []int64) ([]models.StageTiming, error) {
	// Mesmo recorte do estágio 1 (usuários ativos da organização).
	membershipsFrom := "memberships ms JOIN users u ON u.id = ms.user_id"
	membershipsWhere := "u.deleted_at IS NULL"
	var membershipsArgs []any
	if orgID > 0 {
		membershipsWhere += " AND ms.organization_id = ?"
		membershipsArgs = append(membershipsArgs, orgID)
	}

	resumesWhere, resumesArgs := fnFunnelWhere("resumes", orgID, "")
	matchesWhere, matchesArgs := fnFunnelWhere("matches", orgID, "")

	mSignup, err := fnFunnelMedian(db, "ms.created_at", membershipsFrom, membershipsWhere, membershipsArgs...)
	if err != nil {
		return nil, err
	}
	mResume, err := fnFunnelMedian(db, "uploaded_at", "resumes", resumesWhere, resumesArgs...)
	if err != nil {
		return nil, err
	}
	mMatch, err := fnFunnelMedian(db, "analyzed_at", "matches", matchesWhere, matchesArgs...)
	if err != nil {
		return nil, err
	}
	mMatchApplied, err := fnFunnelMedian(db, "analyzed_at", "matches",
		matchesWhere+" AND applied = TRUE", matchesArgs...)
	if err != nil {
		return nil, err
	}
	mAppliedAt, err := fnFunnelMedian(db, "applied_at", "matches",
		matchesWhere+" AND applied = TRUE AND applied_at IS NOT NULL", matchesArgs...)
	if err != nil {
		return nil, err
	}

	return []models.StageTiming{
		{Stage: "signup_to_resume", MedianDays: fnFunnelMedianDays(mResume, mSignup), Samples: counts[1]},
		{Stage: "resume_to_match", MedianDays: fnFunnelMedianDays(mMatch, mResume), Samples: counts[2]},
		{Stage: "match_to_apply", MedianDays: fnFunnelMedianDays(mAppliedAt, mMatchApplied), Samples: counts[3]},
	}, nil
}

// fnFunnelWhere devolve o WHERE (e os args) de <table> com o escopo opcional
// de organização: orgID == 0 agrega todas as organizações; orgID > 0 filtra
// por organization_id. <extra> é acrescentado ao WHERE com AND.
func fnFunnelWhere(table string, orgID uint, extra string) (string, []any) {
	where := "1 = 1"
	var args []any
	if orgID > 0 {
		where = table + ".organization_id = ?"
		args = append(args, orgID)
	}
	if extra != "" {
		where += " AND " + extra
	}
	return where, args
}

// fnFunnelMedian devolve a mediana (epoch segundos) da coluna no recorte,
// usando o padrão de window functions de matching.go (ROW_NUMBER/COUNT OVER
// sobre a coluna ordenada). Recorte vazio vira 0 (COALESCE).
func fnFunnelMedian(db *gorm.DB, column, from, where string, args ...any) (float64, error) {
	query := fmt.Sprintf(`
		SELECT COALESCE((
			SELECT AVG(v)
			FROM (
				SELECT UNIX_TIMESTAMP(%s) AS v,
					ROW_NUMBER() OVER (ORDER BY %s) AS rn,
					COUNT(*) OVER () AS total
				FROM %s
				WHERE %s
			) ranked
			WHERE rn IN (FLOOR((total + 1) / 2), CEIL((total + 1) / 2))
		), 0) AS median`, column, column, from, where)

	var med sql.NullFloat64
	if err := db.Raw(query, args...).Scan(&med).Error; err != nil {
		return 0, err
	}
	if !med.Valid {
		return 0, nil
	}
	return med.Float64, nil
}

// fnFunnelMedianDays converte a diferença de duas medianas (epoch segundos)
// em dias, arredondada em 1 casa decimal e truncada em 0 (nunca negativa).
func fnFunnelMedianDays(from, to float64) float64 {
	d := (to - from) / 86400
	if d < 0 {
		d = 0
	}
	return fnFunnelRound1(d)
}

// fnFunnelDropOffs devolve até 2 estágios (T060) com as MENORES razões
// estágio-a-estágio (count_i / count_{i-1}), em ordem crescente. Só entram os
// estágios 2..4 — o estágio 1 (signups) nunca é candidato. Denominador 0 vira
// razão 0 (queda total); empates preservam a ordem do funil.
func fnFunnelDropOffs(stages []models.FunnelStageMetric) []string {
	type candidate struct {
		stage string
		ratio float64
	}
	candidates := make([]candidate, 0, len(stages)-1)
	for i := 1; i < len(stages); i++ {
		r := 0.0
		if stages[i-1].Count > 0 {
			r = float64(stages[i].Count) / float64(stages[i-1].Count)
		}
		candidates = append(candidates, candidate{stage: stages[i].Stage, ratio: r})
	}
	sort.SliceStable(candidates, func(a, b int) bool {
		return candidates[a].ratio < candidates[b].ratio
	})

	limit := 2
	if len(candidates) < limit {
		limit = len(candidates)
	}
	out := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		out = append(out, candidates[i].stage)
	}
	return out
}

// fnFunnelRound1 arredonda para 1 casa decimal (mesmo critério de round2 em
// base.go: arredondamento para o inteiro mais próximo, sinal preservado).
func fnFunnelRound1(v float64) float64 {
	if v < 0 {
		return float64(int64(v*10-0.5)) / 10
	}
	return float64(int64(v*10+0.5)) / 10
}
