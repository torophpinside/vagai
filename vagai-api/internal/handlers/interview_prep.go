package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/anomalyco/vagai-api/internal/models"
	"github.com/anomalyco/vagai-api/internal/services"
	"github.com/gin-gonic/gin"
	"log"
)

type interviewPrepProgressDTO struct {
	Total     int `json:"total"`
	Practiced int `json:"practiced"`
	Mastered  int `json:"mastered"`
	Answered  int `json:"answered"`
	Remaining int `json:"remaining"`
}

type interviewPrepEntryDTO struct {
	ID               uint      `json:"id"`
	QuestionID       uint      `json:"question_id"`
	Answer           string    `json:"answer"`
	SelfAssessment   string    `json:"self_assessment"`
	TimeSpentSeconds int       `json:"time_spent_seconds"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type interviewPrepVerificationDTO struct {
	Status     string     `json:"status"`
	Score      float64    `json:"score"`
	Feedback   string     `json:"feedback"`
	Source     string     `json:"source"`
	VerifiedAt *time.Time `json:"verified_at"`
}

// interviewPrepHistoryDTO é uma tentativa concluída (evolução da nota).
type interviewPrepHistoryDTO struct {
	Score      float64   `json:"score"`
	Source     string    `json:"source"`
	VerifiedAt time.Time `json:"verified_at"`
}

type interviewPrepQuestionDTO struct {
	ID          uint   `json:"id"`
	Category    string `json:"category"`
	Topic       string `json:"topic"`
	Text        string `json:"text"`
	AnswerGuide string `json:"answer_guide"`
	Status      string `json:"status"`
	Position    int    `json:"position"`
}

// interviewPrepAnalysisDTO é o feedback individual da correção por IA de uma
// resposta, exibido abaixo do quadro da resposta de cada pergunta.
type interviewPrepAnalysisDTO struct {
	ID         uint    `json:"id"`
	QuestionID uint    `json:"question_id"`
	Text       string  `json:"text"`
	Score      float64 `json:"score"`
	Status     string  `json:"status"`
	Source     string  `json:"source"`
}

type interviewPrepDTO struct {
	ID             uint                              `json:"id"`
	OrganizationID uint                              `json:"organization_id"`
	MatchID        uint                              `json:"match_id"`
	JobID          uint                              `json:"job_id"`
	Title          string                            `json:"title"`
	Company        string                            `json:"company"`
	Description    string                            `json:"description"`
	TechStack      []string                          `json:"tech_stack"`
	Source         string                            `json:"source"`
	GeneratedAt    time.Time                         `json:"generated_at"`
	Questions      []interviewPrepQuestionDTO        `json:"questions,omitempty"`
	Entries        map[uint]interviewPrepEntryDTO    `json:"entries"`
	Progress       interviewPrepProgressDTO          `json:"progress"`
	Verification   *interviewPrepVerificationDTO     `json:"verification,omitempty"`
	Analyses       map[uint]interviewPrepAnalysisDTO `json:"analyses"`
	History        []interviewPrepHistoryDTO         `json:"history"`
}

func newInterviewPrepDTO(detail *services.PreparationDetail) interviewPrepDTO {
	prep := detail.Preparation
	questions := detail.Questions

	dto := interviewPrepDTO{
		ID:             prep.ID,
		OrganizationID: prep.OrganizationID,
		Title:          prep.Title,
		Company:        prep.Company,
		Description:    prep.Description,
		Source:         string(prep.Source),
		GeneratedAt:    prep.GeneratedAt,
		TechStack:      []string{},
	}
	if prep.MatchID != nil {
		dto.MatchID = *prep.MatchID
	}
	if prep.JobID != nil {
		dto.JobID = *prep.JobID
	}
	_ = json.Unmarshal([]byte(prep.TechStack), &dto.TechStack)
	if dto.TechStack == nil {
		dto.TechStack = []string{}
	}

	total, practiced, mastered := progressCounts(questions)
	answered := 0
	for _, e := range detail.Entries {
		if strings.TrimSpace(e.Answer) != "" {
			answered++
		}
	}
	dto.Progress = interviewPrepProgressDTO{
		Total:     total,
		Practiced: practiced,
		Mastered:  mastered,
		Answered:  answered,
		Remaining: total - answered,
	}

	dto.Entries = make(map[uint]interviewPrepEntryDTO, len(detail.Entries))
	for qid, e := range detail.Entries {
		dto.Entries[qid] = interviewPrepEntryDTO{
			ID:               e.ID,
			QuestionID:       e.QuestionID,
			Answer:           e.Answer,
			SelfAssessment:   e.SelfAssessment,
			TimeSpentSeconds: e.TimeSpentSeconds,
			UpdatedAt:        e.UpdatedAt,
		}
	}

	dto.Questions = make([]interviewPrepQuestionDTO, 0, len(questions))
	for _, q := range questions {
		dto.Questions = append(dto.Questions, interviewPrepQuestionDTO{
			ID:          q.ID,
			Category:    string(q.Category),
			Topic:       q.Topic,
			Text:        q.Text,
			AnswerGuide: q.AnswerGuide,
			Status:      string(q.Status),
			Position:    q.Position,
		})
	}

	if detail.Verification != nil {
		v := newVerificationDTO(detail.Verification)
		dto.Verification = &v
	}

	dto.History = make([]interviewPrepHistoryDTO, 0, len(detail.History))
	for _, h := range detail.History {
		dto.History = append(dto.History, interviewPrepHistoryDTO{
			Score:      h.Score,
			Source:     string(h.Source),
			VerifiedAt: h.VerifiedAt,
		})
	}

	dto.Analyses = make(map[uint]interviewPrepAnalysisDTO, len(detail.Analyses))
	for qid, a := range detail.Analyses {
		dto.Analyses[qid] = interviewPrepAnalysisDTO{
			ID:         a.ID,
			QuestionID: a.QuestionID,
			Text:       a.AnalysisText,
			Score:      a.AnalysisScore,
			Status:     string(a.Status),
			Source:     string(a.Source),
		}
	}

	return dto
}

func newVerificationDTO(v *models.PreparationVerification) interviewPrepVerificationDTO {
	return interviewPrepVerificationDTO{
		Status:     string(v.Status),
		Score:      v.Score,
		Feedback:   v.Feedback,
		Source:     string(v.Source),
		VerifiedAt: v.VerifiedAt,
	}
}

func progressCounts(questions []models.InterviewQuestion) (total, practiced, mastered int) {
	total = len(questions)
	for _, q := range questions {
		switch q.Status {
		case models.QuestionStatusPracticed:
			practiced++
		case models.QuestionStatusMastered:
			mastered++
		}
	}
	return
}

func CreateInterviewPrep(c *gin.Context) {
	db := getDB(c)
	orgID := c.GetUint("org_id")

	var req struct {
		MatchID  uint   `json:"match_id"`
		Random   string `json:"random"`
		AtRandom string `json:"@random"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.MatchID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
		return
	}

	// 006-fixed-count (FR-005): seed opcional para ordenação reproduzível.
	seed := req.AtRandom
	if seed == "" {
		seed = req.Random
	}

	prep, _, err := services.CreateOrRegeneratePreparation(db, orgID, req.MatchID, seed)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrMatchNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Match não encontrado"})
		case errors.Is(err, services.ErrMatchNotApplied):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Vaga não candidatada"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao gerar preparação"})
		}
		return
	}

	detail, derr := services.GetPreparationDetail(db, orgID, prep.ID)
	if derr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao carregar preparação"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"preparation": newInterviewPrepDTO(detail)})
}

// CreateSpontaneousPreparation gera e PERSISTE uma preparação avulsa
// (InterviewPreparation + 15 perguntas) a partir do conteúdo bruto de uma vaga,
// no escopo da organização autenticada. Requer JWT e org_id válidos; a
// preparação criada é retornada com id real e nenhuma estrutura paralela é
// montada (reusa newInterviewPrepDTO).
func CreateSpontaneousPreparation(c *gin.Context) {
	db := getDB(c)
	orgID := c.GetUint("org_id")
	if orgID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Organização inválida"})
		return
	}

	var req struct {
		Title       string `json:"title"`
		Company     string `json:"company"`
		Description string `json:"description"`
		Random      string `json:"random"`
		AtRandom    string `json:"@random"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
		return
	}

	// Seed: random tem precedência; @random é alias quando random está ausente.
	seed := req.Random
	if seed == "" {
		seed = req.AtRandom
	}

	prep, err := services.CreatePersistedSpontaneousPreparation(db, orgID, req.Title, req.Company, req.Description, seed)
	if err != nil {
		if errors.Is(err, services.ErrSpontaneousDescriptionRequired) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Conteúdo da vaga é obrigatório"})
			return
		}
		log.Printf("CreateSpontaneousPreparation (org=%d) error: %v", orgID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao gerar preparação"})
		return
	}

	detail, derr := services.GetPreparationDetail(db, orgID, prep.ID)
	if derr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao carregar preparação"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"preparation": newInterviewPrepDTO(detail)})
}

func ListInterviewPreps(c *gin.Context) {
	db := getDB(c)
	orgID := c.GetUint("org_id")

	var matchID *uint
	if v := c.Query("match_id"); v != "" {
		id, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "match_id inválido"})
			return
		}
		mid := uint(id)
		matchID = &mid
	}

	preps, err := services.ListPreparations(db, orgID, matchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar preparações"})
		return
	}

	data := make([]interviewPrepDTO, 0, len(preps))
	for i := range preps {
		detail, err := services.GetPreparationDetail(db, orgID, preps[i].ID)
		if err != nil {
			continue
		}
		data = append(data, newInterviewPrepDTO(detail))
	}

	c.JSON(http.StatusOK, gin.H{"data": data, "total": len(data)})
}

func GetInterviewPrep(c *gin.Context) {
	db := getDB(c)
	orgID := c.GetUint("org_id")

	prepID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || prepID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	detail, err := services.GetPreparationDetail(db, orgID, uint(prepID))
	if err != nil {
		if errors.Is(err, services.ErrPrepNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Preparação não encontrada"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar preparação"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"preparation": newInterviewPrepDTO(detail)})
}

// VerifyInterviewPrep inicia (ou re-inicia) a verificação assíncrona das
// respostas da preparação. Retorna o estado da verificação; a nota/feedback
// chegam no GET enquanto roda em background.
func VerifyInterviewPrep(c *gin.Context) {
	db := getDB(c)
	orgID := c.GetUint("org_id")

	prepID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || prepID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	_, err = services.VerifyPreparation(db, orgID, uint(prepID))
	if err != nil {
		var inc *services.PrepIncompleteError
		switch {
		case errors.As(err, &inc):
			c.JSON(http.StatusBadRequest, gin.H{
				"error":     "Preparação sem respostas — responda ao menos uma questão",
				"remaining": inc.Remaining,
			})
		case errors.Is(err, services.ErrPrepNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Preparação não encontrada"})
		default:
			log.Printf("VerifyPrep (prep=%d) error: %v", prepID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao verificar preparação"})
		}
		return
	}

	ver, err := services.GetPreparationVerification(db, orgID, uint(prepID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao verificar preparação"})
		return
	}

	response := gin.H{"preparation_id": uint(prepID)}
	if ver != nil {
		response["verification"] = newVerificationDTO(ver)
	}
	c.JSON(http.StatusOK, response)
}

func UpdateQuestionStatus(c *gin.Context) {
	db := getDB(c)
	orgID := c.GetUint("org_id")

	prepID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || prepID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	questionID, err := strconv.ParseUint(c.Param("questionId"), 10, 32)
	if err != nil || questionID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID da pergunta inválido"})
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil && req.Status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
		return
	}

	q, err := services.UpdateQuestionStatus(db, orgID, uint(prepID), uint(questionID), models.InterviewQuestionStatus(req.Status))
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidQuestionStatus):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Status inválido"})
		case errors.Is(err, services.ErrPrepNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Preparação não encontrada"})
		case errors.Is(err, services.ErrQuestionNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Pergunta não encontrada"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar status"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"question": interviewPrepQuestionDTO{
		ID:          q.ID,
		Category:    string(q.Category),
		Topic:       q.Topic,
		Text:        q.Text,
		AnswerGuide: q.AnswerGuide,
		Status:      string(q.Status),
		Position:    q.Position,
	}})
}

func SaveQuestionAnswer(c *gin.Context) {
	db := getDB(c)
	orgID := c.GetUint("org_id")

	prepID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || prepID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	questionID, err := strconv.ParseUint(c.Param("questionId"), 10, 32)
	if err != nil || questionID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID da pergunta inválido"})
		return
	}

	var req struct {
		Answer           string `json:"answer"`
		SelfAssessment   string `json:"self_assessment"`
		TimeSpentSeconds int    `json:"time_spent_seconds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
		return
	}

	entry, err := services.SaveQuestionAnswer(db, orgID, uint(prepID), uint(questionID), req.Answer, req.SelfAssessment, req.TimeSpentSeconds)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidSelfAssessment):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Autoavaliação inválida"})
		case errors.Is(err, services.ErrPrepNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Preparação não encontrada"})
		case errors.Is(err, services.ErrQuestionNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Pergunta não encontrada"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar resposta"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"practice_entry": gin.H{
		"id":                 entry.ID,
		"question_id":        entry.QuestionID,
		"answer":             entry.Answer,
		"self_assessment":    entry.SelfAssessment,
		"time_spent_seconds": entry.TimeSpentSeconds,
	}})
}
