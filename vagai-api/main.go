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
		api.GET("/me", handlers.GetMe)
		api.PATCH("/me", handlers.UpdateProfile)
		api.POST("/me/change-password", handlers.ChangePassword)
		api.GET("/jobs", handlers.ListJobs)
		api.GET("/jobs/:id", handlers.GetJob)
		api.PATCH("/jobs/:id", handlers.UpdateJobStatus)

		api.GET("/matches", handlers.ListMatches)
		api.PATCH("/matches/:id", handlers.UpdateMatch)
		api.DELETE("/matches/:id", handlers.DeleteMatch)

		api.GET("/sites", handlers.ListSites)
		api.POST("/sites", handlers.AddSite)
		api.DELETE("/sites/:id", handlers.DeleteSite)

		api.GET("/resumes", handlers.ListResumes)
		api.POST("/resumes/upload", handlers.UploadResume)
		api.POST("/resumes/analyze", handlers.AnalyzeResume)
	}

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  180 * time.Second,
		WriteTimeout: 180 * time.Second,
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
		&models.AgentLog{},
	)

	seedPlans(db)
}

func seedPlans(db *gorm.DB) {
	var count int64
	db.Model(&models.Plan{}).Count(&count)
	if count > 0 {
		return
	}

	plans := []models.Plan{
		{
			Name: "Free", Slug: "free", PriceMonthly: 0, PriceYearly: 0,
			MaxJobs: 100, MaxResumes: 3, MaxSites: 5, MaxCrawlsPerDay: 10,
			Features: `["100 vagas", "3 currículos", "5 sites", "Matching básico"]`,
		},
		{
			Name: "Pro", Slug: "pro", PriceMonthly: 4900, PriceYearly: 49000,
			MaxJobs: 1000, MaxResumes: 10, MaxSites: 20, MaxCrawlsPerDay: 50,
			Features: `["1000 vagas", "10 currículos", "20 sites", "Matching AI avançado", "Alertas por email", "Análise de currículo"]`,
		},
		{
			Name: "Enterprise", Slug: "enterprise", PriceMonthly: 14900, PriceYearly: 149000,
			MaxJobs: -1, MaxResumes: -1, MaxSites: -1, MaxCrawlsPerDay: -1,
			Features: `["Vagas ilimitadas", "Currículos ilimitados", "Sites ilimitados", "Matching AI avançado", "Alertas por email", "Análise de currículo", "API access", "Webhooks", "Suporte prioritário"]`,
		},
	}

	for _, plan := range plans {
		db.Where("slug = ?", plan.Slug).FirstOrCreate(&plan)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
