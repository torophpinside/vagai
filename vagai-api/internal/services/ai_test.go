package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAI_SystemPromptLanguage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		systemPromptFound := false
		for _, msg := range req.Messages {
			if msg.Role == "system" {
				if strings.Contains(msg.Content, "Português do Brasil") || strings.Contains(msg.Content, "PT-BR") {
					systemPromptFound = true
					break
				}
			}
		}

		if !systemPromptFound {
			t.Errorf("System prompt does not contain PT-BR instruction. Found messages: %+v", req.Messages)
		}

		resp := ChatResponse{
			Choices: []struct {
				Message Message `json:"message"`
			}{{Message: Message{Role: "assistant", Content: "Resposta em PT-BR"}}},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	originalBaseURL := baseURL
	baseURL = server.URL
	defer func() { baseURL = originalBaseURL }()

	_, err := AnalyzeResumeWithAI("Test English Resume Content")
	if err != nil {
		t.Fatalf("AnalyzeResumeWithAI failed: %v", err)
	}
}
