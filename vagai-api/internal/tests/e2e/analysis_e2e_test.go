package e2e

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

// TestPerQuestionAnalysis_Wiring cobre a correção por IA de cada resposta:
// responder tudo -> verificar -> cada pergunta respondida ganha uma linha de
// análise exposta em GET /interview-prep/:id; nova correção substitui.
func TestPerQuestionAnalysis_Wiring(t *testing.T) {
	token, _ := registerUser(t, "Analysis User", "analysis@example.com", "password123", "Analysis Org")
	orgID := getOrgID(t, token)

	matchID := setupAppliedMatch(t, token, orgID,
		"https://example.com/job/analysis", "Backend Developer", fixedCountTechDescription)

	prep := createFixedCountPrep(t, token, matchID, "")
	prepID := prep.Preparation.ID
	questions := prep.Preparation.Questions
	if len(questions) == 0 {
		t.Fatal("expected questions in preparation")
	}

	// Responde todas as perguntas.
	for _, q := range questions {
		resp := doRequest(t, "POST",
			fmt.Sprintf("/api/interview-prep/%d/questions/%d/answers", prepID, q.ID),
			map[string]string{"answer": "resposta de teste para correção", "self_assessment": "medium"}, token)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("POST answer for question %d failed: status=%d", q.ID, resp.StatusCode)
		}
		resp.Body.Close()
	}

	// Pede a correção.
	resp := doRequest(t, "POST", fmt.Sprintf("/api/interview-prep/%d/verify", prepID), nil, token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /verify failed: status=%d", resp.StatusCode)
	}
	resp.Body.Close()

	// Aguarda as linhas de análise por pergunta (pipeline assíncrono).
	analyses := waitForAnalyses(t, token, prepID, len(questions))

	for _, q := range questions {
		a, ok := analyses[fmt.Sprint(q.ID)]
		if !ok {
			t.Errorf("missing analysis row for answered question %d", q.ID)
			continue
		}
		if a.Status == "" {
			t.Errorf("analysis for question %d has empty status", q.ID)
		}
		// Com LM Studio fora do ar o status é error; com IA ativa, completed + texto.
		if a.Status == "completed" && a.Text == "" {
			t.Errorf("completed analysis for question %d has empty text", q.ID)
		}
	}

	// Nova correção substitui: continua exatamente 1 linha por pergunta.
	resp = doRequest(t, "POST", fmt.Sprintf("/api/interview-prep/%d/verify", prepID), nil, token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("second POST /verify failed: status=%d", resp.StatusCode)
	}
	resp.Body.Close()

	analyses = waitForAnalyses(t, token, prepID, len(questions))
	if len(analyses) != len(questions) {
		t.Fatalf("after re-verify: expected exactly %d analysis rows (replaced), got %d",
			len(questions), len(analyses))
	}

	// Item 8 — histórico anexa as tentativas concluídas.
	detail := getPrepDetail(t, token, prepID)
	if len(detail.Preparation.History) < 1 {
		t.Fatalf("expected at least 1 history entry after verifications, got %d", len(detail.Preparation.History))
	}
	for _, h := range detail.Preparation.History {
		if h.Score < 0 || h.Score > 10 {
			t.Errorf("history score out of range: %v", h.Score)
		}
	}

	// Item 1 — regenerar limpa verificação, feedbacks e histórico.
	resp = doRequest(t, "POST", "/api/interview-prep", map[string]interface{}{"match_id": matchID}, token)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST /interview-prep (regenerate) failed: status=%d", resp.StatusCode)
	}
	resp.Body.Close()

	detail = getPrepDetail(t, token, prepID)
	if detail.Preparation.Verification != nil {
		t.Errorf("after regenerate: expected no verification, got %+v", detail.Preparation.Verification)
	}
	if len(detail.Preparation.Analyses) != 0 {
		t.Errorf("after regenerate: expected no analyses, got %d", len(detail.Preparation.Analyses))
	}
	if len(detail.Preparation.History) != 0 {
		t.Errorf("after regenerate: expected no history, got %d", len(detail.Preparation.History))
	}
}

// Item 4 — correção parcial: 1 resposta basta para agendar a correção.
func TestPerQuestionAnalysis_PartialVerify(t *testing.T) {
	token, _ := registerUser(t, "Partial User", "partial-verify@example.com", "password123", "Partial Org")
	orgID := getOrgID(t, token)

	matchID := setupAppliedMatch(t, token, orgID,
		"https://example.com/job/partial-verify", "Backend Developer", fixedCountTechDescription)

	prep := createFixedCountPrep(t, token, matchID, "")
	prepID := prep.Preparation.ID
	questions := prep.Preparation.Questions
	if len(questions) < 2 {
		t.Fatalf("expected at least 2 questions, got %d", len(questions))
	}

	// Responde só a primeira.
	resp := doRequest(t, "POST",
		fmt.Sprintf("/api/interview-prep/%d/questions/%d/answers", prepID, questions[0].ID),
		map[string]string{"answer": "resposta parcial", "self_assessment": "medium"}, token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST answer failed: status=%d", resp.StatusCode)
	}
	resp.Body.Close()

	// Correção parcial deve agendar (antes: 400 exigindo 100%).
	resp = doRequest(t, "POST", fmt.Sprintf("/api/interview-prep/%d/verify", prepID), nil, token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("partial POST /verify failed: status=%d (want 200)", resp.StatusCode)
	}
	resp.Body.Close()

	analyses := waitForAnalyses(t, token, prepID, 1)
	if len(analyses) < 1 {
		t.Fatalf("expected analysis for the answered question, got %d rows", len(analyses))
	}
}

type analysisDTO struct {
	ID         uint    `json:"id"`
	QuestionID uint    `json:"question_id"`
	Text       string  `json:"text"`
	Score      float64 `json:"score"`
	Status     string  `json:"status"`
	Source     string  `json:"source"`
}

type prepDetailDTO struct {
	Preparation struct {
		Analyses map[string]analysisDTO `json:"analyses"`
		History  []struct {
			Score      float64 `json:"score"`
			Source     string  `json:"source"`
			VerifiedAt string  `json:"verified_at"`
		} `json:"history"`
		Verification *struct {
			Status string `json:"status"`
		} `json:"verification"`
	} `json:"preparation"`
}

func getPrepDetail(t *testing.T, token string, prepID uint) prepDetailDTO {
	t.Helper()

	resp := doRequest(t, "GET", fmt.Sprintf("/api/interview-prep/%d", prepID), nil, token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/interview-prep/%d failed: status=%d", prepID, resp.StatusCode)
	}
	var detail prepDetailDTO
	parseBody(t, resp, &detail)
	return detail
}

// T001 [US1] Regressão: salvar resposta idêntica (navegar na sabatina) em prep
// verificada NÃO agenda nova correção — status segue verified e nada é recriado.
func TestNoReverifyOnIdenticalSave(t *testing.T) {
	token, _ := registerUser(t, "NoReverify User", "noreverify@example.com", "password123", "NoReverify Org")
	orgID := getOrgID(t, token)

	matchID := setupAppliedMatch(t, token, orgID,
		"https://example.com/job/no-reverify", "Backend Developer", fixedCountTechDescription)

	prep := createFixedCountPrep(t, token, matchID, "")
	prepID := prep.Preparation.ID
	questions := prep.Preparation.Questions

	const answerText = "resposta final de teste"
	for _, q := range questions {
		resp := doRequest(t, "POST",
			fmt.Sprintf("/api/interview-prep/%d/questions/%d/answers", prepID, q.ID),
			map[string]string{"answer": answerText, "self_assessment": "medium"}, token)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("POST answer for question %d failed: status=%d", q.ID, resp.StatusCode)
		}
		resp.Body.Close()
	}

	resp := doRequest(t, "POST", fmt.Sprintf("/api/interview-prep/%d/verify", prepID), nil, token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /verify failed: status=%d", resp.StatusCode)
	}
	resp.Body.Close()

	waitForVerified(t, token, prepID)
	waitForAnalyses(t, token, prepID, len(questions))
	before := getPrepDetail(t, token, prepID)
	historyLen := len(before.Preparation.History)

	// Save IDÊNTICO (simula navegar/voltar na sabatina com autosave).
	resp = doRequest(t, "POST",
		fmt.Sprintf("/api/interview-prep/%d/questions/%d/answers", prepID, questions[0].ID),
		map[string]string{"answer": answerText, "self_assessment": "medium"}, token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST identical answer failed: status=%d", resp.StatusCode)
	}
	resp.Body.Close()

	// O claim de nova verificação é síncrono: se re-agendou, o GET imediato
	// mostra running/outdated em vez de verified.
	after := getPrepDetail(t, token, prepID)
	if after.Preparation.Verification == nil || after.Preparation.Verification.Status != "verified" {
		status := "<nil>"
		if after.Preparation.Verification != nil {
			status = after.Preparation.Verification.Status
		}
		t.Fatalf("identical save re-triggered verification: status=%s (want verified)", status)
	}

	// Sem pipeline eventual: histórico não cresce.
	time.Sleep(5 * time.Second)
	after = getPrepDetail(t, token, prepID)
	if len(after.Preparation.History) != historyLen {
		t.Fatalf("identical save appended history: %d -> %d rows (want unchanged)",
			historyLen, len(after.Preparation.History))
	}
}

// T004 [US2]: editar o texto de uma resposta em prep verificada marca
// outdated SEM re-corrigir sozinha (correção manual posterior decide).
func TestEditMarksOutdatedWithoutReverify(t *testing.T) {
	token, _ := registerUser(t, "EditOutdated User", "edit-outdated@example.com", "password123", "EditOutdated Org")
	orgID := getOrgID(t, token)

	matchID := setupAppliedMatch(t, token, orgID,
		"https://example.com/job/edit-outdated", "Backend Developer", fixedCountTechDescription)

	prep := createFixedCountPrep(t, token, matchID, "")
	prepID := prep.Preparation.ID
	questions := prep.Preparation.Questions

	const answerText = "resposta original"
	for _, q := range questions {
		resp := doRequest(t, "POST",
			fmt.Sprintf("/api/interview-prep/%d/questions/%d/answers", prepID, q.ID),
			map[string]string{"answer": answerText, "self_assessment": "medium"}, token)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("POST answer for question %d failed: status=%d", q.ID, resp.StatusCode)
		}
		resp.Body.Close()
	}

	resp := doRequest(t, "POST", fmt.Sprintf("/api/interview-prep/%d/verify", prepID), nil, token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /verify failed: status=%d", resp.StatusCode)
	}
	resp.Body.Close()

	waitForVerified(t, token, prepID)
	before := getPrepDetail(t, token, prepID)
	historyLen := len(before.Preparation.History)

	// Edita o TEXTO da resposta.
	resp = doRequest(t, "POST",
		fmt.Sprintf("/api/interview-prep/%d/questions/%d/answers", prepID, questions[0].ID),
		map[string]string{"answer": "resposta editada de verdade", "self_assessment": "medium"}, token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST edited answer failed: status=%d", resp.StatusCode)
	}
	resp.Body.Close()

	// Marca outdated, mas NÃO re-corrige sozinha.
	after := getPrepDetail(t, token, prepID)
	if after.Preparation.Verification == nil || after.Preparation.Verification.Status != "outdated" {
		status := "<nil>"
		if after.Preparation.Verification != nil {
			status = after.Preparation.Verification.Status
		}
		t.Fatalf("edited answer: status=%s (want outdated)", status)
	}

	time.Sleep(5 * time.Second)
	after = getPrepDetail(t, token, prepID)
	if st := after.Preparation.Verification.Status; st != "outdated" {
		t.Fatalf("edited answer auto re-verified: status=%s (want outdated)", st)
	}
	if len(after.Preparation.History) != historyLen {
		t.Fatalf("edited answer appended history: %d -> %d rows (want unchanged)",
			historyLen, len(after.Preparation.History))
	}
}

func waitForVerified(t *testing.T, token string, prepID uint) {
	t.Helper()

	deadline := time.Now().Add(60 * time.Second)
	for {
		detail := getPrepDetail(t, token, prepID)
		if detail.Preparation.Verification != nil && detail.Preparation.Verification.Status == "verified" {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for verified status on prep %d", prepID)
		}
		time.Sleep(time.Second)
	}
}

func waitForAnalyses(t *testing.T, token string, prepID uint, want int) map[string]analysisDTO {
	t.Helper()

	deadline := time.Now().Add(30 * time.Second)
	for {
		resp := doRequest(t, "GET", fmt.Sprintf("/api/interview-prep/%d", prepID), nil, token)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("GET /api/interview-prep/%d failed: status=%d", prepID, resp.StatusCode)
		}
		var detail struct {
			Preparation struct {
				Analyses map[string]analysisDTO `json:"analyses"`
			} `json:"preparation"`
		}
		parseBody(t, resp, &detail)
		if len(detail.Preparation.Analyses) >= want {
			return detail.Preparation.Analyses
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %d analysis rows, got %d", want, len(detail.Preparation.Analyses))
		}
		time.Sleep(time.Second)
	}
}
