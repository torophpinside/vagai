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
