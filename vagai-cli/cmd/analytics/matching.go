package analytics

import (
	"database/sql"
	"fmt"

	"github.com/anomalyco/vagai-cli/internal/db"
	"github.com/spf13/cobra"
)

var matchingCmd = &cobra.Command{
	Use:   "matching-distribution",
	Short: "Relatório de distribuição de scores de matching",
	Long: `Histograma de scores de similaridade (0-1), média/mediana,
taxa de conversão match→candidatura e top títulos por taxa de sucesso.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		org, err := mustSetup(cmd)
		if err != nil {
			return err
		}
		printHeader("Matching Effectiveness", org)

		type bucketRow struct {
			Bucket   string
			Cnt      int64
			AvgScore float64
		}
		var rows []bucketRow
		if err := db.DB.Raw(`
			SELECT
				CASE
					WHEN similarity_score < 20  THEN '0.0-0.2'
					WHEN similarity_score < 40  THEN '0.2-0.4'
					WHEN similarity_score < 60  THEN '0.4-0.6'
					WHEN similarity_score < 80  THEN '0.6-0.8'
					ELSE '0.8-1.0'
				END AS bucket,
				COUNT(*) AS cnt,
				AVG(similarity_score) AS avg_score
			FROM matches
			WHERE organization_id = ?
			GROUP BY bucket`, org).Scan(&rows).Error; err != nil {
			return err
		}

		var total, applied int64
		db.DB.Table("matches").Where("organization_id = ?", org).Count(&total)
		db.DB.Table("matches").Where("organization_id = ? AND applied = ?", org, true).Count(&applied)

		if total == 0 {
			fmt.Println("Nenhum match registrado ainda.")
			fmt.Println("Dica: execute o matching de vagas com seus currículos para gerar este relatório.")
			return nil
		}

		byBucket := map[string]bucketRow{}
		var weighted float64
		for _, r := range rows {
			byBucket[r.Bucket] = r
			weighted += float64(r.Cnt) * r.AvgScore
		}
		fmt.Println("\nDistribuição de scores:")
		for _, name := range []string{"0.0-0.2", "0.2-0.4", "0.4-0.6", "0.6-0.8", "0.8-1.0"} {
			row, ok := byBucket[name]
			if !ok {
				row = bucketRow{Bucket: name}
			}
			fmt.Printf("  %-9s %8d  (%5.1f%%)\n", name, row.Cnt, pct64(row.Cnt, total))
		}

		var med sql.NullFloat64
		db.DB.Raw(`
			SELECT AVG(similarity_score) / 100.0
			FROM (
				SELECT similarity_score,
					ROW_NUMBER() OVER (ORDER BY similarity_score) AS rn,
					COUNT(*) OVER () AS total
				FROM matches WHERE organization_id = ?
			) ranked
			WHERE rn IN (FLOOR((total + 1) / 2), CEIL((total + 1) / 2))`, org).Scan(&med)

		fmt.Printf("\nMédia: %.2f | Mediana: %.2f | Total: %d\n",
			(weighted/float64(total))/100, med.Float64, total)
		fmt.Printf("Conversão match→candidatura: %.1f%% (%d/%d)\n",
			pct64(applied, total), applied, total)

		type titleRow struct {
			Title   string
			Total   int64
			Applied int64
		}
		var titles []titleRow
		if err := db.DB.Raw(`
			SELECT j.title AS title, COUNT(*) AS total, COALESCE(SUM(m.applied), 0) AS applied
			FROM matches m
			JOIN jobs j ON j.id = m.job_id
			WHERE m.organization_id = ?
			GROUP BY j.title
			ORDER BY COALESCE(SUM(m.applied), 0) / COUNT(*) DESC, COUNT(*) DESC
			LIMIT 5`, org).Scan(&titles).Error; err != nil {
			return err
		}
		if len(titles) > 0 {
			fmt.Println("\nTop títulos por taxa de sucesso:")
			for _, t := range titles {
				fmt.Printf("  %-40s %5.1f%% (%d/%d)\n", t.Title, pct64(t.Applied, t.Total), t.Applied, t.Total)
			}
		}
		return nil
	},
}

func pct64(part, total int64) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total) * 100
}

func init() {
	Cmd.AddCommand(matchingCmd)
}
