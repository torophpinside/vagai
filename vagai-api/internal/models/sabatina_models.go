package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SabatinaSession representa uma sessão de sabatina criada automaticamente
// quando o usuário clica no botão de preparação.
type SabatinaSession struct {
	ID         uuid.UUID `gorm:"type:char(36);default:(UUID());primaryKey" json:"id"`
	UserID     uuid.UUID `gorm:"type:char(36);not null;index" json:"user_id"`
	CandidateID uuid.UUID `gorm:"type:char(36);not null;index" json:"candidate_id"`
	Status     string    `gorm:"size:20;default:'in_progress';not null" json:"status"` // in_progress | completed
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// BeforeCreate gera UUID se não fornecido.
func (s *SabatinaSession) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// Question representa uma pergunta situacional vinculada a uma sessão.
type Question struct {
	ID        uuid.UUID `gorm:"type:char(36);default:(UUID());primaryKey" json:"id"`
	SessionID uuid.UUID `gorm:"type:char(36);not null;index" json:"session_id"`
	Text      string    `gorm:"type:text;not null" json:"text"`
	Position  int       `gorm:"not null" json:"position"` // 1 ou 2
	CreatedAt time.Time `json:"created_at"`
}

func (q *Question) BeforeCreate(tx *gorm.DB) error {
	if q.ID == uuid.Nil {
		q.ID = uuid.New()
	}
	return nil
}
