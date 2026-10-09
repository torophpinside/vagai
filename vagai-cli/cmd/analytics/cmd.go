package analytics

import (
	"errors"
	"fmt"

	"github.com/anomalyco/vagai-cli/internal/db"
	"github.com/anomalyco/vagai-cli/internal/models"
	"github.com/spf13/cobra"
	"gorm.io/gorm/logger"
)

var userID uint

// Cmd é o comando raiz `vagai analytics <relatório>` com os subcomandos de
// relatórios (research.md #3).
var Cmd = &cobra.Command{
	Use:   "analytics",
	Short: "Gera relatórios de analytics do banco do VagAI",
	Long: `Gera relatórios de analytics direto do banco de dados:
matching-distribution, resume-insights, job-trends, user-metrics,
interview-prep, ai-quality, time-to-hire, skills-gap, funnel e geographic.`,
	SilenceUsage: true,
}

// setup conecta ao banco, silencia o logger do GORM (saída limpa do relatório)
// e resolve o usuário (--user; auto-detecção quando há exatamente uma).
func setup(cmd *cobra.Command) (uint, error) {
	if err := db.Init(); err != nil {
		return 0, err
	}
	db.DB.Logger = logger.Default.LogMode(logger.Silent)

	if userID > 0 {
		return userID, nil
	}

	var users []models.User
	if err := db.DB.Order("id").Find(&users).Error; err != nil {
		return 0, err
	}
	switch len(users) {
	case 0:
		return 0, errors.New("nenhum usuário encontrado no banco")
	case 1:
		return users[0].ID, nil
	default:
		return 0, fmt.Errorf("%d usuários encontrados: use --user <id> para escolher", len(users))
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
	// PersistentFlags para o --user estar disponível em TODOS os subcomandos
	// (ex.: `vagai analytics matching-distribution --user 17`).
	Cmd.PersistentFlags().UintVar(&userID, "user", 0, "ID do usuário (auto-detectada se houver apenas uma)")
}

// printHeader imprime o cabeçalho padrão dos relatórios.
func printHeader(title string, user uint) {
	fmt.Printf("=== %s (usuário %d) ===\n", title, user)
}
