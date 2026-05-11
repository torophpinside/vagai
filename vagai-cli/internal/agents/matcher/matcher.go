package matcher

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"regexp"
	"strings"
	"sync"

	"github.com/anomalyco/vagai-cli/internal/agents/lmstudio"
	"github.com/anomalyco/vagai-cli/internal/db"
	"github.com/anomalyco/vagai-cli/internal/models"
)

var useLMStudio = true
var idf map[string]float64

const maxParallelAI = 2

type jobTask struct {
	job    models.Job
	resume models.Resume
}

type matchResult struct {
	jobID         uint
	resumeID      uint
	score         float64
	keywords      []string
	reason        string
	err           error
	alreadyExists bool
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

	var resumes []models.Resume
	db.DB.Find(&resumes)

	if len(resumes) == 0 {
		log.Println("Nenhum currículo encontrado")
		return nil
	}

	var jobs []models.Job
	query := db.DB.Where("status = ?", models.JobStatusNew).Joins("JOIN sites ON sites.id = jobs.site_id").Where("sites.active = ?", true)
	if force {
		query = db.DB.Joins("JOIN sites ON sites.id = jobs.site_id").Where("sites.active = ?", true)
	}
	query.Find(&jobs)

	if len(jobs) == 0 {
		log.Println("Nenhuma vaga nova para processar")
		return nil
	}

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
				result := processTask(task.job, task.resume, threshold)
				if result.err != nil && !result.alreadyExists {
					log.Printf("Aviso no matching job=%d resume=%d: %v", task.job.ID, task.resume.ID, result.err)
				}
				resultsChan <- result
			}
		}(w)
	}

	go func() {
		for _, job := range jobs {
			for _, resume := range resumes {
				jobsChan <- jobTask{job: job, resume: resume}
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
		if result.alreadyExists {
			continue
		}
		if result.score > float64(threshold) {
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
			if err := db.DB.Create(&match).Error; err != nil {
				log.Printf("Erro ao salvar match: %v", err)
			} else {
				matchCount++
				var job models.Job
				db.DB.First(&job, result.jobID)
				db.DB.Model(&job).Update("status", models.JobStatusMatched)
				log.Printf("Match salvo: job=%d resume=%d score=%.2f", result.jobID, result.resumeID, result.score)
			}
		}
	}

	log.Printf("Matcher Agent finalizado. %d matches encontrados", matchCount)
	return nil
}

func processTask(job models.Job, resume models.Resume, threshold int) matchResult {
	result := matchResult{
		jobID:    job.ID,
		resumeID: resume.ID,
	}

	var existing models.Match
	err := db.DB.Where("job_id = ? AND resume_id = ?", job.ID, resume.ID).First(&existing).Error
	if err == nil {
		result.alreadyExists = true
		return result
	}

	score, keywords, reason, err := calculateMatchAI(job.Title, job.Description, resume.Content)
	if err != nil {
		result.err = err
		return result
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

func calculateMatchAI(jobTitle, jobDesc, resumeContent string) (float64, []string, string, error) {
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

	prompt := fmt.Sprintf(`Analise se o currículo é adequado para a vaga. 
VAGA: %s - %s
RESUMO CURRÍCULO: %s
Responda APENAS um objeto JSON no formato: {"score": 0-100, "reason": "sua explicação curta"}`, jobTitle, jobDesc, resumeForPrompt)

	log.Printf("Chamando LM Studio para análise...")
	response, err := lmstudio.Chat(prompt, "Você é um especialista em recruitment tech. Responda sempre em JSON.")
	if err != nil {
		log.Printf("⚠️ Erro ao chamar LM Studio: %v. Usando fallback algorítmico.", err)
		return calculateMatchFallback(jobTitle, jobDesc, resumeContent)
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
		return calculateMatchFallback(jobTitle, jobDesc, resumeContent)
	}

	log.Printf("✅ Análise AI concluída com sucesso. Score: %.2f", aiResp.Score)
	keywords := extractKeywordsLMStudio(jobDesc, resumeContent)

	return aiResp.Score, keywords, aiResp.Reason, nil
}

func calculateMatchFallback(jobTitle, jobDesc, resumeContent string) (float64, []string, string, error) {
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

	loc := locationMatch(jobDesc, resumeContent)

	score := 0.6*textual + 0.3*skillScore + 0.1*loc
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

func locationMatch(text1, text2 string) float64 {
	locations := []string{
		"remote", "remoto", "presencial", "hibrido", "híbrido",
		"são paulo", "sp", "rio de janeiro", "rj",
		"belo horizonte", "bh", "curitiba", "porto alegre",
		"brasil", "brazil", "lisboa", "lisbon",
	}

	t1 := strings.ToLower(text1)
	t2 := strings.ToLower(text2)

	for _, loc := range locations {
		if strings.Contains(t1, loc) && strings.Contains(t2, loc) {
			return 100
		}
	}
	return 0
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
