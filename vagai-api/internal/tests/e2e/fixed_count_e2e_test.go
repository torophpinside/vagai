package e2e

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/anomalyco/vagai-api/internal/models"
)

// Helpers compartilhados dos testes de interview-prep (006-fixed-count).
// O ambiente de E2E não tem LM Studio; todos os testes forçam o fallback
// determinístico via LMSTUDIO_URL apontando para uma porta morta.

func getOrgID(t *testing.T, token string) uint {
	t.Helper()

	resp := doRequest(t, "GET", "/api/me", nil, token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/me failed: status=%d", resp.StatusCode)
	}
	var body struct {
		Organization struct {
			ID uint `json:"id"`
		} `json:"organization"`
	}
	parseBody(t, resp, &body)
	if body.Organization.ID == 0 {
		t.Fatal("GET /api/me returned empty organization id")
	}
	return body.Organization.ID
}

func insertResume(t *testing.T, orgID uint, name, content string) uint {
	t.Helper()

	resume := models.Resume{
		OrganizationID: orgID,
		Name:           name,
		Content:        content,
		// MySQL rejeita string vazia em coluna JSON; "{}" desfaz para
		// ResumeData vazio e o matching usa o Content como fallback.
		Data:       "{}",
		UploadedAt: time.Now(),
	}
	if err := testDB.Create(&resume).Error; err != nil {
		t.Fatalf("inserting resume: %v", err)
	}
	return resume.ID
}

func listMatchIDsForJob(t *testing.T, token string, jobID uint) []uint {
	t.Helper()

	resp := doRequest(t, "GET", "/api/matches?threshold=1", nil, token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/matches failed: status=%d", resp.StatusCode)
	}
	var body struct {
		Data []struct {
			ID    uint `json:"id"`
			JobID uint `json:"job_id"`
		} `json:"data"`
	}
	parseBody(t, resp, &body)

	var ids []uint
	for _, m := range body.Data {
		if m.JobID == jobID {
			ids = append(ids, m.ID)
		}
	}
	return ids
}

func applyMatch(t *testing.T, token string, matchID uint) {
	t.Helper()

	resp := doRequest(t, "PATCH", fmt.Sprintf("/api/matches/%d", matchID), map[string]bool{"applied": true}, token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("PATCH /api/matches/%d applied=true failed: status=%d", matchID, resp.StatusCode)
	}
	resp.Body.Close()
}

// setupAppliedMatch cria vaga -> match -> applied e retorna o matchID.
// A descrição define as tecnologias detectadas; a URL deve ser única por teste.
func setupAppliedMatch(t *testing.T, token string, orgID uint, url, title, description string) uint {
	t.Helper()

	insertResume(t, orgID, "Resume "+url,
		"Desenvolvedor backend com Go, MySQL, Redis, Docker, Kubernetes, React, testes, APIs REST, CI/CD e observabilidade.")

	jobID := createJob(t, token, map[string]interface{}{
		"url":         url,
		"title":       title,
		"description": description,
	})

	createMatch(t, token, jobID, 0)

	ids := listMatchIDsForJob(t, token, jobID)
	if len(ids) == 0 {
		t.Fatalf("expected at least 1 match for job %d", jobID)
	}
	applyMatch(t, token, ids[0])
	return ids[0]
}

type fixedCountQuestionDTO struct {
	ID          uint   `json:"id"`
	Category    string `json:"category"`
	Topic       string `json:"topic"`
	Text        string `json:"text"`
	AnswerGuide string `json:"answer_guide"`
	Status      string `json:"status"`
	Position    int    `json:"position"`
}

type fixedCountPrepResponse struct {
	Preparation struct {
		ID        uint                    `json:"id"`
		Source    string                  `json:"source"`
		Questions []fixedCountQuestionDTO `json:"questions"`
		Progress  struct {
			Total     int `json:"total"`
			Answered  int `json:"answered"`
			Remaining int `json:"remaining"`
		} `json:"progress"`
	} `json:"preparation"`
}

func createFixedCountPrep(t *testing.T, token string, matchID uint, seed string) fixedCountPrepResponse {
	t.Helper()

	t.Setenv("LMSTUDIO_URL", "http://127.0.0.1:59999")

	body := map[string]interface{}{"match_id": matchID}
	if seed != "" {
		body["@random"] = seed
	}
	resp := doRequest(t, "POST", "/api/interview-prep", body, token)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST /api/interview-prep failed: status=%d", resp.StatusCode)
	}
	var prep fixedCountPrepResponse
	parseBody(t, resp, &prep)
	return prep
}

func assertValidQuestions(t *testing.T, qs []fixedCountQuestionDTO, requireAllCategories bool) {
	t.Helper()

	seen := map[string]bool{}
	cats := map[string]bool{}
	for _, q := range qs {
		if q.Text == "" {
			t.Errorf("question %d has empty text", q.ID)
		}
		if q.AnswerGuide == "" {
			t.Errorf("question %d has empty answer_guide", q.ID)
		}
		switch q.Category {
		case "technology", "foundations", "architecture":
			cats[q.Category] = true
		default:
			t.Errorf("question %d has invalid category %q", q.ID, q.Category)
		}
		if seen[q.Text] {
			t.Errorf("duplicated question text %q", q.Text)
		}
		seen[q.Text] = true
	}
	// Truncagem preserva a ordem e pode remover categorias finais (FR-003);
	// só exige as 3 categorias quando o conjunto não foi truncado.
	if requireAllCategories && !(cats["technology"] && cats["foundations"] && cats["architecture"]) {
		t.Errorf("expected all 3 categories present, got %v", cats)
	}
}

const fixedCountTechDescription = "Vaga backend com Go, MySQL, Redis e Docker em produção. " +
	"Experiência com testes, APIs REST, CI/CD e observabilidade."

// T005 [US1] Geração padrão retorna exatamente 15 perguntas.
func TestFixedCount_DefaultGeneration(t *testing.T) {
	token, _ := registerUser(t, "Fixed Default", "fixed-default@example.com", "password123", "Fixed Default Org")
	orgID := getOrgID(t, token)

	matchID := setupAppliedMatch(t, token, orgID,
		"https://example.com/job/fixed-default", "Backend Developer", fixedCountTechDescription)

	prep := createFixedCountPrep(t, token, matchID, "")

	if n := len(prep.Preparation.Questions); n != 15 {
		t.Fatalf("expected exactly 15 questions, got %d", n)
	}
	if prep.Preparation.Progress.Total != 15 {
		t.Errorf("expected progress.total=15, got %d", prep.Preparation.Progress.Total)
	}
	assertValidQuestions(t, prep.Preparation.Questions, true)
}

// T006 [US1] Truncagem: banco com 19 perguntas resulta em exatamente 15.
func TestFixedCount_TruncatesWhenMoreThan15(t *testing.T) {
	token, _ := registerUser(t, "Fixed Truncate", "fixed-truncate@example.com", "password123", "Fixed Truncate Org")
	orgID := getOrgID(t, token)

	desc := "Stack: Go, MySQL, Redis, Docker, Kubernetes, Vue, React e Python em produção."
	matchID := setupAppliedMatch(t, token, orgID,
		"https://example.com/job/fixed-truncate", "Fullstack Developer", desc)

	prep := createFixedCountPrep(t, token, matchID, "")

	if n := len(prep.Preparation.Questions); n != 15 {
		t.Fatalf("expected exactly 15 questions after truncation, got %d", n)
	}
	assertValidQuestions(t, prep.Preparation.Questions, false)
}

// T007 [US1] Suplementação: banco com 7 perguntas resulta em exatamente 15.
func TestFixedCount_SupplementsWhenFewerThan15(t *testing.T) {
	token, _ := registerUser(t, "Fixed Supplement", "fixed-supplement@example.com", "password123", "Fixed Supplement Org")
	orgID := getOrgID(t, token)

	// Descrição sem tecnologias detectadas (pool de 7), mas com termos em
	// comum com o currículo para gerar match (score > 0).
	matchID := setupAppliedMatch(t, token, orgID,
		"https://example.com/job/fixed-supplement", "Estagiário Administrativo",
		"Vaga para desenvolvedor backend com foco em testes e observabilidade em produção.")

	prep := createFixedCountPrep(t, token, matchID, "")

	if n := len(prep.Preparation.Questions); n != 15 {
		t.Fatalf("expected exactly 15 questions after supplementation, got %d", n)
	}
	assertValidQuestions(t, prep.Preparation.Questions, true)
}

// T008 [US1] Fallback de template (sem IA) retorna exatamente 15 perguntas.
func TestFixedCount_TemplateFallback(t *testing.T) {
	token, _ := registerUser(t, "Fixed Fallback", "fixed-fallback@example.com", "password123", "Fixed Fallback Org")
	orgID := getOrgID(t, token)

	matchID := setupAppliedMatch(t, token, orgID,
		"https://example.com/job/fixed-fallback", "Backend Developer", fixedCountTechDescription)

	prep := createFixedCountPrep(t, token, matchID, "")

	if prep.Preparation.Source != "template" {
		t.Errorf("expected source=template without LM Studio, got %q", prep.Preparation.Source)
	}
	if n := len(prep.Preparation.Questions); n != 15 {
		t.Fatalf("expected exactly 15 questions from template fallback, got %d", n)
	}
}

// T009 [US1] Mesma seed @random produz a mesma ordem de perguntas.
func TestFixedCount_ReproducibleOrderingWithSeed(t *testing.T) {
	token, _ := registerUser(t, "Fixed Seed", "fixed-seed@example.com", "password123", "Fixed Seed Org")
	orgID := getOrgID(t, token)

	matchA := setupAppliedMatch(t, token, orgID,
		"https://example.com/job/fixed-seed-a", "Backend Developer", fixedCountTechDescription)
	matchB := setupAppliedMatch(t, token, orgID,
		"https://example.com/job/fixed-seed-b", "Backend Developer", fixedCountTechDescription)

	prepA := createFixedCountPrep(t, token, matchA, "seed123")
	prepB := createFixedCountPrep(t, token, matchB, "seed123")

	if len(prepA.Preparation.Questions) != 15 || len(prepB.Preparation.Questions) != 15 {
		t.Fatalf("expected 15 questions in both preps, got %d and %d",
			len(prepA.Preparation.Questions), len(prepB.Preparation.Questions))
	}
	for i := range prepA.Preparation.Questions {
		a := prepA.Preparation.Questions[i].Text
		b := prepB.Preparation.Questions[i].Text
		if a != b {
			t.Fatalf("position %d differs with same seed: %q vs %q", i, a, b)
		}
	}
}
