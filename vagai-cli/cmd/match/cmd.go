package match

import (
	"github.com/anomalyco/vagai-cli/internal/agents/matcher"
	"github.com/spf13/cobra"
)

var (
	threshold int
	force     bool
)

var Cmd = &cobra.Command{
	Use:   "match",
	Short: "Executa o matching de vagas com currículos",
	Long:  `Compara vagas coletadas com os currículos disponíveis`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return matcher.Run(threshold, force)
	},
}

func init() {
	Cmd.Flags().IntVarP(&threshold, "threshold", "t", 70, "Threshold mínimo de similaridade (0-100)")
	Cmd.Flags().BoolVarP(&force, "force", "f", false, "Força re-matching de todas as vagas")
}
