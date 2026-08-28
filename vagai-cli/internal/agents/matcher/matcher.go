package matcher

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/anomalyco/vagai-cli/internal/agents/lmstudio"
	"github.com/anomalyco/vagai-cli/internal/db"
	"github.com/anomalyco/vagai-cli/internal/models"
	"gorm.io/gorm/clause"
)

var useLMStudio = true
var idf map[string]float64

const maxParallelAI = 2

type jobTask struct {
	job      models.Job
	resume   models.Resume
	userCity string
}

type matchResult struct {
	jobID    uint
	resumeID uint
	score    float64
	keywords []string
	reason   string
	err      error
}

// JobLocation armazena a localização analisada de uma vaga
type JobLocation struct {
	Type  string // "remote", "presencial", "hybrid", "unknown"
	City  string // cidade extraída da descrição
	State string // estado extraído
}

func Run(threshold int, force bool) error {
	log.Println("Iniciando Matcher Agent...")

	if err := db.Init(); err != nil {
		return fmt.Errorf("falha ao inicializar banco: %w", err)
	}

	log.Printf("Verificando conexão com LM Studio...")
	_, err := lmstudio.Chat("ping", "Responda apenas pong")
	if err != nil {
		log.Printf("🚨 LM Studio inacessível em %s: %v", lmstudio.GetBaseURL(), err)
		log.Printf("O Agente de Matching continuará usando o Fallback Algorítmico (menos preciso).")
	} else {
		log.Printf("🚀 Conexão com LM Studio estabelecida com sucesso!")
	}

	// Buscar todas as vagas novas de sites ativos, agrupadas por organização
	var jobs []models.Job
	query := db.DB.Where("status = ?", models.JobStatusNew).
		Joins("JOIN sites ON sites.id = jobs.site_id").
		Where("sites.active = ?", true).
		Order("jobs.organization_id ASC")
	if force {
		query = db.DB.Joins("JOIN sites ON sites.id = jobs.site_id").
			Where("sites.active = ?", true).
			Order("jobs.organization_id ASC")
	}
	query.Find(&jobs)

	if len(jobs) == 0 {
		log.Println("Nenhuma vaga nova para processar")
		return nil
	}

	// Agrupar vagas por organização para derivar currículo e cidade de cada uma
	orgJobs := make(map[uint][]models.Job)
	for _, job := range jobs {
		if job.OrganizationID == 0 {
			// Sem org definida (ex.: fetch manual), pula sem organização associada
			log.Printf("Vaga %d sem organization_id definida, ignorada", job.ID)
			continue
		}
		orgJobs[job.OrganizationID] = append(orgJobs[job.OrganizationID], job)
	}

	var totalMatches int
	for orgID, orgJobList := range orgJobs {
		matches := runForOrg(orgID, orgJobList, threshold)
		totalMatches += matches
	}

	log.Printf("Matcher Agent finalizado. %d matches encontrados", totalMatches)
	return nil
}

func runForOrg(orgID uint, jobs []models.Job, threshold int) int {
	log.Printf("Processando %d vagas da organização %d", len(jobs), orgID)

	// Buscar cidade do usuário da organização
	var userCity string
	var users []models.User
	db.DB.Where("organization_id = ?", orgID).Find(&users)
	for _, u := range users {
		if u.City != "" {
			userCity = u.City
			break
		}
	}
	if userCity != "" {
		log.Printf("📍 Cidade do usuário (org %d): %s", orgID, userCity)
	} else {
		log.Printf("📍 Nenhuma cidade configurada na organização %d. Filtro de localização desativado.", orgID)
	}

	// Usar apenas o currículo mais recente da organização
	var resume models.Resume
	if err := db.DB.Where("organization_id = ?", orgID).Order("uploaded_at DESC").First(&resume).Error; err != nil {
		log.Printf("Nenhum currículo encontrado para org %d", orgID)
		return 0
	}
	resumes := []models.Resume{resume}

	log.Printf("Construindo corpus TF-IDF com %d vagas e %d currículos...", len(jobs), len(resumes))
	corpus := buildCorpus(jobs, resumes)
	idf = computeIDF(corpus)
	log.Printf("IDF calculado para %d termos únicos", len(idf))

	log.Printf("Processando %d vagas com %d currículos (max %d paralelas)", len(jobs), len(resumes), maxParallelAI)

	jobsChan := make(chan jobTask, len(jobs)*len(resumes))
	resultsChan := make(chan matchResult, len(jobs)*len(resumes))

	var wg sync.WaitGroup

	for w := 0; w < maxParallelAI; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for task := range jobsChan {
				result := processTask(task.job, task.resume, task.userCity, threshold)
				resultsChan <- result
			}
		}(w)
	}

	go func() {
		for _, job := range jobs {
			for _, resume := range resumes {
				jobsChan <- jobTask{job: job, resume: resume, userCity: userCity}
			}
		}
		close(jobsChan)
	}()

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	var matchCount int
	for result := range resultsChan {
		if result.err != nil {
			log.Printf("Aviso no matching job=%d resume=%d: %v", result.jobID, result.resumeID, result.err)
			continue
		}
		// Buscar organization_id do job
		var job models.Job
		db.DB.First(&job, result.jobID)

		match := models.Match{
			OrganizationID:  job.OrganizationID,
			JobID:           result.jobID,
			ResumeID:        result.resumeID,
			SimilarityScore: result.score,
			KeywordsMatched: fmt.Sprintf(`["%s"]`, strings.Join(result.keywords, `", "`)),
			AIReason:        result.reason,
		}
		tx := db.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&match)
		if tx.Error != nil {
			log.Printf("Erro ao salvar match: %v", tx.Error)
		} else if tx.RowsAffected == 0 {
			log.Printf("Match duplicado ignorado: job=%d resume=%d", result.jobID, result.resumeID)
		} else {
			matchCount++
			newStatus := models.JobStatusAnalyzed
			if result.score > float64(threshold) {
				newStatus = models.JobStatusMatched
			}
			db.DB.Model(&job).Update("status", newStatus)
			log.Printf("Match salvo: job=%d resume=%d score=%.2f status=%s", result.jobID, result.resumeID, result.score, newStatus)
		}
	}

	log.Printf("Matcher Agent finalizado para org %d. %d matches encontrados", orgID, matchCount)
	return matchCount
}

func processTask(job models.Job, resume models.Resume, userCity string, threshold int) matchResult {
	result := matchResult{
		jobID:    job.ID,
		resumeID: resume.ID,
	}

	// Analisar localização da vaga
	jobLoc := analyzeJobLocation(job.Title+" "+job.Description, userCity)

	score, keywords, reason, err := calculateMatchAI(job.Title, job.Description, resume.Content, userCity, jobLoc)
	if err != nil {
		result.err = err
		return result
	}

	// Aplicar penalidade de localização para vagas presenciais/hibridas fora da cidade
	if userCity != "" && jobLoc.Type != "remote" && jobLoc.Type != "unknown" {
		penalty := locationPenalty(jobLoc, userCity)
		if penalty > 0 {
			score = score - penalty
			if score < 0 {
				score = 0
			}
			reason += fmt.Sprintf(" [Penalidade localização: -%.0f pts]", penalty)
		}
	}

	result.score = score
	result.keywords = keywords
	result.reason = reason
	return result
}

type AIResponse struct {
	Score  float64 `json:"score"`
	Reason string  `json:"reason"`
}

func calculateMatchAI(jobTitle, jobDesc, resumeContent, userCity string, jobLoc JobLocation) (float64, []string, string, error) {
	if resumeContent == "" {
		return 0, nil, "", fmt.Errorf("currículo vazio")
	}

	if jobDesc == "" {
		jobDesc = jobTitle
	}

	const maxLen = 2000
	const resumePreviewLen = 500
	if len(resumeContent) > maxLen {
		resumeContent = resumeContent[:maxLen]
	}
	if len(jobDesc) > maxLen {
		jobDesc = jobDesc[:maxLen]
	}

	log.Printf("Calculando match AI: job_title=%s, resume_len=%d", jobTitle, len(resumeContent))

	resumeForPrompt := resumeContent
	if len(resumeForPrompt) > 500 {
		resumeForPrompt = resumeForPrompt[:500]
	}

	// Construir prompt com contexto de localização
	locationContext := ""
	if userCity != "" {
		locationContext = fmt.Sprintf(`
LOCALIZAÇÃO DO USUARIO: %s
TIPO DA VAGA: %s
CIDADE DA VAGA: %s`, userCity, jobLoc.Type, jobLoc.City)
	}

	prompt := fmt.Sprintf(`Analise se o currículo é adequado para a vaga.
VAGA: %s - %s%s
RESUMO CURRÍCULO: %s
Considere se a vaga é remota, presencial ou híbrida, e se a localização é compatível com a do usuário.
Responda APENAS um objeto JSON no formato: {"score": 0-100, "reason": "sua explicação curta"}`, jobTitle, jobDesc, locationContext, resumeForPrompt)

	log.Printf("Chamando LM Studio para análise...")
	response, err := lmstudio.Chat(prompt, "Você é um especialista em recruitment tech. Responda sempre em JSON.")
	if err != nil {
		log.Printf("⚠️ Erro ao chamar LM Studio: %v. Usando fallback algorítmico.", err)
		return calculateMatchFallback(jobTitle, jobDesc, resumeContent, userCity, jobLoc)
	}

	log.Printf("Resposta da AI recebida. Processando...")

	first := strings.Index(response, "{")
	last := strings.LastIndex(response, "}")
	if first != -1 && last != -1 && last > first {
		response = response[first : last+1]
	}

	var aiResp AIResponse
	if err := json.Unmarshal([]byte(response), &aiResp); err != nil {
		log.Printf("⚠️ Erro ao decodificar JSON da AI: %v. Resposta bruta: %s. Usando fallback.", err, response)
		return calculateMatchFallback(jobTitle, jobDesc, resumeContent, userCity, jobLoc)
	}

	log.Printf("✅ Análise AI concluída com sucesso. Score: %.2f", aiResp.Score)
	keywords := extractKeywordsLMStudio(jobDesc, resumeContent)

	return aiResp.Score, keywords, aiResp.Reason, nil
}

func calculateMatchFallback(jobTitle, jobDesc, resumeContent, userCity string, jobLoc JobLocation) (float64, []string, string, error) {
	jobWords := extractWords(jobDesc)
	resumeWords := extractWords(resumeContent)

	if len(jobWords) == 0 {
		return 0, nil, "", fmt.Errorf("vaga sem conteúdo textual válido (id: %s)", jobTitle)
	}
	if len(resumeWords) == 0 {
		return 0, nil, "", fmt.Errorf("currículo sem conteúdo textual válido")
	}

	textual := cosineSimilarityTFIDF(jobDesc, resumeContent)

	skills := extractKeywordsLMStudio(jobDesc, resumeContent)
	skillScore := float64(len(skills)) * (100.0 / 15.0)
	if skillScore > 100 {
		skillScore = 100
	}

	loc := locationScore(jobLoc, userCity) / 100

	// Ajustar peso da localização baseado no tipo da vaga
	locWeight := 0.1
	if userCity != "" && jobLoc.Type == "presencial" {
		locWeight = 0.25
	} else if userCity != "" && jobLoc.Type == "hybrid" {
		locWeight = 0.15
	}

	score := 0.6*textual + 0.3*skillScore + locWeight*loc*100
	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}

	reason := fmt.Sprintf("TF-IDF: %.1f, Skills: %s, Local: %.1f", textual, strings.Join(skills, ", "), loc)
	return score, skills, reason, nil
}

func buildCorpus(jobs []models.Job, resumes []models.Resume) []string {
	corpus := make([]string, 0, len(jobs)+len(resumes))
	for _, j := range jobs {
		corpus = append(corpus, j.Title+" "+j.Description)
	}
	for _, r := range resumes {
		corpus = append(corpus, r.Content)
	}
	return corpus
}

func computeIDF(corpus []string) map[string]float64 {
	docCount := make(map[string]int)
	for _, doc := range corpus {
		words := unique(extractWords(doc))
		for _, w := range words {
			docCount[w]++
		}
	}
	N := float64(len(corpus))
	result := make(map[string]float64, len(docCount))
	for w, count := range docCount {
		result[w] = math.Log(N / float64(count))
	}
	return result
}

func termFrequency(words []string) map[string]float64 {
	tf := make(map[string]float64)
	for _, w := range words {
		tf[w]++
	}
	total := float64(len(words))
	if total == 0 {
		return tf
	}
	for w, c := range tf {
		tf[w] = c / total
	}
	return tf
}

func cosineSimilarityTFIDF(text1, text2 string) float64 {
	if idf == nil {
		return 0
	}

	words1 := extractWords(text1)
	words2 := extractWords(text2)

	tf1 := termFrequency(words1)
	tf2 := termFrequency(words2)

	vocab := make(map[string]bool)
	for w := range tf1 {
		vocab[w] = true
	}
	for w := range tf2 {
		vocab[w] = true
	}

	var dot, norm1, norm2 float64
	for w := range vocab {
		idfW := idf[w]
		v1 := tf1[w] * idfW
		v2 := tf2[w] * idfW
		dot += v1 * v2
		norm1 += v1 * v1
		norm2 += v2 * v2
	}

	if norm1 == 0 || norm2 == 0 {
		return 0
	}
	return dot / (math.Sqrt(norm1) * math.Sqrt(norm2)) * 100
}

// stripAccents remove acentos para permitir comparação insensível
// (ex.: "Sao Paulo" == "são paulo")
var accentReplacer = strings.NewReplacer(
	"á", "a", "à", "a", "ã", "a", "â", "a", "ä", "a",
	"é", "e", "è", "e", "ê", "e", "ë", "e",
	"í", "i", "ì", "i", "î", "i", "ï", "i",
	"ó", "o", "ò", "o", "õ", "o", "ô", "o", "ö", "o",
	"ú", "u", "ù", "u", "û", "u", "ü", "u",
	"ç", "c", "ñ", "n",
)

func stripAccents(s string) string {
	return accentReplacer.Replace(strings.ToLower(s))
}

type cityPattern struct {
	re        *regexp.Regexp
	canonical string
}

// cityAliases mapeia grafias e siglas para o nome canônico da cidade.
var cityAliases = []struct{ alias, canonical string }{
	{"sao paulo", "sao paulo"}, {"sp", "sao paulo"},
	{"rio de janeiro", "rio de janeiro"}, {"rj", "rio de janeiro"},
	{"belo horizonte", "belo horizonte"}, {"bh", "belo horizonte"},
	{"porto alegre", "porto alegre"}, {"poa", "porto alegre"},
	{"florianopolis", "florianopolis"}, {"floripa", "florianopolis"},
	{"curitiba", "curitiba"}, {"cwb", "curitiba"},
	{"campinas", "campinas"},
	{"brasilia", "brasilia"}, {"bsb", "brasilia"},
	{"salvador", "salvador"}, {"ssa", "salvador"},
	{"fortaleza", "fortaleza"},
	{"recife", "recife"}, {"rec", "recife"},
	{"manaus", "manaus"}, {"mao", "manaus"},
	{"lisboa", "lisboa"}, {"lisbon", "lisboa"},
	{"porto", "porto"},
}

// Padrões ordenados por especificidade (mais longos primeiro) para que
// "porto alegre" vença "porto" e "são paulo" vença "sp".
var cityPatterns = buildCityPatterns()

func buildCityPatterns() []cityPattern {
	patterns := make([]cityPattern, 0, len(cityAliases))
	for _, entry := range cityAliases {
		re, err := regexp.Compile(`\b` + regexp.QuoteMeta(entry.alias) + `\b`)
		if err != nil {
			continue
		}
		patterns = append(patterns, cityPattern{re: re, canonical: entry.canonical})
	}
	sort.Slice(patterns, func(i, j int) bool {
		return len(patterns[i].canonical) > len(patterns[j].canonical)
	})
	return patterns
}

// normalizeUserCity padroniza a cidade configurada pelo usuário
// (ex.: "SP", "São Paulo" e "Sao Paulo" -> "sao paulo")
func normalizeUserCity(userCity string) string {
	stripped := stripAccents(strings.TrimSpace(userCity))
	for _, entry := range cityAliases {
		if stripped == entry.alias {
			return entry.canonical
		}
	}
	return stripped
}

var workModelMarkers = []struct {
	marker string
	model  string
}{
	{"hibrido", "hybrid"},
	{"hybrid", "hybrid"},
	{"presencial", "presencial"},
	{"on-site", "presencial"},
	{"onsite", "presencial"},
	{"no escritorio", "presencial"},
	{"home office", "remote"},
	{"trabalho remoto", "remote"},
	{"100% remote", "remote"},
	{"remoto", "remote"},
	{"remota", "remote"},
	{"remotas", "remote"},
	{"remotos", "remote"},
	{"remote", "remote"},
}

// analyzeJobLocation analisa a descrição da vaga para determinar o tipo de localização.
// A detecção usa fronteira de palavra e texto sem acentos para evitar falsos
// positivos como "responsável" -> "sp" ou "informações" -> "for(taleza)".
func analyzeJobLocation(description, _ string) JobLocation {
	desc := stripAccents(description)

	loc := JobLocation{Type: "unknown"}
	for _, m := range workModelMarkers {
		if strings.Contains(desc, m.marker) {
			loc.Type = m.model
			break
		}
	}

	for _, p := range cityPatterns {
		if p.re.MatchString(desc) {
			loc.City = p.canonical
			break
		}
	}

	return loc
}

// locationPenalty calcula os pontos a deduzir do score quando a vaga é
// presencial/híbrida e não é compatível com a cidade do usuário.
func locationPenalty(jobLoc JobLocation, userCity string) float64 {
	locScore := locationScore(jobLoc, userCity)

	if jobLoc.City == "" {
		// Vaga presencial/híbrida sem cidade identificável: penalidade parcial,
		// pois não há como confirmar compatibilidade com a cidade configurada.
		return 12
	}
	if locScore < 50 {
		return (50 - locScore) * 0.8
	}
	return 0
}

// locationScore compara a localização da vaga com a cidade do usuário
func locationScore(jobLoc JobLocation, userCity string) float64 {
	if jobLoc.Type == "remote" || jobLoc.Type == "unknown" {
		return 100
	}

	if jobLoc.City == "" {
		return 50
	}

	userCityLower := normalizeUserCity(userCity)
	jobCityLower := stripAccents(jobLoc.City)

	if userCityLower == jobCityLower {
		return 100
	}

	// Verificar se está no mesmo estado (simplificado)
	stateMap := map[string][]string{
		"são paulo":      {"campinas", "sorocaba", "santos", "rio preto", "ribeirão preto"},
		"rio de janeiro": {"niterói", "petrópolis"},
		"minas gerais":   {"belo horizonte", "uberlandia", "juiz de fora"},
	}

	for state, cities := range stateMap {
		if userCityLower == state || jobCityLower == state {
			for _, city := range cities {
				if userCityLower == city || jobCityLower == city {
					return 80
				}
			}
		}
	}

	return 20
}

func extractKeywordsLMStudio(jobDesc, resumeContent string) []string {
	commonKeywords := []string{
		"python", "javascript", "typescript", "java", "go", "rust", "c++", "c#",
		"react", "vue", "angular", "node", "django", "flask", "spring",
		"docker", "kubernetes", "aws", "azure", "gcp", "linux",
		"sql", "mysql", "postgresql", "mongodb", "redis",
		"git", "ci/cd", "devops", "agile", "scrum",
		"api", "rest", "graphql", "microservices", "mqtt", "kafka",
		"machine learning", "data science", "ai", "deep learning", "tensorflow", "pytorch",
	}

	jobLower := strings.ToLower(jobDesc)
	resumeLower := strings.ToLower(resumeContent)

	var found []string
	for _, kw := range commonKeywords {
		if strings.Contains(jobLower, kw) && strings.Contains(resumeLower, kw) {
			found = append(found, kw)
		}
	}

	return found
}

func extractWords(text string) []string {
	reg := regexp.MustCompile(`\b[a-zA-Z+#.]{2,}\b`)
	words := reg.FindAllString(strings.ToLower(text), -1)
	return unique(words)
}

func unique(words []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0)
	for _, w := range words {
		if !seen[w] {
			seen[w] = true
			result = append(result, w)
		}
	}
	return result
}

func intersection(a, b []string) []string {
	setB := make(map[string]bool)
	for _, w := range b {
		setB[w] = true
	}
	var result []string
	for _, w := range a {
		if setB[w] {
			result = append(result, w)
		}
	}
	return result
}
