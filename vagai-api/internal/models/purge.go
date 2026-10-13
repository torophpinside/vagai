package models

import (
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// PurgeDeletedUsers expurga contas cuja janela de reativação já venceu
// (FR-007, research.md R3): anonimiza dados pessoais, cascateia o conteúdo
// por user_id e remove a linha. Registros financeiros (Stripe) e AuditLog
// não são tocados. Retorna o número de contas expurgadas.
func PurgeDeletedUsers(db *gorm.DB, retention time.Duration) int {
	deadline := time.Now().Add(-retention)

	var due []User
	if err := db.Unscoped().
		Where("deleted_at IS NOT NULL AND purge_after IS NOT NULL AND purge_after < ?", deadline).
		Find(&due).Error; err != nil {
		log.Printf("purge: falha ao buscar contas expiradas: %v", err)
		return 0
	}

	purged := 0
	for i := range due {
		u := due[i]
		if err := purgeOne(db, &u); err != nil {
			log.Printf("purge: falha ao expurgar usuário %d: %v", u.ID, err)
			continue
		}
		purged++
	}
	if purged > 0 {
		log.Printf("purge: %d conta(s) expurgada(s)", purged)
	}
	return purged
}

func purgeOne(db *gorm.DB, u *User) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// Conteúdo pessoal por usuário (ordem respeita dependências por ids).
		for _, model := range []interface{}{
			&InterviewPreparation{}, &InterviewQuestion{}, &PracticeEntry{},
			&PreparationVerification{}, &PreparationVerificationHistory{},
			&AIAnalysisResult{}, &ResumeAnalysis{}, &Match{}, &Resume{},
			&Job{}, &Site{}, &AgentLog{}, &ApiKey{},
		} {
			if err := tx.Where("user_id = ?", u.ID).Unscoped().Delete(model).Error; err != nil {
				return fmt.Errorf("cascata %T: %w", model, err)
			}
		}

		// Anonimização dos dados pessoais antes da remoção (spec FR-007).
		anonEmail := fmt.Sprintf("removido+%d@eliminado.local", u.ID)
		if err := tx.Unscoped().Model(&User{}).Where("id = ?", u.ID).Updates(map[string]interface{}{
			"name":              "removido",
			"email":             anonEmail,
			"avatar":            "",
			"city":              "",
			"timezone":          "",
			"password_hash":     "removido",
			"stripe_customer_id": "",
			"negative_keywords": "[]",
		}).Error; err != nil {
			return fmt.Errorf("anonimizar: %w", err)
		}

		if err := tx.Unscoped().Delete(&User{}, u.ID).Error; err != nil {
			return fmt.Errorf("remover usuário: %w", err)
		}
		return nil
	})
}

// StartPurgeJob dispara o expurgo imediatamente e depois em ticker diário
// (research.md R3). Contexto desanexado: não cancela com o request HTTP.
func StartPurgeJob(db *gorm.DB, retention time.Duration) {
	go func() {
		PurgeDeletedUsers(db, retention)
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			PurgeDeletedUsers(db, retention)
		}
	}()
}
