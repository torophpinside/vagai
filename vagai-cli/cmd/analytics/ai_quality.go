package analytics

import (
	"fmt"

	"github.com/anomalyco/vagai-cli/internal/db"
	"github.com/spf13/cobra"
)

var aiQualityCmd = &cobra.Command{
	Use:   "ai-quality",
	Short: "Relatório de qualidade de análises de IA",
	Long: `Insights de qualidade das análises de IA: distribuição de scores,
taxa de sucesso por fonte e tempo médio de processamento.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		org, err := mustSetup(cmd)
		if err != nil {
			return err
		}
		printHeader("AI Analysis Quality", org)

		type bucketRow struct {
			Bucket string
			Cnt    int64
		}
		var buckets []bucketRow
		qb := db.DB.Table("ai_analysis_results").Select(`
			CASE
				WHEN analysis_score < 2.0 THEN '0-2'
				WHEN analysis_score < 4.0 THEN '2-4'
				WHEN analysis_score < 6.0 THEN '4-6'
				WHEN analysis_score < 8.0 THEN '6-8'
				ELSE '8-10'
			END AS bucket,
			COUNT(*) AS cnt
		`)
		if org > 0 {
			qb = qb.Where("organization_id = ?", org)
		}
		qb = qb.Group("bucket")
		if err := qb.Scan(&buckets).Error; err != nil {
			return err
		}

		var total int64
		qt := db.DB.Table("ai_analysis_results")
		if org > 0 {
			qt = qt.Where("organization_id = ?", org)
		}
		if err := qt.Count(&total).Error; err != nil {
			return err
		}
		if total == 0 {
			fmt.Println("Nenhuma análise de IA disponível.")
			fmt.Println("Dica: execute análises de IA para gerar este relatório.")
			return nil
		}

		fmt.Printf("\nTotal de análises: %d\n", total)

		// Histograma
		order := []string{"0-2", "2-4", "4-6", "6-8", "8-10"}
		m := make(map[string]int64, len(buckets))
		for _, b := range buckets {
			m[b.Bucket] = b.Cnt
		}
		fmt.Println("\nDistribuição de scores:")
		for _, k := range order {
			c := m[k]
			fmt.Printf("  %-6s %8d  (%5.1f%%)\n", k, c, pct64(c, total))
		}

		// Média
		var avg float64
		qa := db.DB.Table("ai_analysis_results").Select("AVG(analysis_score) AS avg")
		if org > 0 {
			qa = qa.Where("organization_id = ?", org)
		}
		if err := qa.Scan(&avg).Error; err != nil {
			return err
		}
		if avg > 10 {
			avg = 10
		}
		if avg < 0 {
			avg = 0
		}
		fmt.Printf("\nPontuação média: %.2f (0-10)\n", aiqRound2(avg))

		// Status
		type st struct {
			Status string
			Cnt    int64
		}
		var sts []st
		qs := db.DB.Table("ai_analysis_results").Select("status, COUNT(*) AS cnt")
		if org > 0 {
			qs = qs.Where("organization_id = ?", org)
		}
		qs = qs.Group("status")
		if err := qs.Scan(&sts).Error; err != nil {
			return err
		}
		fmt.Println("\nStatus das análises:")
		for _, s := range sts {
			fmt.Printf("  %-12s %8d\n", s.Status, s.Cnt)
		}

		// Por fonte
		type src struct {
			Source    string
			Completed int64
			Error     int64
			Pending   int64
		}
		var srcs []src
		qsrc := db.DB.Table("ai_analysis_results").Select(`
			source,
			SUM(CASE WHEN status='completed' THEN 1 ELSE 0 END) AS completed,
			SUM(CASE WHEN status='error' THEN 1 ELSE 0 END) AS error,
			SUM(CASE WHEN status='pending' THEN 1 ELSE 0 END) AS pending
		`)
		if org > 0 {
			qsrc = qsrc.Where("organization_id = ?", org)
		}
		qsrc = qsrc.Group("source")
		if err := qsrc.Scan(&srcs).Error; err != nil {
			return err
		}
		if len(srcs) > 0 {
			fmt.Println("\nTaxa de sucesso por fonte:")
			for _, s := range srcs {
				den := s.Completed + s.Error
				sr := 0.0
				if den > 0 {
					sr = float64(s.Completed) / float64(den) * 100
				}
				fmt.Printf("  %-12s compl=%-4d err=%-4d pend=%-4d  succ=%.2f%%\n", s.Source, s.Completed, s.Error, s.Pending, aiqRound2(sr))
			}
		}

		// Processamento
		var proc float64
		qp := db.DB.Table("ai_analysis_results").Select("AVG(UNIX_TIMESTAMP(updated_at) - UNIX_TIMESTAMP(created_at)) AS proc")
		if org > 0 {
			qp = qp.Where("organization_id = ?", org)
		}
		qp = qp.Where("status IN ('completed','error')")
		if err := qp.Scan(&proc).Error; err != nil {
			return err
		}
		fmt.Printf("\nTempo médio de processamento: %.1fs (concluídos/erros)\n", aiqRound1(proc))

		return nil
	},
}

func aiqRound2(v float64) float64 {
	if v < 0 {
		return float64(int64(v*100-0.5)) / 100
	}
	return float64(int64(v*100+0.5)) / 100
}

func aiqRound1(v float64) float64 {
	if v < 0 {
		return float64(int64(v*10-0.5)) / 10
	}
	return float64(int64(v*10+0.5)) / 10
}

func init() {
	Cmd.AddCommand(aiQualityCmd)
}
