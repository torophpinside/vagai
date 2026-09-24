package services

import (
	"context"
	"strings"
	"testing"

	"github.com/anomalyco/vagai-api/internal/models"
)

// forceSpontaneousTemplateFallback aponta a IA para uma porta morta, tornando o
// fallback determinístico e rápido (constituição IV). Usado por todos os testes
// de geração avulsa.
func forceSpontaneousTemplateFallback(t *testing.T) {
	t.Helper()
	t.Setenv("LMSTUDIO_URL", "http://127.0.0.1:59999")
}

const spontaneousTestDescription = "Vaga backend com Go, MySQL e Docker para serviços em produção."

// T004/T007 [US1] Geração avulsa: 15 perguntas categorizadas sem IA.
func TestGenerateSpontaneousPreparation_Returns15Categorized(t *testing.T) {
	forceSpontaneousTemplateFallback(t)

	res, err := GenerateSpontaneousPreparation(context.Background(), "Backend Developer", "Tech Corp", spontaneousTestDescription, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Questions) != TargetQuestionCount {
		t.Fatalf("expected %d questions, got %d", TargetQuestionCount, len(res.Questions))
	}
	if res.Source != models.SourceTemplate {
		t.Errorf("expected source=template without AI, got %q", res.Source)
	}
	for _, want := range []string{"Go", "MySQL", "Docker"} {
		if !containsString(res.TechStack, want) {
			t.Errorf("expected tech %q detected, got %v", want, res.TechStack)
		}
	}
	if res.Title != "Backend Developer" || res.Company != "Tech Corp" {
		t.Errorf("expected title/company preserved, got %q/%q", res.Title, res.Company)
	}
	if res.Description != spontaneousTestDescription {
		t.Errorf("expected description preserved, got %q", res.Description)
	}
	if res.GeneratedAt.IsZero() {
		t.Error("expected generated_at to be set")
	}

	cats := map[string]bool{}
	seen := map[string]bool{}
	for _, q := range res.Questions {
		cats[string(q.Category)] = true
		if q.Text == "" || q.AnswerGuide == "" {
			t.Errorf("question has empty text/answer_guide: %+v", q)
		}
		if q.Status != models.QuestionStatusPending {
			t.Errorf("expected pending status, got %q", q.Status)
		}
		if seen[q.Text] {
			t.Errorf("duplicated question text %q", q.Text)
		}
		seen[q.Text] = true
	}
	if !(cats["technology"] && cats["foundations"] && cats["architecture"]) {
		t.Errorf("expected all 3 categories, got %v", cats)
	}
}

// T005/T011 [US1] Descrição vazia (após trim) é rejeitada.
func TestGenerateSpontaneousPreparation_EmptyDescription(t *testing.T) {
	forceSpontaneousTemplateFallback(t)

	if _, err := GenerateSpontaneousPreparation(context.Background(), "t", "c", "   \n\t ", ""); err == nil {
		t.Fatal("expected error for blank description")
	} else if !strings.Contains(err.Error(), "obrigatório") {
		t.Errorf("expected obrigatório error, got %v", err)
	}
}

// T010 [US2] Mesma seed produz a mesma ordem; seed vazia preserva a ordem.
func TestGenerateSpontaneousPreparation_SeedReproducible(t *testing.T) {
	forceSpontaneousTemplateFallback(t)

	a, err := GenerateSpontaneousPreparation(context.Background(), "Backend", "Corp", spontaneousTestDescription, "seed123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b, err := GenerateSpontaneousPreparation(context.Background(), "Backend", "Corp", spontaneousTestDescription, "seed123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(a.Questions) != len(b.Questions) {
		t.Fatalf("question counts differ: %d vs %d", len(a.Questions), len(b.Questions))
	}
	for i := range a.Questions {
		if a.Questions[i].Text != b.Questions[i].Text {
			t.Fatalf("same seed must produce same order: position %d differs (%q vs %q)",
				i, a.Questions[i].Text, b.Questions[i].Text)
		}
	}
}

// T015 [US3] Descrição alterada reflete as tecnologias detectadas.
func TestGenerateSpontaneousPreparation_ChangedDescriptionReflectsTech(t *testing.T) {
	forceSpontaneousTemplateFallback(t)

	goRes, err := GenerateSpontaneousPreparation(context.Background(), "Backend", "Corp",
		"Vaga backend com Go e Docker.", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pyRes, err := GenerateSpontaneousPreparation(context.Background(), "Data", "Corp",
		"Vaga de dados com Python e PostgreSQL.", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !containsString(goRes.TechStack, "Go") || containsString(goRes.TechStack, "Python") {
		t.Errorf("expected Go-only stack, got %v", goRes.TechStack)
	}
	if !containsString(pyRes.TechStack, "Python") || containsString(pyRes.TechStack, "Go") {
		t.Errorf("expected Python-only stack, got %v", pyRes.TechStack)
	}
}

// T015 [US3] Serviço é stateless: não recebe DB e chamadas não compartilham estado.
func TestGenerateSpontaneousPreparation_StatelessNoSharedState(t *testing.T) {
	forceSpontaneousTemplateFallback(t)

	first, err := GenerateSpontaneousPreparation(context.Background(), "Backend", "Corp", spontaneousTestDescription, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(first.TechStack) == 0 {
		t.Fatal("precondition: expected detected technologies")
	}
	first.TechStack[0] = "MUTATED"
	first.Questions[0].Text = "MUTATED"

	second, err := GenerateSpontaneousPreparation(context.Background(), "Backend", "Corp", spontaneousTestDescription, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if second.TechStack[0] == "MUTATED" || second.Questions[0].Text == "MUTATED" {
		t.Error("generation leaked shared state across calls")
	}
}

// T004 [US1] Descrição longa: o texto completo é retornado e apenas o prefixo
// rune-safe vai para o prompt.
func TestGenerateSpontaneousPreparation_LongDescriptionRetainsFullText(t *testing.T) {
	forceSpontaneousTemplateFallback(t)

	long := strings.Repeat("a", spontaneousMaxPromptRunes+100) + " Go Docker"
	res, err := GenerateSpontaneousPreparation(context.Background(), "Backend", "Corp", long, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Description != long {
		t.Errorf("expected full description retained (%d runes), got %d runes",
			len([]rune(long)), len([]rune(res.Description)))
	}
	prefix := truncateRunes(res.Description, spontaneousMaxPromptRunes)
	if got := len([]rune(prefix)); got != spontaneousMaxPromptRunes {
		t.Errorf("expected prompt prefix of %d runes, got %d", spontaneousMaxPromptRunes, got)
	}
	if !containsString(res.TechStack, "Go") || !containsString(res.TechStack, "Docker") {
		t.Errorf("expected technologies detected from long description, got %v", res.TechStack)
	}
}
