package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
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
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
}

type ChatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Error any `json:"error"`
}

func callLMStudio(prompt string) (string, error) {
	client := &http.Client{
		Timeout: 240 * time.Second,
	}

	messages := []Message{
		{Role: "system", Content: "Você é um assistente de IA especializado em RH e recrutamento."},
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

func ProcessResumeContent(rawContent string) (string, error) {
	if rawContent == "" {
		return "", fmt.Errorf("conteúdo vazio")
	}

	client := &http.Client{
		Timeout: 240 * time.Second,
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

	// Busca o HTML real da página
	htmlContent, err := fetchHTML(url)
	if err != nil {
		log.Printf("Erro ao buscar HTML de %s: %v, usando fallback", url, err)
		return getDefaultSelectors(), nil
	}

	// Limita o tamanho do HTML para evitar tokens excessivos
	if len(htmlContent) > 50000 {
		htmlContent = htmlContent[:50000]
	}

	prompt := fmt.Sprintf(`Analise o HTML abaixo de uma página de vagas de emprego e retorne os seletores CSS necessários para extrair os dados.

URL: %s

HTML DA PÁGINA:
%s

TAREFA:
Identifique os seletores CSS para:
- LINKS: elementos <a> que levam às páginas das vagas (seletor para links de vagas)
- COMPANY: elemento que contém o nome da empresa
- DESCRIPTION: elemento que contém a descrição da vaga

Responda ESTRITAMENTE em JSON válido no formato:
{
  "selector_links": "seletor css para links",
  "selector_company": "seletor css para empresa",
  "selector_description": "seletor css para descrição"
}`, url, htmlContent)

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

// fetchHTML busca o conteúdo HTML de uma URL
func fetchHTML(rawURL string) (string, error) {
	// SSRF guard: bloqueia esquemas não-HTTP e hosts apontando para redes
	// privadas/loopback/link-local antes de qualquer conexão ser aberta.
	if err := validateFetchURL(rawURL); err != nil {
		return "", err
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return fmt.Errorf("muitos redirecionamentos")
			}
			// Cada hop de redirecionamento também precisa passar pela mesma
			// validação, para impedir redirect para 127.0.0.1/metadados.
			return validateFetchURL(req.URL.String())
		},
	}

	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// getDefaultSelectors retorna seletores padrão quando a IA falha
func getDefaultSelectors() map[string]string {
	return map[string]string{
		"selector_links":       "a[href*='job'], a.job-link, article a",
		"selector_company":     ".company, .company-name, [class*='company']",
		"selector_description": ".description, .job-description, article, main",
	}
}

func AnalyzeResumeWithAI(resumeContent string) (map[string]interface{}, error) {
	if resumeContent == "" {
		return nil, fmt.Errorf("conteúdo vazio")
	}

	client := &http.Client{
		Timeout: 240 * time.Second,
	}

	prompt := fmt.Sprintf(`Analise o currículo abaixo como um Analista de RH Sênior.

CURRÍCULO:
%s

Escreva sua análise usando EXATAMENTE estes cabeçalhos de seção:

PONTOS FORTES:
- item 1
- item 2

PONTOS DE ATENÇÃO:
- item 1
- item 2

SUGESTÕES DE MELHORIA:
- item 1
- item 2

ANÁLISE COMPLETA:
texto livre aqui`, resumeContent)

	messages := []Message{
		{Role: "system", Content: "Você é um Analista de RH Sênior especializado em tecnologia. Use os cabeçalhos PONTOS FORTES, PONTOS DE ATENÇÃO, SUGESTÕES DE MELHORIA e ANÁLISE COMPLETA. Use bullets (-) para listar itens."},
		{Role: "user", Content: prompt},
	}

	body, err := json.Marshal(ChatRequest{
		Model:       "local-model",
		Messages:    messages,
		Temperature: 0.1,
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

	rawText := strings.TrimSpace(result.Choices[0].Message.Content)

	// A resposta pode vir em JSON estruturado (com ou sem code fences). Se for,
	// usa direto — cobre respostas no formato {"strengths": [...], ...}.
	if parsed, ok := parseStructuredAnalysis(rawText); ok {
		return parsed, nil
	}

	analysis := map[string]interface{}{
		"strengths":   extractBullets(rawText, "PONTOS FORTES", "PONTOS DE ATENÇÃO"),
		"weaknesses":  extractBullets(rawText, "PONTOS DE ATENÇÃO", "SUGESTÕES DE MELHORIA"),
		"suggestions": extractBullets(rawText, "SUGESTÕES DE MELHORIA", "ANÁLISE COMPLETA"),
	}

	analiseIdx := strings.Index(rawText, "ANÁLISE COMPLETA")
	if analiseIdx != -1 {
		after := rawText[analiseIdx+len("ANÁLISE COMPLETA"):]
		analysis["fullAnalysis"] = strings.TrimSpace(after)
	} else {
		analysis["fullAnalysis"] = rawText
	}

	return analysis, nil
}

// parseStructuredAnalysis tenta extrair um objeto JSON da resposta da IA.
// Suporta respostas puras, em code fences (```json ... ```) ou embutidas em texto.
func parseStructuredAnalysis(rawText string) (map[string]interface{}, bool) {
	s := strings.TrimSpace(rawText)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	s = strings.TrimSpace(s)

	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start == -1 || end == -1 || end <= start {
		return nil, false
	}

	var analysis map[string]interface{}
	if err := json.Unmarshal([]byte(s[start:end+1]), &analysis); err != nil {
		return nil, false
	}

	if _, ok := analysis["fullAnalysis"]; !ok {
		if _, ok := analysis["strengths"]; !ok {
			return nil, false
		}
	}

	return analysis, true
}

func ExtractJobFromURL(url string) (map[string]string, error) {
	htmlContent, err := fetchHTML(url)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar página: %w", err)
	}

	pageText := stripHTMLTags(htmlContent)

	lines := strings.Split(pageText, "\n")
	var cleaned []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleaned = append(cleaned, line)
		}
	}
	pageText = strings.Join(cleaned, "\n")

	aiInput := pageText
	if len(aiInput) > 6000 {
		aiInput = aiInput[:6000]
	}

	prompt := fmt.Sprintf(`Extraia APENAS o titulo e a empresa da vaga abaixo.

TEXTO:
%s

URL: %s

Responda em JSON:
{"title": "titulo", "company": "empresa"}

Se nao encontrar um campo, deixe string vazia.`, aiInput, url)

	messages := []Message{
		{Role: "system", Content: "Você extrai titulo e empresa de paginas de emprego. Retorne apenas JSON."},
		{Role: "user", Content: prompt},
	}

	client := &http.Client{Timeout: 120 * time.Second}
	body, err := json.Marshal(ChatRequest{
		Model:       "local-model",
		Messages:    messages,
		Temperature: 0.1,
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

	var data map[string]string
	if err := json.Unmarshal([]byte(responseText), &data); err != nil {
		return nil, fmt.Errorf("erro ao parsear JSON da AI: %w", err)
	}

	if data["title"] == "" {
		title, ok := extractTitleFallback(htmlContent)
		if ok {
			data["title"] = title
		}
	}

	descMax := 50000
	if len(pageText) > descMax {
		pageText = pageText[:descMax]
	}
	data["description"] = pageText

	return data, nil
}

func stripHTMLTags(html string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return html
	}

	doc.Find("script, style, noscript, iframe, svg, meta, link, head").Remove()

	var text string
	doc.Find("body").Each(func(_ int, s *goquery.Selection) {
		text = s.Text()
	})
	if text == "" {
		text = doc.Text()
	}

	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&amp;", "&")
	reSpace := regexp.MustCompile(`\s+`)
	text = reSpace.ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

func extractTitleFallback(html string) (string, bool) {
	re := regexp.MustCompile(`<title[^>]*>([^<]+)</title>`)
	matches := re.FindStringSubmatch(html)
	if len(matches) >= 2 {
		title := strings.TrimSpace(matches[1])
		if title != "" {
			parts := strings.Split(title, "|")
			title = strings.TrimSpace(parts[0])
			parts2 := strings.Split(title, "-")
			title = strings.TrimSpace(parts2[0])
			if title != "" {
				return title, true
			}
		}
	}
	return "", false
}

func extractBullets(text, sectionStart, sectionEnd string) []string {
	start := strings.Index(text, sectionStart)
	if start == -1 {
		return []string{}
	}
	content := text[start+len(sectionStart):]
	if sectionEnd != "" {
		end := strings.Index(content, sectionEnd)
		if end != -1 {
			content = content[:end]
		}
	}
	lines := strings.Split(content, "\n")
	var items []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "- ") {
			items = append(items, strings.TrimPrefix(line, "- "))
		} else if strings.HasPrefix(line, "* ") {
			items = append(items, strings.TrimPrefix(line, "* "))
		} else if matched, _ := regexp.MatchString(`^\d+[\.\)]\s`, line); matched {
			re := regexp.MustCompile(`^\d+[\.\)]\s`)
			items = append(items, re.ReplaceAllString(line, ""))
		}
	}
	if len(items) == 0 {
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "**") && !strings.HasPrefix(line, "#") {
				items = append(items, line)
			}
		}
	}
	return items
}

func GenerateInterviewPrep(jobDescription string, resumeContent string) (string, error) {
	prompt := fmt.Sprintf(`You are an expert interview coach. Based on the following job description and candidate resume, generate a comprehensive interview preparation guide.

Job Description:
%s

Candidate Resume:
%s

Please provide:
1. Top 5 most likely technical questions and suggested high-impact answers.
2. Top 3 behavioral questions based on the job's requirements and the candidate's experience.
3. A "Red Flag" section: gaps between the resume and job requirements and how to address them.
4. 3 strategic questions the candidate should ask the interviewer to show deep interest and competence.

Format the output in clear Markdown.`, jobDescription, resumeContent)

	result, err := callLMStudio(prompt)
	if err != nil {
		return "", err
	}

	return result, nil
}
