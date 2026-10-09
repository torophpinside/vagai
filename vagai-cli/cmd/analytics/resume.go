package analytics

import (
	"fmt"

	"github.com/anomalyco/vagai-cli/internal/db"
	"github.com/spf13/cobra"
)

var resumeInsightsCmd = &cobra.Command{
	Use:   "resume-insights",
	Short: "Relatório de qualidade de currículos",
	Long: `Forças, fraquezas e sugestões mais frequentes das análises de
currículos, além da frequência de skills nos currículos parseados.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		org, err := mustSetup(cmd)
		if err != nil {
			return err
		}
		printHeader("Resume Quality Insights", org)

		type itemCount struct {
			Item string
			Cnt  int64
		}
		freq := func(column string, limit int) ([]itemCount, error) {
			q := fmt.Sprintf(`
				SELECT jt.item AS item, COUNT(*) AS cnt
				FROM resume_analyses t,
					JSON_TABLE(t.%s, '$[*]' COLUMNS (item VARCHAR(500) PATH '$')) jt
				WHERE t.organization_id = ? AND JSON_VALID(t.%s)
				GROUP BY jt.item ORDER BY cnt DESC LIMIT %d`, column, column, limit)
			var rows []itemCount
			err := db.DB.Raw(q, org).Scan(&rows).Error
			return rows, err
		}

		var total int64
		db.DB.Table("resume_analyses").Where("organization_id = ?", org).Count(&total)
		if total == 0 {
			fmt.Println("Nenhuma análise de currículo disponível.")
			fmt.Println("Dica: analise currículos (resumes/analyze) para gerar este relatório.")
			return nil
		}

		for _, section := range []struct {
			title, column string
		}{
			{"Top forças", "strengths"},
			{"Top fraquezas", "weaknesses"},
			{"Top sugestões", "suggestions"},
		} {
			rows, err := freq(section.column, 10)
			if err != nil {
				return err
			}
			fmt.Printf("\n%s:\n", section.title)
			if len(rows) == 0 {
				fmt.Println("  (sem dados)")
				continue
			}
			for _, r := range rows {
				fmt.Printf("  %-50s %d\n", r.Item, r.Cnt)
			}
		}

		var skills []itemCount
		if err := db.DB.Raw(`
			SELECT LOWER(jt.item) AS item, COUNT(*) AS cnt
			FROM resumes r,
				JSON_TABLE(r.data, '$.skills[*]' COLUMNS (item VARCHAR(200) PATH '$')) jt
			WHERE r.organization_id = ? AND JSON_VALID(r.data)
			GROUP BY LOWER(jt.item) ORDER BY cnt DESC LIMIT 20`, org).Scan(&skills).Error; err != nil {
			return err
		}
		fmt.Println("\nSkills mais frequentes:")
		if len(skills) == 0 {
			fmt.Println("  (sem dados)")
		}
		for _, s := range skills {
			fmt.Printf("  %-30s %d\n", s.Item, s.Cnt)
		}

		fmt.Printf("\nTotal de análises: %d\n", total)
		return nil
	},
}

func init() {
	Cmd.AddCommand(resumeInsightsCmd)
}
