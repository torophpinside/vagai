package lmstudio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

var baseURL = "http://127.0.0.1:1234"

func SetBaseURL(url string) {
	if url != "" {
		baseURL = url
	}
}

func GetBaseURL() string {
	return baseURL
}

type EmbeddingRequest struct {
	Input string `json:"input"`
}

type EmbeddingResponse struct {
	Model string `json:"model"`
	Data  []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
	Error any `json:"error"`
}

func GetEmbedding(text string) ([]float64, error) {
	if text == "" {
		return make([]float64, 0), nil
	}

	client := &http.Client{
		Timeout: 120 * time.Second,
	}

	body, err := json.Marshal(EmbeddingRequest{Input: text})
	if err != nil {
		return nil, fmt.Errorf("erro ao criar request: %w", err)
	}

	req, err := http.NewRequest("POST", baseURL+"/v1/embeddings", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar no LM Studio: %w", err)
	}
	defer resp.Body.Close()

	var result EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("erro do LM Studio: %v", result.Error)
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("nenhum embedding retornado")
	}

	return result.Data[0].Embedding, nil
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Error any `json:"error"`
}

func Chat(prompt string, systemPrompt string) (string, error) {
	client := &http.Client{
		Timeout: 300 * time.Second,
	}

	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: prompt},
	}

	body, err := json.Marshal(ChatRequest{
		Model:    "local-model",
		Messages: messages,
	})
	if err != nil {
		return "", fmt.Errorf("erro ao criar request: %w", err)
	}

	req, err := http.NewRequest("POST", baseURL+"/v1/chat/completions", bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("erro ao criar request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("erro ao conectar no LM Studio: %w", err)
	}
	defer resp.Body.Close()

	var result ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("erro ao decodificar resposta: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("erro do LM Studio: %v", result.Error)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("nenhuma resposta retornada")
	}

	return result.Choices[0].Message.Content, nil
}

func init() {
	if url := os.Getenv("LMSTUDIO_URL"); url != "" {
		baseURL = url
	}
	// Using fmt instead of log to avoid circular dependency or init order issues
	fmt.Printf("AI Agent: Usando LM Studio em %s\n", baseURL)
}
