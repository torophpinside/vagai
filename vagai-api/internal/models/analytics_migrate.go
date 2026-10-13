package models

import (
	"fmt"
	"log"
	"strings"

	"gorm.io/gorm"
)

// MigrateAnalytics cria os índices críticos para as queries de analytics
// (data-model.md → Index Recommendations). Idempotente: índices já existentes
// (por nome ou por mesma sequência de colunas) são ignorados.
func MigrateAnalytics(db *gorm.DB) {
	ensureIndex(db, "matches", "idx_match_user", "user_id")
	ensureIndex(db, "resumes", "idx_resume_user", "user_id")
	ensureIndex(db, "ai_analysis_results", "idx_ai_user_prep", "user_id", "preparation_id")
	ensureIndex(db, "interview_preparations", "idx_interview_user", "user_id")
	ensureIndex(db, "preparation_verification_histories", "idx_verification_user", "user_id")
}

// ensureIndex cria o índice se ainda não existir um índice com o mesmo nome
// ou com a mesma sequência de colunas (evita duplicar índices do AutoMigrate).
func ensureIndex(db *gorm.DB, table, index string, cols ...string) {
	if indexExists(db, table, index) || indexExistsForColumns(db, table, cols...) {
		return
	}

	stmt := fmt.Sprintf("CREATE INDEX `%s` ON `%s` (%s)",
		index, table, "`"+strings.Join(cols, "`, `")+"`")
	if err := db.Exec(stmt).Error; err != nil {
		log.Printf("analytics migrate: falha ao criar índice %s em %s: %v", index, table, err)
	}
}

func indexExists(db *gorm.DB, table, index string) bool {
	var count int64
	db.Raw(`SELECT COUNT(*) FROM information_schema.statistics
		WHERE table_schema = DATABASE() AND table_name = ? AND index_name = ?`,
		table, index).Scan(&count)
	return count > 0
}

// indexExistsForColumns verifica se algum índice da tabela cobre exatamente a
// mesma sequência de colunas (comparação por nome+ordem das colunas).
func indexExistsForColumns(db *gorm.DB, table string, cols ...string) bool {
	type idxCol struct {
		IndexName  string
		ColumnName string
		SeqInIndex int
	}
	var rows []idxCol
	db.Raw(`SELECT index_name, column_name, seq_in_index
		FROM information_schema.statistics
		WHERE table_schema = DATABASE() AND table_name = ?
		ORDER BY index_name, seq_in_index`, table).Scan(&rows)

	byIndex := map[string][]string{}
	for _, r := range rows {
		byIndex[r.IndexName] = append(byIndex[r.IndexName], r.ColumnName)
	}
	for _, ic := range byIndex {
		if len(ic) == len(cols) {
			match := true
			for i := range cols {
				if ic[i] != cols[i] {
					match = false
					break
				}
			}
			if match {
				return true
			}
		}
	}
	return false
}
