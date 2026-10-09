package analytics

import (
	"context"
	"fmt"

	"github.com/anomalyco/vagai-api/internal/models"
	"gorm.io/gorm"
)

// Mensagens de estado vazio do resume-insights (T020).
const (
	EmptyResumeMessage  = "Nenhuma análise de currículo disponível"
	EmptyResumeGuidance = "Envie e analise currículos (POST /resumes/analyze) para extrair forças, fraquezas e sugestões de melhoria."
)

// ResumeInsights extrai os insights de qualidade dos currículos: forças,
// fraquezas e sugestões mais frequentes (top 10) e frequência de skills
// combinando technical + behaviorals do JSON de currículos parseados
// (FR-011, US2 — T017/T018).
//
// Retorna nil quando a organização não possui análises de currículo.
func (s *Service) ResumeInsights(ctx context.Context, orgID uint) (*models.ResumeQualityInsights, error) {
	db, ctx, cancel := s.ctxDB(ctx)
	defer cancel()

	key := fmt.Sprintf("org:%d:resume-insights", orgID)
	return cachedValue(ctx, s, key, func(ctx context.Context) (*models.ResumeQualityInsights, error) {
		return loadResumeInsights(db, orgID)
	})
}

func loadResumeInsights(db *gorm.DB, orgID uint) (*models.ResumeQualityInsights, error) {
	// T017 — frequência de forças/fragilidades/sugestões nos JSON das análises.
	strengths, err := jsonColumnFrequency(db, "resume_analyses", "strengths", orgID, 10)
	if err != nil {
		return nil, err
	}
	weaknesses, err := jsonColumnFrequency(db, "resume_analyses", "weaknesses", orgID, 10)
	if err != nil {
		return nil, err
	}
	suggestions, err := jsonColumnFrequency(db, "resume_analyses", "suggestions", orgID, 10)
	if err != nil {
		return nil, err
	}

	var total int64
	qTotal := db.Table("resume_analyses")
	if orgID > 0 {
		qTotal = qTotal.Where("organization_id = ?", orgID)
	}
	if err := qTotal.Count(&total).Error; err != nil {
		return nil, err
	}
	if total == 0 {
		return nil, nil
	}

	// T018 — frequência de skills (lista skills[] do JSON de currículos).
	skills, err := skillFrequency(db, orgID, 50)
	if err != nil {
		return nil, err
	}
	if skills == nil {
		skills = map[string]int{}
	}

	return &models.ResumeQualityInsights{
		TopStrengths:   toSkillCounts(strengths, 10),
		TopWeaknesses:  toSkillCounts(weaknesses, 10),
		TopSuggestions: toSuggestionCounts(suggestions, 10),
		SkillFrequency: skills,
		TotalAnalyses:  total,
	}, nil
}

// jsonColumnFrequency conta as ocorrências de cada item de um array JSON de
// uma coluna, escopado por organization_id. Usa JSON_TABLE (MySQL 8) com
// guarda JSON_VALID para tolerar colunas malformadas.
func jsonColumnFrequency(db *gorm.DB, table, column string, orgID uint, limit int) ([]itemCount, error) {
	where := " AND t.organization_id = ?"
	if orgID == 0 {
		where = ""
	}
	query := fmt.Sprintf(`
		SELECT jt.item AS item, COUNT(*) AS cnt
		FROM %s t,
			JSON_TABLE(t.%s, '$[*]' COLUMNS (item VARCHAR(500) PATH '$')) jt
		WHERE JSON_VALID(t.%s)%s
		GROUP BY jt.item
		ORDER BY cnt DESC
		LIMIT %d`, table, column, column, where, limit)

	var args []any
	if orgID > 0 {
		args = append(args, orgID)
	}
	var rows []itemCount
	if err := db.Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// skillFrequency agrega a lista skills[] do JSON de currículos parseado,
// normalizada para minúsculas.
func skillFrequency(db *gorm.DB, orgID uint, limit int) (map[string]int, error) {
	where := " WHERE"
	var args []any
	if orgID > 0 {
		where = " WHERE r.organization_id = ? AND"
		args = append(args, orgID)
	}
	var rows []itemCount
	err := db.Raw(`
		SELECT LOWER(jt.item) AS item, COUNT(*) AS cnt
		FROM resumes r,
			JSON_TABLE(r.data, '$.skills[*]' COLUMNS (item VARCHAR(200) PATH '$')) jt
		`+where+` JSON_VALID(r.data)
		GROUP BY LOWER(jt.item)
		ORDER BY cnt DESC
		LIMIT ?`, append(args, limit)...).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[string]int, len(rows))
	for _, r := range rows {
		if r.Item == "" {
			continue
		}
		out[r.Item] = int(r.Cnt)
	}
	return out, nil
}

type itemCount struct {
	Item string
	Cnt  int64
}

func toSkillCounts(rows []itemCount, limit int) []models.SkillCount {
	out := make([]models.SkillCount, 0, minInt(len(rows), limit))
	for _, r := range rows {
		if len(out) >= limit {
			break
		}
		out = append(out, models.SkillCount{Skill: r.Item, Count: r.Cnt})
	}
	return out
}

func toSuggestionCounts(rows []itemCount, limit int) []models.SuggestionCount {
	out := make([]models.SuggestionCount, 0, minInt(len(rows), limit))
	for _, r := range rows {
		if len(out) >= limit {
			break
		}
		out = append(out, models.SuggestionCount{Suggestion: r.Item, Count: r.Cnt})
	}
	return out
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
