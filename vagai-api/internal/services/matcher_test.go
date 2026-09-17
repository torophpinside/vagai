package services

import (
	"testing"
)

func TestMatchResumeToJob_SkillOverlap(t *testing.T) {
	data := ResumeData{
		Summary: "Desenvolvedor Go com experiencia em APIs REST.",
		Skills:  []string{"Go", "MySQL", "Docker", "Kubernetes"},
	}
	res := MatchResumeToJob(data, "", "Desenvolvedor Go Senior", "Requisitos: Go, MySQL, Docker, Kubernetes. Experiencia com APIs REST e microservicos.", nil)

	if res.Score < 60 || res.Score > 100 {
		t.Fatalf("expected high score, got %.2f", res.Score)
	}
	for _, want := range []string{"go", "mysql", "docker", "kubernetes"} {
		if !containsString(res.KeywordsMatched, want) {
			t.Errorf("expected keyword %q in %v", want, res.KeywordsMatched)
		}
	}
}

func TestMatchResumeToJob_NoOverlap(t *testing.T) {
	data := ResumeData{
		Summary: "Desenvolvedor Java backend.",
		Skills:  []string{"Java"},
	}
	res := MatchResumeToJob(data, "", "Gerente de Vendas", "Responsavel pelo crescimento de receita e gestao da equipe comercial.", nil)

	if res.Score != 0 {
		t.Errorf("expected score 0, got %.2f", res.Score)
	}
	if len(res.KeywordsMatched) != 0 {
		t.Errorf("expected no keywords, got %v", res.KeywordsMatched)
	}
}

func TestMatchResumeToJob_AccentInsensitive(t *testing.T) {
	data := ResumeData{
		Summary: "Analise de dados com Python.",
		Skills:  []string{"Python"},
	}
	res := MatchResumeToJob(data, "", "Analista de Dados", "Requisitos: análise de dados e Python.", nil)

	if res.Score <= 0 {
		t.Fatalf("expected positive score, got %.2f", res.Score)
	}
	for _, want := range []string{"dados", "analise", "python"} {
		if !containsString(res.KeywordsMatched, want) {
			t.Errorf("expected keyword %q in %v", want, res.KeywordsMatched)
		}
	}
}

func TestMatchResumeToJob_MultiWordSkill(t *testing.T) {
	data := ResumeData{
		Summary: "Engenheiro com atuacao em IA.",
		Skills:  []string{"Machine Learning"},
	}
	res := MatchResumeToJob(data, "", "Machine Learning Engineer", "Vaga para engenheiro de machine learning.", nil)

	if res.Score <= 0 {
		t.Fatalf("expected positive score, got %.2f", res.Score)
	}
	if !containsString(res.KeywordsMatched, "machine learning") {
		t.Errorf("expected keyword %q in %v", "machine learning", res.KeywordsMatched)
	}
}

func TestMatchResumeToJob_TitleWeightsMore(t *testing.T) {
	data := ResumeData{Summary: "Desenvolvedor Java."}

	titleHit := MatchResumeToJob(data, "", "Desenvolvedor Java", "Stack backend.", nil)
	descHit := MatchResumeToJob(data, "", "Desenvolvedor Backend", "Stack Java.", nil)

	if titleHit.Score <= descHit.Score {
		t.Errorf("expected title match score > description match score, got %.2f <= %.2f", titleHit.Score, descHit.Score)
	}
}

func TestMatchResumeToJob_RawContentFallback(t *testing.T) {
	data := ResumeData{}
	rawContent := "Conhecimento em PostgreSQL e Redis para cache."
	res := MatchResumeToJob(data, rawContent, "Backend Developer", "Requisitos: PostgreSQL, Redis.", nil)

	if res.Score <= 0 {
		t.Fatalf("expected positive score using raw content, got %.2f", res.Score)
	}
	for _, want := range []string{"postgresql", "redis"} {
		if !containsString(res.KeywordsMatched, want) {
			t.Errorf("expected keyword %q in %v", want, res.KeywordsMatched)
		}
	}
}

func TestMatchResumeToJob_EmptyJob(t *testing.T) {
	data := ResumeData{Summary: "Desenvolvedor Go.", Skills: []string{"Go"}}
	res := MatchResumeToJob(data, "", "", "", nil)

	if res.Score != 0 {
		t.Errorf("expected score 0 for empty job, got %.2f", res.Score)
	}
	if len(res.KeywordsMatched) != 0 {
		t.Errorf("expected no keywords, got %v", res.KeywordsMatched)
	}
}

func TestMatchResumeToJob_DeduplicatesKeywords(t *testing.T) {
	data := ResumeData{
		Summary: "Go developer.",
		Skills:  []string{"Go", "go"},
	}
	res := MatchResumeToJob(data, "", "Go Developer", "Desenvolvedor Go.", nil)

	seen := map[string]int{}
	for _, kw := range res.KeywordsMatched {
		seen[kw]++
	}
	for kw, n := range seen {
		if n > 1 {
			t.Errorf("keyword %q duplicated %d times", kw, n)
		}
	}
}

func TestMatchResumeToJob_NegativeKeywordsPenalty(t *testing.T) {
	data := ResumeData{Summary: "Desenvolvedor Go.", Skills: []string{"Go"}}

	base := MatchResumeToJob(data, "", "Desenvolvedor Go", "Linguagem Go. Aceita freela.", nil)
	pen := MatchResumeToJob(data, "", "Desenvolvedor Go", "Linguagem Go. Aceita freela.", []string{"freela"})

	if base.Score-pen.Score != 3 {
		t.Fatalf("expected penalty of exactly 3 pts, base=%.2f pen=%.2f", base.Score, pen.Score)
	}
}

func TestMatchResumeToJob_NegativeKeywordsAccumulateDistinct(t *testing.T) {
	data := ResumeData{Summary: "Desenvolvedor Go.", Skills: []string{"Go"}}

	base := MatchResumeToJob(data, "", "Desenvolvedor Go", "Freela e contrato PJ.", nil)
	pen := MatchResumeToJob(data, "", "Desenvolvedor Go", "Freela e contrato PJ.", []string{"freela", "PJ"})

	if base.Score-pen.Score != 6 {
		t.Fatalf("expected -6 for two distinct keywords, got base=%.2f pen=%.2f", base.Score, pen.Score)
	}
}

func TestMatchResumeToJob_NegativeKeywordWordBoundary(t *testing.T) {
	data := ResumeData{Summary: "Desenvolvedor Go.", Skills: []string{"Go"}}

	base := MatchResumeToJob(data, "", "Desenvolvedor Plataforma", "Desenvolvedor Golang senior com governanca de dados.", nil)
	pen := MatchResumeToJob(data, "", "Desenvolvedor Plataforma", "Desenvolvedor Golang senior com governanca de dados.", []string{"go"})

	if pen.Score != base.Score {
		t.Fatalf("expected no penalty for substring hit, base=%.2f pen=%.2f", base.Score, pen.Score)
	}
}

func TestMatchResumeToJob_NegativeKeywordsClampedAtZero(t *testing.T) {
	data := ResumeData{Summary: ""}

	res := MatchResumeToJob(data, "", "Freela", "Freela PJ remoto.", []string{"freela", "pj", "remoto", "contrato"})

	if res.Score != 0 {
		t.Fatalf("expected score clamped at 0, got %.2f", res.Score)
	}
}

func TestNegativeKeywordPenalty_Unit(t *testing.T) {
	cases := []struct {
		name string
		text string
		kw   []string
		want float64
	}{
		{"empty text", "", []string{"freela"}, 0},
		{"keyword absent", "desenvolvedor senior", []string{"freela"}, 0},
		{"one distinct despite multiple occurrences", "freela remoto freela pj", []string{"freela"}, 3},
		{"two distinct keywords", "freela e contrato pj", []string{"freela", "contrato"}, 6},
		{"blank keyword ignored", "freela", []string{""}, 0},
		{"duplicates normalized once", "freela", []string{"Freela", "freela"}, 3},
		{"multi-word phrase", "vaga para machine learning senior", []string{"machine learning"}, 3},
		{"no false positive substring", "desenvolvedor golang compreensivo", []string{"go", "compra"}, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := negativeKeywordPenalty(tc.text, tc.kw)
			if got != tc.want {
				t.Errorf("negativeKeywordPenalty(%q, %v) = %.2f, want %.2f", tc.text, tc.kw, got, tc.want)
			}
		})
	}
}

func containsString(items []string, want string) bool {
	for _, it := range items {
		if it == want {
			return true
		}
	}
	return false
}
