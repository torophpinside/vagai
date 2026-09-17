package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/anomalyco/vagai-api/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Sentinels da feature — mapeados para códigos HTTP no handler.
var (
	ErrMatchNotFound         = errors.New("match não encontrado")
	ErrMatchNotApplied       = errors.New("vaga não candidatada")
	ErrPrepNotFound          = errors.New("preparação não encontrada")
	ErrQuestionNotFound      = errors.New("pergunta não encontrada")
	ErrInvalidQuestionStatus = errors.New("status inválido")
	ErrInvalidSelfAssessment = errors.New("autoavaliação inválida")
)

// PrepIncompleteError indica que não há nada para corrigir ainda (nenhuma
// resposta), carregando o nº de questões que faltam responder.
type PrepIncompleteError struct {
	Remaining int
}

func (e *PrepIncompleteError) Error() string {
	return "preparação sem respostas — responda ao menos uma questão"
}

// AIQuestion é o contrato do JSON gerado pela IA, espelhando
// services.TemplateQuestion (mesmo schema nos dois caminhos).
type AIQuestion struct {
	Category    string `json:"category"`
	Topic       string `json:"topic"`
	Text        string `json:"text"`
	AnswerGuide string `json:"answer_guide"`
}

const (
	minPrepQuestions       = 7
	maxPrepQuestions       = 19
	maxAITechQuestions     = techQuestionsPerTech * maxDetectedTechs // 12
	prepQuestionCategories = 3
	// TargetQuestionCount é o total exato de perguntas de toda preparação
	// (006-fixed-count, FR-001): gerar menos suplementa, gerar mais trunca.
	TargetQuestionCount      = 15
	verificationTimeout      = 60 * time.Second
	verificationMaxRetries   = 1
	selfHealRunningThreshold = 5 * time.Minute
)

// currentLMStudioURL resolve a URL da IA a cada chamada (lazy), permitindo
// que testes forcem fallback via env var após o init do pacote.
func currentLMStudioURL() string {
	if url := os.Getenv("LMSTUDIO_URL"); url != "" {
		return url
	}
	return baseURL
}

// generateInterviewQuestionsAI pede à IA perguntas estruturadas e valida o
// contrato (3 categorias, campos preenchidos, limites). Qualquer falha é
// enviada ao chamador, que cai no fallback determinístico (template).
func generateInterviewQuestionsAI(jobTitle, company, description string, techs []string) ([]AIQuestion, error) {
	prompt := fmt.Sprintf(`Gere uma preparação para entrevista técnica da vaga abaixo.

CARGO: %s
EMPRESA: %s
DESCRIÇÃO DA VAGA:
%s

Tecnologias detectadas na descrição: %s

Retorne APENAS JSON válido, SEM texto adicional, no formato:
{
  "questions": [
    {
      "category": "technology | foundations | architecture",
      "topic": "tópico curto (ex.: Go, SOLID, Design de sistemas)",
      "text": "pergunta para o candidato",
      "answer_guide": "resposta-guia com pontos-chave"
    }
  ]
}

Regras:
- Categorias: technology (2 perguntas por tecnologia detectada, máx. %d), foundations (princípios como SOLID/DRY/KISS, 4), architecture (design de sistemas, escalabilidade, cache, 3).
- Total entre %d e %d perguntas.
- Sem duplicar o texto de uma pergunta.
- Sem texto fora do JSON.`,
		jobTitle, company, description, strings.Join(techs, ", "),
		maxAITechQuestions, minPrepQuestions, maxPrepQuestions)

	messages := []Message{
		{Role: "system", Content: "Você é um assistente de IA especializado em RH e recrutamento técnico. Retorne apenas JSON válido."},
		{Role: "user", Content: prompt},
	}

	body, err := json.Marshal(ChatRequest{
		Model:       "local-model",
		Messages:    messages,
		Temperature: 0.3,
	})
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 240 * time.Second}
	req, err := http.NewRequest("POST", currentLMStudioURL()+"/v1/chat/completions", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("LM Studio indisponível: %w", err)
	}
	defer resp.Body.Close()

	var result ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, fmt.Errorf("erro AI: %v", result.Error)
	}
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("sem resposta da AI")
	}

	responseText := strings.TrimPrefix(result.Choices[0].Message.Content, "```json")
	responseText = strings.TrimPrefix(responseText, "```")
	responseText = strings.TrimSuffix(responseText, "```")
	responseText = strings.TrimSpace(responseText)

	var payload struct {
		Questions []AIQuestion `json:"questions"`
	}
	if err := json.Unmarshal([]byte(responseText), &payload); err != nil {
		log.Printf("Erro ao parsear JSON da AI (prep): %v", err)
		return nil, err
	}

	if !validAIQuestionSet(payload.Questions) {
		return nil, fmt.Errorf("conjunto de perguntas da IA inválido")
	}
	return payload.Questions, nil
}

// validAIQuestionSet garante o contrato único da API para o caminho IA:
// 3 categorias presentes, campos preenchidos e sem duplicação. O limite
// superior foi removido (006-fixed-count, FR-003): conjuntos maiores que
// TargetQuestionCount são truncados por EnforceQuestionCount em vez de
// descartados para o fallback.
func validAIQuestionSet(qs []AIQuestion) bool {
	if len(qs) < minPrepQuestions {
		return false
	}
	if qs == nil {
		return false
	}
	cats := map[string]bool{}
	seenText := map[string]bool{}
	for _, q := range qs {
		switch q.Category {
		case "technology", "foundations", "architecture":
			cats[q.Category] = true
		default:
			return false
		}
		if strings.TrimSpace(q.Topic) == "" || strings.TrimSpace(q.Text) == "" {
			return false
		}
		if seenText[q.Text] {
			return false
		}
		seenText[q.Text] = true
	}
	return cats["technology"] && cats["foundations"] && cats["architecture"]
}

// progressOf calcula o resumo de progresso de uma preparação.
func progressOf(db *gorm.DB, orgID, prepID uint) (total, practiced, mastered int64, err error) {
	if err = db.Model(&models.InterviewQuestion{}).
		Where("organization_id = ? AND preparation_id = ?", orgID, prepID).
		Count(&total).Error; err != nil {
		return
	}
	if err = db.Model(&models.InterviewQuestion{}).
		Where("organization_id = ? AND preparation_id = ? AND status = ?", orgID, prepID, models.QuestionStatusPracticed).
		Count(&practiced).Error; err != nil {
		return
	}
	err = db.Model(&models.InterviewQuestion{}).
		Where("organization_id = ? AND preparation_id = ? AND status = ?", orgID, prepID, models.QuestionStatusMastered).
		Count(&mastered).Error
	return
}

// CreateOrRegeneratePreparation (upsert) gera e persiste o conjunto de
// perguntas de uma vaga candidatada, numa única transação. O Match é carregado
// por id + organization_id (G3). Fallback determinístico quando a IA falha.
// O conjunto final tem sempre exatamente TargetQuestionCount perguntas
// (006-fixed-count). A seed opcional (variádica, primeiro valor) embaralha o
// resultado de forma reproduzível (FR-005).
func CreateOrRegeneratePreparation(db *gorm.DB, orgID, matchID uint, seeds ...string) (*models.InterviewPreparation, []models.InterviewQuestion, error) {
	var match models.Match
	if err := db.Where("id = ? AND organization_id = ?", matchID, orgID).First(&match).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrMatchNotFound
		}
		return nil, nil, err
	}
	if !match.Applied {
		return nil, nil, ErrMatchNotApplied
	}

	var job models.Job
	if err := db.Where("id = ? AND organization_id = ?", match.JobID, orgID).First(&job).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrMatchNotFound
		}
		return nil, nil, err
	}

	// Geração: IA com fallback determinístico (constituição IV).
	techs := DetectTechnologies(job.Description)
	source := models.SourceTemplate
	questions := templateQuestionsToModels(BuildTemplateQuestions(techs))

	if aiQs, aiErr := generateInterviewQuestionsAI(job.Title, job.Company, job.Description, techs); aiErr == nil {
		source = models.SourceAI
		questions = aiQuestionsToModels(aiQs)
	}

	// 006-fixed-count: garante exatamente TargetQuestionCount perguntas nos
	// dois caminhos (FR-001/FR-002/FR-003) + ordenação reproduzível (FR-005).
	seed := ""
	if len(seeds) > 0 {
		seed = seeds[0]
	}
	before := len(questions)
	questions = EnforceQuestionCount(questions, techs, seed)
	log.Printf("CreateOrRegeneratePreparation (match=%d, source=%s): %d -> %d perguntas (seed=%q)",
		matchID, source, before, len(questions), seed)

	var prep *models.InterviewPreparation
	err := db.Transaction(func(tx *gorm.DB) error {
		var p models.InterviewPreparation
		if err := tx.Where("match_id = ? AND organization_id = ?", matchID, orgID).First(&p).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			p = models.InterviewPreparation{
				OrganizationID: orgID,
				MatchID:        matchID,
				JobID:          job.ID,
			}
		}

		// Regeneração: substitui perguntas, respostas, verificação e feedbacks
		// anteriores (FR-008) — nada de conjunto antigo sobrevive.
		if p.ID != 0 {
			var qids []uint
			if err := tx.Model(&models.InterviewQuestion{}).
				Where("organization_id = ? AND preparation_id = ?", orgID, p.ID).
				Pluck("id", &qids).Error; err != nil {
				return err
			}
			if len(qids) > 0 {
				if err := tx.Where("organization_id = ? AND question_id IN ?", orgID, qids).
					Delete(&models.PracticeEntry{}).Error; err != nil {
					return err
				}
			}
			if err := tx.Where("organization_id = ? AND preparation_id = ?", orgID, p.ID).
				Delete(&models.InterviewQuestion{}).Error; err != nil {
				return err
			}
			if err := tx.Where("organization_id = ? AND preparation_id = ?", orgID, p.ID).
				Delete(&models.PreparationVerification{}).Error; err != nil {
				return err
			}
			if err := tx.Where("organization_id = ? AND preparation_id = ?", orgID, p.ID).
				Delete(&models.AIAnalysisResult{}).Error; err != nil {
				return err
			}
			if err := tx.Where("organization_id = ? AND preparation_id = ?", orgID, p.ID).
				Delete(&models.PreparationVerificationHistory{}).Error; err != nil {
				return err
			}
		}

		p.MatchID = matchID
		p.JobID = job.ID
		p.Title = job.Title
		p.Company = job.Company
		p.Description = job.Description
		p.TechStack = mustJSON(techs)
		p.Source = source
		p.GeneratedAt = time.Now()

		if err := tx.Save(&p).Error; err != nil {
			return err
		}

		for i, q := range questions {
			q.OrganizationID = orgID
			q.PreparationID = p.ID
			q.Position = i + 1
			if err := tx.Create(&q).Error; err != nil {
				return err
			}
		}

		prep = &p
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	// Recarrega as perguntas com posições determinísticas + progresso.
	return GetPreparation(db, orgID, prep.ID)
}

// ListPreparations lista as preparações do tenant, com filtro opcional por match.
func ListPreparations(db *gorm.DB, orgID uint, matchID *uint) ([]models.InterviewPreparation, error) {
	query := db.Where("organization_id = ?", orgID)
	if matchID != nil && *matchID > 0 {
		query = query.Where("match_id = ?", *matchID)
	}

	var preps []models.InterviewPreparation
	if err := query.Order("generated_at DESC").Find(&preps).Error; err != nil {
		return nil, err
	}
	return preps, nil
}

// GetPreparation carrega uma preparação do tenant com todas as perguntas
// ordenadas pela sequência da sabatina.
func GetPreparation(db *gorm.DB, orgID, prepID uint) (*models.InterviewPreparation, []models.InterviewQuestion, error) {
	var prep models.InterviewPreparation
	if err := db.Where("id = ? AND organization_id = ?", prepID, orgID).First(&prep).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrPrepNotFound
		}
		return nil, nil, err
	}

	var questions []models.InterviewQuestion
	if err := db.Where("organization_id = ? AND preparation_id = ?", orgID, prep.ID).
		Order("position ASC").
		Find(&questions).Error; err != nil {
		return nil, nil, err
	}
	return &prep, questions, nil
}

// UpdateQuestionStatus atualiza o status de autoavaliação de uma pergunta,
// validando o enum e o escopo de tenant (preparação → pergunta).
func UpdateQuestionStatus(db *gorm.DB, orgID, prepID, questionID uint, status models.InterviewQuestionStatus) (*models.InterviewQuestion, error) {
	switch status {
	case models.QuestionStatusPending, models.QuestionStatusPracticed, models.QuestionStatusMastered:
	default:
		return nil, ErrInvalidQuestionStatus
	}

	var prep models.InterviewPreparation
	if err := db.Where("id = ? AND organization_id = ?", prepID, orgID).First(&prep).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPrepNotFound
		}
		return nil, err
	}

	var q models.InterviewQuestion
	if err := db.Where("id = ? AND organization_id = ? AND preparation_id = ?", questionID, orgID, prepID).
		First(&q).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrQuestionNotFound
		}
		return nil, err
	}

	q.Status = status
	if err := db.Save(&q).Error; err != nil {
		return nil, err
	}
	return &q, nil
}

// validateSelfAssessment valida a autoavaliação informada. Aceita "" (rascunho
// sem avaliação) e "low"/"medium"/"high". Valores não-vazios inválidos retornam
// ErrInvalidSelfAssessment.
func validateSelfAssessment(sa string) error {
	if sa == "" {
		return nil
	}
	switch sa {
	case "low", "medium", "high":
		return nil
	default:
		return ErrInvalidSelfAssessment
	}
}

// SaveQuestionAnswer persiste (upsert) a resposta escrita + autoavaliação da
// pergunta. Sobrescreve a resposta anterior (rascunho mais recente).
// Quando selfAssessment é vazio, a autoavaliação já existente é preservada
// (FR-003: mantém a mais recente informada pelo candidato).
// timeSpentSecs opcional acumula o tempo de resposta (modo simulado).
func SaveQuestionAnswer(db *gorm.DB, orgID, prepID, questionID uint, answer string, selfAssessment string, timeSpentSecs ...int) (*models.PracticeEntry, error) {
	if err := validateSelfAssessment(selfAssessment); err != nil {
		return nil, err
	}

	var prep models.InterviewPreparation
	if err := db.Where("id = ? AND organization_id = ?", prepID, orgID).First(&prep).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPrepNotFound
		}
		return nil, err
	}

	var q models.InterviewQuestion
	if err := db.Where("id = ? AND organization_id = ? AND preparation_id = ?", questionID, orgID, prepID).
		First(&q).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrQuestionNotFound
		}
		return nil, err
	}

	var entry models.PracticeEntry
	err := db.Where("question_id = ? AND organization_id = ?", questionID, orgID).
		First(&entry).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Estado pré-save para decidir outdated/auto-disparo: só há mudança
	// relevante se o texto da resposta mudou ou se este save completa a prep.
	oldAnswer := entry.Answer
	answerChanged := strings.TrimSpace(answer) != strings.TrimSpace(oldAnswer)
	_, _, remainingBefore, cerr := PreparationCompletion(db, orgID, prepID)
	if cerr != nil {
		return nil, cerr
	}

	entry.OrganizationID = orgID
	entry.QuestionID = questionID
	entry.Answer = answer
	// FR-003: auto-save preserva a autoavaliação já existente quando não
	// fornecida; apenas sobrescreve quando explicitamente informada.
	if selfAssessment != "" {
		entry.SelfAssessment = selfAssessment
	}
	// Modo simulado: acumula o tempo gasto respondendo (em segundos).
	if len(timeSpentSecs) > 0 && timeSpentSecs[0] > 0 {
		entry.TimeSpentSeconds += timeSpentSecs[0]
	}

	if err := db.Save(&entry).Error; err != nil {
		return nil, err
	}

	// US2/FR-007: editar o TEXTO da resposta após 'verified' sinaliza o
	// resultado antigo como desatualizado (imediatamente). Save idêntico
	// (autosave ao navegar) preserva o 'verified' — retomar a sabatina não
	// re-corrige o já corrigido.
	if answerChanged {
		ver, verr := GetPreparationVerification(db, orgID, prepID)
		if verr != nil {
			return nil, verr
		}
		if ver != nil && ver.Status == models.VerificationVerified {
			if err := db.Model(&models.PreparationVerification{}).
				Where("id = ? AND organization_id = ? AND status = ?", ver.ID, orgID, models.VerificationVerified).
				Updates(map[string]interface{}{"status": models.VerificationOutdated}).Error; err != nil {
				return nil, err
			}
		}
	}

	// US1/FR-001: dispara a verificação assíncrona automaticamente SOMENTE na
	// transição para completa (última questão respondida). Saves sem mudança
	// de completude — ex.: navegar com autosave — nunca re-agendam correção.
	_, _, remainingAfter, cerr := PreparationCompletion(db, orgID, prepID)
	if cerr != nil {
		return nil, cerr
	}
	if remainingBefore > 0 && remainingAfter == 0 {
		_, _ = VerifyPreparation(db, orgID, prepID)
	}

	return &entry, nil
}

func mustJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "[]"
	}
	return string(b)
}

func templateQuestionsToModels(in []TemplateQuestion) []models.InterviewQuestion {
	out := make([]models.InterviewQuestion, 0, len(in))
	for _, q := range in {
		out = append(out, models.InterviewQuestion{
			Category:    models.InterviewCategory(q.Category),
			Topic:       q.Topic,
			Text:        q.Text,
			AnswerGuide: q.AnswerGuide,
			Status:      models.QuestionStatusPending,
		})
	}
	return out
}

func aiQuestionsToModels(in []AIQuestion) []models.InterviewQuestion {
	out := make([]models.InterviewQuestion, 0, len(in))
	for _, q := range in {
		out = append(out, models.InterviewQuestion{
			Category:    models.InterviewCategory(q.Category),
			Topic:       q.Topic,
			Text:        q.Text,
			AnswerGuide: q.AnswerGuide,
			Status:      models.QuestionStatusPending,
		})
	}
	return out
}

// ShuffleQuestionsWithSeed embaralha as perguntas de forma determinística
// (006-fixed-count, FR-005): a mesma seed produz sempre a mesma ordem.
// Seed vazia retorna uma cópia sem embaralhar.
func ShuffleQuestionsWithSeed(qs []models.InterviewQuestion, seed string) []models.InterviewQuestion {
	out := make([]models.InterviewQuestion, len(qs))
	copy(out, qs)
	if seed == "" || len(out) < 2 {
		return out
	}
	h := fnv.New64a()
	_, _ = h.Write([]byte(seed))
	r := rand.New(rand.NewSource(int64(h.Sum64())))
	r.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}

// EnforceQuestionCount garante exatamente TargetQuestionCount perguntas
// (006-fixed-count, FR-001/FR-002/FR-003): trunca o excedente preservando a
// ordem determinística, ou suplementa do banco de templates quando faltam.
// Quando seed é informada, o conjunto final é embaralhado de forma
// reproduzível (FR-005).
func EnforceQuestionCount(qs []models.InterviewQuestion, techs []string, seed string) []models.InterviewQuestion {
	if len(qs) > TargetQuestionCount {
		log.Printf("EnforceQuestionCount: truncando %d perguntas para %d", len(qs), TargetQuestionCount)
		qs = qs[:TargetQuestionCount]
	} else if len(qs) < TargetQuestionCount {
		existing := make([]TemplateQuestion, 0, len(qs))
		for _, q := range qs {
			existing = append(existing, TemplateQuestion{
				Category:    string(q.Category),
				Topic:       q.Topic,
				Text:        q.Text,
				AnswerGuide: q.AnswerGuide,
			})
		}
		extra := SupplementTemplateQuestions(existing, techs, TargetQuestionCount)
		log.Printf("EnforceQuestionCount: suplementando %d perguntas com %d do banco de templates", len(qs), len(extra))
		qs = append(qs, templateQuestionsToModels(extra)...)
	}
	return ShuffleQuestionsWithSeed(qs, seed)
}

// ---------------------------------------------------------------------------
// Verificação das respostas com IA + fallback determinístico (003-feature).
// ---------------------------------------------------------------------------

// AIVerification é o contrato do JSON retornado pela IA na verificação.
type AIVerification struct {
	Score    float64 `json:"score"`
	Feedback string  `json:"feedback"`
}

// verificationQA é a unidade avaliada: pergunta + resposta do candidato + guia.
type verificationQA struct {
	Question   models.InterviewQuestion
	Answer     string
	Guide      string
	Assessment string
}

// claimAllows define os estados que podem ser reivindicados por uma nova
// verificação. Inclui 'verified' para permitir re-verificação explícita
// (FR-007 / contracts/api.md).
func claimAllows(status models.PreparationVerificationStatus) bool {
	switch status {
	case models.VerificationPending, models.VerificationOutdated, models.VerificationVerified:
		return true
	}
	return false
}

// isStaleRunning detecta uma verificação 'running' órfã (processo morto /
// tela fechada), candidata ao self-heal. Como o orçamento de IA é ≤120s e o
// limiar é 5 min, nunca compete com uma tentativa viva.
func isStaleRunning(v *models.PreparationVerification, maxAge time.Duration, now time.Time) bool {
	return v != nil && v.Status == models.VerificationRunning && v.StartedAt != nil &&
		now.Sub(*v.StartedAt) > maxAge
}

// PreparationCompletion conta as questões da preparação e quantas possuem
// resposta escrita não-vazia (FR-008). answered = questões com resposta;
// remaining = total - answered.
func PreparationCompletion(db *gorm.DB, orgID, prepID uint) (total, answered, remaining int, err error) {
	var totalCount int64
	if err = db.Model(&models.InterviewQuestion{}).
		Where("organization_id = ? AND preparation_id = ?", orgID, prepID).
		Count(&totalCount).Error; err != nil {
		return
	}
	total = int(totalCount)

	var answeredCount int64
	if err = db.Model(&models.InterviewQuestion{}).
		Joins("JOIN practice_entries pe ON pe.question_id = interview_questions.id AND pe.organization_id = ?", orgID).
		Where("interview_questions.organization_id = ? AND interview_questions.preparation_id = ? AND TRIM(pe.answer) <> ''", orgID, prepID).
		Count(&answeredCount).Error; err != nil {
		return
	}
	answered = int(answeredCount)
	remaining = total - answered
	if remaining < 0 {
		remaining = 0
	}
	return
}

// GetPreparationVerification carrega o resultado atual (1:1) da verificação,
// dentro do escopo do tenant. Retorna (nil, nil) se nunca houve verificação.
func GetPreparationVerification(db *gorm.DB, orgID, prepID uint) (*models.PreparationVerification, error) {
	var v models.PreparationVerification
	if err := db.Where("organization_id = ? AND preparation_id = ?", orgID, prepID).
		First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

// PreparationDetail agrega tudo o que os handlers precisam exibir: a
// preparação, as perguntas ordenadas, as respostas escritas (por pergunta) e o
// resultado da verificação.
type PreparationDetail struct {
	Preparation  *models.InterviewPreparation
	Questions    []models.InterviewQuestion
	Entries      map[uint]models.PracticeEntry
	Verification *models.PreparationVerification
	// Analyses é o feedback individual da correção por IA, por pergunta.
	Analyses map[uint]models.AIAnalysisResult
	// History são as tentativas de verificação concluídas (evolução da nota).
	History []models.PreparationVerificationHistory
}

// GetPreparationDetail carrega o detalhe completo de uma preparação. Executa o
// self-heal na leitura (T022): verificação 'running' órfã retoma sem bloquear.
func GetPreparationDetail(db *gorm.DB, orgID, prepID uint) (*PreparationDetail, error) {
	prep, questions, err := GetPreparation(db, orgID, prepID)
	if err != nil {
		return nil, err
	}

	detail := &PreparationDetail{
		Preparation: prep,
		Questions:   questions,
		Entries:     map[uint]models.PracticeEntry{},
		Analyses:    map[uint]models.AIAnalysisResult{},
	}

	if len(questions) > 0 {
		ids := make([]uint, 0, len(questions))
		for _, q := range questions {
			ids = append(ids, q.ID)
		}
		var entries []models.PracticeEntry
		if err := db.Where("organization_id = ? AND question_id IN ?", orgID, ids).
			Find(&entries).Error; err != nil {
			return nil, err
		}
		for _, e := range entries {
			detail.Entries[e.QuestionID] = e
		}
	}

	// Feedback individual da correção por IA, por pergunta respondida.
	var analyses []models.AIAnalysisResult
	if err := db.Where("organization_id = ? AND preparation_id = ?", orgID, prepID).
		Find(&analyses).Error; err != nil {
		return nil, err
	}
	for _, a := range analyses {
		detail.Analyses[a.QuestionID] = a
	}

	var history []models.PreparationVerificationHistory
	if err := db.Where("organization_id = ? AND preparation_id = ?", orgID, prepID).
		Order("verified_at DESC").Limit(10).
		Find(&history).Error; err != nil {
		return nil, err
	}
	detail.History = history

	ver, err := GetPreparationVerification(db, orgID, prepID)
	if err != nil {
		return nil, err
	}
	if isStaleRunning(ver, selfHealRunningThreshold, time.Now()) {
		if _, err := VerifyPreparation(db, orgID, prepID); err == nil {
			ver, err = GetPreparationVerification(db, orgID, prepID)
			if err != nil {
				return nil, err
			}
		}
	}
	detail.Verification = ver

	return detail, nil
}

// VerifyPreparation inicia (ou re-inicia) a verificação assíncrona da
// preparação. Permite correção parcial: basta 1+ resposta para agendar (as
// respondidas são corrigidas; o progresso segue informando as faltantes).
// Retorna 0, nil quando agendada (ou já em andamento) e um
// *PrepIncompleteError quando nada há para corrigir. O processamento real roda
// em goroutine (IA → retry → fallback; orçamento total ≤120s).
func VerifyPreparation(db *gorm.DB, orgID, prepID uint) (int, error) {
	var prep models.InterviewPreparation
	if err := db.Where("id = ? AND organization_id = ?", prepID, orgID).First(&prep).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, ErrPrepNotFound
		}
		return 0, err
	}

	_, answered, remaining, err := PreparationCompletion(db, orgID, prepID)
	if err != nil {
		return 0, err
	}
	if answered == 0 {
		return remaining, &PrepIncompleteError{Remaining: remaining}
	}

	now := time.Now()

	// Garante a linha 1:1 (única por preparação) com insert atômico — se duas
	// solicitações concorrentes (auto-trigger + POST /verify) tentarem criar,
	// o conflito no uniqueIndex é ignorado e recarregamos a linha existente.
	ver := &models.PreparationVerification{
		OrganizationID: orgID,
		PreparationID:  prepID,
		Status:         models.VerificationPending,
	}
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(ver).Error; err != nil {
		return 0, err
	}
	if err := db.Where("organization_id = ? AND preparation_id = ?", orgID, prepID).
		First(ver).Error; err != nil {
		return 0, err
	}

	// Self-heal antes do claim: 'running' órfão antigo volta a 'pending'.
	if isStaleRunning(ver, selfHealRunningThreshold, now) {
		if err := db.Model(&models.PreparationVerification{}).
			Where("id = ? AND organization_id = ? AND status = ?", ver.ID, orgID, models.VerificationRunning).
			Updates(map[string]interface{}{"status": models.VerificationPending, "started_at": nil}).Error; err != nil {
			return 0, err
		}
		ver.Status = models.VerificationPending
	}

	// Claim otimista: pending|outdated|verified → running.
	res := db.Model(&models.PreparationVerification{}).
		Where("id = ? AND organization_id = ? AND status IN ?", ver.ID, orgID, []models.PreparationVerificationStatus{
			models.VerificationPending, models.VerificationOutdated, models.VerificationVerified,
		}).
		Updates(map[string]interface{}{"status": models.VerificationRunning, "started_at": now})
	if res.Error != nil {
		return 0, res.Error
	}
	if res.RowsAffected == 0 {
		// Já existe verificação em andamento.
		return 0, nil
	}

	go runVerificationPipeline(db, orgID, prepID, ver.ID)
	return 0, nil
}

// runVerificationPipeline executa a verificação em background: carrega as
// respostas, avalia (IA → retry → fallback) e persiste o resultado atomically.
// Dispara em paralelo a correção por pergunta (feedback individual por questão).
func runVerificationPipeline(db *gorm.DB, orgID, prepID, verID uint) {
	qas, err := loadVerificationInput(db, orgID, prepID)
	if err != nil {
		log.Printf("Verificação (prep=%d): %v", prepID, err)
		return
	}

	// Correção por IA de cada resposta (feedback individual abaixo da resposta).
	// Roda em paralelo à verificação geral e substitui o feedback anterior.
	analysisQs := make([]models.InterviewQuestion, 0, len(qas))
	for _, qa := range qas {
		analysisQs = append(analysisQs, qa.Question)
	}
	NewAIAnalysisService(db).StartAsyncAnalysis(context.Background(), prepID, analysisQs)

	score, feedback, source, verr := runVerification(qas)
	if verr != nil {
		log.Printf("Verificação (prep=%d) falhou: %v", prepID, verr)
		return
	}

	now := time.Now()
	db.Model(&models.PreparationVerification{}).
		Where("id = ? AND organization_id = ? AND status = ?", verID, orgID, models.VerificationRunning).
		Updates(map[string]interface{}{
			"status":      models.VerificationVerified,
			"score":       score,
			"feedback":    feedback,
			"source":      source,
			"verified_at": now,
			"started_at":  nil,
		})

	// Histórico de evolução: anexa a tentativa concluída (imutável).
	if herr := db.Create(&models.PreparationVerificationHistory{
		OrganizationID: orgID,
		PreparationID:  prepID,
		Score:          score,
		Feedback:       feedback,
		Source:         source,
		VerifiedAt:     now,
	}).Error; herr != nil {
		log.Printf("Verificação (prep=%d): falha ao anexar histórico: %v", prepID, herr)
	}
}

// loadVerificationInput monta a lista pergunta+resposta+guia para avaliar.
func loadVerificationInput(db *gorm.DB, orgID, prepID uint) ([]verificationQA, error) {
	var questions []models.InterviewQuestion
	if err := db.Where("organization_id = ? AND preparation_id = ?", orgID, prepID).
		Order("position ASC").Find(&questions).Error; err != nil {
		return nil, err
	}
	if len(questions) == 0 {
		return nil, errors.New("preparação sem perguntas para verificar")
	}

	ids := make([]uint, 0, len(questions))
	for _, q := range questions {
		ids = append(ids, q.ID)
	}
	var entries []models.PracticeEntry
	if err := db.Where("organization_id = ? AND question_id IN ?", orgID, ids).
		Find(&entries).Error; err != nil {
		return nil, err
	}

	byQ := make(map[uint]models.PracticeEntry, len(entries))
	for _, e := range entries {
		byQ[e.QuestionID] = e
	}

	var qas []verificationQA
	for _, q := range questions {
		e, ok := byQ[q.ID]
		if !ok || strings.TrimSpace(e.Answer) == "" {
			continue
		}
		qas = append(qas, verificationQA{
			Question:   q,
			Answer:     e.Answer,
			Guide:      q.AnswerGuide,
			Assessment: e.SelfAssessment,
		})
	}
	if len(qas) == 0 {
		return nil, errors.New("preparação sem respostas para verificar")
	}
	return qas, nil
}

// runVerification tenta a IA (1 tentativa + 1 retry, cada uma ≤60s) e cai no
// fallback determinístico se todas as tentativas falharem (constituição IV).
func runVerification(qas []verificationQA) (float64, string, models.VerificationSource, error) {
	for attempt := 0; attempt < verificationMaxRetries+1; attempt++ {
		v, aerr := verifyPreparationAI(qas)
		if aerr == nil {
			return v.Score, v.Feedback, models.VerificationSourceAI, nil
		}
		log.Printf("Verificação via IA (tentativa %d) falhou: %v", attempt+1, aerr)
	}
	fb, ferr := verifyPreparationFallback(qas)
	if ferr != nil {
		return 0, "", "", ferr
	}
	return fb.Score, fb.Feedback, models.VerificationSourceTemplate, nil
}

// verifyPreparationAI pede à IA a nota (0–10) e o feedback e valida o contrato.
func verifyPreparationAI(qas []verificationQA) (AIVerification, error) {
	var b strings.Builder
	b.WriteString("Avalie as respostas do candidato em uma preparação para entrevista técnica.\n\n")
	for i, qa := range qas {
		fmt.Fprintf(&b, "%d. Pergunta: %s\nResposta do candidato: %s\nGuia esperado: %s\nAutoavaliação do candidato: %s\n\n",
			i+1, qa.Question.Text, qa.Answer, qa.Guide, qa.Assessment)
	}
	b.WriteString(`Retorne APENAS JSON válido, sem texto adicional, no formato:
{"score": 7.5, "feedback": "Ponto forte: ...; Sugestão: ..."}

Regras:
- score: 0.0 a 10.0, com 1 casa decimal.
- feedback: obrigatório, com pelo menos um ponto forte e uma sugestão de melhoria.`)

	messages := []Message{
		{Role: "system", Content: "Você é um avaliador técnico. Retorne apenas JSON válido."},
		{Role: "user", Content: b.String()},
	}

	body, err := json.Marshal(ChatRequest{
		Model:       "local-model",
		Messages:    messages,
		Temperature: 0.2,
	})
	if err != nil {
		return AIVerification{}, err
	}

	client := &http.Client{Timeout: verificationTimeout}
	req, err := http.NewRequest("POST", currentLMStudioURL()+"/v1/chat/completions", bytes.NewBuffer(body))
	if err != nil {
		return AIVerification{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return AIVerification{}, fmt.Errorf("LM Studio indisponível: %w", err)
	}
	defer resp.Body.Close()

	var result ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return AIVerification{}, err
	}
	if result.Error != nil {
		return AIVerification{}, fmt.Errorf("erro AI: %v", result.Error)
	}
	if len(result.Choices) == 0 {
		return AIVerification{}, errors.New("sem resposta da AI")
	}

	responseText := strings.TrimPrefix(result.Choices[0].Message.Content, "```json")
	responseText = strings.TrimPrefix(responseText, "```")
	responseText = strings.TrimSuffix(responseText, "```")
	responseText = strings.TrimSpace(responseText)

	var v AIVerification
	if err := json.Unmarshal([]byte(responseText), &v); err != nil {
		return AIVerification{}, err
	}
	if !validVerificationPayload(v) {
		return AIVerification{}, errors.New("payload de verificação da IA inválido")
	}
	return v, nil
}

// validVerificationPayload valida o contrato mínimo do JSON da IA.
func validVerificationPayload(v AIVerification) bool {
	if v.Score < 0 || v.Score > 10 {
		return false
	}
	return len(strings.TrimSpace(v.Feedback)) >= 20
}

// verifyPreparationFallback avalia por sobreposição léxica entre resposta e
// guia (normalizeText/tokenSet), ponderada pela autoavaliação. O feedback é um
// template determinístico com ponto forte e sugestão (source=template).
func verifyPreparationFallback(qas []verificationQA) (AIVerification, error) {
	if len(qas) == 0 {
		return AIVerification{}, errors.New("sem respostas para avaliar")
	}

	type scoredQA struct {
		qa    verificationQA
		score float64
	}

	scored := make([]scoredQA, 0, len(qas))
	var sum float64
	for _, qa := range qas {
		s := fallbackQuestionScore(qa)
		sum += s
		scored = append(scored, scoredQA{qa: qa, score: s})
	}
	avg := clamp0to10(math.Round(sum/float64(len(scored))*10) / 10)

	strong, weak := scored[0], scored[0]
	for _, s := range scored {
		if s.score > strong.score {
			strong = s
		}
		if s.score < weak.score {
			weak = s
		}
	}

	feedback := fmt.Sprintf(
		"Ponto forte: %s (%s). Sugestão: %s — revise os pontos-chave do guia de resposta e pratique mais.",
		strong.qa.Question.Topic, strong.qa.Question.Text, weak.qa.Question.Topic)

	return AIVerification{Score: avg, Feedback: feedback}, nil
}

// fallbackQuestionScore calcula o score de uma única resposta no fallback.
func fallbackQuestionScore(qa verificationQA) float64 {
	guide := tokenSet(qa.Guide)
	ans := tokenSet(qa.Answer)

	if len(guide) == 0 {
		switch qa.Assessment {
		case "low":
			return 4.0
		case "high":
			return 8.0
		default:
			return 6.0
		}
	}
	if len(ans) == 0 {
		return 0.5
	}

	hits := 0
	for tok := range guide {
		if ans[tok] {
			hits++
		}
	}
	coverage := float64(hits) / float64(len(guide))
	if coverage > 1 {
		coverage = 1
	}
	switch qa.Assessment {
	case "low":
		coverage *= 0.85
	case "high":
		coverage *= 1.05
	}
	if coverage > 1 {
		coverage = 1
	}
	return clamp0to10(math.Round(coverage*10*10) / 10)
}

func clamp0to10(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 10 {
		return 10
	}
	return x
}
