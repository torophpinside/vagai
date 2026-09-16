package matcher

import (
	"strings"
	"testing"

	"github.com/anomalyco/vagai-cli/internal/models"
)

const sampleResumeDataJSON = `{
	"personal_info": {"name": "João Silva", "location": "São Paulo, SP"},
	"summary": "Desenvolvedor Go com experiência em APIs REST.",
	"experience": [{"role": "Backend Developer", "company": "Acme", "description": "Docker e Kubernetes em produção."}],
	"education": [{"degree": "Bacharelado", "field": "Ciência da Computação", "institution": "USP"}],
	"skills": ["Go", "PostgreSQL"],
	"languages": ["Inglês"],
	"certifications": ["AWS"]
}`

func TestResumeText_PrefersRawContent(t *testing.T) {
	resume := models.Resume{Content: "Conteudo bruto extraído", Data: sampleResumeDataJSON}
	if got := resumeText(resume); got != resume.Content {
		t.Errorf("resumeText = %q, want raw content %q", got, resume.Content)
	}
}

func TestResumeText_DerivesFromData(t *testing.T) {
	resume := models.Resume{Data: sampleResumeDataJSON}
	got := resumeText(resume)

	for _, want := range []string{
		"João Silva", "São Paulo, SP", "Desenvolvedor Go", "Backend Developer",
		"Docker e Kubernetes", "Go", "PostgreSQL", "Inglês", "AWS",
		"Ciência da Computação", "USP",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("resumeText() deveria conter %q, got %q", want, got)
		}
	}
}

func TestResumeText_EmptyData(t *testing.T) {
	if got := resumeText(models.Resume{}); got != "" {
		t.Errorf("resumeText() = %q, want empty", got)
	}
	if got := resumeText(models.Resume{Data: "invalid json"}); got != "" {
		t.Errorf("resumeText() = %q for invalid JSON, want empty", got)
	}
}

func TestAnalyzeJobLocation_TypeDetection(t *testing.T) {
	tests := []struct {
		name string
		desc string
		want string
	}{
		{"presencial explicito", "Trabalho presencial no escritorio de Curitiba", "presencial"},
		{"hibrido com home office", "Modelo hibrido, home office 2x por semana", "hybrid"},
		{"home office e remoto", "Trabalhe de home office em qualquer lugar do Brasil", "remote"},
		{"pacote office nao e presencial", "Requisitos: Pacote Office avancado, Excel e Word", "unknown"},
		{"remoto simples", "Vaga remota para toda a equipe", "remote"},
		{"sem sinal", "Desenvolvedor backend com experiencia em Go", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc := analyzeJobLocation(tt.desc, "")
			if loc.Type != tt.want {
				t.Errorf("analyzeJobLocation(%q).Type = %q, want %q", tt.desc, loc.Type, tt.want)
			}
		})
	}
}

func TestAnalyzeJobLocation_CityExtraction(t *testing.T) {
	tests := []struct {
		name string
		desc string
		want string
	}{
		{"cidade com acento", "Vaga presencial em São Paulo, região central", "sao paulo"},
		{"sigla sp", "Escritório localizado na região de SP", "sao paulo"},
		{"responsavel nao extrai sp", "Profissional responsável pelo time de dados", ""},
		{"informacoes nao extrai fortaleza", "Envie suas informações atualizadas", ""},
		{"porto alegre vence porto", "Vaga em Porto Alegre, bairro Moinhos de Vento", "porto alegre"},
		{"formato cidade/sigla", "Presencial em Florianópolis/SC", "florianopolis"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc := analyzeJobLocation(tt.desc, "")
			if loc.City != tt.want {
				t.Errorf("analyzeJobLocation(%q).City = %q, want %q", tt.desc, loc.City, tt.want)
			}
		})
	}
}

func TestLocationScore(t *testing.T) {
	tests := []struct {
		name     string
		jobLoc   JobLocation
		userCity string
		want     float64
	}{
		{"mesma cidade com acento diferente", JobLocation{Type: "presencial", City: "sao paulo"}, "São Paulo", 100},
		{"cidades diferentes", JobLocation{Type: "presencial", City: "curitiba"}, "são paulo", 20},
		{"hibrido sem cidade", JobLocation{Type: "hybrid"}, "recife", 50},
		{"remote sempre 100", JobLocation{Type: "remote", City: "curitiba"}, "são paulo", 100},
		{"unknown sem cidade", JobLocation{Type: "unknown"}, "são paulo", 100},
		{"unknown cidade diferente penaliza", JobLocation{Type: "unknown", City: "lisboa"}, "são paulo", 20},
		{"unknown mesma cidade", JobLocation{Type: "unknown", City: "sao paulo"}, "São Paulo", 100},
		{"mesmo estado via UF", JobLocation{Type: "presencial", City: "parana"}, "Curitiba", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := locationScore(tt.jobLoc, tt.userCity)
			if got != tt.want {
				t.Errorf("locationScore(%+v, %q) = %.0f, want %.0f", tt.jobLoc, tt.userCity, got, tt.want)
			}
		})
	}
}

func TestLocationPenalty(t *testing.T) {
	tests := []struct {
		name     string
		jobLoc   JobLocation
		userCity string
		want     float64
	}{
		{"fora da cidade", JobLocation{Type: "presencial", City: "curitiba"}, "são paulo", 24},
		{"presencial sem cidade identificavel", JobLocation{Type: "presencial"}, "são paulo", 12},
		{"hibrido sem cidade identificavel", JobLocation{Type: "hybrid"}, "são paulo", 12},
		{"mesma cidade", JobLocation{Type: "presencial", City: "sao paulo"}, "sp", 0},
		{"remote sem penalidade", JobLocation{Type: "remote", City: "curitiba"}, "são paulo", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := locationPenalty(tt.jobLoc, tt.userCity)
			if got != tt.want {
				t.Errorf("locationPenalty(%+v, %q) = %.2f, want %.2f", tt.jobLoc, tt.userCity, got, tt.want)
			}
		})
	}
}

func TestCalculateMatchFallback_UsesJobLocation(t *testing.T) {
	resume := "Desenvolvedor Go com experiencia em Docker, Kubernetes e PostgreSQL."
	jobDesc := "Vaga para desenvolvedor Go. Requisitos: Docker, Kubernetes, PostgreSQL."

	sameCityLoc := JobLocation{Type: "presencial", City: "sao paulo"}
	farCityLoc := JobLocation{Type: "presencial", City: "curitiba"}

	scoreSame, _, _, err := calculateMatchFallback(jobDesc, jobDesc, resume, "São Paulo", sameCityLoc)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	scoreFar, _, _, err := calculateMatchFallback(jobDesc, jobDesc, resume, "São Paulo", farCityLoc)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if scoreFar >= scoreSame {
		t.Errorf("vaga fora da cidade deveria pontuar menos: mesma cidade=%.2f, outra cidade=%.2f", scoreSame, scoreFar)
	}

	expectedDelta := (100 - 20) * 0.25
	got := scoreSame - scoreFar
	if got < expectedDelta-1 || got > expectedDelta+1 {
		t.Errorf("delta de localização = %.2f, esperado ~%.2f", got, expectedDelta)
	}
}

func TestExtractAIResult_MultilineReason(t *testing.T) {
	raw := `{
  "score": 75,
  "reason": "Aqui está uma análise
  com quebras de linha
  no texto."
}`
	score, reason := extractAIResult(raw)
	if score != 75 {
		t.Errorf("score = %v, esperado 75", score)
	}
	if reason == "" {
		t.Error("reason deveria ser extraída mesmo com quebras de linha")
	}
}

func TestExtractAIResult_StrictJSON(t *testing.T) {
	raw := `{"score": 88.5, "reason": "ótimo match"}`
	score, reason := extractAIResult(raw)
	if score != 88.5 {
		t.Errorf("score = %v, esperado 88.5", score)
	}
	if reason != "ótimo match" {
		t.Errorf("reason = %q, esperado ótimo match", reason)
	}
}

func TestExtractAIResult_InvalidScore(t *testing.T) {
	score, _ := extractAIResult(`{"score": "alto", "reason": "x"}`)
	if score != 0 {
		t.Errorf("score = %v, esperado 0", score)
	}
}

func TestNegativeKeywordPenalty(t *testing.T) {
	tests := []struct {
		name string
		text string
		kw   []string
		want float64
	}{
		{"texto vazio", "", []string{"freela"}, 0},
		{"keyword ausente", "desenvolvedor senior", []string{"freela"}, 0},
		{"uma por distinta apesar de multiplas ocorrencias", "freela remoto freela pj", []string{"freela"}, 3},
		{"duas distintas acumulam", "freela e contrato pj", []string{"freela", "contrato"}, 6},
		{"keyword em branco ignorada", "freela", []string{""}, 0},
		{"normalizacao ignora acentos e caixa", "São Paulo", []string{"sao paulo"}, 3},
		{"duplicatas normalizadas contam uma vez", "freela", []string{"Freela", "freela"}, 3},
		{"frase multipalavra", "vaga para machine learning senior", []string{"machine learning"}, 3},
		{"sem falso positivo por substring", "desenvolvedor golang compreensivo", []string{"go", "compra"}, 0},
		{"pontuacao nao quebra fronteira", "Trabalho remoto; presencial opcional.", []string{"remoto"}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := negativeKeywordPenalty(tt.text, tt.kw); got != tt.want {
				t.Errorf("negativeKeywordPenalty(%q, %v) = %.1f, want %.1f", tt.text, tt.kw, got, tt.want)
			}
		})
	}
}

func TestLoadNegativeKeywords(t *testing.T) {
	if got := loadNegativeKeywords(""); got != nil {
		t.Errorf("loadNegativeKeywords(\"\") = %v, want nil", got)
	}
	if got := loadNegativeKeywords("invalid json"); got != nil {
		t.Errorf("loadNegativeKeywords(invalid) = %v, want nil", got)
	}

	got := loadNegativeKeywords(`["freela", "pj"]`)
	if len(got) != 2 || got[0] != "freela" || got[1] != "pj" {
		t.Errorf("loadNegativeKeywords valid = %v, want [freela pj]", got)
	}
}

func TestLocationAllowedToMatch(t *testing.T) {
	tests := []struct {
		name     string
		jobLoc   JobLocation
		userCity string
		want     bool
	}{
		{"presencial cidade diferente bloqueia", JobLocation{Type: "presencial", City: "sao paulo"}, "Curitiba", false},
		{"hibrido cidade diferente bloqueia", JobLocation{Type: "hybrid", City: "porto alegre"}, "Curitiba", false},
		{"presencial mesma cidade libera", JobLocation{Type: "presencial", City: "curitiba"}, "Curitiba", true},
		{"hibrido mesmo estado libera", JobLocation{Type: "hybrid", City: "parana"}, "Curitiba", true},
		{"presencial sem cidade libera com penalidade", JobLocation{Type: "presencial"}, "Curitiba", true},
		{"hibrido sem cidade libera com penalidade", JobLocation{Type: "hybrid"}, "Curitiba", true},
		{"remoto libera qualquer cidade", JobLocation{Type: "remote", City: "sao paulo"}, "Curitiba", true},
		{"unknown libera (penalidade no score)", JobLocation{Type: "unknown", City: "porto alegre"}, "Curitiba", true},
		{"sem cidade configurada nao filtra", JobLocation{Type: "presencial", City: "sao paulo"}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := locationAllowedToMatch(tt.jobLoc, tt.userCity); got != tt.want {
				t.Errorf("locationAllowedToMatch(%+v, %q) = %v, want %v", tt.jobLoc, tt.userCity, got, tt.want)
			}
		})
	}
}
