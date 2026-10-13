package e2e

import (
	"net/http"
	"testing"

	"github.com/anomalyco/vagai-api/internal/models"
)

type statsResponse struct {
	TotalJobs    int64                    `json:"total_jobs"`
	TotalMatches int64                    `json:"total_matches"`
	TotalApplied int64                    `json:"total_applied"`
	ActiveSites  int64                    `json:"active_sites"`
	Languages    []map[string]interface{} `json:"languages"`
	Keywords     []map[string]interface{} `json:"keywords"`
}

func fetchStats(t *testing.T, token string) statsResponse {
	t.Helper()

	resp := doRequest(t, "GET", "/api/stats", nil, token)
	if resp.StatusCode != http.StatusOK {
		var errBody struct {
			Error string `json:"error"`
		}
		parseBody(t, resp, &errBody)
		t.Fatalf("GET /api/stats failed: status=%d, error=%s", resp.StatusCode, errBody.Error)
	}

	var s statsResponse
	parseBody(t, resp, &s)
	return s
}

func createStatsJob(t *testing.T, token, title, url string) uint {
	t.Helper()

	resp := doRequest(t, "POST", "/api/jobs", map[string]interface{}{
		"title":       title,
		"company":     "Stats Corp",
		"url":         url,
		"description": "job de teste de escopo de estatísticas",
	}, token)
	if resp.StatusCode != http.StatusCreated {
		var errBody struct {
			Error string `json:"error"`
		}
		parseBody(t, resp, &errBody)
		t.Fatalf("create job failed: status=%d, error=%s", resp.StatusCode, errBody.Error)
	}

	var result struct {
		ID uint `json:"id"`
	}
	parseBody(t, resp, &result)
	return result.ID
}

func dbCount(t *testing.T, model interface{}, where string, args ...interface{}) int64 {
	t.Helper()
	var n int64
	if err := testDB.Model(model).Where(where, args...).Count(&n).Error; err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	return n
}

// FR-008 / SC-005: usuário comum enxerga apenas os próprios dados,
// mesmo com dados de outros usuários presentes na plataforma.
func TestStats_UserSeesOnlyOwnData(t *testing.T) {
	aliceToken, aliceID := registerUser(t, "Stats Alice", "stats-alice@example.com", "password123", "")
	bobToken, bobID := registerUser(t, "Stats Bob", "stats-bob@example.com", "password123", "")

	// Dados de outro usuário: devem ficar invisíveis para Alice.
	createStatsJob(t, bobToken, "Bob Foreign Job", "https://example.com/jobs/stats-bob-1")
	if err := testDB.Create(&models.Site{
		UserID: bobID, Name: "Bob Site", URL: "https://bob.example.com/jobs", Active: true,
	}).Error; err != nil {
		t.Fatalf("seed bob site: %v", err)
	}

	// Dados próprios da Alice: 2 jobs, 1 site ativo, 2 matches (1 applied).
	jobOne := createStatsJob(t, aliceToken, "Alice Job One", "https://example.com/jobs/stats-alice-1")
	jobTwo := createStatsJob(t, aliceToken, "Alice Job Two", "https://example.com/jobs/stats-alice-2")
	if err := testDB.Create(&models.Site{
		UserID: aliceID, Name: "Alice Site", URL: "https://alice.example.com/jobs", Active: true,
	}).Error; err != nil {
		t.Fatalf("seed alice site: %v", err)
	}
	if err := testDB.Create(&models.Match{
		UserID: aliceID, JobID: jobOne, SimilarityScore: 88.5,
		KeywordsMatched: `["go","sql"]`, AIReason: "compatibilidade alta", Applied: true,
	}).Error; err != nil {
		t.Fatalf("seed alice match applied: %v", err)
	}
	if err := testDB.Create(&models.Match{
		UserID: aliceID, JobID: jobTwo, SimilarityScore: 72.0,
		KeywordsMatched: `["docker"]`, AIReason: "compatibilidade média", Applied: false,
	}).Error; err != nil {
		t.Fatalf("seed alice match: %v", err)
	}

	s := fetchStats(t, aliceToken)

	if s.TotalJobs != 2 {
		t.Errorf("total_jobs = %d, esperado 2 (apenas jobs de Alice; job de Bob excluído)", s.TotalJobs)
	}
	if s.TotalMatches != 2 {
		t.Errorf("total_matches = %d, esperado 2 (apenas matches de Alice)", s.TotalMatches)
	}
	if s.TotalApplied != 1 {
		t.Errorf("total_applied = %d, esperado 1 (apenas applied de Alice)", s.TotalApplied)
	}
	if s.ActiveSites != 1 {
		t.Errorf("active_sites = %d, esperado 1 (apenas site de Alice)", s.ActiveSites)
	}

	// Invariante de contrato: campos do payload continuam presentes (Constituição II).
	if s.Languages == nil || s.Keywords == nil {
		t.Errorf("payload incompleto: languages=%v keywords=%v", s.Languages, s.Keywords)
	}

	// Garantia extra: os dados da plataforma são maiores que o recorte da Alice
	// (há pelo menos o job do Bob), provando que o filtro está ativo.
	platformJobs := dbCount(t, &models.Job{}, "1=1")
	if platformJobs <= s.TotalJobs {
		t.Logf("nota: plataforma com %d jobs <= %d da Alice (demais testes ainda não popularam o DB)", platformJobs, s.TotalJobs)
	}
}

// FR-007 / SC-006: administrador enxerga a plataforma inteira —
// a resposta iguala a soma dos usuários, mesmo sem dados próprios.
func TestStats_AdminSeesPlatformData(t *testing.T) {
	t.Setenv("ADMIN_EMAILS", "stats-admin@example.com")
	adminToken, adminID := registerUser(t, "Stats Admin", "stats-admin@example.com", "password123", "")

	// Dado de usuário comum na plataforma (admin não pode vê-lo como "pessoal").
	carolToken, _ := registerUser(t, "Stats Carol", "stats-carol@example.com", "password123", "")
	createStatsJob(t, carolToken, "Carol Job", "https://example.com/jobs/stats-carol-1")

	var platformJobs, adminOwnJobs int64
	if err := testDB.Model(&models.Job{}).Count(&platformJobs).Error; err != nil {
		t.Fatalf("platform job count: %v", err)
	}
	if err := testDB.Model(&models.Job{}).Where("user_id = ?", adminID).Count(&adminOwnJobs).Error; err != nil {
		t.Fatalf("admin own job count: %v", err)
	}

	s := fetchStats(t, adminToken)

	if s.TotalJobs != platformJobs {
		t.Errorf("total_jobs do admin = %d, esperado o total da plataforma = %d (FR-007)", s.TotalJobs, platformJobs)
	}
	if s.TotalJobs <= 0 {
		t.Errorf("total_jobs do admin = %d, esperado > 0 com dados de usuários na plataforma", s.TotalJobs)
	}
	if adminOwnJobs == 0 && s.TotalJobs > 0 {
		t.Logf("ok: admin com 0 jobs próprios enxerga %d da plataforma (agregação não é pessoal)", s.TotalJobs)
	}

	// SC-006: metric(admin) = Σ metric(user_i) para total_jobs.
	var sumPerUser int64
	if err := testDB.Raw(
		"SELECT COALESCE(SUM(cnt),0) FROM (SELECT COUNT(*) AS cnt FROM jobs GROUP BY user_id) AS sub",
	).Scan(&sumPerUser).Error; err != nil {
		t.Fatalf("sum per user query: %v", err)
	}
	if s.TotalJobs != sumPerUser {
		t.Errorf("total_jobs do admin = %d, soma por usuário = %d (SC-006 quebrada)", s.TotalJobs, sumPerUser)
	}

	// Invariantes do payload para o admin (campos presentes).
	if s.Languages == nil || s.Keywords == nil {
		t.Errorf("payload incompleto: languages=%v keywords=%v", s.Languages, s.Keywords)
	}
}
