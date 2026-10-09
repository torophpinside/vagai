package analytics

import (
	"fmt"

	"github.com/anomalyco/vagai-cli/internal/db"
	"github.com/spf13/cobra"
)

var jobTrendsPeriod string

var jobTrendsCmd = &cobra.Command{
	Use:   "job-trends",
	Short: "Relatório de tendências de vagas",
	Long: `Distribuição de vagas por status, top organizações por volume,
volume de postagens por período e empresas ativas.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		org, err := mustSetup(cmd)
		if err != nil {
			return err
		}
		printHeader("Job Market Trends", org)

		type statusRow struct {
			Status string
			Cnt    int64
		}
		var statuses []statusRow
		q := db.DB.Table("jobs").Select("status, COUNT(*) AS cnt").Group("status")
		if org > 0 {
			q = q.Where("organization_id = ?", org)
		}
		if err := q.Scan(&statuses).Error; err != nil {
			return err
		}
		var total int64
		for _, s := range statuses {
			total += s.Cnt
		}
		if total == 0 {
			fmt.Println("Nenhuma vaga registrada ainda.")
			fmt.Println("Dica: cadastre fontes e execute o crawler para gerar este relatório.")
			return nil
		}

		fmt.Println("\nVagas por status:")
		for _, s := range statuses {
			fmt.Printf("  %-12s %6d  (%5.1f%%)\n", s.Status, s.Cnt, pct64(s.Cnt, total))
		}

		type orgRow struct {
			Name     string
			JobCount int64
		}
		var orgs []orgRow
		orgQ := db.DB.Table("jobs").
			Select("organizations.name, COUNT(*) AS job_count").
			Joins("JOIN organizations ON organizations.id = jobs.organization_id").
			Group("organizations.name").Order("job_count DESC").Limit(5)
		if org > 0 {
			orgQ = orgQ.Where("jobs.organization_id = ?", org)
		}
		if err := orgQ.Scan(&orgs).Error; err != nil {
			return err
		}
		fmt.Println("\nTop organizações:")
		for _, o := range orgs {
			fmt.Printf("  %-40s %6d vagas\n", o.Name, o.JobCount)
		}

		period := periodOrDefaultCLI(jobTrendsPeriod)
		var src, expr string
		src = "COALESCE(posted_date, collected_at)"
		switch period {
		case "week":
			expr = fmt.Sprintf("DATE_FORMAT(%s, '%%x-W%%v')", src)
		case "quarter":
			expr = fmt.Sprintf("CONCAT(YEAR(%s), '-Q', QUARTER(%s))", src, src)
		case "year":
			expr = fmt.Sprintf("DATE_FORMAT(%s, '%%Y')", src)
		default:
			expr = fmt.Sprintf("DATE_FORMAT(%s, '%%Y-%%m')", src)
		}
		type volRow struct {
			Period string
			Cnt    int64
		}
		var vol []volRow
		volQ := db.DB.Table("jobs").
			Select(fmt.Sprintf("%s AS period, COUNT(*) AS cnt", expr)).
			Where("COALESCE(posted_date, collected_at) IS NOT NULL")
		if org > 0 {
			volQ = volQ.Where("organization_id = ?", org)
		}
		if err := volQ.Group("period").Order("period DESC").Limit(12).Scan(&vol).Error; err != nil {
			return err
		}
		fmt.Printf("\nVolume de postagens (por %s, últimos 12):\n", period)
		for _, v := range vol {
			fmt.Printf("  %-10s %6d\n", v.Period, v.Cnt)
		}

		var active int64
		activeQ := db.DB.Table("jobs").Select("COUNT(DISTINCT NULLIF(TRIM(company), ''))")
		if org > 0 {
			activeQ = activeQ.Where("organization_id = ?", org)
		}
		activeQ.Scan(&active)

		fmt.Printf("\nTotal: %d vagas | Empresas ativas: %d\n", total, active)
		return nil
	},
}

func periodOrDefaultCLI(period string) string {
	switch period {
	case "week", "month", "quarter", "year":
		return period
	default:
		return "month"
	}
}

func init() {
	Cmd.AddCommand(jobTrendsCmd)
	jobTrendsCmd.Flags().StringVarP(&jobTrendsPeriod, "period", "p", "month", "Período: week, month, quarter, year")
}
