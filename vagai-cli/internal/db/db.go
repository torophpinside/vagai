package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/anomalyco/vagai-cli/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		getEnv("DB_USER", "vagai"),
		getEnv("DB_PASSWORD", "vagai"),
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "3306"),
		getEnv("DB_NAME", "vagai"),
	)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return fmt.Errorf("falha ao conectar ao banco: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return autoMigrate()
}

func autoMigrate() error {
	log.Println("Executando auto migrate...")
	return DB.AutoMigrate(
		&models.Organization{},
		&models.Plan{},
		&models.Site{},
		&models.Job{},
		&models.Resume{},
		&models.Match{},
		&models.AgentLog{},
		&models.Schedule{},
		&models.ScheduleLog{},
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func Log(orgID uint, agentName, action, details string) {
	logEntry := models.AgentLog{
		OrganizationID: orgID,
		AgentName:      agentName,
		Action:         action,
		Details:        details,
	}
	DB.Create(&logEntry)
}

func CheckJobsLimit(orgID uint) (bool, int, int, error) {
	var org models.Organization
	if err := DB.First(&org, orgID).Error; err != nil {
		return false, 0, 0, err
	}

	var plan models.Plan
	if err := DB.Where("slug = ?", org.Plan).First(&plan).Error; err != nil {
		return false, 0, 0, err
	}

	if plan.MaxJobs == -1 {
		return true, 0, 0, nil
	}

	var count int64
	DB.Model(&models.Job{}).Where("organization_id = ?", orgID).Count(&count)

	return int(count) < plan.MaxJobs, int(count), plan.MaxJobs, nil
}

func CheckSitesLimit(orgID uint) (bool, int, int, error) {
	if orgID == 0 {
		return true, 0, 0, nil
	}

	var org models.Organization
	if err := DB.First(&org, orgID).Error; err != nil {
		return false, 0, 0, err
	}

	var plan models.Plan
	if err := DB.Where("slug = ?", org.Plan).First(&plan).Error; err != nil {
		return false, 0, 0, err
	}

	if plan.MaxSites == -1 {
		return true, 0, 0, nil
	}

	var count int64
	DB.Model(&models.Site{}).Where("organization_id = ?", orgID).Count(&count)

	return int(count) < plan.MaxSites, int(count), plan.MaxSites, nil
}
