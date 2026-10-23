package services

import (
	"github.com/anomalyco/vagai-api/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreateQuestions persiste uma lista de perguntas vinculadas à sessão.
func CreateQuestions(db *gorm.DB, qs []models.Question) error {
	return db.Create(&qs).Error
}

// GetQuestionsBySession busca todas as perguntas de uma sessão, ordenadas por posição.
func GetQuestionsBySession(db *gorm.DB, sessionID uuid.UUID) ([]models.Question, error) {
	var qs []models.Question
	err := db.Where("session_id = ?", sessionID).Order("position ASC").Find(&qs).Error
	return qs, err
}
