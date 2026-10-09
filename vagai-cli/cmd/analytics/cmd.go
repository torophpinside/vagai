package analytics

import (
	"errors"
	"fmt"

	"github.com/anomalyco/vagai-cli/internal/db"
	"github.com/anomalyco/vagai-cli/internal/models"
	"github.com/spf13/cobra"
	"gorm.io/gorm/logger"
)

var orgID uint

// Cmd é o comando raiz `vagai analytics <relatório>` com os subcomandos de
// relatórios (research.md #3).
var Cmd = &cobra.Command{
	Use:   "analytics",
	Short: "Gera relatórios de analytics do banco do VagAI",
	Long: `Gera relatórios de analytics direto do banco de dados:
matching-distribution, resume-insights, job-trends, org-metrics,
interview-prep, ai-quality, time-to-hire, skills-gap, funnel e geographic.`,
	SilenceUsage: true,
}

// setup conecta ao banco, silencia o logger do GORM (saída limpa do relatório)
// e resolve a organização (--org; auto-detecção quando há exatamente uma).
func setup(cmd *cobra.Command) (uint, error) {
	if err := db.Init(); err != nil {
		return 0, err
	}
	db.DB.Logger = logger.Default.LogMode(logger.Silent)

	if orgID > 0 {
		return orgID, nil
	}

	var orgs []models.Organization
	if err := db.DB.Order("id").Find(&orgs).Error; err != nil {
		return 0, err
	}
	switch len(orgs) {
	case 0:
		return 0, errors.New("nenhuma organização encontrada no banco")
	case 1:
		return orgs[0].ID, nil
	default:
		return 0, fmt.Errorf("%d organizações encontradas: use --org <id> para escolher", len(orgs))
	}
}

// mustSetup falha o comando com a mensagem de erro adequada.
func mustSetup(cmd *cobra.Command) (uint, error) {
	id, err := setup(cmd)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func init() {
	// PersistentFlags para o --org estar disponível em TODOS os subcomandos
	// (ex.: `vagai analytics matching-distribution --org 17`).
	Cmd.PersistentFlags().UintVar(&orgID, "org", 0, "ID da organização (auto-detectada se houver apenas uma)")
}

// printHeader imprime o cabeçalho padrão dos relatórios.
func printHeader(title string, org uint) {
	fmt.Printf("=== %s (organização %d) ===\n", title, org)
}
