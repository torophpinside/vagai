package models

import (
	"log"

	"gorm.io/gorm"
)

// BackfillPlans preenche as colunas de gestão de planos (type/active/is_default)
// em bancos que já existiam antes da spec 012. É idempotente: pode rodar a cada
// boot sem efeitos colaterais (FR-012).
func BackfillPlans(db *gorm.DB) {
	// 0. Renomeia a coluna legada max_resumes para max_resume_analyses: o
	// limite deixou de ser "quantidade de currículos" e passou a ser a quota
	// mensal de análises de currículo.
	if db.Migrator().HasColumn(&Plan{}, "max_resumes") {
		if db.Migrator().HasColumn(&Plan{}, "max_resume_analyses") {
			if err := db.Exec("UPDATE plans SET max_resume_analyses = max_resumes").Error; err != nil {
				log.Printf("backfill plans: falha ao copiar max_resumes: %v", err)
			}
		}
		if err := db.Migrator().DropColumn(&Plan{}, "max_resumes"); err != nil {
			log.Printf("backfill plans: falha ao remover max_resumes: %v", err)
		}
	}

	// 1. Tipo derivado do preço. Corrige tanto linhas legadas sem tipo quanto o
	// estado inválido criado pelo AutoMigrate (coluna `type` NOT NULL DEFAULT
	// 'free' preenche linhas existentes com 'free', inclusive planos pagos).
	// Impor preço↔tipo é seguro: é o invariante V2/V3 (FR-012/SC-006).
	if err := db.Exec(`UPDATE plans SET type = CASE
		WHEN price_monthly = 0 AND price_yearly = 0 THEN 'free'
		ELSE 'paid' END
		WHERE type IS NULL OR type = ''
		   OR (type = 'free' AND NOT (price_monthly = 0 AND price_yearly = 0))
		   OR (type = 'paid' AND price_monthly = 0 AND price_yearly = 0)`).Error; err != nil {
		log.Printf("backfill plans: falha ao derivar type: %v", err)
	}

	// 2. Todo plano existente começa ativo.
	db.Exec("UPDATE plans SET active = 1 WHERE active IS NULL")

	// 3. Garante exatamente um plano padrão de cadastro (o gratuito).
	var count int64
	if err := db.Model(&Plan{}).Where("is_default = ?", true).Count(&count).Error; err != nil {
		log.Printf("backfill plans: falha ao contar defaults: %v", err)
		return
	}
	if count == 0 {
		var free Plan
		if err := db.Where("type = ? AND active = ?", PlanTypeFree, true).
			Order("id").First(&free).Error; err == nil {
			db.Model(&free).Update("is_default", true)
			log.Printf("backfill plans: plano %q marcado como padrão de cadastro", free.Slug)
		}
	}
}
