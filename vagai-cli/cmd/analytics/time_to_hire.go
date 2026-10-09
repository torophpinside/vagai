package analytics

import (
	"database/sql"
	"fmt"

	"github.com/anomalyco/vagai-cli/internal/db"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var timeToHireCmd = &cobra.Command{
	Use:   "time-to-hire",
	Short: "Relatório de tempo até a contratação",
	Long: `Tempo médio e mediano (em dias) entre as etapas do pipeline:
coleta→match, match→análise IA e coleta→candidatura, com nº de amostras.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		org, err := mustSetup(cmd)
		if err != nil {
			return err
		}
		printHeader("Time to Hire", org)

		var total int64
		if err := db.DB.Table("matches").
			Where("organization_id = ?", org).
			Count(&total).Error; err != nil {
			return err
		}
		if total == 0 {
			fmt.Println("Nenhum match disponível para calcular o tempo até a contratação.")
			fmt.Println("Dica: execute o matching de vagas com seus currículos para gerar este relatório.")
			return nil
		}

		// T047 — coleta → match.
		match, err := tthStage(db.DB,
			"GREATEST(0, DATEDIFF(matches.analyzed_at, jobs.collected_at))",
			"matches JOIN jobs ON jobs.id = matches.job_id",
			"matches.organization_id = ?", org)
		if err != nil {
			return err
		}

		// T048 — match → análise IA.
		analysis, err := tthStage(db.DB,
			"GREATEST(0, DATEDIFF(ai_analysis_results.created_at, m.analyzed_at))",
			"ai_analysis_results"+
				" JOIN interview_preparations ip ON ip.id = ai_analysis_results.preparation_id"+
				" JOIN matches m ON m.id = ip.match_id",
			"m.organization_id = ?", org)
		if err != nil {
			return err
		}

		// T049 — pipeline completo: coleta → candidatura.
		apply, err := tthStage(db.DB,
			"GREATEST(0, DATEDIFF(matches.applied_at, jobs.collected_at))",
			"matches JOIN jobs ON jobs.id = matches.job_id",
			"matches.applied = TRUE AND matches.applied_at IS NOT NULL AND matches.organization_id = ?",
			org)
		if err != nil {
			return err
		}

		stages := []struct {
			Label string
			Stats tthStageStats
		}{
			{"Coleta → match (time_to_match)", match},
			{"Match → análise IA (time_to_analysis)", analysis},
			{"Coleta → candidatura (time_to_apply)", apply},
		}

		fmt.Println("\nTempo por etapa (dias):")
		for _, s := range stages {
			fmt.Printf("  %-42s média %6.2f | mediana %6.2f | %d amostras\n",
				s.Label, s.Stats.Avg, s.Stats.Median, s.Stats.Samples)
		}
		fmt.Printf("\nTotal de matches avaliados: %d\n", total)
		return nil
	},
}

// tthStageStats é o resultado agregado de uma etapa do time-to-hire.
type tthStageStats struct {
	Avg     float64
	Median  float64
	Samples int64
}

// tthStage calcula média, mediana e amostras (COUNT(*) do JOIN) de uma etapa
// em dias. daysExpr é a expressão GREATEST(0, DATEDIFF(...)), fromSQL o
// FROM/JOIN e whereCond a cláusula WHERE completa (no CLI o org é sempre > 0).
// O relatório imprime os valores com 2 casas (%.2f); sem linhas tudo sai 0.
func tthStage(gdb *gorm.DB, daysExpr, fromSQL, whereCond string, args ...any) (tthStageStats, error) {
	var out tthStageStats
	whereSQL := " WHERE " + whereCond

	// Média + amostras numa única passada.
	type tthAggRow struct {
		AvgDays float64
		Samples int64
	}
	var agg tthAggRow
	aggSQL := fmt.Sprintf(`
		SELECT COALESCE(AVG(%s), 0) AS avg_days, COUNT(*) AS samples
		FROM %s%s`, daysExpr, fromSQL, whereSQL)
	if err := gdb.Raw(aggSQL, args...).Scan(&agg).Error; err != nil {
		return out, err
	}

	// Mediana via window functions (MySQL 8) — mesmo padrão do matching.
	var med sql.NullFloat64
	medSQL := fmt.Sprintf(`
		SELECT COALESCE(AVG(v), 0) AS median
		FROM (
			SELECT %s AS v,
				ROW_NUMBER() OVER (ORDER BY %s) AS rn,
				COUNT(*) OVER () AS total
			FROM %s%s
		) t
		WHERE rn IN (FLOOR((total + 1) / 2), CEIL((total + 1) / 2))`,
		daysExpr, daysExpr, fromSQL, whereSQL)
	if err := gdb.Raw(medSQL, args...).Scan(&med).Error; err != nil {
		return out, err
	}

	out.Avg = agg.AvgDays
	if med.Valid {
		out.Median = med.Float64
	}
	out.Samples = agg.Samples
	return out, nil
}

func init() {
	Cmd.AddCommand(timeToHireCmd)
}
