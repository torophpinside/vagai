package e2e

import (
	"net/http"
	"testing"
)

func createMatch(t *testing.T, token string, jobID, resumeID uint) {
	t.Helper()

	resp := doRequest(t, "PATCH", "/api/jobs/"+intToStr(jobID), map[string]string{
		"status": "matched",
	}, token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("update job status failed: status=%d", resp.StatusCode)
	}
}

func TestListMatches_Empty(t *testing.T) {
	token, _ := registerUser(t, "Empty Match", "emptymatch@example.com", "password123", "Empty Org")

	resp := doRequest(t, "GET", "/api/matches", nil, token)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}

	var result struct {
		Data []interface{} `json:"data"`
	}
	parseBody(t, resp, &result)

	if len(result.Data) != 0 {
		t.Errorf("expected 0 matches, got %d", len(result.Data))
	}
}

func TestListMatches_AfterMatchingJob(t *testing.T) {
	token, _ := registerUser(t, "Match List", "matchlist@example.com", "password123", "MatchList Org")

	jobID := createJob(t, token, map[string]interface{}{
		"url": "https://example.com/job/formatch",
	})

	createMatch(t, token, jobID, 0)

	resp := doRequest(t, "GET", "/api/matches", nil, token)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}

	var result struct {
		Data []struct {
			ID              uint   `json:"id"`
			SimilarityScore float64 `json:"similarity_score"`
		} `json:"data"`
	}
	parseBody(t, resp, &result)
}

func TestListMatches_AppliedFilter(t *testing.T) {
	token, _ := registerUser(t, "Applied Filter", "appliedfilter@example.com", "password123", "Applied Org")

	jobID := createJob(t, token, map[string]interface{}{
		"url": "https://example.com/job/applied",
	})
	createMatch(t, token, jobID, 0)

	resp := doRequest(t, "GET", "/api/matches", nil, token)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}

	var result struct {
		Data []struct {
			ID      uint `json:"id"`
			Applied bool `json:"applied"`
		} `json:"data"`
	}
	parseBody(t, resp, &result)

	if len(result.Data) > 0 && result.Data[0].Applied {
		t.Error("expected match to be not applied by default")
	}
}
