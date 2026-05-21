package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/anomalyco/vagai-api/internal/handlers"
	"github.com/anomalyco/vagai-api/internal/middleware"
	"github.com/anomalyco/vagai-api/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func main() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		getEnv("DB_USER", "vagai"),
		getEnv("DB_PASSWORD", "vagai"),
		getEnv("DB_HOST", "mysql"),
		getEnv("DB_PORT", "3306"),
		getEnv("DB_NAME", "vagai"),
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Falha ao conectar ao banco: %v", err)
	}
	DB = db

	autoMigrate(db)
	dropResumeAnalysisFK(db)
	handlers.SetDB(db)

	r := gin.Default()

	r.Use(middleware.RateLimit(100, time.Minute))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	auth := r.Group("/api/auth")
	{
		auth.POST("/register", handlers.Register)
		auth.POST("/login", handlers.Login)
	}

	api := r.Group("/api")
	api.Use(middleware.JWTAuth())
	api.Use(middleware.ScopedDB(db))
	{
		api.GET("/stats", handlers.GetStats)
		api.GET("/plans", handlers.ListPlans)
		api.GET("/me", handlers.GetMe)
		api.PATCH("/me", handlers.UpdateProfile)
		api.POST("/me/change-password", handlers.ChangePassword)
		api.POST("/me/plan", handlers.ChangePlan)
		api.GET("/jobs", handlers.ListJobs)
		api.POST("/jobs", handlers.CreateJob)
		api.POST("/jobs/extract", handlers.ExtractJob)
		api.GET("/jobs/:id", handlers.GetJob)
		api.PATCH("/jobs/:id", handlers.UpdateJobStatus)

		api.GET("/matches", handlers.ListMatches)
		api.PATCH("/matches/:id", handlers.UpdateMatch)
		api.DELETE("/matches/:id", handlers.DeleteMatch)

		api.GET("/sites", handlers.ListSites)
		api.POST("/sites", handlers.AddSite)
		api.PATCH("/sites/:id", handlers.UpdateSite)
		api.DELETE("/sites/:id", handlers.DeleteSite)

		api.GET("/resumes", handlers.ListResumes)
		api.POST("/resumes/upload", handlers.UploadResume)
		api.POST("/resumes/analyze", handlers.AnalyzeResume)
		api.GET("/resume-analyses", handlers.ListResumeAnalyses)
		api.GET("/resume-analyses/:id", handlers.GetResumeAnalysis)
		api.DELETE("/resume-analyses/:id", handlers.DeleteResumeAnalysis)
	}

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  240 * time.Second,
		WriteTimeout: 240 * time.Second,
	}

	log.Println("VagAI API rodando em :8080")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Falha ao iniciar servidor: %v", err)
	}
}

func autoMigrate(db *gorm.DB) {
	log.Println("Running migrations...")
	db.AutoMigrate(
		&models.Organization{},
		&models.User{},
		&models.Membership{},
		&models.Plan{},
		&models.Subscription{},
		&models.ApiKey{},
		&models.AuditLog{},
		&models.Site{},
		&models.Job{},
		&models.Resume{},
		&models.Match{},
		&models.ResumeAnalysis{},
		&models.AgentLog{},
	)

	seedPlans(db)
}

func dropResumeAnalysisFK(db *gorm.DB) {
	db.Exec("ALTER TABLE resume_analyses DROP FOREIGN KEY fk_resume_analyses_resume")
	db.Exec("ALTER TABLE resume_analyses DROP INDEX fk_resume_analyses_resume")
}

func seedPlans(db *gorm.DB) {
	plans := []models.Plan{
		{
			Name: "Free", Slug: "free", PriceMonthly: 0, PriceYearly: 0,
			MaxJobs: 1000, MaxResumes: 1, MaxSites: 5,
			Features: `["1000 vagas", "1 currículo", "5 fontes", "Matching básico"]`,
		},
		{
			Name: "Pro", Slug: "pro", PriceMonthly: 4900, PriceYearly: 49000,
			MaxJobs: -1, MaxResumes: 3, MaxSites: 25,
			Features: `["Vagas ilimitadas", "3 currículos", "25 fontes", "Matching por IA"]`,
		},
	}

	for _, plan := range plans {
		var existing models.Plan
		if err := db.Where("slug = ?", plan.Slug).First(&existing).Error; err != nil {
			db.Create(&plan)
		} else {
			plan.ID = existing.ID
			db.Model(&existing).Updates(plan)
		}
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
