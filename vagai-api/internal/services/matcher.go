package services

import (
	"math"
	"regexp"
	"sort"
	"strings"
)

type MatchResult struct {
	Score           float64  `json:"score"`
	KeywordsMatched []string `json:"keywords_matched"`
}

var (
	accentFold = strings.NewReplacer(
		"á", "a", "à", "a", "â", "a", "ã", "a", "ä", "a",
		"é", "e", "è", "e", "ê", "e", "ë", "e",
		"í", "i", "ì", "i", "î", "i", "ï", "i",
		"ó", "o", "ò", "o", "ô", "o", "õ", "o", "ö", "o",
		"ú", "u", "ù", "u", "û", "u", "ü", "u",
		"ç", "c", "ñ", "n",
	)
	nonWord = regexp.MustCompile(`[^a-z0-9]+`)

	stopWords = func() map[string]bool {
		words := strings.Fields(`a o e de do da dos das em um uma uns umas para por com sem sob sobre entre
que se na no nas nos ao aos pelo pela pelos pelas até como mais menos muito muitos muita muitas
todo toda todos todas outro outra outros outras esse essa esses essas este esta estes estas
aquele aquela aqueles aquelas seu sua seus suas nosso nossa nossos nossas eu tu ele ela voce
nos eles elas me te lhe nos vos them you he she it we they i my your his her our their
the and of to in for on with from by at an be or is are was were will not but this that
as so if then than too any can has had have been being some all each other most own
практически o das ser estar fazer ter ter de forma ainda também ja já não nao so depois
curso vaga cargo area empresa trabalho profissional experiencia conhecimento requisitos
desejavel diferencial responsabilidades atuacao habilidades exigidos ativos
tem tenho ter sendo ainda porque pois assim conforme entre
procuramos buscamos necessitamos atuando atuar`)
		out := make(map[string]bool, len(words))
		for _, w := range words {
			out[w] = true
		}
		return out
	}()
)

const (
	jobTitleWeight   = 3.0
	jobDescWeight    = 1.0
	conceptHitWeight = 4.0
)

func normalizeText(s string) string {
	s = strings.ToLower(s)
	return nonWord.ReplaceAllString(accentFold.Replace(s), " ")
}

func tokenSet(text string) map[string]bool {
	tokens := make(map[string]bool)
	for _, tok := range strings.Fields(normalizeText(text)) {
		if len(tok) < 2 || stopWords[tok] {
			continue
		}
		tokens[tok] = true
	}
	return tokens
}

func collectConcepts(data ResumeData) []string {
	concepts := make([]string, 0, len(data.Skills)+len(data.Certifications)+len(data.Languages))
	concepts = append(concepts, data.Skills...)
	concepts = append(concepts, data.Certifications...)
	concepts = append(concepts, data.Languages...)
	return concepts
}

// ResumeDataToText gera um texto plano a partir dos dados estruturados do
// curriculo, para alimentar o matching quando nao ha texto bruto extraido
// (curriculos criados/editados pelo editor web).
func ResumeDataToText(data ResumeData) string {
	var parts []string
	write := func(s string) {
		if s != "" {
			parts = append(parts, s)
		}
	}
	write(data.PersonalInfo.Name)
	write(data.PersonalInfo.Location)
	write(data.Summary)
	parts = append(parts, collectConcepts(data)...)
	for _, exp := range data.Experience {
		write(exp.Role)
		write(exp.Company)
		write(exp.Description)
	}
	for _, edu := range data.Education {
		write(edu.Degree)
		write(edu.Field)
		write(edu.Institution)
	}
	return strings.Join(parts, ". ")
}

func phraseInText(title, description, phrase string) bool {
	normalized := normalizeText(phrase)
	if normalized == "" {
		return false
	}
	re := regexp.MustCompile(`\b` + regexp.QuoteMeta(normalized) + `\b`)
	return re.MatchString(title + " " + description)
}

const negativeKeywordScorePenalty = 3.0

// negativeKeywordPenalty calcula o desconto por palavras-chave negativas.
// Cada palavra-chave distinta presente na vaga (título + descrição) desconta
// 3 pontos, uma única vez. Usa fronteira de palavra para evitar falsos
// positivos (ex.: "go" dentro de "golang" ou "governança").
func negativeKeywordPenalty(combinedJobText string, keywords []string) float64 {
	if combinedJobText == "" {
		return 0
	}
	seen := make(map[string]bool, len(keywords))
	var penalty float64
	for _, kw := range keywords {
		normKw := normalizeText(kw)
		if normKw == "" || seen[normKw] {
			continue
		}
		seen[normKw] = true
		if phraseInText("", combinedJobText, normKw) {
			penalty += negativeKeywordScorePenalty
		}
	}
	return penalty
}

// MatchResumeToJob calcula a compatibilidade entre uma vaga e o curriculo usando
// os dados estruturados gravados no banco (Resume.Data) como fonte principal de
// termos do curriculo. O resultado e deterministico e baseado em sobreposicao
// de termos ponderada: termos do titulo valem mais, e habilidades/idiomas/
// certificacoes do curriculo que aparecem na vaga recebem bonus.
func MatchResumeToJob(data ResumeData, rawContent, jobTitle, jobDescription string, negativeKeywords []string) MatchResult {
	titleN := normalizeText(jobTitle)
	descN := normalizeText(jobDescription)

	var resumeParts []string
	if data.Summary != "" {
		resumeParts = append(resumeParts, data.Summary)
	}
	for _, exp := range data.Experience {
		resumeParts = append(resumeParts, exp.Role, exp.Company, exp.Description)
	}
	for _, edu := range data.Education {
		resumeParts = append(resumeParts, edu.Degree, edu.Field, edu.Institution)
	}
	if rawContent != "" {
		resumeParts = append(resumeParts, rawContent)
	}
	resumeParts = append(resumeParts, collectConcepts(data)...)
	resumeTokens := tokenSet(strings.Join(resumeParts, " "))
	titleTokens := tokenSet(titleN)
	jobTerms := tokenSet(titleN + " " + descN)

	var totalWeight, matchedWeight float64
	keywords := make([]string, 0, len(jobTerms))

	for term := range jobTerms {
		weight := jobDescWeight
		if titleTokens[term] {
			weight = jobTitleWeight
		}
		totalWeight += weight
		if resumeTokens[term] {
			matchedWeight += weight
			keywords = append(keywords, term)
		}
	}

	conceptHits := 0
	for _, concept := range collectConcepts(data) {
		if phraseInText(titleN, descN, concept) {
			conceptHits++
			keywords = append(keywords, normalizeText(concept))
		}
	}

	scoredWeight := matchedWeight + conceptHitWeight*float64(conceptHits)
	totalScored := totalWeight + conceptHitWeight*float64(conceptHits)

	var score float64
	if totalScored > 0 {
		score = scoredWeight / totalScored * 100
	}

	// Apply negative keywords penalty
	combinedJobText := titleN + " " + descN
	score -= negativeKeywordPenalty(combinedJobText, negativeKeywords)
	if score < 0 {
		score = 0
	}
	score = math.Round(score*100) / 100
	if score > 100 {
		score = 100
	}

	keywords = dedupKeywords(keywords)
	keywords = limitKeywords(keywords)

	return MatchResult{
		Score:           score,
		KeywordsMatched: keywords,
	}
}

func dedupKeywords(keywords []string) []string {
	seen := make(map[string]bool, len(keywords))
	out := make([]string, 0, len(keywords))
	for _, kw := range keywords {
		if kw == "" || seen[kw] {
			continue
		}
		seen[kw] = true
		out = append(out, kw)
	}
	sort.Strings(out)
	return out
}

func limitKeywords(keywords []string) []string {
	const maxKeywords = 20
	if len(keywords) <= maxKeywords {
		return keywords
	}
	return keywords[:maxKeywords]
}
