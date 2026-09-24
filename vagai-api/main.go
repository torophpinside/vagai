package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
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
	// Segurança: JWT_SECRET é obrigatório (mínimo 32 caracteres). Sem fallback
	// hardcoded — aborta a inicialização se não estiver configurado.
	if len(os.Getenv("JWT_SECRET")) < 32 {
		log.Fatal("JWT_SECRET não configurado. Defina JWT_SECRET (mínimo 32 caracteres) para iniciar a API. Abortando por segurança.")
	}

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
	r.MaxMultipartMemory = 16 << 20

	// Segurança: nenhum proxy é confiável por padrão. Se estiver atrás de um
	// proxy reverso, defina TRUSTED_PROXIES com os CIDRs dele (separados por
	// vírgula) para que c.ClientIP() use o X-Forwarded-For de fontes confiáveis.
	if proxies := os.Getenv("TRUSTED_PROXIES"); strings.TrimSpace(proxies) != "" {
		var cidrs []string
		for _, p := range strings.Split(proxies, ",") {
			if p = strings.TrimSpace(p); p != "" {
				cidrs = append(cidrs, p)
			}
		}
		if err := r.SetTrustedProxies(cidrs); err != nil {
			log.Fatalf("TRUSTED_PROXIES inválido: %v", err)
		}
	} else {
		_ = r.SetTrustedProxies(nil)
	}

	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.CORS(parseOrigins(os.Getenv("CORS_ALLOWED_ORIGINS"))))
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
		api.GET("/negative-keywords", handlers.GetNegativeKeywords)
		api.PUT("/negative-keywords", handlers.UpdateNegativeKeywords)
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
		api.POST("/matches/rematch", handlers.RematchMatches)

		api.GET("/sites", handlers.ListSites)
		api.POST("/sites", handlers.AddSite)
		api.PATCH("/sites/:id", handlers.UpdateSite)
		api.DELETE("/sites/:id", handlers.DeleteSite)

		api.GET("/resumes", handlers.ListResumes)
		api.POST("/resumes", handlers.CreateResume)
		api.DELETE("/resumes/:id", handlers.DeleteResume)
		api.POST("/resumes/upload", handlers.UploadResume)
		api.POST("/resumes/analyze", handlers.AnalyzeResume)
		api.GET("/resume-analyses", handlers.ListResumeAnalyses)
		api.GET("/resume-analyses/:id", handlers.GetResumeAnalysis)
		api.DELETE("/resume-analyses/:id", handlers.DeleteResumeAnalysis)

		api.POST("/resumes/parse", handlers.ParseResume)
		api.GET("/resumes/:id/data", handlers.GetResumeData)
		api.PUT("/resumes/:id/data", handlers.UpdateResumeData)
		api.POST("/resumes/:id/generate-pdf", handlers.GenerateResumePDFHandler)

		api.GET("/interview-prep", handlers.ListInterviewPreps)
		api.POST("/interview-prep", handlers.CreateInterviewPrep)
		api.POST("/interview-prep/spontaneous", handlers.CreateSpontaneousPreparation)
		api.GET("/interview-prep/:id", handlers.GetInterviewPrep)
		api.PATCH("/interview-prep/:id/questions/:questionId", handlers.UpdateQuestionStatus)
		api.POST("/interview-prep/:id/questions/:questionId/answers", handlers.SaveQuestionAnswer)
		api.POST("/interview-prep/:id/verify", handlers.VerifyInterviewPrep)
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
		&models.InterviewPreparation{},
		&models.InterviewQuestion{},
		&models.PracticeEntry{},
		&models.PreparationVerification{},
		&models.AIAnalysisResult{},
		&models.PreparationVerificationHistory{},
	)

	seedPlans(db)
	ensureNullablePrepColumns(db)
}

// ensureNullablePrepColumns relaxa match_id/job_id de interview_preparations
// para NULL em bancos já existentes. O AutoMigrate cria as colunas corretamente
// em bancos novos, mas NÃO remove o NOT NULL de colunas antigas — sem isso,
// preparações avulsas (sem match/job) falhariam com ERRCONN 1048 na primeira
// inserção.
func ensureNullablePrepColumns(db *gorm.DB) {
	check := `SELECT IS_NULLABLE FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'interview_preparations' AND COLUMN_NAME = ?`
	for _, col := range []string{"match_id", "job_id"} {
		var nullable string
		if err := db.Raw(check, col).Scan(&nullable).Error; err != nil || nullable == "YES" {
			continue
		}
		if err := db.Exec("ALTER TABLE interview_preparations MODIFY "+col+" bigint unsigned NULL").Error; err != nil {
			log.Printf("migrate: falha ao relaxar %s: %v", col, err)
		} else {
			log.Printf("migrate: %s relaxado para NULL", col)
		}
	}
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

// parseOrigins converte uma lista separada por vírgula de origens permitidas
// em um slice. Retorna nil quando vazio (nenhuma origem cross-origin liberada).
func parseOrigins(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var origins []string
	for _, o := range strings.Split(raw, ",") {
		if o = strings.TrimSpace(o); o != "" {
			origins = append(origins, o)
		}
	}
	return origins
}
