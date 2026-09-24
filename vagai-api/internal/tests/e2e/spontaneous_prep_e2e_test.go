package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/anomalyco/vagai-api/internal/models"
)

type spontaneousQuestionDTO struct {
	ID          uint   `json:"id"`
	Category    string `json:"category"`
	Topic       string `json:"topic"`
	Text        string `json:"text"`
	AnswerGuide string `json:"answer_guide"`
	Status      string `json:"status"`
	Position    int    `json:"position"`
}

type spontaneousPrepDTO struct {
	Preparation struct {
		ID             uint                       `json:"id"`
		OrganizationID uint                       `json:"organization_id"`
		MatchID        uint                       `json:"match_id"`
		JobID          uint                       `json:"job_id"`
		Title          string                     `json:"title"`
		Company        string                     `json:"company"`
		Description    string                     `json:"description"`
		TechStack      []string                   `json:"tech_stack"`
		Source         string                     `json:"source"`
		Questions      []spontaneousQuestionDTO   `json:"questions"`
		Entries        map[string]json.RawMessage `json:"entries"`
		Analyses       map[string]json.RawMessage `json:"analyses"`
		History        []json.RawMessage          `json:"history"`
		Progress       struct {
			Total     int `json:"total"`
			Practiced int `json:"practiced"`
			Mastered  int `json:"mastered"`
			Answered  int `json:"answered"`
			Remaining int `json:"remaining"`
		} `json:"progress"`
	} `json:"preparation"`
}

type spontaneousListedDTO struct {
	ID          uint                       `json:"id"`
	OrganizationID uint                    `json:"organization_id"`
	MatchID     uint                       `json:"match_id"`
	JobID       uint                       `json:"job_id"`
	Title       string                     `json:"title"`
	Company     string                     `json:"company"`
	Description string                     `json:"description"`
	TechStack   []string                   `json:"tech_stack"`
	Source      string                     `json:"source"`
	Questions   []spontaneousQuestionDTO   `json:"questions"`
	Entries     map[string]json.RawMessage `json:"entries"`
	Analyses    map[string]json.RawMessage `json:"analyses"`
	History     []json.RawMessage          `json:"history"`
	Progress    struct {
		Total     int `json:"total"`
		Practiced int `json:"practiced"`
		Mastered  int `json:"mastered"`
		Answered  int `json:"answered"`
		Remaining int `json:"remaining"`
	} `json:"progress"`
}

const spontaneousE2EDescription = "Vaga backend com Go, MySQL e Docker em produção, com testes e APIs REST."

func containsTech(list []string, want string) bool {
	for _, v := range list {
		if v == want {
			return true
		}
	}
	return false
}

func spontaneousRequestBody(description, seed string) map[string]interface{} {
	body := map[string]interface{}{
		"title":       "Backend Developer",
		"company":     "Tech Corp",
		"description": description,
	}
	if seed != "" {
		body["random"] = seed
	}
	return body
}

// postSpontaneous força a IA para uma porta morta (fallback determinístico) e
// envia a requisição de preparação avulsa persistida.
func postSpontaneous(t *testing.T, token, description, seed string) *http.Response {
	t.Helper()
	t.Setenv("LMSTUDIO_URL", "http://127.0.0.1:59999")
	return doRequest(t, "POST", "/api/interview-prep/spontaneous", spontaneousRequestBody(description, seed), token)
}

func countRows(t *testing.T, model interface{}, orgID uint) int64 {
	t.Helper()
	var count int64
	if err := testDB.Model(model).Where("organization_id = ?", orgID).Count(&count).Error; err != nil {
		t.Fatalf("counting rows: %v", err)
	}
	return count
}

// T007 [US2] Geração bem-sucedida persiste a preparação (201, id real, 15
// perguntas com ids reais) e a expõe na listagem.
func TestSpontaneousPrep_PersistsPreparation(t *testing.T) {
	token, _ := registerUser(t, "Spont Persist", "spont-persist@example.com", "password123", "Spont Persist Org")
	orgID := getOrgID(t, token)

	resp := postSpontaneous(t, token, spontaneousE2EDescription, "")
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	var body spontaneousPrepDTO
	parseBody(t, resp, &body)
	prep := body.Preparation

	if prep.ID == 0 {
		t.Errorf("expected real id > 0, got %d", prep.ID)
	}
	if prep.OrganizationID != orgID {
		t.Errorf("expected organization_id=%d, got %d", orgID, prep.OrganizationID)
	}
	if prep.MatchID != 0 || prep.JobID != 0 {
		t.Errorf("expected match_id=0 job_id=0 for avulsa, got match=%d job=%d", prep.MatchID, prep.JobID)
	}
	if prep.Title != "Backend Developer" || prep.Company != "Tech Corp" {
		t.Errorf("expected echoed title/company, got %q/%q", prep.Title, prep.Company)
	}
	if len(prep.Questions) != 15 {
		t.Fatalf("expected exactly 15 questions, got %d", len(prep.Questions))
	}
	if prep.Progress.Total != 15 || prep.Progress.Remaining != 15 || prep.Progress.Answered != 0 {
		t.Errorf("unexpected progress: %+v", prep.Progress)
	}
	if prep.Entries == nil || prep.Analyses == nil || prep.History == nil {
		t.Error("expected initialized entries/analyses/history collections")
	}
	seen := map[uint]bool{}
	for i, q := range prep.Questions {
		if q.ID == 0 {
			t.Errorf("question %d expected real id > 0", i)
		}
		if seen[q.ID] {
			t.Errorf("duplicate question id %d", q.ID)
		}
		seen[q.ID] = true
		if q.Status != "pending" || q.Position != i+1 {
			t.Errorf("question %d: status=%q position=%d", i, q.Status, q.Position)
		}
		if q.Text == "" || q.AnswerGuide == "" || q.Topic == "" {
			t.Errorf("question %d has empty fields", i)
		}
		switch q.Category {
		case "technology", "foundations", "architecture":
		default:
			t.Errorf("question %d invalid category %q", i, q.Category)
		}
	}

	// Persistência real: linhas na tabela de preparações e perguntas.
	if got := countRows(t, &models.InterviewPreparation{}, orgID); got != 1 {
		t.Errorf("expected 1 preparation row, got %d", got)
	}
	if got := countRows(t, &models.InterviewQuestion{}, orgID); got != 15 {
		t.Errorf("expected 15 question rows, got %d", got)
	}
}

// T007 [US2] A preparação salva aparece na listagem (FR-004/FR-007).
func TestSpontaneousPrep_AppearsInList(t *testing.T) {
	token, _ := registerUser(t, "Spont List", "spont-list@example.com", "password123", "Spont List Org")

	createResp := postSpontaneous(t, token, spontaneousE2EDescription, "list-seed")
	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", createResp.StatusCode)
	}
	var created spontaneousPrepDTO
	parseBody(t, createResp, &created)
	createdID := created.Preparation.ID

	resp := doRequest(t, "GET", "/api/interview-prep", nil, token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 listing, got %d", resp.StatusCode)
	}
	var list struct {
		Data  []spontaneousListedDTO `json:"data"`
		Total int                    `json:"total"`
	}
	parseBody(t, resp, &list)
	if list.Total == 0 {
		t.Fatal("expected at least 1 preparation in listing")
	}
	found := false
	for _, item := range list.Data {
		if item.ID == createdID {
			found = true
			if item.Progress.Total != 15 || item.Progress.Answered != 0 {
				t.Errorf("listed prep progress wrong: %+v", item.Progress)
			}
			if item.Source != "template" {
				t.Errorf("expected listed source=template (fallback), got %q", item.Source)
			}
			break
		}
	}
	if !found {
		t.Fatalf("persisted preparation %d not found in listing", createdID)
	}
}

// T007 [US2] IA indisponível => fallback de template que ainda persiste.
func TestSpontaneousPrep_TemplateFallbackStillPersists(t *testing.T) {
	token, _ := registerUser(t, "Spont Fallback", "spont-fallback@example.com", "password123", "Spont Fallback Org")
	orgID := getOrgID(t, token)

	resp := postSpontaneous(t, token, spontaneousE2EDescription, "")
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	var body spontaneousPrepDTO
	parseBody(t, resp, &body)

	if body.Preparation.Source != "template" {
		t.Errorf("expected source=template without AI, got %q", body.Preparation.Source)
	}
	if len(body.Preparation.Questions) != 15 {
		t.Fatalf("expected exactly 15 questions from fallback, got %d", len(body.Preparation.Questions))
	}
	// Mesmo em fallback, o resultado é persistido (FR-006).
	if got := countRows(t, &models.InterviewPreparation{}, orgID); got != 1 {
		t.Errorf("expected 1 persisted preparation even on fallback, got %d", got)
	}
}

// T007 [US2] Entrada inválida (sem descrição) => 400 e ZERO registros.
func TestSpontaneousPrep_MissingDescriptionPersistsNothing(t *testing.T) {
	token, _ := registerUser(t, "Spont Invalid", "spont-invalid@example.com", "password123", "Spont Invalid Org")
	orgID := getOrgID(t, token)
	t.Setenv("LMSTUDIO_URL", "http://127.0.0.1:59999")

	resp := doRequest(t, "POST", "/api/interview-prep/spontaneous", map[string]interface{}{"title": "Backend"}, token)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}

	if got := countRows(t, &models.InterviewPreparation{}, orgID); got != 0 {
		t.Errorf("empty description must persist nothing, got %d prep rows", got)
	}
	if got := countRows(t, &models.InterviewQuestion{}, orgID); got != 0 {
		t.Errorf("empty description must persist nothing, got %d question rows", got)
	}
}

// T007 [US2] Sem autenticação => 401.
func TestSpontaneousPrep_MissingAuth(t *testing.T) {
	t.Setenv("LMSTUDIO_URL", "http://127.0.0.1:59999")

	resp := doRequest(t, "POST", "/api/interview-prep/spontaneous", spontaneousRequestBody(spontaneousE2EDescription, ""), "")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

// T007 [US2] Repetição cria preparações independentes (FR-008/SC-005):
// mesma seed reproduz a ordem; descrição alterada reflete a stack.
func TestSpontaneousPrep_RepetitionCreatesIndependentPreps(t *testing.T) {
	token, _ := registerUser(t, "Spont Seed", "spont-seed@example.com", "password123", "Spont Seed Org")
	orgID := getOrgID(t, token)

	a := postSpontaneous(t, token, spontaneousE2EDescription, "seed123")
	if a.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", a.StatusCode)
	}
	var first spontaneousPrepDTO
	parseBody(t, a, &first)

	b := postSpontaneous(t, token, spontaneousE2EDescription, "seed123")
	if b.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", b.StatusCode)
	}
	var second spontaneousPrepDTO
	parseBody(t, b, &second)

	// São preparações distintas: ids reais diferentes (FR-008).
	if first.Preparation.ID == second.Preparation.ID || first.Preparation.ID == 0 || second.Preparation.ID == 0 {
		t.Fatalf("expected two distinct real ids, got %d and %d",
			first.Preparation.ID, second.Preparation.ID)
	}

	// Reproduzibilidade: mesma seed => mesma ordem de perguntas.
	if len(first.Preparation.Questions) != len(second.Preparation.Questions) {
		t.Fatalf("question counts differ: %d vs %d",
			len(first.Preparation.Questions), len(second.Preparation.Questions))
	}
	for i := range first.Preparation.Questions {
		if first.Preparation.Questions[i].Text != second.Preparation.Questions[i].Text {
			t.Fatalf("same seed must produce same order: position %d differs", i)
		}
	}

	// Duas preparações persistidas; a primeira intacta.
	if got := countRows(t, &models.InterviewPreparation{}, orgID); got != 2 {
		t.Errorf("expected 2 persisted preparation rows, got %d", got)
	}
	if got := countRows(t, &models.InterviewQuestion{}, orgID); got != 30 {
		t.Errorf("expected 30 persisted question rows, got %d", got)
	}

	// Reflexo da stack na descrição.
	goResp := postSpontaneous(t, token, "Vaga backend com Go e Docker.", "")
	var goPrep spontaneousPrepDTO
	parseBody(t, goResp, &goPrep)
	pyResp := postSpontaneous(t, token, "Vaga de dados com Python e PostgreSQL.", "")
	var pyPrep spontaneousPrepDTO
	parseBody(t, pyResp, &pyPrep)

	if !containsTech(goPrep.Preparation.TechStack, "Go") || containsTech(goPrep.Preparation.TechStack, "Python") {
		t.Errorf("expected Go-only stack, got %v", goPrep.Preparation.TechStack)
	}
	if !containsTech(pyPrep.Preparation.TechStack, "Python") || containsTech(pyPrep.Preparation.TechStack, "Go") {
		t.Errorf("expected Python-only stack, got %v", pyPrep.Preparation.TechStack)
	}
}

// T012 [US3] Preparação avulsa salva é continuada como qualquer outra:
// reabertura, resposta com progresso e verificação por IA.
func TestSpontaneousPrep_ContinueSavedPrep(t *testing.T) {
	token, _ := registerUser(t, "Spont Continue", "spont-continue@example.com", "password123", "Spont Continue Org")

	createResp := postSpontaneous(t, token, spontaneousE2EDescription, "cont-seed")
	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", createResp.StatusCode)
	}
	var created spontaneousPrepDTO
	parseBody(t, createResp, &created)
	prepID := created.Preparation.ID
	firstQuestion := created.Preparation.Questions[0]

	// Reabertura (FR-005).
	getResp := doRequest(t, "GET", fmt.Sprintf("/api/interview-prep/%d", prepID), nil, token)
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 reopening, got %d", getResp.StatusCode)
	}
	var reopened spontaneousPrepDTO
	parseBody(t, getResp, &reopened)
	if len(reopened.Preparation.Questions) != 15 {
		t.Fatalf("expected 15 questions on reopen, got %d", len(reopened.Preparation.Questions))
	}
	for i, q := range reopened.Preparation.Questions {
		if q.AnswerGuide == "" || q.Topic == "" {
			t.Errorf("question %d missing guide/topic on reopen", i)
		}
	}

	// Responde a primeira pergunta (FR-005).
	ansResp := doRequest(t, "POST",
		fmt.Sprintf("/api/interview-prep/%d/questions/%d/answers", prepID, firstQuestion.ID),
		map[string]string{"answer": "Minha resposta de teste", "self_assessment": "medium"}, token)
	if ansResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 answering, got %d", ansResp.StatusCode)
	}
	ansResp.Body.Close()

	// Progresso atualizado (FR-005/US3 acceptance 2).
	afterResp := doRequest(t, "GET", fmt.Sprintf("/api/interview-prep/%d", prepID), nil, token)
	var after spontaneousPrepDTO
	parseBody(t, afterResp, &after)
	if after.Preparation.Progress.Answered != 1 || after.Preparation.Progress.Remaining != 14 {
		t.Errorf("progress after one answer: %+v", after.Preparation.Progress)
	}
}

// T012 [US3] Usuário de outra organização não acessa nem altera a preparação
// avulsa alheia (SC-006/Constitution III).
func TestSpontaneousPrep_TenantIsolation(t *testing.T) {
	tokenA, _ := registerUser(t, "Spont OrgA", "spont-orga@example.com", "password123", "Spont OrgA")
	tokenB, _ := registerUser(t, "Spont OrgB", "spont-orgb@example.com", "password123", "Spont OrgB")

	createResp := postSpontaneous(t, tokenA, spontaneousE2EDescription, "iso-seed")
	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", createResp.StatusCode)
	}
	var created spontaneousPrepDTO
	parseBody(t, createResp, &created)
	prepID := created.Preparation.ID

	// GET: 404 para org B.
	got404 := func(name string, method, path string, body interface{}) {
		t.Helper()
		resp := doRequest(t, method, path, body, tokenB)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("%s: expected 404 for other tenant, got %d", name, resp.StatusCode)
		}
	}
	got404("get", "GET", fmt.Sprintf("/api/interview-prep/%d", prepID), nil)
	got404("answer", "POST",
		fmt.Sprintf("/api/interview-prep/%d/questions/%d/answers", prepID, created.Preparation.Questions[0].ID),
		map[string]string{"answer": "invadir", "self_assessment": "medium"})
	got404("verify", "POST", fmt.Sprintf("/api/interview-prep/%d/verify", prepID), nil)
}