package models

import (
	"time"
)

// PreparationVerificationHistory guarda cada verificação concluída de uma
// preparação (evolução da nota ao longo das tentativas). Linhas imutáveis:
// nova verificação anexa, nunca atualiza. Limpo na regeneração da preparação.
type PreparationVerificationHistory struct {
	ID             uint               `gorm:"primaryKey" json:"id"`
	OrganizationID uint               `gorm:"index;not null" json:"organization_id"`
	PreparationID  uint               `gorm:"index;not null" json:"preparation_id"`
	Score          float64            `gorm:"type:decimal(4,1)" json:"score"`
	Feedback       string             `gorm:"type:text" json:"-"`
	Source         VerificationSource `gorm:"size:20" json:"source"`
	VerifiedAt     time.Time          `gorm:"index" json:"verified_at"`
}
