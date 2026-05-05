package schedule

import (
	"fmt"

	"github.com/anomalyco/vagai-cli/internal/scheduler"
	"github.com/spf13/cobra"
)

var (
	scheduleName    string
	scheduleCommand string
	scheduleCron    string
)

var Cmd = &cobra.Command{
	Use:   "schedule",
	Short: "Gerencia tarefas agendadas",
	Long:  `Adiciona, lista ou remove schedules`,
}

var AddCmd = &cobra.Command{
	Use:   "add",
	Short: "Adiciona um novo schedule",
	RunE: func(cmd *cobra.Command, args []string) error {
		if scheduleName == "" || scheduleCommand == "" || scheduleCron == "" {
			return fmt.Errorf("nome, comando e schedule são obrigatórios")
		}
		return scheduler.AddSchedule(scheduleName, scheduleCommand, scheduleCron)
	},
}

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lista todos os schedules",
	RunE: func(cmd *cobra.Command, args []string) error {
		return scheduler.ListSchedules()
	},
}

var RemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove um schedule",
	RunE: func(cmd *cobra.Command, args []string) error {
		if scheduleName == "" {
			return fmt.Errorf("nome é obrigatório")
		}
		return scheduler.RemoveSchedule(scheduleName)
	},
}

var RunCmd = &cobra.Command{
	Use:   "run",
	Short: "Executa o scheduler em background",
	RunE: func(cmd *cobra.Command, args []string) error {
		return scheduler.RunScheduler()
	},
}

func init() {
	Cmd.AddCommand(AddCmd)
	Cmd.AddCommand(ListCmd)
	Cmd.AddCommand(RemoveCmd)
	Cmd.AddCommand(RunCmd)

	AddCmd.Flags().StringVarP(&scheduleName, "name", "n", "", "Nome do schedule")
	AddCmd.Flags().StringVarP(&scheduleCommand, "command", "c", "", "Comando a executar")
	AddCmd.Flags().StringVarP(&scheduleCron, "schedule", "s", "", "Expressão cron (ex: @hourly, */30 * * * *)")
	AddCmd.MarkFlagRequired("name")
	AddCmd.MarkFlagRequired("command")
	AddCmd.MarkFlagRequired("schedule")

	RemoveCmd.Flags().StringVarP(&scheduleName, "name", "n", "", "Nome do schedule")
	RemoveCmd.MarkFlagRequired("name")
}
