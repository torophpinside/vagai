package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/anomalyco/vagai-api/internal/models"
	"github.com/anomalyco/vagai-api/internal/services"
	"github.com/anomalyco/vagai-api/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateSabatinaQuestions cria até duas perguntas situacionais para uma sessão existente.
// POST /api/sabatinas/{id}/questions
func CreateSabatinaQuestions(c *gin.Context) {
	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sessão inválida"})
		return
	}

	var input []struct {
		Text string `json:"text" binding:"required"`
	}
	if err := json.NewDecoder(c.Request.Body).Decode(&input); err != nil || len(input) < 1 || len(input) > 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "envie 1 ou 2 perguntas no body"})
		return
	}

	qs := make([]models.Question, 0, len(input))
	for i, q := range input {
		qs = append(qs, models.Question{
			SessionID: sessionID,
			Text:      utils.SanitizeHTML(q.Text),
			Position:  i + 1,
		})
	}

	if err := services.CreateQuestions(DB, qs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "não foi possível salvar perguntas"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"questions": qs})
}

// GetSabatinaQuestions retorna as perguntas de uma sessão.
// GET /api/sabatinas/{id}/questions
func GetSabatinaQuestions(c *gin.Context) {
	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sessão inválida"})
		return
	}

	qs, err := services.GetQuestionsBySession(DB, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "não foi possível recuperar perguntas"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"questions": qs})
}
