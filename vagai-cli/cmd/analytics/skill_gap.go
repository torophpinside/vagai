package analytics

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/anomalyco/vagai-cli/internal/db"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

// Mensagens de estado vazio (duplicadas do serviço — o CLI não importa o
// código da API).
const (
	sgEmptyMessage  = "Nenhuma análise de gap de skills disponível"
	sgEmptyGuidance = "Cadastre vagas, execute o matching e envie currículos (POST /resumes/analyze) para comparar as skills exigidas às do candidato."
)

// Vocabulário curado de skills — idêntico ao do serviço
// (vagai-api/internal/services/analytics/skill_gap.go).
var (
	sgTechnicalSkills = []string{
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

	sgSoftSkills = []string{
		"comunicacao", "comunicacao escrita", "lideranca", "trabalho em equipe",
		"proatividade", "adaptabilidade", "resolucao de problemas", "organizacao",
		"empatia", "criatividade", "analise critica", "negociacao",
		"gestao de tempo", "orientacao a resultados", "pensamento critico", "atencao a detalhes",
		"colaboracao", "iniciativa", "networking", "apresentacao de resultados",
	}

	sgBehavioralSkills = []string{
		"escuta ativa", "tomada de decisao", "resiliencia", "etica profissional",
		"responsabilidade", "disciplina", "pontualidade", "trabalho sob pressao",
		"inteligencia emocional", "gerenciamento de conflitos", "autoconfianca", "crescimento profissional",
		"foco", "comprometimento",
	}

	sgTermVariants = map[string][]string{
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

var skillsGapMinFreq int

var skillsGapCmd = &cobra.Command{
	Use:   "skills-gap",
	Short: "Relatório de gap de skills",
	Long: `Compara as skills exigidas nas vagas com as skills dos candidatos
(keywords dos matches e skills dos currículos): skills em comum, skills em
demanda e taxa de match por categoria.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		user, err := mustSetup(cmd)
		if err != nil {
			return err
		}
		printHeader("Skills Gap Analysis", user)

		minFreq := skillsGapMinFreq
		if minFreq < 1 {
			minFreq = 1
		}

		var sources int64
		for _, table := range []string{"jobs", "matches", "resumes"} {
			var cnt int64
			if err := db.DB.Table(table).Where("user_id = ?", user).Count(&cnt).Error; err != nil {
				return err
			}
			sources += cnt
		}
		if sources == 0 {
			fmt.Println(sgEmptyMessage)
			fmt.Println(sgEmptyGuidance)
			return nil
		}

		jobFreq, err := sgJobSkillFrequencies(db.DB, user)
		if err != nil {
			return err
		}
		candFreq, err := sgCandidateSkillFrequencies(db.DB, user)
		if err != nil {
			return err
		}

		jobScoped := sgScopedByMinFreq(jobFreq, minFreq)
		if len(jobScoped) == 0 && len(candFreq) == 0 {
			fmt.Println(sgEmptyMessage)
			fmt.Println(sgEmptyGuidance)
			return nil
		}

		jobSkills := sgSortedSkills(jobScoped, 50)
		candidateSkills := sgSortedSkills(candFreq, 50)
		overlapping := sgOverlappingSkills(jobScoped, candFreq)
		inDemand := sgInDemandSkills(jobScoped, candFreq, 10)

		fmt.Printf("\nTop skills nas vagas (frequência ≥ %d):\n", minFreq)
		sgPrintFreqList(jobSkills, jobScoped)

		fmt.Println("\nSkills dos candidatos:")
		sgPrintFreqList(candidateSkills, candFreq)

		fmt.Println("\nSkills em comum (vaga ↔ candidato):")
		sgPrintNameList(overlapping)

		fmt.Println("\nSkills em demanda (exigidas nas vagas, ausentes no candidato):")
		sgPrintFreqList(inDemand, jobScoped)

		rates := sgMatchRatesByCategory(jobScoped, candFreq)
		fmt.Println("\nTaxa de match por categoria:")
		fmt.Printf("  %-14s %5.1f%%\n", "Técnico", rates["technical"])
		fmt.Printf("  %-14s %5.1f%%\n", "Comportamental", rates["behavioral"])
		fmt.Printf("  %-14s %5.1f%%\n", "Soft skills", rates["soft_skills"])
		return nil
	},
}

// sgPrintFreqList imprime skills com frequência; entradas ausentes do mapa
// aparecem como 0.
func sgPrintFreqList(skills []string, freq map[string]int64) {
	if len(skills) == 0 {
		fmt.Println("  (sem dados)")
		return
	}
	for _, s := range skills {
		fmt.Printf("  %-30s %6d\n", s, freq[s])
	}
}

// sgPrintNameList imprime apenas os nomes das skills.
func sgPrintNameList(skills []string) {
	if len(skills) == 0 {
		fmt.Println("  (sem dados)")
		return
	}
	fmt.Println("  " + strings.Join(skills, ", "))
}

// sgJobSkillFrequencies conta as menções do vocabulário em title+description
// das vagas da organização, em uma única query com agregados condicionais.
func sgJobSkillFrequencies(gormDB *gorm.DB, user uint) (map[string]int64, error) {
	terms := sgAllTerms()
	text := "LOWER(CONCAT(COALESCE(title, ''), ' ', COALESCE(description, '')))"

	parts := make([]string, len(terms))
	aliasTerm := make(map[string]string, len(terms))
	for i, term := range terms {
		alias := fmt.Sprintf("s%d", i)
		aliasTerm[alias] = term

		conds := make([]string, 0, 2)
		for _, pattern := range sgMatchPatterns(term) {
			conds = append(conds, fmt.Sprintf("%s LIKE '%%%s%%'", text, pattern))
		}
		parts[i] = fmt.Sprintf("SUM(%s) AS %s", strings.Join(conds, " OR "), alias)
	}

	query := "SELECT " + strings.Join(parts, ", ") + " FROM jobs WHERE user_id = ?"
	rows, err := gormDB.Raw(query, user).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		if rows.Err() != nil {
			return nil, rows.Err()
		}
		return map[string]int64{}, nil
	}
	vals := make([]sql.NullFloat64, len(cols))
	scanArgs := make([]any, len(cols))
	for i := range vals {
		scanArgs[i] = &vals[i]
	}
	if err := rows.Scan(scanArgs...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return map[string]int64{}, nil
		}
		return nil, err
	}

	out := make(map[string]int64, len(terms))
	for i, col := range cols {
		term, ok := aliasTerm[col]
		if !ok || !vals[i].Valid {
			continue
		}
		if n := int64(vals[i].Float64); n > 0 {
			out[term] = n
		}
	}
	return out, nil
}

// sgCandidateSkillFrequencies agrega matches.keywords_matched e
// resumes.data $.skills[*] por JSON_TABLE e soma as contagens em Go.
func sgCandidateSkillFrequencies(gormDB *gorm.DB, user uint) (map[string]int64, error) {
	out := make(map[string]int64)

	var matchRows []sgSkillCountRow
	if err := gormDB.Raw(`
		SELECT LOWER(TRIM(jt.kw)) AS skill, COUNT(*) AS cnt
		FROM matches m,
			JSON_TABLE(m.keywords_matched, '$[*]' COLUMNS (kw VARCHAR(200) PATH '$')) jt
		WHERE m.user_id = ? AND JSON_VALID(m.keywords_matched)
		GROUP BY LOWER(TRIM(jt.kw))`, user).Scan(&matchRows).Error; err != nil {
		return nil, err
	}
	sgMergeCounts(out, matchRows)

	var resumeRows []sgSkillCountRow
	if err := gormDB.Raw(`
		SELECT LOWER(TRIM(jt.item)) AS skill, COUNT(*) AS cnt
		FROM resumes r,
			JSON_TABLE(r.data, '$.skills[*]' COLUMNS (item VARCHAR(200) PATH '$')) jt
		WHERE r.user_id = ? AND JSON_VALID(r.data)
		GROUP BY LOWER(TRIM(jt.item))`, user).Scan(&resumeRows).Error; err != nil {
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

// sgNormalizeSkill normaliza para minúsculas, sem acentos e sem espaços extras.
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
	all := make([]string, 0, len(sgTechnicalSkills)+len(sgSoftSkills)+len(sgBehavioralSkills))
	all = append(all, sgTechnicalSkills...)
	all = append(all, sgSoftSkills...)
	all = append(all, sgBehavioralSkills...)
	return all
}

// sgMatchPatterns devolve as formas (base + acentuada) de um termo para o
// matching LIKE.
func sgMatchPatterns(term string) []string {
	patterns := make([]string, 0, 2)
	patterns = append(patterns, term)
	patterns = append(patterns, sgTermVariants[term]...)
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

	out := make([]string, 0, min(len(pairs), limit))
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

// sgOverlappingSkills devolve JobSkills ∩ CandidateSkills ordenadas pela
// frequência nas vagas.
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

// sgInDemandSkills devolve JobSkills − CandidateSkills, top limit por
// frequência.
func sgInDemandSkills(job, candidate map[string]int64, limit int) []string {
	pairs := make([]sgSkillFreq, 0)
	for term, count := range job {
		if _, ok := candidate[term]; !ok {
			pairs = append(pairs, sgSkillFreq{Skill: term, Count: count})
		}
	}
	sgSortSkillFreqs(pairs)

	out := make([]string, 0, min(len(pairs), limit))
	for _, p := range pairs {
		if len(out) >= limit {
			break
		}
		out = append(out, p.Skill)
	}
	return out
}

// sgMatchRatesByCategory taxa por categoria (technical, behavioral,
// soft_skills), 0–100 com 1 casa decimal.
func sgMatchRatesByCategory(jobScoped, candidate map[string]int64) map[string]float64 {
	rates := make(map[string]float64, 3)
	rates["technical"] = sgCategoryRate(sgTechnicalSkills, jobScoped, candidate)
	rates["behavioral"] = sgCategoryRate(sgBehavioralSkills, jobScoped, candidate)
	rates["soft_skills"] = sgCategoryRate(sgSoftSkills, jobScoped, candidate)
	return rates
}

// sgCategoryRate computa numer/denom*100 (0 quando denom é 0).
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
	return sgRound1(pct64(numer, denom))
}

// sgRound1 arredonda para 1 casa decimal.
func sgRound1(v float64) float64 {
	if v < 0 {
		return float64(int64(v*10-0.5)) / 10
	}
	return float64(int64(v*10+0.5)) / 10
}

func init() {
	Cmd.AddCommand(skillsGapCmd)
	skillsGapCmd.Flags().IntVar(&skillsGapMinFreq, "min-frequency", 1,
		"Frequência mínima da skill nas vagas para considerá-la na análise")
}
