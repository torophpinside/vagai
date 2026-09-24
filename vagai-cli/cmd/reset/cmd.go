package reset

import (
	"log"
	"time"

	"github.com/anomalyco/vagai-cli/internal/db"
	"github.com/anomalyco/vagai-cli/internal/models"
	"github.com/spf13/cobra"
)

var (
	Cmd = &cobra.Command{
		Use:   "reset",
		Short: "Reseta vagas coletadas em uma data para 'new' sem score",
		Long: `Define as vagas coletadas em uma data (padrao: hoje) como 'new' e remove
os matches (score) associados a elas, permitindo que o matching seja refeito.`,
		Args: cobra.NoArgs,
		Run:  run,
	}
	resetDate string
	orgID     uint
	force     bool
)

func init() {
	Cmd.Flags().StringVar(&resetDate, "date", "", "Data das vagas a resetar (YYYY-MM-DD). Padrao: hoje")
	Cmd.Flags().UintVar(&orgID, "org", 0, "Limitar a uma organizacao (padrao: todas)")
	Cmd.Flags().BoolVarP(&force, "force", "f", false, "Aplicar de fato (sem isso, apenas mostra o que sera feito)")
}

func run(cmd *cobra.Command, args []string) {
	if err := db.Init(); err != nil {
		log.Printf("Erro ao conectar banco: %v", err)
		return
	}

	base := time.Now()
	if resetDate != "" {
		parsed, err := time.Parse("2006-01-02", resetDate)
		if err != nil {
			log.Printf("Data invalida: %v (use YYYY-MM-DD)", err)
			return
		}
		base = parsed
	}
	start := time.Date(base.Year(), base.Month(), base.Day(), 0, 0, 0, 0, time.Local)
	end := start.Add(24 * time.Hour)

	q := db.DB.Model(&models.Job{}).Where("collected_at >= ? AND collected_at < ?", start, end)
	if orgID != 0 {
		q = q.Where("organization_id = ?", orgID)
	}

	var jobIDs []uint
	if err := q.Pluck("id", &jobIDs).Error; err != nil {
		log.Printf("Erro ao buscar vagas: %v", err)
		return
	}

	var matchCount int64
	if len(jobIDs) > 0 {
		db.DB.Model(&models.Match{}).Where("job_id IN ?", jobIDs).Count(&matchCount)
	}

	log.Printf("Vagas coletadas em %s: %d | matches (scores) associados: %d", start.Format("2006-01-02"), len(jobIDs), matchCount)
	if len(jobIDs) == 0 {
		log.Println("Nenhuma vaga no periodo. Nada a fazer.")
		return
	}

	if !force {
		log.Println("Modo previsualizacao. Use --force para aplicar (status -> new, matches removidos).")
		return
	}

	if err := db.DB.Where("job_id IN ?", jobIDs).Delete(&models.Match{}).Error; err != nil {
		log.Printf("Erro ao remover matches: %v", err)
		return
	}

	if err := db.DB.Model(&models.Job{}).Where("id IN ?", jobIDs).Update("status", models.JobStatusNew).Error; err != nil {
		log.Printf("Erro ao resetar status das vagas: %v", err)
		return
	}

	log.Printf("Concluido: %d vagas agora sao 'new' sem score", len(jobIDs))
}