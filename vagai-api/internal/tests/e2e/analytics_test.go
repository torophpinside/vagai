package e2e

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/anomalyco/vagai-api/internal/models"
)

// analyticsEnvelope espelha o shape plano de analytics.Envelope (sem wrapper
// "data"): {user, metrics, pagination?, empty_state?}.
type analyticsEnvelope struct {
	User struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	} `json:"user"`
	Metrics struct {
		JobsPosted int64 `json:"jobs_posted"`
	} `json:"metrics"`
}

func getAnalytics(t *testing.T, path, token string) (int, analyticsEnvelope) {
	t.Helper()

	resp := doRequest(t, "GET", path, nil, token)
	var env analyticsEnvelope
	if resp.StatusCode == http.StatusOK {
		parseBody(t, resp, &env)
	} else {
		resp.Body.Close()
	}
	return resp.StatusCode, env
}

// Regra de escopo de Analytics: usuário comum acessa e vê apenas os próprios
// dados; admin vê dados globais (todos os usuários).
func TestAnalytics_RoleScoping(t *testing.T) {
	aliceToken, aliceID := registerUser(t, "An Alice", "an-alice@example.com", "password123", "")
	bobToken, bobID := registerUser(t, "An Bob", "an-bob@example.com", "password123", "")

	createStatsJob(t, aliceToken, "An Alice Job", "https://example.com/jobs/an-alice-1")

	// 1) Usuário comum consegue acessar analytics (era 403 admin-only).
	status, _ := getAnalytics(t, "/api/v1/analytics/matching-distribution", aliceToken)
	if status != http.StatusOK {
		t.Errorf("user comum em matching-distribution: status=%d, esperado 200", status)
	}

	// 2) Usuário comum pedindo dados de outro usuário recebe os próprios.
	status, env := getAnalytics(t, fmt.Sprintf("/api/v1/analytics/user/%d", bobID), aliceToken)
	if status != http.StatusOK {
		t.Errorf("user comum em /user/%d: status=%d, esperado 200", bobID, status)
	}
	if env.User.ID != aliceID {
		t.Errorf("user comum recebeu user.id=%d, esperado o próprio id=%d (escopo forçado)", env.User.ID, aliceID)
	}

	// 3) Admin vê dados globais em /user/0 (soma de todos os usuários).
	t.Setenv("ADMIN_EMAILS", "an-admin@example.com")
	adminToken, _ := registerUser(t, "An Admin", "an-admin@example.com", "password123", "")

	createStatsJob(t, bobToken, "An Bob Job", "https://example.com/jobs/an-bob-1")

	var platformJobs int64
	if err := testDB.Model(&models.Job{}).Count(&platformJobs).Error; err != nil {
		t.Fatalf("platform job count: %v", err)
	}

	status, env = getAnalytics(t, "/api/v1/analytics/user/0", adminToken)
	if status != http.StatusOK {
		t.Errorf("admin em /user/0 (globais): status=%d, esperado 200", status)
	}
	if env.Metrics.JobsPosted != platformJobs {
		t.Errorf("admin /user/0 jobs_posted=%d, esperado o total da plataforma=%d", env.Metrics.JobsPosted, platformJobs)
	}
}
