package root

import (
	"os"

	"github.com/anomalyco/vagai-cli/cmd/crawl"
	"github.com/anomalyco/vagai-cli/cmd/delete"
	"github.com/anomalyco/vagai-cli/cmd/match"
	"github.com/anomalyco/vagai-cli/cmd/schedule"
	"github.com/anomalyco/vagai-cli/cmd/sites"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "vagai",
	Short: "CLI para busca automatizada de vagas de trabalho com IA",
	Long:  `Sistema VagAI para crawler de vagas e matching com currículos usando inteligência artificial`,
}

func Execute() {
	RootCmd.AddCommand(
		crawl.Cmd,
		match.Cmd,
		sites.Cmd,
		schedule.Cmd,
		delete.Cmd,
	)
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
