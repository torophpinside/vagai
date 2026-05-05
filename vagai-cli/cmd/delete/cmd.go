package delete

import (
	"fmt"
	"log"
	"strings"

	"github.com/anomalyco/vagai-cli/internal/db"
	"github.com/spf13/cobra"
)

var (
	Cmd = &cobra.Command{
		Use:   "delete [jobs|matches|all]",
		Short: "Limpa dados do banco",
		Long:  `Comando para limpar jobs, matches ou todos os dados`,
		Args:  cobra.MinimumNArgs(1),
		Run:   run,
	}
	force bool
)

func init() {
	Cmd.Flags().BoolVarP(&force, "force", "f", false, "Confirmar sem perguntar")
}

func run(cmd *cobra.Command, args []string) {
	target := args[0]

	if !force {
		log.Printf("Tem certeza que deseja excluir %s? Use --force para confirmar", target)
		return
	}

	if err := db.Init(); err != nil {
		log.Printf("Erro ao conectar banco: %v", err)
		return
	}

	switch strings.ToLower(target) {
	case "jobs":
		result := db.DB.Exec("DELETE FROM jobs")
		log.Printf("Jobs excluídos: %d", result.RowsAffected)
	case "matches":
		result := db.DB.Exec("DELETE FROM matches")
		log.Printf("Matches excluídos: %d", result.RowsAffected)
	case "all":
		db.DB.Exec("DELETE FROM matches")
		db.DB.Exec("DELETE FROM jobs")
		log.Println("Todos os jobs e matches foram excluídos")
	default:
		fmt.Println("Comando inválido. Use: jobs, matches ou all")
	}
}
