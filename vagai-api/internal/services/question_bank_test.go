package services

import (
	"strings"
	"testing"
)

// T009 [P] [US1] Unit: tech detection + template bank.

func TestDetectTechnologies_ExtractsKnownStack(t *testing.T) {
	desc := "Desenvolvedor backend com Go, MySQL e Redis. Conhecimento em Docker e Kubernetes."
	got := DetectTechnologies(desc)

	for _, want := range []string{"Go", "MySQL", "Redis", "Docker", "Kubernetes"} {
		if !containsString(got, want) {
			t.Errorf("expected tech %q to be detected, got %v", want, got)
		}
	}
}

func TestDetectTechnologies_WordBoundaryNoFalsePositive(t *testing.T) {
	desc := "Vaga para engenheiro de dados com experience em governança e logging."
	got := DetectTechnologies(desc)

	if containsString(got, "Go") {
		t.Errorf("expected no 'Go' false positive from 'governança'/'logging', got %v", got)
	}
}

func TestDetectTechnologies_MaxSix(t *testing.T) {
	desc := "Stack: Go, MySQL, Redis, Docker, Kubernetes, Vue, React, Python, Java e TypeScript."
	got := DetectTechnologies(desc)

	if len(got) > 6 {
		t.Errorf("expected at most 6 detected techs, got %d: %v", len(got), got)
	}
}

func TestDetectTechnologies_Deduplicates(t *testing.T) {
	desc := "Precisamos de alguém com Go e também experiência com Golang em projetos reais."
	got := DetectTechnologies(desc)

	if !containsString(got, "Go") {
		t.Fatalf("expected Go to be detected, got %v", got)
	}
	if countString(got, "Go") != 1 {
		t.Errorf("expected Go to appear exactly once, got %v", got)
	}
}

func TestDetectTechnologies_AccentAndCaseInsensitive(t *testing.T) {
	desc := "Requisitos: KUBERNETES, JAVASCRIPT e DOCKER em ambiente de produção."
	got := DetectTechnologies(desc)

	for _, want := range []string{"Kubernetes", "JavaScript", "Docker"} {
		if !containsString(got, want) {
			t.Errorf("expected %q to be detected regardless of case, got %v", want, got)
		}
	}
}

func TestBuildTemplateQuestions_ThreeCategoriesPresent(t *testing.T) {
	qs := BuildTemplateQuestions([]string{"Go", "MySQL"})

	var techs, foundations, arch int
	for _, q := range qs {
		switch q.Category {
		case "technology":
			techs++
		case "foundations":
			foundations++
		case "architecture":
			arch++
		default:
			t.Errorf("unexpected category %q", q.Category)
		}
	}

	if techs != 4 {
		t.Errorf("expected 4 technology questions (2 per tech), got %d", techs)
	}
	if foundations != 4 {
		t.Errorf("expected 4 foundations questions, got %d", foundations)
	}
	if arch != 3 {
		t.Errorf("expected 3 architecture questions, got %d", arch)
	}
	if len(qs) != 11 {
		t.Errorf("expected 11 questions total, got %d", len(qs))
	}
}

func TestBuildTemplateQuestions_NoDuplicateQuestions(t *testing.T) {
	qs := BuildTemplateQuestions([]string{"Go", "MySQL", "Go", "MySQL", "Redis"})

	// Tecnologias repetidas devem gerar apenas 2 perguntas por tech (FR-008:
	// a regeneração substitui sem duplicar). As 2 perguntas da mesma tech têm
	// o mesmo Topic mas enunciados distintos.
	if len(qs) != 13 { // (3 techs * 2) + 4 + 3
		t.Errorf("expected 13 questions from deduped techs, got %d", len(qs))
	}

	seen := map[string]bool{}
	for _, q := range qs {
		if seen[q.Text] {
			t.Errorf("duplicated question text %q", q.Text)
		}
		seen[q.Text] = true
	}

	techTopics := 0
	for _, q := range qs {
		if q.Category == "technology" && q.Topic == "Go" {
			techTopics++
		}
	}
	if techTopics != 2 {
		t.Errorf("expected exactly 2 questions for Go, got %d", techTopics)
	}
}

func TestBuildTemplateQuestions_LimitsTechsToSix(t *testing.T) {
	qs := BuildTemplateQuestions([]string{"Go", "MySQL", "Redis", "Docker", "Vue", "React", "Python", "Java"})

	if len(qs) != 19 {
		t.Errorf("expected 19 questions with 8 techs (capped at 6), got %d", len(qs))
	}
}

func TestBuildTemplateQuestions_AllQuestionsHaveText(t *testing.T) {
	qs := BuildTemplateQuestions([]string{"Go"})

	for _, q := range qs {
		if strings.TrimSpace(q.Text) == "" {
			t.Errorf("question for topic %q has empty text", q.Topic)
		}
		if strings.TrimSpace(q.AnswerGuide) == "" {
			t.Errorf("question for topic %q has empty answer guide", q.Topic)
		}
	}
}

// T010 [P] [US1] Unit: empty/short description -> only foundations+architecture.

func TestDetectTechnologies_EmptyDescription(t *testing.T) {
	if got := DetectTechnologies(""); len(got) != 0 {
		t.Errorf("expected no technologies for empty description, got %v", got)
	}
}

func TestBuildTemplateQuestions_EmptyDescription(t *testing.T) {
	qs := BuildTemplateQuestions(nil)

	for _, q := range qs {
		if q.Category == "technology" {
			t.Errorf("expected no technology questions when no techs detected, got %v", q)
		}
	}
	if len(qs) != 7 {
		t.Errorf("expected 7 questions (4 foundations + 3 architecture), got %d", len(qs))
	}
}

func TestBuildTemplateQuestions_ShortDescription(t *testing.T) {
	desc := "Estágio na área administrativa."
	techs := DetectTechnologies(desc)
	if len(techs) != 0 {
		t.Fatalf("expected no techs for short generic description, got %v", techs)
	}

	qs := BuildTemplateQuestions(techs)
	for _, q := range qs {
		if q.Category == "technology" {
			t.Errorf("expected no technology questions, got %v", q)
		}
	}
}

func countString(list []string, want string) int {
	n := 0
	for _, v := range list {
		if v == want {
			n++
		}
	}
	return n
}
