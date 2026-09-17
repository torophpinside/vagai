package models

import (
	"time"
)

type AIAnalysisStatus string

const (
	AIAnalysisStatusPending   AIAnalysisStatus = "pending"
	AIAnalysisStatusCompleted AIAnalysisStatus = "completed"
	AIAnalysisStatusError     AIAnalysisStatus = "error"
)

type AIAnalysisResult struct {
	ID             uint               `gorm:"primaryKey" json:"id"`
	OrganizationID uint               `gorm:"index;not null" json:"organization_id"`
	PreparationID  uint               `gorm:"not null;index" json:"preparation_id"`
	QuestionID     uint               `gorm:"not null;index" json:"question_id"`
	AnalysisText   string             `gorm:"type:text" json:"analysis_text"`
	AnalysisScore  float64            `gorm:"type:float" json:"analysis_score"`
	Status         AIAnalysisStatus   `gorm:"size:20;not null;index" json:"status"`
	Source         VerificationSource `gorm:"size:20" json:"source"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
}
