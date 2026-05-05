package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProcessResumeContent_EmptyContent(t *testing.T) {
	_, err := ProcessResumeContent("")
	if err == nil {
		t.Error("Expected error for empty content, got nil")
	}
}

func TestAnalyzeResumeWithAI_EmptyContent(t *testing.T) {
	_, err := AnalyzeResumeWithAI("")
	if err == nil {
		t.Error("Expected error for empty content, got nil")
	}
}

func TestProcessResumeContent_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	originalBaseURL := baseURL
	baseURL = server.URL
	defer func() { baseURL = originalBaseURL }()

	_, err := ProcessResumeContent("test content")
	if err == nil {
		t.Error("Expected error for server error, got nil")
	}
}

func TestProcessResumeContent_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("Expected /v1/chat/completions path, got %s", r.URL.Path)
		}
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", contentType)
		}

		response := ChatResponse{
			Choices: []struct {
				Message Message `json:"message"`
			}{
				{
					Message: Message{
						Role:    "assistant",
						Content: "Processed resume content here",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	originalBaseURL := baseURL
	baseURL = server.URL
	defer func() { baseURL = originalBaseURL }()

	result, err := ProcessResumeContent("John Doe - Software Engineer")
	if err != nil {
		t.Fatalf("ProcessResumeContent() error = %v", err)
	}
	if result != "Processed resume content here" {
		t.Errorf("ProcessResumeContent() = %v, expected 'Processed resume content here'", result)
	}
}

func TestAnalyzeResumeWithAI_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		analysis := map[string]interface{}{
			"strengths":    []string{"5+ years Go experience", "Strong API design skills"},
			"weaknesses":   []string{"No cloud experience mentioned"},
			"suggestions":  []string{"Add AWS certification", "Include metrics"},
			"fullAnalysis": "Strong candidate with solid backend experience",
		}

		response := ChatResponse{
			Choices: []struct {
				Message Message `json:"message"`
			}{
				{
					Message: Message{
						Role:    "assistant",
						Content: toJSON(analysis),
					},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	originalBaseURL := baseURL
	baseURL = server.URL
	defer func() { baseURL = originalBaseURL }()

	result, err := AnalyzeResumeWithAI("Resume content here")
	if err != nil {
		t.Fatalf("AnalyzeResumeWithAI() error = %v", err)
	}

	if _, ok := result["strengths"]; !ok {
		t.Error("Expected 'strengths' key in result")
	}
	if _, ok := result["weaknesses"]; !ok {
		t.Error("Expected 'weaknesses' key in result")
	}
	if _, ok := result["suggestions"]; !ok {
		t.Error("Expected 'suggestions' key in result")
	}
	if _, ok := result["fullAnalysis"]; !ok {
		t.Error("Expected 'fullAnalysis' key in result")
	}
}

func TestAnalyzeResumeWithAI_MarkdownResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := ChatResponse{
			Choices: []struct {
				Message Message `json:"message"`
			}{
				{
					Message: Message{
						Role:    "assistant",
						Content: "```json\n{\"strengths\": [\"Go expertise\"], \"weaknesses\": [], \"suggestions\": [], \"fullAnalysis\": \"Good profile\"}\n```",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	originalBaseURL := baseURL
	baseURL = server.URL
	defer func() { baseURL = originalBaseURL }()

	result, err := AnalyzeResumeWithAI("Resume content")
	if err != nil {
		t.Fatalf("AnalyzeResumeWithAI() error = %v", err)
	}

	strengths, ok := result["strengths"].([]interface{})
	if !ok {
		t.Errorf("Expected strengths to be array, got %T", result["strengths"])
	}
	if len(strengths) != 1 {
		t.Errorf("Expected 1 strength, got %d", len(strengths))
	}
}

func TestAnalyzeResumeWithAI_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := ChatResponse{
			Choices: []struct {
				Message Message `json:"message"`
			}{
				{
					Message: Message{
						Role:    "assistant",
						Content: "This is not valid JSON at all",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	originalBaseURL := baseURL
	baseURL = server.URL
	defer func() { baseURL = originalBaseURL }()

	result, err := AnalyzeResumeWithAI("Resume content")
	if err != nil {
		t.Fatalf("AnalyzeResumeWithAI() error = %v", err)
	}

	if _, ok := result["fullAnalysis"]; !ok {
		t.Error("Expected fallback 'fullAnalysis' key in result")
	}
}

func TestAnalyzeResumeWithAI_EmbeddedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := ChatResponse{
			Choices: []struct {
				Message Message `json:"message"`
			}{
				{
					Message: Message{
						Role:    "assistant",
						Content: "Here is the analysis: {\"strengths\": [\"Python\"], \"weaknesses\": [], \"suggestions\": [], \"fullAnalysis\": \"Good\"}",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	originalBaseURL := baseURL
	baseURL = server.URL
	defer func() { baseURL = originalBaseURL }()

	result, err := AnalyzeResumeWithAI("Resume content")
	if err != nil {
		t.Fatalf("AnalyzeResumeWithAI() error = %v", err)
	}

	if _, ok := result["strengths"]; !ok {
		t.Error("Expected 'strengths' key in result from embedded JSON")
	}
}

func TestProcessResumeContent_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := ChatResponse{
			Error: map[string]interface{}{
				"message": "Model not found",
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	originalBaseURL := baseURL
	baseURL = server.URL
	defer func() { baseURL = originalBaseURL }()

	_, err := ProcessResumeContent("test content")
	if err == nil {
		t.Error("Expected error for API error response, got nil")
	}
}

func TestProcessResumeContent_NoChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := ChatResponse{
			Choices: []struct {
				Message Message `json:"message"`
			}{},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	originalBaseURL := baseURL
	baseURL = server.URL
	defer func() { baseURL = originalBaseURL }()

	_, err := ProcessResumeContent("test content")
	if err == nil {
		t.Error("Expected error for empty choices, got nil")
	}
}

func toJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
