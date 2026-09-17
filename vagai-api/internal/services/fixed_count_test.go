package services

import (
	"testing"

	"github.com/anomalyco/vagai-api/internal/models"
)

func mkModelQs(n int, category string) []models.InterviewQuestion {
	qs := make([]models.InterviewQuestion, 0, n)
	for i := 0; i < n; i++ {
		qs = append(qs, models.InterviewQuestion{
			Category:    models.InterviewCategory(category),
			Topic:       "Tópico",
			Text:        "Pergunta modelo",
			AnswerGuide: "Guia",
			Status:      models.QuestionStatusPending,
		})
		// Texto único por pergunta para detectar duplicação indevida.
		qs[i].Text = "Pergunta " + string(rune('A'+i%26)) + string(rune('0'+i/26))
	}
	return qs
}

// T002 — truncagem preserva a ordem e o total exato.
func TestEnforceQuestionCount_TruncatesTo15(t *testing.T) {
	qs := BuildTemplateQuestions([]string{"Go", "MySQL", "Redis", "Docker", "Kubernetes", "Vue"})
	models := templateQuestionsToModels(qs)
	if len(models) != 19 {
		t.Fatalf("precondition: expected 19 template questions, got %d", len(models))
	}

	got := EnforceQuestionCount(models, []string{"Go"}, "")
	if len(got) != TargetQuestionCount {
		t.Fatalf("expected %d questions, got %d", TargetQuestionCount, len(got))
	}
	for i := range got {
		if got[i].Text != models[i].Text {
			t.Fatalf("truncation must preserve order: position %d differs", i)
		}
	}
}

// T002/T003 — suplementação completa para 15 sem duplicar textos.
func TestEnforceQuestionCount_SupplementsTo15(t *testing.T) {
	qs := BuildTemplateQuestions(nil) // 7 perguntas (4 foundations + 3 arch)
	models := templateQuestionsToModels(qs)
	if len(models) != 7 {
		t.Fatalf("precondition: expected 7 template questions, got %d", len(models))
	}

	got := EnforceQuestionCount(models, nil, "")
	if len(got) != TargetQuestionCount {
		t.Fatalf("expected %d questions, got %d", TargetQuestionCount, len(got))
	}

	seen := map[string]bool{}
	cats := map[string]bool{}
	for _, q := range got {
		if q.Text == "" || q.AnswerGuide == "" {
			t.Errorf("supplemented question has empty text/guide: %+v", q)
		}
		if seen[q.Text] {
			t.Errorf("duplicated question text %q", q.Text)
		}
		seen[q.Text] = true
		cats[string(q.Category)] = true
	}
	if !(cats["technology"] && cats["foundations"] && cats["architecture"]) {
		t.Errorf("expected all 3 categories after supplementation, got %v", cats)
	}
}

// T002 — conjunto exato permanece intacto sem seed.
func TestEnforceQuestionCount_Exact15Unchanged(t *testing.T) {
	qs := mkModelQs(15, "technology")
	got := EnforceQuestionCount(qs, nil, "")
	if len(got) != 15 {
		t.Fatalf("expected 15 questions, got %d", len(got))
	}
	for i := range got {
		if got[i].Text != qs[i].Text {
			t.Fatalf("exact set must keep order without seed: position %d differs", i)
		}
	}
}

// T004 — mesma seed, mesma ordem; seed vazia preserva a ordem.
func TestShuffleQuestionsWithSeed_Reproducible(t *testing.T) {
	qs := mkModelQs(15, "technology")

	a := ShuffleQuestionsWithSeed(qs, "seed123")
	b := ShuffleQuestionsWithSeed(qs, "seed123")
	for i := range a {
		if a[i].Text != b[i].Text {
			t.Fatalf("same seed must produce same order: position %d differs", i)
		}
	}

	plain := ShuffleQuestionsWithSeed(qs, "")
	for i := range plain {
		if plain[i].Text != qs[i].Text {
			t.Fatalf("empty seed must preserve order: position %d differs", i)
		}
	}
}

// T003 — suplemento retorna exatamente o necessário, sem repetir existentes.
func TestSupplementTemplateQuestions_ExactNeed(t *testing.T) {
	existing := BuildTemplateQuestions(nil) // 7
	extra := SupplementTemplateQuestions(existing, nil, TargetQuestionCount)
	if len(extra) != TargetQuestionCount-len(existing) {
		t.Fatalf("expected %d supplements, got %d", TargetQuestionCount-len(existing), len(extra))
	}

	seen := map[string]bool{}
	for _, q := range existing {
		seen[q.Text] = true
	}
	for _, q := range extra {
		if seen[q.Text] {
			t.Errorf("supplement duplicates existing text %q", q.Text)
		}
		seen[q.Text] = true
	}

	if got := SupplementTemplateQuestions(existing, nil, len(existing)); len(got) != 0 {
		t.Errorf("no supplementation needed: expected 0, got %d", len(got))
	}
}
