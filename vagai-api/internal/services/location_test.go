package services

import "testing"

func TestAnalyzeJobLocation_APIType(t *testing.T) {
	tests := []struct {
		name string
		desc string
		want string
	}{
		{"presencial explicito", "Trabalho presencial no escritorio de Curitiba", "presencial"},
		{"hibrido com home office", "Modelo hibrido, home office 2x por semana", "hybrid"},
		{"remoto simples", "Vaga remota para toda a equipe", "remote"},
		{"pacote office nao e presencial", "Requisitos: Pacote Office avancado, Excel e Word", "unknown"},
		{"sem sinal", "Desenvolvedor backend com experiencia em Go", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc := AnalyzeJobLocation(tt.desc)
			if loc.Type != tt.want {
				t.Errorf("AnalyzeJobLocation(%q).Type = %q, want %q", tt.desc, loc.Type, tt.want)
			}
		})
	}
}

func TestAnalyzeJobLocation_APICity(t *testing.T) {
	tests := []struct {
		name string
		desc string
		want string
	}{
		{"cidade com acento", "Vaga presencial em São Paulo, região central", "sao paulo"},
		{"formato cidade/sigla", "Presencial em Florianópolis/SC", "florianopolis"},
		{"responsavel nao extrai sp", "Profissional responsável pelo time de dados", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc := AnalyzeJobLocation(tt.desc)
			if loc.City != tt.want {
				t.Errorf("AnalyzeJobLocation(%q).City = %q, want %q", tt.desc, loc.City, tt.want)
			}
		})
	}
}

func TestLocationAllowedToMatch_API(t *testing.T) {
	tests := []struct {
		name     string
		loc      JobLocation
		userCity string
		want     bool
	}{
		{"presencial cidade diferente bloqueia", JobLocation{Type: "presencial", City: "sao paulo"}, "Curitiba", false},
		{"hibrido cidade diferente bloqueia", JobLocation{Type: "hybrid", City: "porto alegre"}, "Curitiba", false},
		{"presencial mesma cidade libera", JobLocation{Type: "presencial", City: "curitiba"}, "Curitiba", true},
		{"hibrido mesmo estado libera", JobLocation{Type: "hybrid", City: "parana"}, "Curitiba", true},
		{"hibrido sem cidade libera com penalidade", JobLocation{Type: "hybrid"}, "Curitiba", true},
		{"remoto libera qualquer cidade", JobLocation{Type: "remote", City: "sao paulo"}, "Curitiba", true},
		{"unknown libera", JobLocation{Type: "unknown", City: "porto alegre"}, "Curitiba", true},
		{"sem cidade configurada nao filtra", JobLocation{Type: "presencial", City: "sao paulo"}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LocationAllowedToMatch(tt.loc, tt.userCity); got != tt.want {
				t.Errorf("LocationAllowedToMatch(%+v, %q) = %v, want %v", tt.loc, tt.userCity, got, tt.want)
			}
		})
	}
}
