package e2e

import (
	"fmt"
	"net/http"
	"testing"
)

// TestInterviewPrep_AnsweredCountReactivity ensures that progress.answered
// correctly tracks non-empty answers and ignores blank ones, matching the
// frontend's expectation for the "respondidas" counter.
func TestInterviewPrep_AnsweredCountReactivity(t *testing.T) {
	token, _ := registerUser(t, "Reactive User", "reactive@example.com", "password123", "Reactive Org")
	orgID := getOrgID(t, token)

	matchID := setupAppliedMatch(t, token, orgID,
		"https://example.com/job/reactive", "Backend Developer", fixedCountTechDescription)

	prep := createFixedCountPrep(t, token, matchID, "")
	prepID := prep.Preparation.ID
	questions := prep.Preparation.Questions
	if len(questions) < 2 {
		t.Fatalf("expected at least 2 questions for reactivity test, got %d", len(questions))
	}

	// Helper to fetch detail and return answered count
	getAnswered := func() int {
		resp := doRequest(t, "GET", fmt.Sprintf("/api/interview-prep/%d", prepID), nil, token)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("GET /api/interview-prep/%d failed: status=%d", prepID, resp.StatusCode)
		}
		var detail struct {
			Preparation struct {
				Progress struct {
					Answered int `json:"answered"`
				} `json:"progress"`
			} `json:"preparation"`
		}
		parseBody(t, resp, &detail)
		return detail.Preparation.Progress.Answered
	}

	saveAnswer := func(qID uint, answer, assessment string) {
		t.Helper()
		resp := doRequest(t, "POST",
			fmt.Sprintf("/api/interview-prep/%d/questions/%d/answers", prepID, qID),
			map[string]string{"answer": answer, "self_assessment": assessment}, token)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("POST answer for question %d failed: status=%d", qID, resp.StatusCode)
		}
		resp.Body.Close()
	}

	// Start: answered should be 0 (no entries saved yet)
	if got := getAnswered(); got != 0 {
		t.Fatalf("Initial answered count: expected 0, got %d", got)
	}

	// --- Answer Q1 with non-empty text ---
	q1 := questions[0]
	saveAnswer(q1.ID, "first answer", "medium")

	// After saving Q1 non-empty: answered should be 1
	if got := getAnswered(); got != 1 {
		t.Fatalf("After Q1 non-empty: expected answered=1, got %d", got)
	}

	// --- Answer Q2 with empty text (should not increase answered) ---
	q2 := questions[1]
	saveAnswer(q2.ID, "", "")

	// After saving Q2 empty: answered should still be 1 (blank ignored)
	if got := getAnswered(); got != 1 {
		t.Fatalf("After Q2 empty: expected answered=1 (blank ignored), got %d", got)
	}

	// --- Re-answer Q1 with empty text (should decrement answered) ---
	saveAnswer(q1.ID, "", "")

	// After re-saving Q1 empty: answered should be 0
	if got := getAnswered(); got != 0 {
		t.Fatalf("After Q1 empty: expected answered=0, got %d", got)
	}

	// --- Re-answer Q1 with non-empty text (should increment answered) ---
	saveAnswer(q1.ID, "reactivated answer", "high")

	// After re-saving Q1 non-empty: answered should be 1
	if got := getAnswered(); got != 1 {
		t.Fatalf("After Q1 reactivated: expected answered=1, got %d", got)
	}
}
