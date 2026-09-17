package services

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/anomalyco/vagai-api/internal/models"
)

func questionOf(text, guide string) verificationQA {
	return verificationQA{
		Question:   models.InterviewQuestion{Text: text, AnswerGuide: guide, Topic: "Go"},
		Answer:     "Meu candidato responde sobre goroutines e channels.",
		Guide:      "goroutines, channels, concorrência, waitgroups",
		Assessment: "medium",
	}
}

// T003 — fallback determinístico: nota 0–10 com 1 casa decimal, feedback com
// ponto forte + sugestão (SC-003/SC-004).
func TestVerifyPreparationFallback_ScoreAndFeedback(t *testing.T) {
	qas := []verificationQA{questionOf("O que são goroutines?", "goroutines, channels, concorrência, waitgroups")}
	v, err := verifyPreparationFallback(qas)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Score < 0 || v.Score > 10 {
		t.Errorf("score must be within 0..10, got %v", v.Score)
	}
	if !strings.Contains(v.Feedback, "Ponto forte") || !strings.Contains(v.Feedback, "Sugestão") {
		t.Errorf("feedback must contain strength + suggestion, got %q", v.Feedback)
	}
	// 1 casa decimal
	if v.Score*10 != float64(int(v.Score*10)) {
		t.Errorf("score must have 1 decimal place, got %v", v.Score)
	}
}

// T003 — autoavaliação influencia a nota no fallback: high >= medium >= low.
func TestVerifyPreparationFallback_SelfAssessmentWeighing(t *testing.T) {
	base := questionOf("O que são goroutines?", "goroutines, channels, concorrência, waitgroups")
	base.Answer = "goroutines channels concorrência waitgroups"

	low := base
	low.Assessment = "low"
	med := base
	med.Assessment = "medium"
	high := base
	high.Assessment = "high"

	vl, _ := verifyPreparationFallback([]verificationQA{low})
	vm, _ := verifyPreparationFallback([]verificationQA{med})
	vh, _ := verifyPreparationFallback([]verificationQA{high})

	if vh.Score < vm.Score || vm.Score < vl.Score {
		t.Errorf("expected high>=medium>=low, got high=%v medium=%v low=%v", vh.Score, vm.Score, vl.Score)
	}
}

// T003 — validVerificationPayload rejeita score fora de 0..10 e feedback vazio.
func TestValidVerificationPayload(t *testing.T) {
	if validVerificationPayload(AIVerification{Score: 11, Feedback: "Ponto forte: x; Sugestão: y"}) {
		t.Error("score 11 should be rejected")
	}
	if validVerificationPayload(AIVerification{Score: -1, Feedback: "Ponto forte: x; Sugestão: y"}) {
		t.Error("score -1 should be rejected")
	}
	if validVerificationPayload(AIVerification{Score: 5, Feedback: " "}) {
		t.Error("empty feedback should be rejected")
	}
	if !validVerificationPayload(AIVerification{Score: 7.5, Feedback: "Ponto forte: Go; Sugestão: estudar cache"}) {
		t.Error("valid payload should be accepted")
	}
}

// T004 — claim otimista permite pending|outdated|verified, bloqueia running.
func TestClaimAllows(t *testing.T) {
	cases := []struct {
		status models.PreparationVerificationStatus
		want   bool
	}{
		{models.VerificationPending, true},
		{models.VerificationOutdated, true},
		{models.VerificationVerified, true},
		{models.VerificationRunning, false},
	}
	for _, c := range cases {
		if got := claimAllows(c.status); got != c.want {
			t.Errorf("claimAllows(%s) = %v, want %v", c.status, got, c.want)
		}
	}
}

// T004 — self-heal: running antigo (>5 min) é stale; recente/outros não.
func TestIsStaleRunning(t *testing.T) {
	now := time.Now()
	staleStart := now.Add(-(selfHealRunningThreshold + time.Minute))
	freshStart := now.Add(-time.Minute)

	older := &models.PreparationVerification{Status: models.VerificationRunning, StartedAt: &staleStart}
	if !isStaleRunning(older, selfHealRunningThreshold, now) {
		t.Error("running started 6 minutes ago should be stale")
	}

	fresher := &models.PreparationVerification{Status: models.VerificationRunning, StartedAt: &freshStart}
	if isStaleRunning(fresher, selfHealRunningThreshold, now) {
		t.Error("running started 1 minute ago should NOT be stale (tentativa viva)")
	}

	verified := &models.PreparationVerification{Status: models.VerificationVerified, StartedAt: &staleStart}
	if isStaleRunning(verified, selfHealRunningThreshold, now) {
		t.Error("verified is not stale-running")
	}

	if isStaleRunning(nil, selfHealRunningThreshold, now) {
		t.Error("nil should never be stale")
	}
}

// T004 — clamp0to10 limita a faixa da nota.
func TestClamp0to10(t *testing.T) {
	if got := clamp0to10(12.4); got != 10 {
		t.Errorf("clamp high = %v, want 10", got)
	}
	if got := clamp0to10(-3.0); got != 0 {
		t.Errorf("clamp low = %v, want 0", got)
	}
	if got := clamp0to10(5.5); got != 5.5 {
		t.Errorf("clamp mid = %v, want 5.5", got)
	}
}

// T003 — auto-save aceita rascunho sem autoavaliação (FR-008).
func TestValidateSelfAssessment_EmptyAllowed(t *testing.T) {
	if err := validateSelfAssessment(""); err != nil {
		t.Errorf("empty self_assessment (draft) should be allowed, got error: %v", err)
	}
}

// T003 — valores válidos continuam aceitos.
func TestValidateSelfAssessment_ValidValues(t *testing.T) {
	for _, v := range []string{"low", "medium", "high"} {
		if err := validateSelfAssessment(v); err != nil {
			t.Errorf("self_assessment %q should be allowed, got error: %v", v, err)
		}
	}
}

// T003 — valores não-vazios inválidos continuam rejeitados.
func TestValidateSelfAssessment_InvalidValues(t *testing.T) {
	for _, v := range []string{"max", "none", " ", "LOW"} {
		if err := validateSelfAssessment(v); err == nil {
			t.Errorf("self_assessment %q should be rejected", v)
		} else if !errors.Is(err, ErrInvalidSelfAssessment) {
			t.Errorf("self_assessment %q should return ErrInvalidSelfAssessment, got: %v", v, err)
		}
	}
}
