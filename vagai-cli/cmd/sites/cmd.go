package sites

import (
	"fmt"

	"github.com/anomalyco/vagai-cli/internal/agents/registry"
	"github.com/spf13/cobra"
)

var (
	siteName     string
	siteURL      string
	siteSelector string
)

var Cmd = &cobra.Command{
	Use:   "sites",
	Short: "Gerencia sites de vagas",
	Long:  `Adiciona, lista ou remove sites de vagas`,
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Adiciona um novo site",
	RunE: func(cmd *cobra.Command, args []string) error {
		if siteName == "" || siteURL == "" {
			return fmt.Errorf("nome e URL são obrigatórios")
		}
		return registry.AddSite(siteName, siteURL, siteSelector)
	},
}

func init() {
	Cmd.AddCommand(addCmd)
	addCmd.Flags().StringVarP(&siteName, "name", "n", "", "Nome do site")
	addCmd.Flags().StringVarP(&siteURL, "url", "u", "", "URL do site")
	addCmd.Flags().StringVarP(&siteSelector, "selector", "s", "", "Seletor CSS para links de vagas")
}
