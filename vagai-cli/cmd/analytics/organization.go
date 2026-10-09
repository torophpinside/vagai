package analytics

import (
	"fmt"

	"github.com/anomalyco/vagai-cli/internal/db"
	"github.com/spf13/cobra"
)

var orgMetricsCmd = &cobra.Command{
	Use:   "org-metrics",
	Short: "Relatório de indicadores da organização",
	Long: `Volumes de vagas, currículos, matches e candidaturas, taxa de
conversão, utilização de features, assinatura e tempos médios até
match/análise.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		org, err := mustSetup(cmd)
		if err != nil {
			return err
		}
		printHeader("Organization Metrics", org)

		// Textos de estado vazio (espelham EmptyOrgMetricsMessage/Guidance do
		// serviço — os dois módulos Go são independentes).
		const emptyMessage = "Nenhum indicador da organização disponível"
		const emptyGuidance = "Cadastre vagas, envie currículos e execute o matching para começar a acompanhar os indicadores da organização."
		printEmpty := func() {
			fmt.Println(emptyMessage)
			fmt.Println(emptyGuidance)
		}

		type orgInfoRow struct {
			Name string
			Plan string
		}
		var orgInfo []orgInfoRow
		if err := db.DB.Table("organizations").
			Select("name, plan").
			Where("id = ? AND deleted_at IS NULL", org).
			Scan(&orgInfo).Error; err != nil {
			return err
		}
		if len(orgInfo) == 0 {
			printEmpty()
			return nil
		}

		type countsRow struct {
			JobsPosted       int64
			ResumesUploaded  int64
			MatchesMade      int64
			ApplicationsMade int64
			AIAnalyses       int64
			InterviewPreps   int64
		}
		var counts countsRow
		if err := db.DB.Raw(`
			SELECT
				(SELECT COUNT(*) FROM jobs WHERE organization_id = ?) AS jobs_posted,
				(SELECT COUNT(*) FROM resumes WHERE organization_id = ?) AS resumes_uploaded,
				(SELECT COUNT(*) FROM matches WHERE organization_id = ?) AS matches_made,
				(SELECT COUNT(*) FROM matches WHERE organization_id = ? AND applied = ?) AS applications_made,
				(SELECT COUNT(*) FROM ai_analysis_results WHERE organization_id = ?) AS ai_analyses,
				(SELECT COUNT(*) FROM interview_preparations WHERE organization_id = ?) AS interview_preps`,
			org, org, org, org, true, org, org).
			Scan(&counts).Error; err != nil {
			return err
		}
		if counts.JobsPosted == 0 && counts.ResumesUploaded == 0 && counts.MatchesMade == 0 {
			printEmpty()
			return nil
		}

		type subRow struct {
			Status string
		}
		var subs []subRow
		if err := db.DB.Table("subscriptions").
			Select("status").
			Where("organization_id = ? AND deleted_at IS NULL", org).
			Order("created_at DESC").
			Limit(1).
			Scan(&subs).Error; err != nil {
			return err
		}
		subStatus := ""
		if len(subs) > 0 {
			subStatus = subs[0].Status
		}

		var timeToMatch, timeToAnalysis float64
		if err := db.DB.Raw(`
			SELECT COALESCE(AVG(GREATEST(0, DATEDIFF(m.analyzed_at, j.collected_at))), 0)
			FROM matches m
			JOIN jobs j ON j.id = m.job_id
			WHERE m.organization_id = ?`, org).Scan(&timeToMatch).Error; err != nil {
			return err
		}
		if err := db.DB.Raw(`
			SELECT COALESCE(AVG(GREATEST(0, DATEDIFF(ai.created_at, m.analyzed_at))), 0)
			FROM matches m
			JOIN interview_preparations ip ON ip.match_id = m.id
			JOIN ai_analysis_results ai ON ai.preparation_id = ip.id
			WHERE m.organization_id = ?`, org).Scan(&timeToAnalysis).Error; err != nil {
			return err
		}

		yesNo := func(b bool) string {
			if b {
				return "sim"
			}
			return "não"
		}

		fmt.Printf("\nOrganização: %s | Plano: %s\n", orgInfo[0].Name, orgInfo[0].Plan)
		if subStatus == "" {
			fmt.Println("Assinatura: (nenhuma)")
		} else {
			fmt.Printf("Assinatura: %s | Trial→Pago: %s\n", subStatus, yesNo(subStatus == "active"))
		}

		fmt.Println("\nVolumes:")
		fmt.Printf("  Vagas postadas:        %d\n", counts.JobsPosted)
		fmt.Printf("  Currículos enviados:   %d\n", counts.ResumesUploaded)
		fmt.Printf("  Matches realizados:    %d\n", counts.MatchesMade)
		fmt.Printf("  Candidaturas:          %d\n", counts.ApplicationsMade)
		fmt.Printf("  Taxa de conversão:     %.2f%% (%d/%d)\n",
			pct64(counts.ApplicationsMade, counts.MatchesMade),
			counts.ApplicationsMade, counts.MatchesMade)

		fmt.Println("\nUtilização de features:")
		fmt.Printf("  Matching:              %s\n", yesNo(counts.MatchesMade > 0))
		fmt.Printf("  Análise IA:            %s\n", yesNo(counts.AIAnalyses > 0))
		fmt.Printf("  Preparação entrevista: %s\n", yesNo(counts.InterviewPreps > 0))
		fmt.Printf("  Skill Gap:             %s\n",
			yesNo(counts.ResumesUploaded > 0 && counts.MatchesMade > 0))

		fmt.Println("\nTempos médios:")
		fmt.Printf("  Até o match:           %.2f dias\n", timeToMatch)
		fmt.Printf("  Até a análise:         %.2f dias\n", timeToAnalysis)
		return nil
	},
}

func init() {
	Cmd.AddCommand(orgMetricsCmd)
}
