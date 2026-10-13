package analytics

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/anomalyco/vagai-api/internal/models"
	"gorm.io/gorm"
)

// Mensagens de estado vazio do skills-gap (US8 — T058, espelham
// EmptyResumeMessage/EmptyResumeGuidance).
const (
	EmptySkillGapMessage  = "Nenhuma análise de gap de skills disponível"
	EmptySkillGapGuidance = "Cadastre vagas, execute o matching e envie currículos (POST /resumes/analyze) para comparar as skills exigidas às do candidato."
)

// Vocabulário curado de skills (T053). Os jobs não têm coluna de skills: as
// skills das vagas são inferidas por LIKE sobre título + descrição. Termos em
// minúsculas, sem % ou _ (o SQL é gerado somente a partir destas constantes,
// sem entrada do usuário). ~97 termos no total.
var (
	// us8TechnicalSkills cobre linguagens, frameworks, infra e ferramentas.
	us8TechnicalSkills = []string{
		"python", "java", "javascript", "typescript", "react", "vue",
		"angular", "node", "go", "golang", "php", "ruby",
		"rails", "django", "flask", "spring", "kotlin", "swift",
		"c#", "c++", "rust", "scala", "sql", "mysql",
		"postgresql", "mongodb", "redis", "elasticsearch", "firebase", "html",
		"css", "jquery", "next.js", "docker", "kubernetes", "aws",
		"azure", "gcp", "linux", "git", "ci/cd", "terraform",
		"ansible", "jenkins", "nginx", "kafka", "graphql", "rest",
		"microservices", "machine learning", "data science", "analytics", "power bi", "tableau",
		"excel", "sap", "salesforce", "jira", "agile", "scrum",
		"flutter", "react native", "selenium",
	}

	// us8SoftSkills cobre competências interpessoais (categoria soft_skills).
	us8SoftSkills = []string{
		"comunicacao", "comunicacao escrita", "lideranca", "trabalho em equipe",
		"proatividade", "adaptabilidade", "resolucao de problemas", "organizacao",
		"empatia", "criatividade", "analise critica", "negociacao",
		"gestao de tempo", "orientacao a resultados", "pensamento critico", "atencao a detalhes",
		"colaboracao", "iniciativa", "networking", "apresentacao de resultados",
	}

	// us8BehavioralSkills cobre comportamentos observáveis (categoria behavioral).
	us8BehavioralSkills = []string{
		"escuta ativa", "tomada de decisao", "resiliencia", "etica profissional",
		"responsabilidade", "disciplina", "pontualidade", "trabalho sob pressao",
		"inteligencia emocional", "gerenciamento de conflitos", "autoconfianca", "crescimento profissional",
		"foco", "comprometimento",
	}

	// us8TermVariants adiciona formas acentuadas para o matching LIKE barato
	// dos termos escritos em ASCII (ex.: "comunicacao" também casa
	// "comunicação" no texto da vaga).
	us8TermVariants = map[string][]string{
		"comunicacao":                {"comunicação"},
		"comunicacao escrita":        {"comunicação escrita"},
		"lideranca":                  {"liderança"},
		"resolucao de problemas":     {"resolução de problemas"},
		"organizacao":                {"organização"},
		"analise critica":            {"análise crítica", "analise crítica"},
		"negociacao":                 {"negociação"},
		"gestao de tempo":            {"gestão de tempo"},
		"orientacao a resultados":    {"orientação a resultados"},
		"pensamento critico":         {"pensamento crítico"},
		"atencao a detalhes":         {"atenção a detalhes"},
		"apresentacao de resultados": {"apresentação de resultados"},
		"tomada de decisao":          {"tomada de decisão"},
		"etica profissional":         {"ética profissional"},
		"trabalho sob pressao":       {"trabalho sob pressão"},
		"inteligencia emocional":     {"inteligência emocional"},
		"autoconfianca":              {"autoconfiança"},
	}
)

// SkillGap compara as skills exigidas nas vagas com as skills dos candidatos
// (keywords dos matches + lista skills dos currículos) e devolve sobreposição,
// skills em demanda e a taxa de match por categoria (FR-013, US8 — T052–T058).
//
// minFreq é a frequência mínima da skill nas vagas para entrar na análise
// (default 1). Retorna nil quando não há dados suficientes (estado vazio).
func (s *Service) SkillGap(ctx context.Context, userID uint, minFreq int) (*models.SkillGapAnalysis, error) {
	db, ctx, cancel := s.ctxDB(ctx)
	defer cancel()

	if minFreq < 1 {
		minFreq = 1
	}
	key := fmt.Sprintf("org:%d:skills-gap:%d", userID, minFreq)
	return cachedValue(ctx, s, key, func(ctx context.Context) (*models.SkillGapAnalysis, error) {
		return loadSkillGap(db, userID, minFreq)
	})
}

func loadSkillGap(db *gorm.DB, userID uint, minFreq int) (*models.SkillGapAnalysis, error) {
	// Estado vazio: sem jobs, matches nem currículos (T058).
	empty, err := sgNoSourceData(db, userID)
	if err != nil {
		return nil, err
	}
	if empty {
		return nil, nil
	}

	// T053 — skills das vagas via vocabulário curado (agregados condicionais).
	jobFreq, err := sgJobSkillFrequencies(db, userID)
	if err != nil {
		return nil, err
	}

	// T054 — skills dos candidatos: matches.keywords_matched + resumes.data.
	candFreq, err := sgCandidateSkillFrequencies(db, userID)
	if err != nil {
		return nil, err
	}

	// T058 — sem skills em nenhuma das duas fontes → estado vazio.
	jobScoped := sgScopedByMinFreq(jobFreq, minFreq)
	if len(jobScoped) == 0 && len(candFreq) == 0 {
		return nil, nil
	}

	// T055 — listas de saída (freq desc, cap 50) e operações de conjunto.
	return &models.SkillGapAnalysis{
		JobSkills:           sgSortedSkills(jobScoped, 50),
		CandidateSkills:     sgSortedSkills(candFreq, 50),
		OverlappingSkills:   sgOverlappingSkills(jobScoped, candFreq),
		InDemandSkills:      sgInDemandSkills(jobScoped, candFreq, 10),
		MatchRateByCategory: sgMatchRatesByCategory(jobScoped, candFreq),
	}, nil
}

// sgNoSourceData reports se a organização não tem jobs, matches nem resumes.
func sgNoSourceData(db *gorm.DB, userID uint) (bool, error) {
	var total int64
	for _, table := range []string{"jobs", "matches", "resumes"} {
		var cnt int64
		q := db.Table(table)
		if userID > 0 {
			q = q.Where("user_id = ?", userID)
		}
		if err := q.Count(&cnt).Error; err != nil {
			return false, err
		}
		total += cnt
	}
	return total == 0, nil
}

// sgJobSkillFrequencies conta quantas vagas mencionam cada termo do
// vocabulário em title+description (T053). A busca LIKE é feita em Go
// (strings.Contains, SIMD) sobre o texto normalizado UMA vez por linha:
// LIKE '%termo%' no MySQL é O(linha×termo×bytes) e estoura o budget de 5s em
// descrições longas, mesmo com derived table.
func sgJobSkillFrequencies(db *gorm.DB, userID uint) (map[string]int64, error) {
	type jobText struct {
		Title       string
		Description string
	}
	q := db.Model(&models.Job{}).Select("title, description")
	if userID > 0 {
		q = q.Where("user_id = ?", userID)
	}
	var jobs []jobText
	if err := q.Find(&jobs).Error; err != nil {
		return nil, err
	}

	terms := sgAllTerms()
	out := make(map[string]int64, len(terms))
	for _, job := range jobs {
		text := strings.ToLower(job.Title + " " + job.Description)
		for _, term := range terms {
			if out[term] > 0 {
				continue
			}
			for _, pattern := range sgMatchPatterns(term) {
				if strings.Contains(text, pattern) {
					out[term]++
					break
				}
			}
		}
	}
	return out, nil
}

// sgCandidateSkillFrequencies agrega as duas fontes de skills do candidato em
// SQL e soma as contagens por chave normalizada em Go (T054):
//
//	(a) matches.keywords_matched — array JSON via JSON_TABLE;
//	(b) resumes.data $.skills[*] — mesmo padrão de skillFrequency.
func sgCandidateSkillFrequencies(db *gorm.DB, userID uint) (map[string]int64, error) {
	out := make(map[string]int64)

	// (a) keywords dos matches.
	matchQuery := `
		SELECT LOWER(TRIM(jt.kw)) AS skill, COUNT(*) AS cnt
		FROM matches m,
			JSON_TABLE(m.keywords_matched, '$[*]' COLUMNS (kw VARCHAR(200) PATH '$')) jt
		WHERE JSON_VALID(m.keywords_matched)`
	var matchArgs []any
	if userID > 0 {
		matchQuery += " AND m.user_id = ?"
		matchArgs = append(matchArgs, userID)
	}
	matchQuery += " GROUP BY LOWER(TRIM(jt.kw))"

	var matchRows []sgSkillCountRow
	if err := db.Raw(matchQuery, matchArgs...).Scan(&matchRows).Error; err != nil {
		return nil, err
	}
	sgMergeCounts(out, matchRows)

	// (b) skills dos currículos parseados.
	resumeQuery := `
		SELECT LOWER(TRIM(jt.item)) AS skill, COUNT(*) AS cnt
		FROM resumes r,
			JSON_TABLE(r.data, '$.skills[*]' COLUMNS (item VARCHAR(200) PATH '$')) jt
		WHERE JSON_VALID(r.data)`
	var resumeArgs []any
	if userID > 0 {
		resumeQuery += " AND r.user_id = ?"
		resumeArgs = append(resumeArgs, userID)
	}
	resumeQuery += " GROUP BY LOWER(TRIM(jt.item))"

	var resumeRows []sgSkillCountRow
	if err := db.Raw(resumeQuery, resumeArgs...).Scan(&resumeRows).Error; err != nil {
		return nil, err
	}
	sgMergeCounts(out, resumeRows)

	return out, nil
}

// sgSkillCountRow é a linha (skill, cnt) das agregações de candidatos.
type sgSkillCountRow struct {
	Skill string
	Cnt   int64
}

// sgMergeCounts soma as contagens em dst pela chave normalizada.
func sgMergeCounts(dst map[string]int64, rows []sgSkillCountRow) {
	for _, r := range rows {
		key := sgNormalizeSkill(r.Skill)
		if key == "" {
			continue
		}
		dst[key] += r.Cnt
	}
}

// sgNormalizeSkill normaliza uma skill para minúsculas, sem acentos e sem
// espaços extras — assim "Comunicação" (currículo) casa com "comunicacao"
// (vocabulário) e as duas fontes de candidato se somam na mesma chave.
func sgNormalizeSkill(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return ""
	}
	return sgFoldAccents(s)
}

// sgFoldAccents converte os caracteres acentuados do português para ASCII.
func sgFoldAccents(s string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case 'á', 'à', 'ã', 'â', 'ä':
			return 'a'
		case 'é', 'ê', 'è', 'ë':
			return 'e'
		case 'í', 'ì', 'î', 'ï':
			return 'i'
		case 'ó', 'ò', 'õ', 'ô', 'ö':
			return 'o'
		case 'ú', 'ù', 'û', 'ü':
			return 'u'
		case 'ç':
			return 'c'
		case 'ñ':
			return 'n'
		}
		return r
	}, s)
}

// sgAllTerms devolve o vocabulário completo (técnico + soft + behavioral).
func sgAllTerms() []string {
	all := make([]string, 0, len(us8TechnicalSkills)+len(us8SoftSkills)+len(us8BehavioralSkills))
	all = append(all, us8TechnicalSkills...)
	all = append(all, us8SoftSkills...)
	all = append(all, us8BehavioralSkills...)
	return all
}

// sgMatchPatterns devolve as formas (base + acentuada) de um termo para o
// matching LIKE.
func sgMatchPatterns(term string) []string {
	patterns := make([]string, 0, 2)
	patterns = append(patterns, term)
	patterns = append(patterns, us8TermVariants[term]...)
	return patterns
}

// sgScopedByMinFreq mantém apenas as skills com frequência >= minFreq.
func sgScopedByMinFreq(freq map[string]int64, minFreq int) map[string]int64 {
	out := make(map[string]int64, len(freq))
	for term, count := range freq {
		if count >= int64(minFreq) {
			out[term] = count
		}
	}
	return out
}

// sgSortedSkills devolve as skills ordenadas por frequência desc (desempate
// alfabético), limitadas a limit.
func sgSortedSkills(freq map[string]int64, limit int) []string {
	pairs := make([]sgSkillFreq, 0, len(freq))
	for term, count := range freq {
		pairs = append(pairs, sgSkillFreq{Skill: term, Count: count})
	}
	sgSortSkillFreqs(pairs)

	out := make([]string, 0, minInt(len(pairs), limit))
	for _, p := range pairs {
		if len(out) >= limit {
			break
		}
		out = append(out, p.Skill)
	}
	return out
}

// sgSkillFreq é o par (skill, frequência) usado nas ordenações.
type sgSkillFreq struct {
	Skill string
	Count int64
}

// sgSortSkillFreqs ordena por frequência desc e nome asc.
func sgSortSkillFreqs(pairs []sgSkillFreq) {
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].Count != pairs[j].Count {
			return pairs[i].Count > pairs[j].Count
		}
		return pairs[i].Skill < pairs[j].Skill
	})
}

// sgOverlappingSkills devolve JobSkills ∩ CandidateSkills (T055), ordenadas
// pela frequência nas vagas.
func sgOverlappingSkills(job, candidate map[string]int64) []string {
	pairs := make([]sgSkillFreq, 0)
	for term, count := range job {
		if _, ok := candidate[term]; ok {
			pairs = append(pairs, sgSkillFreq{Skill: term, Count: count})
		}
	}
	sgSortSkillFreqs(pairs)

	out := make([]string, 0, len(pairs))
	for _, p := range pairs {
		out = append(out, p.Skill)
	}
	return out
}

// sgInDemandSkills devolve JobSkills − CandidateSkills (skills exigidas nas
// vagas e ausentes no candidato), top limit por frequência (T055).
func sgInDemandSkills(job, candidate map[string]int64, limit int) []string {
	pairs := make([]sgSkillFreq, 0)
	for term, count := range job {
		if _, ok := candidate[term]; !ok {
			pairs = append(pairs, sgSkillFreq{Skill: term, Count: count})
		}
	}
	sgSortSkillFreqs(pairs)

	out := make([]string, 0, minInt(len(pairs), limit))
	for _, p := range pairs {
		if len(out) >= limit {
			break
		}
		out = append(out, p.Skill)
	}
	return out
}

// sgMatchRatesByCategory calcula a taxa de match (0–100, 1 casa decimal) de
// cada categoria: denom = skills da categoria com freq >= minFreq nas vagas;
// numer = dessas, quantas o candidato também tem (T056).
func sgMatchRatesByCategory(jobScoped, candidate map[string]int64) map[string]float64 {
	rates := make(map[string]float64, 3)
	rates["technical"] = sgCategoryRate(us8TechnicalSkills, jobScoped, candidate)
	rates["behavioral"] = sgCategoryRate(us8BehavioralSkills, jobScoped, candidate)
	rates["soft_skills"] = sgCategoryRate(us8SoftSkills, jobScoped, candidate)
	return rates
}

// sgCategoryRate computa numer/denom*100 (0 quando denom é 0), arredondado
// para 1 casa decimal.
func sgCategoryRate(vocab []string, jobScoped, candidate map[string]int64) float64 {
	var denom, numer int64
	for _, term := range vocab {
		if _, ok := jobScoped[term]; !ok {
			continue
		}
		denom++
		if _, ok := candidate[term]; ok {
			numer++
		}
	}
	if denom == 0 {
		return 0
	}
	return sgRound1(pct(numer, denom))
}

// sgRound1 arredonda para 1 casa decimal (mesmo estilo de round2).
func sgRound1(v float64) float64 {
	if v < 0 {
		return float64(int64(v*10-0.5)) / 10
	}
	return float64(int64(v*10+0.5)) / 10
}
