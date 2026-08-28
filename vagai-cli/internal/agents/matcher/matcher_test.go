package matcher

import "testing"

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
		{"unknown sempre 100", JobLocation{Type: "unknown", City: "lisboa"}, "são paulo", 100},
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
