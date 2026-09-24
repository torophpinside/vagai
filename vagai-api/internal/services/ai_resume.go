package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

func ParseResumeFields(rawText string) (ResumeData, error) {
	if rawText == "" {
		return ResumeData{}, fmt.Errorf("texto vazio")
	}

	client := &http.Client{
		Timeout: 120 * time.Second,
	}

	prompt := fmt.Sprintf(`Analise o currículo fornecido e extraia as informações estruturadas com precisão.
Retorne APENAS um objeto JSON válido no formato especificado abaixo, sem texto adicional.

CURRICULO:
%s

Retorne EXATAMENTE neste formato JSON:
{
  "personal_info": {
    "name": "nome completo",
    "email": "email",
    "phone": "telefone",
    "location": "cidade, estado",
    "linkedin": "linkedin url ou vazio",
    "website": "website url ou vazio"
  },
  "summary": "resumo profissional ou objetivo",
  "experience": [
    {
      "company": "empresa",
      "role": "cargo",
      "start_date": "MM/YYYY",
      "end_date": "MM/YYYY ou Presente",
      "description": "descricao detalhada das atividades do cargo"
    }
  ],
  "education": [
    {
      "institution": "instituicao",
      "degree": "graduacao/mestrado/etc",
      "field": "area de estudo",
      "start_date": "MM/YYYY",
      "end_date": "MM/YYYY",
      "notes": "observacoes adicionais"
    }
  ],
  "skills": ["habilidade1", "habilidade2"],
  "languages": ["idioma1", "idioma2"],
  "certifications": ["certificacao1"]
}

REGRAS DE EXTRAÇÃO:
1. PRECISÃO: Extraia apenas informações explicitamente presentes no currículo
2. NÃO INVENTAR: Se não tiver certeza, deixe o campo vazio
3. SKILLS:
   - Inclua habilidades técnicas (ex: Java, Python, SQL) e comportamentais (ex: Liderança, Comunicação)
   - Normalize para MINÚSCULAS e SEM ACENTOS (ex: "java", "lideranca")
   - Separe por competência distinta (evite combinar múltiplas skills em uma string)
4. EXPERIENCE/EDUCATION:
   - Para cada entrada, preencha todos os subcampos disponíveis
   - Use formato MM/YYYY para datas
   - Em description, inclua atividades reais realizadas
5. LANGUAGES: Use formato "Idioma - Nível" (ex: "Inglês - Avançado")
6. CAMPOS VAZIOS: Array vazio [] para listas, objeto com strings vazias para personal_info
7. IDIOMA: Mantenha o idioma original do currículo nos campos de texto
8. AMBIGUIDADE: Quando houver dúvida, escolha a interpretação mais conservadora

EXEMPLOS DE SKILLS NORMALIZADAS:
- "JavaScript" → "javascript"
- "Liderança de Equipes" → "lideranca de equipes"
- "AWS / Azure" → ["aws", "azure"]
- "Inglês Avançado" (em skills) → "ingles avanzado" (mas prefira colocar em languages)

IMPORTANTE: A qualidade do matching depende da precisão desta extração. Seja meticuloso.`, rawText)

	messages := []Message{
		{Role: "system", Content: "Voce e um assistente especializado em extrair dados estruturados de curriculos. Retorne apenas JSON valido, sem texto adicional."},
		{Role: "user", Content: prompt},
	}

	body, err := json.Marshal(ChatRequest{
		Model:       "local-model",
		Messages:    messages,
		Temperature: 0.1,
	})
	if err != nil {
		return ResumeData{}, err
	}

	req, err := http.NewRequest("POST", baseURL+"/v1/chat/completions", bytes.NewBuffer(body))
	if err != nil {
		return ResumeData{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return ParseResumeFieldsFallback(rawText), fmt.Errorf("LM Studio indisponivel: %w", err)
	}
	defer resp.Body.Close()

	var result ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ParseResumeFieldsFallback(rawText), err
	}

	if result.Error != nil {
		return ParseResumeFieldsFallback(rawText), fmt.Errorf("erro AI: %v", result.Error)
	}

	if len(result.Choices) == 0 {
		return ParseResumeFieldsFallback(rawText), fmt.Errorf("sem resposta da AI")
	}

	responseText := result.Choices[0].Message.Content
	responseText = strings.TrimPrefix(responseText, "```json")
	responseText = strings.TrimPrefix(responseText, "```")
	responseText = strings.TrimSuffix(responseText, "```")
	responseText = strings.TrimSpace(responseText)

	var data ResumeData
	if err := json.Unmarshal([]byte(responseText), &data); err != nil {
		log.Printf("Erro ao parsear JSON da AI: %v, resposta: %s", err, responseText)
		return ParseResumeFieldsFallback(rawText), fmt.Errorf("erro ao parsear resposta AI: %w", err)
	}

	return data, nil
}

func ParseResumeFieldsFallback(rawText string) ResumeData {
	data := ResumeData{
		PersonalInfo:   PersonalInfo{},
		Experience:     []ExperienceEntry{},
		Education:      []EducationEntry{},
		Skills:         []string{},
		Languages:      []string{},
		Certifications: []string{},
	}

	lines := strings.Split(rawText, "\n")
	var cleanedLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleanedLines = append(cleanedLines, line)
		}
	}

	// Extract email
	emailRe := regexp.MustCompile(`[\w.+-]+@[\w.-]+\.\w+`)
	if match := emailRe.FindString(rawText); match != "" {
		data.PersonalInfo.Email = match
	}

	// Extract phone (Brazilian patterns)
	phoneRe := regexp.MustCompile(`\(?\d{2}\)?\s*\d{4,5}[-\s]?\d{4}`)
	if match := phoneRe.FindString(rawText); match != "" {
		data.PersonalInfo.Phone = strings.TrimSpace(match)
	}

	// Extract LinkedIn
	linkedinRe := regexp.MustCompile(`linkedin\.com/in/[\w-]+`)
	if match := linkedinRe.FindString(rawText); match != "" {
		data.PersonalInfo.Linkedin = "https://" + match
	}

	// Heuristic: first non-empty line is likely the name
	if len(cleanedLines) > 0 {
		firstLine := cleanedLines[0]
		if !emailRe.MatchString(firstLine) && len(firstLine) < 60 {
			data.PersonalInfo.Name = firstLine
		}
	}

	// Detect sections by keywords
	sectionMap := map[string]int{
		"experiencia":    -1,
		"educacao":       -1,
		"habilidades":    -1,
		"skills":         -1,
		"idiomas":        -1,
		"languages":      -1,
		"certificacoes":  -1,
		"certifications": -1,
		"certificacao":   -1,
		"certificação":   -1,
		"certificações":  -1,
		"resumo":         -1,
		"summary":        -1,
		"objetivo":       -1,
		"profile":        -1,
	}

	for i, line := range cleanedLines {
		lower := strings.ToLower(line)
		for key := range sectionMap {
			if strings.Contains(lower, key) || strings.HasPrefix(lower, key) {
				sectionMap[key] = i
			}
		}
	}

	// Extract summary if found
	for _, key := range []string{"resumo", "summary", "objetivo", "profile"} {
		if idx := sectionMap[key]; idx >= 0 && idx+1 < len(cleanedLines) {
			var summaryLines []string
			for j := idx + 1; j < len(cleanedLines); j++ {
				isSection := false
				for k, v := range sectionMap {
					if v == j && k != key {
						isSection = true
						break
					}
				}
				if isSection {
					break
				}
				summaryLines = append(summaryLines, cleanedLines[j])
			}
			data.Summary = strings.Join(summaryLines, " ")
			break
		}
	}

	data.Skills = extractSectionList(cleanedLines, sectionMap, []string{"habilidades", "skills"})
	data.Languages = extractSectionList(cleanedLines, sectionMap, []string{"idiomas", "languages"})
	data.Certifications = extractSectionList(cleanedLines, sectionMap, []string{"certificacoes", "certifications", "certificacao", "certificação", "certificações"})

	return data
}

// extractSectionList coleta os itens de uma secao do curriculo (ex.: skills,
// idiomas, certificacoes) ate o inicio da proxima secao conhecida. Itens sao
// separados por virgula, ponto e virgula, barra ou bullet.
func extractSectionList(cleanedLines []string, sectionMap map[string]int, keys []string) []string {
	var items []string
	for _, key := range keys {
		if idx := sectionMap[key]; idx >= 0 {
			for j := idx + 1; j < len(cleanedLines); j++ {
				isSection := false
				for k, v := range sectionMap {
					if v == j && k != key {
						isSection = true
						break
					}
				}
				if isSection {
					break
				}
				parts := regexp.MustCompile(`[,;|•]`).Split(cleanedLines[j], -1)
				for _, p := range parts {
					p = strings.TrimSpace(p)
					if p != "" {
						items = append(items, p)
					}
				}
			}
			break
		}
	}
	return items
}
