package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockScoreServer(t *testing.T, content string) string {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ChatResponse{
			Choices: []struct {
				Message Message `json:"message"`
			}{{Message: Message{Role: "assistant", Content: content}}},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	t.Cleanup(server.Close)
	t.Setenv("LMSTUDIO_URL", server.URL)
	return server.URL
}

// Item 3 — JSON no contrato gera nota real + feedback.
func TestAnalyzeQuestionAnswerScored_JSONContract(t *testing.T) {
	mockScoreServer(t, `{"score": 7.5, "feedback": "Bom domínio de goroutines; aprofunde channels."}`)

	score, feedback, err := AnalyzeQuestionAnswerScored("O que são goroutines?", "Resposta do candidato")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if score != 7.5 {
		t.Errorf("score = %v, want 7.5", score)
	}
	if feedback == "" {
		t.Error("expected non-empty feedback")
	}
}

// Item 3 — fora do contrato mas com NOTA explícita: usa o texto + nota extraída.
func TestAnalyzeQuestionAnswerScored_NotaFallback(t *testing.T) {
	mockScoreServer(t, "Feedback corrido sobre a resposta. NOTA 8")

	score, feedback, err := AnalyzeQuestionAnswerScored("Pergunta?", "Resposta")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if score != 8 {
		t.Errorf("score = %v, want 8", score)
	}
	if feedback == "" {
		t.Error("expected non-empty feedback")
	}
}

// Item 3 — sem nota parseável: erro para o chamador decidir (retry/fallback).
func TestAnalyzeQuestionAnswerScored_InvalidContract(t *testing.T) {
	mockScoreServer(t, "texto corrido sem nota nenhuma")

	if _, _, err := AnalyzeQuestionAnswerScored("Pergunta?", "Resposta"); err == nil {
		t.Error("expected error for contract-less response without score")
	}
}

// Item 3 — wrapper legado mantém comportamento (só feedback).
func TestAnalyzeQuestionAnswer_Wrapper(t *testing.T) {
	mockScoreServer(t, `{"score": 6, "feedback": "Feedback válido."}`)

	feedback, err := AnalyzeQuestionAnswer("Pergunta?", "Resposta")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if feedback == "" {
		t.Error("expected non-empty feedback")
	}
}

// Item 3 — parseScoreNumber aceita ponto e vírgula, rejeita fora da faixa.
func TestParseScoreNumber(t *testing.T) {
	for in, want := range map[string]float64{"7": 7, "7.5": 7.5, "7,5": 7.5, "10": 10, "0": 0} {
		if got, err := parseScoreNumber(in); err != nil || got != want {
			t.Errorf("parseScoreNumber(%q) = %v, %v; want %v, nil", in, got, err, want)
		}
	}
	for _, in := range []string{"11", "-1", "abc", ""} {
		if _, err := parseScoreNumber(in); err == nil {
			t.Errorf("parseScoreNumber(%q) should fail", in)
		}
	}
}
