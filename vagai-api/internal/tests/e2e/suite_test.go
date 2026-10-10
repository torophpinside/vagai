package e2e

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/anomalyco/vagai-api/internal/handlers"
	"github.com/anomalyco/vagai-api/internal/middleware"
	"github.com/anomalyco/vagai-api/internal/models"
	"github.com/gin-gonic/gin"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	testDB     *gorm.DB
	testServer *httptest.Server
	testCtx    context.Context
	testCancel context.CancelFunc
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	fmt.Fprintf(os.Stderr, "[E2E] Starting MySQL container...\n")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	testCtx = ctx
	testCancel = cancel

	container, db, err := startMySQL(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[E2E] SKIP: MySQL container not available: %v\n", err)
		os.Exit(0)
	}

	fmt.Fprintf(os.Stderr, "[E2E] MySQL ready, running migrations...\n")
	testDB = db

	os.Setenv("JWT_SECRET", "test-secret-for-e2e-tests")

	if err := runMigrations(db); err != nil {
		fmt.Fprintf(os.Stderr, "[E2E] Migration failed: %v\n", err)
		os.Exit(1)
	}
	if err := seedPlans(db); err != nil {
		fmt.Fprintf(os.Stderr, "[E2E] Seed failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "[E2E] Starting test server...\n")
	r := setupRouter(db)
	testServer = httptest.NewServer(r)
	fmt.Fprintf(os.Stderr, "[E2E] Server ready at %s\n", testServer.URL)

	code := m.Run()

	testServer.Close()
	cancel()
	container.Terminate(context.Background())

	os.Unsetenv("JWT_SECRET")
	os.Exit(code)
}

func startMySQL(ctx context.Context) (tc.Container, *gorm.DB, error) {
	req := tc.ContainerRequest{
		Image:        "mysql:8",
		ExposedPorts: []string{"3306/tcp"},
		Env: map[string]string{
			"MYSQL_ROOT_PASSWORD": "testpass",
			"MYSQL_DATABASE":      "vagai_test",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("port: 3306  MySQL Community Server"),
			wait.ForListeningPort("3306/tcp"),
		).WithDeadline(3 * time.Minute),
		SkipReaper: true,
	}

	container, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("starting mysql: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("getting host: %w", err)
	}

	port, err := container.MappedPort(ctx, "3306")
	if err != nil {
		return nil, nil, fmt.Errorf("getting port: %w", err)
	}

	dsn := fmt.Sprintf("root:testpass@tcp(%s:%s)/vagai_test?charset=utf8mb4&parseTime=True&loc=Local", host, port.Port())

	var db *gorm.DB
	var sqlDB *sql.DB
	for i := 0; i < 60; i++ {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		sqlDB, err = db.DB()
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		if err = sqlDB.Ping(); err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}
	if err != nil {
		return nil, nil, fmt.Errorf("connecting to mysql after retries: %w", err)
	}

	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(2)

	return container, db, nil
}

func runMigrations(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Plan{},
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
}

func seedPlans(db *gorm.DB) error {
	plans := []models.Plan{
		{
			Name: "Free", Slug: "free", Type: models.PlanTypeFree, Active: true, IsDefault: true,
			PriceMonthly: 0, PriceYearly: 0,
			MaxJobs: 1000, MaxResumeAnalyses: 1, MaxSites: 5,
			AIMatching: false, InterviewPrep: false,
			Features: `["Matching básico"]`,
		},
		{
			Name: "Pro", Slug: "pro", Type: models.PlanTypePaid, Active: true,
			PriceMonthly: 4900, PriceYearly: 49000,
			MaxJobs: -1, MaxResumeAnalyses: 3, MaxSites: 25,
			AIMatching: true, InterviewPrep: true,
			Features: `[]`,
		},
	}
	for _, plan := range plans {
		if err := db.Where("slug = ?", plan.Slug).FirstOrCreate(&plan).Error; err != nil {
			return err
		}
	}
	return nil
}

func setupRouter(db *gorm.DB) *gin.Engine {
	handlers.SetDB(db)
	r := gin.New()
	r.Use(gin.Recovery())

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
	{
		// Analytics: /api/v1/analytics/* — autenticado; escopo por papel
		// resolvido em middleware.ResolveAnalyticsUserID (admin globais,
		// usuário comum os próprios dados).
		handlers.RegisterAnalyticsRoutes(api.Group("/v1/analytics"))

		// Rotas de gestão de planos (spec 012) — espelha o main.go.
		admin := api.Group("/admin")
		admin.Use(middleware.RequireAdmin())
		{
			admin.GET("/plans", handlers.ListPlansAdmin)
			admin.POST("/plans", handlers.CreatePlan)
			admin.PATCH("/plans/:id", handlers.UpdatePlan)
		}

		api.GET("/me", handlers.GetMe)
		api.PATCH("/me", handlers.UpdateProfile)
		api.POST("/me/change-password", handlers.ChangePassword)
		api.POST("/me/plan", handlers.ChangePlan)
		api.GET("/plans", handlers.ListPlans)
		api.GET("/negative-keywords", handlers.GetNegativeKeywords)
		api.PUT("/negative-keywords", handlers.UpdateNegativeKeywords)
		api.GET("/stats", handlers.GetStats)
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
		api.DELETE("/resumes/:id", handlers.DeleteResume)
		api.POST("/resumes/upload", handlers.UploadResume)
		api.POST("/resumes/analyze", handlers.AnalyzeResume)
		api.GET("/resume-analyses", handlers.ListResumeAnalyses)
		api.GET("/resume-analyses/:id", handlers.GetResumeAnalysis)
		api.DELETE("/resume-analyses/:id", handlers.DeleteResumeAnalysis)
		api.GET("/interview-prep", handlers.ListInterviewPreps)
		api.POST("/interview-prep", handlers.CreateInterviewPrep)
		api.POST("/interview-prep/spontaneous", handlers.CreateSpontaneousPreparation)
		api.GET("/interview-prep/:id", handlers.GetInterviewPrep)
		api.PATCH("/interview-prep/:id/questions/:questionId", handlers.UpdateQuestionStatus)
		api.POST("/interview-prep/:id/questions/:questionId/answers", handlers.SaveQuestionAnswer)
		api.POST("/interview-prep/:id/verify", handlers.VerifyInterviewPrep)
	}

	return r
}

func testURL(path string) string {
	if testServer == nil {
		return ""
	}
	return testServer.URL + path
}

func registerUser(t *testing.T, name, email, password, _ string) (string, uint) {
	t.Helper()

	body := map[string]string{
		"name":     name,
		"email":    email,
		"password": password,
	}

	resp := doRequest(t, "POST", "/api/auth/register", body, "")
	if resp.StatusCode != http.StatusCreated {
		var errBody struct {
			Error string `json:"error"`
		}
		parseBody(t, resp, &errBody)
		t.Fatalf("register failed: status=%d, error=%s", resp.StatusCode, errBody.Error)
	}

	var result struct {
		Token string `json:"token"`
		User  struct {
			ID uint `json:"id"`
		} `json:"user"`
	}
	parseBody(t, resp, &result)

	return result.Token, result.User.ID
}

func doRequest(t *testing.T, method, path string, body interface{}, token string) *http.Response {
	t.Helper()

	var reqBody []byte
	if body != nil {
		var err error
		reqBody, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("marshaling body: %v", err)
		}
	}

	req, err := http.NewRequest(method, testURL(path), bytes.NewReader(reqBody))
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}

	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("executing request: %v", err)
	}

	return resp
}

func parseBody(t *testing.T, resp *http.Response, dest interface{}) {
	t.Helper()
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
}
