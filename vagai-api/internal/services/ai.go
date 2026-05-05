package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var baseURL = "http://127.0.0.1:1234"

func init() {
	if url := os.Getenv("LMSTUDIO_URL"); url != "" {
		baseURL = url
	}
	fmt.Printf("API AI: Usando LM Studio em %s\n", baseURL)
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type ChatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Error any `json:"error"`
}

func ProcessResumeContent(rawContent string) (string, error) {
	if rawContent == "" {
		return "", fmt.Errorf("conteúdo vazio")
	}

	client := &http.Client{
		Timeout: 180 * time.Second,
	}

	prompt := fmt.Sprintf(`Você é um assistente de recrutamento. Abaixo está o texto extraído de um currículo. 
Limpe o texto, remova caracteres estranhos de conversão e organize as informações principais (Experiência, Habilidades, Educação). 
Mantenha o idioma original do currículo.

CURRÍCULO:
%s`, rawContent)

	messages := []Message{
		{Role: "system", Content: "Você é um especialista em processamento de documentos de RH."},
		{Role: "user", Content: prompt},
	}

	body, err := json.Marshal(ChatRequest{
		Model:    "local-model",
		Messages: messages,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", baseURL+"/v1/chat/completions", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("LM Studio indisponível: %w", err)
	}
	defer resp.Body.Close()

	var result ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.Error != nil {
		return "", fmt.Errorf("erro AI: %v", result.Error)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("sem resposta da AI")
	}

	return result.Choices[0].Message.Content, nil
}

func DiscoverSelectorsWithAI(url string) (map[string]string, error) {
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	prompt := fmt.Sprintf(`Analise a página de vagas de emprego neste URL e retorne os seletores CSS necessários para extrair os dados.

URL: %s

TAREFA:
1. Acesse mentalmente a estrutura HTML de sites de vagas similares
2. Identifique os seletores CSS para:
   - LINKS: elementos <a> que levam às páginas das vagas (seletor para links de vagas)
   - COMPANY: elemento que contém o nome da empresa
   - DESCRIPTION: elemento que contém a descrição da vaga

Responda ESTRITAMENTE em JSON válido no formato:
{
  "selector_links": "seletor css para links",
  "selector_company": "seletor css para empresa", 
  "selector_description": "seletor css para descrição"
}

Exemplos de seletores comuns:
- RemoteOK: "td.company a" para links
- WeWorkRemotely: "section.jobs li a[href^='/remote-jobs/']" para links
- LinkedIn: "a.base-card__full-link" para links`, url)

	messages := []Message{
		{Role: "system", Content: "Você é um especialista em web scraping e extração de dados HTML. Retorne apenas JSON válido com os seletores CSS."},
		{Role: "user", Content: prompt},
	}

	body, err := json.Marshal(ChatRequest{
		Model:    "local-model",
		Messages: messages,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", baseURL+"/v1/chat/completions", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("LM Studio indisponível: %w", err)
	}
	defer resp.Body.Close()

	var result ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Error != nil {
		return nil, fmt.Errorf("erro AI: %v", result.Error)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("sem resposta da AI")
	}

	responseText := result.Choices[0].Message.Content
	responseText = strings.TrimPrefix(responseText, "```json")
	responseText = strings.TrimPrefix(responseText, "```")
	responseText = strings.TrimSuffix(responseText, "```")
	responseText = strings.TrimSpace(responseText)

	var selectors map[string]string
	if err := json.Unmarshal([]byte(responseText), &selectors); err != nil {
		log.Printf("Erro ao parsear seletores: %v, resposta: %s", err, responseText)
		return map[string]string{
			"selector_links":       "a.job-link, a[href*='job']",
			"selector_company":     ".company, .company-name",
			"selector_description": ".description, .job-description",
		}, nil
	}

	if selectors["selector_links"] == "" {
		selectors["selector_links"] = "a.job-link, a[href*='job']"
	}
	if selectors["selector_company"] == "" {
		selectors["selector_company"] = ".company, .company-name"
	}
	if selectors["selector_description"] == "" {
		selectors["selector_description"] = ".description, .job-description"
	}

	return selectors, nil
}

func AnalyzeResumeWithAI(resumeContent string) (map[string]interface{}, error) {
	if resumeContent == "" {
		return nil, fmt.Errorf("conteúdo vazio")
	}

	client := &http.Client{
		Timeout: 180 * time.Second,
	}

	prompt := fmt.Sprintf(`Você é um ANALISTA DE RH SÊNIOR com mais de 15 anos de experiência em recrutamento e seleção de profissionais de tecnologia.
Analise o currículo abaixo de forma completa e profissional.

Para cada seção, siga estas diretrizes:
1. PONTOS FORTES: Identifique habilidades técnicas, experiências relevantes, certificações, tecnologias específicas que o candidato domina e que são valorizadas no mercado.
2. PONTOS DE ATENÇÃO: Identifique gaps de experiência, falta de tecnologias importantes, inconsistências no currículo, áreas que precisam de desenvolvimento.
3. SUGESTÕES DE MELHORIA: Recomendações práticas e específicas para melhorar o currículo - mudanças na formatação, tecnologias a adicionar, formas de evidenciar achievements,etc.

Ao final, forneça uma ANÁLISE COMPLETA com visão geral do perfil, recomendações de melhoria e próximos passos.

CURRÍCULO:
%s

Responda estritamente em formato JSON válido com esta estrutura exata:
{"strengths": ["ponto1", "ponto2"], "weaknesses": ["ponto1", "ponto2"], "suggestions": ["sugestão1", "sugestão2"], "fullAnalysis": "texto da análise completa aqui"}`, resumeContent)

	messages := []Message{
		{Role: "system", Content: "Você é um Analista de RH Sênior especializado em tecnologia. Analise currículos de forma profissional e forneça feedback construtivo. Responda sempre em JSON válido com as chaves exactas: strengths, weaknesses, suggestions, fullAnalysis."},
		{Role: "user", Content: prompt},
	}

	body, err := json.Marshal(ChatRequest{
		Model:    "local-model",
		Messages: messages,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", baseURL+"/v1/chat/completions", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("LM Studio indisponível: %w", err)
	}
	defer resp.Body.Close()

	var result ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Error != nil {
		return nil, fmt.Errorf("erro AI: %v", result.Error)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("sem resposta da AI")
	}

	responseText := result.Choices[0].Message.Content

	// Clean up response - remove markdown code blocks if present
	responseText = strings.TrimPrefix(responseText, "```json")
	responseText = strings.TrimPrefix(responseText, "```")
	responseText = strings.TrimSuffix(responseText, "```")
	responseText = strings.TrimSpace(responseText)

	// Try to parse as JSON
	var analysis map[string]interface{}
	if err := json.Unmarshal([]byte(responseText), &analysis); err != nil {
		// Try to find JSON in response if wrapped in quotes
		first := strings.Index(responseText, "{")
		last := strings.LastIndex(responseText, "}")
		if first != -1 && last != -1 && last > first {
			responseText = responseText[first : last+1]
			if err := json.Unmarshal([]byte(responseText), &analysis); err != nil {
				log.Printf("Erro ao parsear JSON: %v", err)
				analysis = map[string]interface{}{
					"fullAnalysis": responseText,
					"strengths":    []string{},
					"weaknesses":   []string{},
					"suggestions":   []string{},
				}
			}
		} else {
			analysis = map[string]interface{}{
				"fullAnalysis": responseText,
				"strengths":    []string{},
				"weaknesses":   []string{},
				"suggestions":   []string{},
			}
		}
	}

	return analysis, nil
}
