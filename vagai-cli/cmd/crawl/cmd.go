package crawl

import (
	"github.com/anomalyco/vagai-cli/internal/agents/crawler"
	"github.com/spf13/cobra"
)

var (
	allSites  bool
	siteName  string
	threshold int
	force     bool
)

var Cmd = &cobra.Command{
	Use:   "crawl",
	Short: "Executa o crawler de vagas",
	Long:  `Coleta vagas de todos os sites ativos ou de um site específico`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return crawler.Run(siteName)
	},
}

func init() {
	Cmd.Flags().BoolVarP(&allSites, "all", "a", false, "Executar crawl em todos os sites ativos")
	Cmd.Flags().StringVarP(&siteName, "site", "s", "", "Executar crawl em um site específico")
}
