package scheduler

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/anomalyco/vagai-cli/internal/db"
	"github.com/anomalyco/vagai-cli/internal/models"
	"github.com/robfig/cron/v3"
)

var c *cron.Cron
var jobs = make(map[string]cron.EntryID)

func AddSchedule(name, command, schedule string) error {
	if err := db.Init(); err != nil {
		return fmt.Errorf("falha ao inicializar banco: %w", err)
	}

	sched := models.Schedule{
		Name:     name,
		Command:  command,
		Schedule: schedule,
		Active:   true,
	}

	if err := db.DB.Create(&sched).Error; err != nil {
		return fmt.Errorf("falha ao adicionar schedule: %w", err)
	}

	log.Printf("Schedule adicionado: %s (%s) - %s", name, schedule, command)

	if c != nil {
		if err := addToCron(name, command, schedule); err != nil {
			log.Printf("Erro ao adicionar ao cron: %v", err)
		}
	}

	return nil
}

func ListSchedules() error {
	if err := db.Init(); err != nil {
		return fmt.Errorf("falha ao inicializar banco: %w", err)
	}

	var schedules []models.Schedule
	db.DB.Find(&schedules)

	fmt.Println("\n=== Schedules ===")
	for _, s := range schedules {
		status := "ATIVO"
		if !s.Active {
			status = "INATIVO"
		}
		fmt.Printf("[%s] %s\n", status, s.Name)
		fmt.Printf("  Comando: %s\n", s.Command)
		fmt.Printf("  Cron: %s\n", s.Schedule)
		if s.NextRun != nil {
			fmt.Printf("  Próxima execução: %s\n", s.NextRun.Format("2006-01-02 15:04:05"))
		}
		fmt.Println()
	}

	return nil
}

func RemoveSchedule(name string) error {
	if err := db.Init(); err != nil {
		return fmt.Errorf("falha ao inicializar banco: %w", err)
	}

	var sched models.Schedule
	if err := db.DB.Where("name = ?", name).First(&sched).Error; err != nil {
		return fmt.Errorf("schedule não encontrado: %w", err)
	}

	db.DB.Delete(&sched)

	if id, ok := jobs[name]; ok {
		c.Remove(id)
		delete(jobs, name)
	}

	log.Printf("Schedule removido: %s", name)
	return nil
}

func RunScheduler() error {
	log.Println("Iniciando Scheduler...")

	if err := db.Init(); err != nil {
		return fmt.Errorf("falha ao inicializar banco: %w", err)
	}

	c = cron.New()

	var schedules []models.Schedule
	db.DB.Where("active = ?", true).Find(&schedules)

	for _, s := range schedules {
		if err := addToCron(s.Name, s.Command, s.Schedule); err != nil {
			log.Printf("Erro ao adicionar %s: %v", s.Name, err)
		}
	}

	log.Printf("Scheduler iniciado com %d jobs", len(schedules))
	c.Start()

	select {}
}

func addToCron(name, command, schedule string) error {
	scheduleNormalized := normalizeSchedule(schedule)
	
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	_, err := parser.Parse(scheduleNormalized)
	if err != nil {
		return fmt.Errorf("expressão cron inválida: %w", err)
	}

	id, err := c.AddFunc(scheduleNormalized, func() {
		runCommand(name, command)
	})
	if err != nil {
		return fmt.Errorf("erro ao adicionar job: %w", err)
	}

	jobs[name] = id
	log.Printf("Job adicionado ao cron: %s", name)

	return nil
}

func normalizeSchedule(schedule string) string {
	schedule = strings.TrimSpace(schedule)
	
	schedule = strings.ReplaceAll(schedule, "@hourly", "0 * * * *")
	schedule = strings.ReplaceAll(schedule, "@daily", "0 0 * * *")
	schedule = strings.ReplaceAll(schedule, "@weekly", "0 0 * * 0")
	schedule = strings.ReplaceAll(schedule, "@monthly", "0 0 1 * *")
	schedule = strings.ReplaceAll(schedule, "@yearly", "0 0 1 1 *")
	schedule = strings.ReplaceAll(schedule, "@annually", "0 0 1 1 *")
	
	if strings.HasPrefix(schedule, "@every ") {
		duration := strings.TrimPrefix(schedule, "@every ")
		parts := strings.Fields(duration)
		if len(parts) >= 1 {
			interval := parts[0]
			switch interval {
			case "1h":
				return "0 * * * *"
			case "2h":
				return "0 */2 * * *"
			case "3h", "3H":
				return "0 */3 * * *"
			case "4h", "4H":
				return "0 */4 * * *"
			case "6h", "6H":
				return "0 */6 * * *"
			case "12h", "12H":
				return "0 */12 * * *"
			case "1m":
				return "*/1 * * * *"
			case "2m":
				return "*/2 * * * *"
			case "5m":
				return "*/5 * * * *"
			case "10m":
				return "*/10 * * * *"
			case "15m":
				return "*/15 * * * *"
			case "30m":
				return "*/30 * * * *"
			}
		}
	}
	
	return schedule
}

func runCommand(name, command string) {
	log.Printf("[%s] Executando: %s", name, command)

	startTime := time.Now()
	startedAt := startTime

	parts := strings.Fields(command)
	cmd := exec.Command(parts[0], parts[1:]...)
	output, err := cmd.CombinedOutput()

	finishedAt := time.Now()

	logEntry := models.ScheduleLog{
		ScheduleID: getScheduleID(name),
		Status:     "success",
		Output:     string(output),
		StartedAt:  startedAt,
		FinishedAt: &finishedAt,
	}

	if err != nil {
		logEntry.Status = "failed"
		log.Printf("[%s] Erro: %v", name, err)
	}

	db.DB.Create(&logEntry)
	db.DB.Model(&models.Schedule{}).Where("name = ?", name).Update("last_run", startTime)
}

func getScheduleID(name string) uint {
	var sched models.Schedule
	db.DB.Where("name = ?", name).First(&sched)
	return sched.ID
}
