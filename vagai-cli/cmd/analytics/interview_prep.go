package analytics

import (
	"database/sql"
	"fmt"

	"github.com/anomalyco/vagai-cli/internal/db"
	"github.com/spf13/cobra"
)

var interviewPrepCmd = &cobra.Command{
	Use:   "interview-prep",
	Short: "Relatório de análise das preparações de entrevista",
	Long: `Distribuição de categorias das perguntas, distribuição de fontes
das preparações (ai/template), nota média de verificação e tendência
de evolução das notas.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		user, err := mustSetup(cmd)
		if err != nil {
			return err
		}
		printHeader("Interview Preparation Analysis", user)

		type catRow struct {
			Category string
			Cnt      int64
		}
		var cats []catRow
		if err := db.DB.Raw(`
			SELECT category, COUNT(*) AS cnt
			FROM interview_questions
			WHERE user_id = ?
			GROUP BY category`, user).Scan(&cats).Error; err != nil {
			return err
		}

		type srcRow struct {
			Source string
			Cnt    int64
		}
		var srcs []srcRow
		if err := db.DB.Raw(`
			SELECT source, COUNT(*) AS cnt
			FROM interview_preparations
			WHERE user_id = ?
			GROUP BY source`, user).Scan(&srcs).Error; err != nil {
			return err
		}

		var total int64
		if err := db.DB.Table("interview_preparations").
			Where("user_id = ?", user).
			Count(&total).Error; err != nil {
			return err
		}

		var avgScore sql.NullFloat64
		if err := db.DB.Raw(`
			SELECT AVG(score) AS avg_score
			FROM preparation_verification_histories
			WHERE user_id = ?`, user).Scan(&avgScore).Error; err != nil {
			return err
		}

		type scoreRow struct {
			Score float64
		}
		const maxTrendRows = 500
		var scoreRows []scoreRow
		if err := db.DB.Raw(`
			SELECT score
			FROM preparation_verification_histories
			WHERE user_id = ?
			ORDER BY verified_at ASC
			LIMIT ?`, user, maxTrendRows).Scan(&scoreRows).Error; err != nil {
			return err
		}

		if total == 0 && len(cats) == 0 && len(scoreRows) == 0 {
			fmt.Println("Nenhuma preparação de entrevista disponível.")
			fmt.Println("Dica: gere uma preparação de entrevista e conclua uma verificação para gerar este relatório.")
			return nil
		}

		fmt.Printf("\nTotal de preparações: %d\n", total)

		fmt.Println("\nCategorias das perguntas:")
		if len(cats) == 0 {
			fmt.Println("  (sem dados)")
		}
		var questionTotal int64
		for _, c := range cats {
			questionTotal += c.Cnt
		}
		for _, c := range cats {
			fmt.Printf("  %-20s %6d  (%5.1f%%)\n", c.Category, c.Cnt, pct64(c.Cnt, questionTotal))
		}

		fmt.Println("\nFontes das preparações:")
		if len(srcs) == 0 {
			fmt.Println("  (sem dados)")
		}
		for _, s := range srcs {
			fmt.Printf("  %-20s %6d  (%5.1f%%)\n", s.Source, s.Cnt, pct64(s.Cnt, total))
		}

		fmt.Printf("\nNota média de verificação: %.1f\n", avgScore.Float64)

		// Tendência: metade mais nova vs. metade mais antiga (±5%), mesmas
		// regras do serviço — o CLI não importa código da API.
		trend := "estável"
		if len(scoreRows) >= 2 {
			avg := func(values []float64) float64 {
				if len(values) == 0 {
					return 0
				}
				var sum float64
				for _, v := range values {
					sum += v
				}
				return sum / float64(len(values))
			}
			mid := len(scoreRows) / 2
			older := make([]float64, 0, mid)
			newer := make([]float64, 0, len(scoreRows)-mid)
			for i, r := range scoreRows {
				if i < mid {
					older = append(older, r.Score)
				} else {
					newer = append(newer, r.Score)
				}
			}
			o, n := avg(older), avg(newer)
			switch {
			case o > 0 && n >= o*1.05:
				trend = "em melhoria"
			case o > 0 && n <= o*0.95:
				trend = "em declínio"
			}
		}
		fmt.Printf("Tendência: %s\n", trend)
		return nil
	},
}

func init() {
	Cmd.AddCommand(interviewPrepCmd)
}
