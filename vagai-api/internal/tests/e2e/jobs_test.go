package e2e

import (
	"net/http"
	"testing"
)

func createJob(t *testing.T, token string, overrides map[string]interface{}) uint {
	t.Helper()

	body := map[string]interface{}{
		"title":       "Software Engineer",
		"company":     "Tech Corp",
		"url":         "https://example.com/job/1",
		"description": "A great job opportunity",
	}
	for k, v := range overrides {
		body[k] = v
	}

	resp := doRequest(t, "POST", "/api/jobs", body, token)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create job failed: status=%d", resp.StatusCode)
	}

	var result struct {
		ID uint `json:"id"`
	}
	parseBody(t, resp, &result)
	return result.ID
}

func TestCreateJob_Success(t *testing.T) {
	token, _ := registerUser(t, "Job Creator", "jobcreator@example.com", "password123", "Job Org")

	id := createJob(t, token, nil)
	if id == 0 {
		t.Error("expected non-zero job ID")
	}
}

func TestCreateJob_DuplicateURL(t *testing.T) {
	token, _ := registerUser(t, "Dupe Job", "dupejob@example.com", "password123", "Dupe Org")

	createJob(t, token, map[string]interface{}{
		"url": "https://example.com/job/dupe",
	})

	resp := doRequest(t, "POST", "/api/jobs", map[string]interface{}{
		"title":       "Another Job",
		"company":     "Tech Corp",
		"url":         "https://example.com/job/dupe",
		"description": "Duplicate URL",
	}, token)
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("expected 409 Conflict for duplicate URL, got %d", resp.StatusCode)
	}
}

func TestCreateJob_WithoutToken(t *testing.T) {
	resp := doRequest(t, "POST", "/api/jobs", map[string]interface{}{
		"title":   "No Auth Job",
		"company": "Corp",
		"url":     "https://example.com/job/noauth",
	}, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d", resp.StatusCode)
	}
}

func TestCreateJob_Validation(t *testing.T) {
	token, _ := registerUser(t, "Val Job", "valjob@example.com", "password123", "Val Org")

	resp := doRequest(t, "POST", "/api/jobs", map[string]interface{}{
		"company": "No Title Corp",
		"url":     "https://example.com/job/notitle",
	}, token)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", resp.StatusCode)
	}

	resp = doRequest(t, "POST", "/api/jobs", map[string]interface{}{
		"title":   "No URL Job",
		"company": "Corp",
	}, token)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for missing URL, got %d", resp.StatusCode)
	}
}

func TestListJobs_Success(t *testing.T) {
	token, _ := registerUser(t, "List Jobs", "listjobs@example.com", "password123", "List Org")

	createJob(t, token, map[string]interface{}{"url": "https://example.com/job/list1"})
	createJob(t, token, map[string]interface{}{"url": "https://example.com/job/list2"})
	createJob(t, token, map[string]interface{}{"url": "https://example.com/job/list3"})

	resp := doRequest(t, "GET", "/api/jobs", nil, token)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}

	var result struct {
		Data []struct {
			ID    uint   `json:"id"`
			Title string `json:"title"`
		} `json:"data"`
	}
	parseBody(t, resp, &result)

	if len(result.Data) != 3 {
		t.Errorf("expected 3 jobs, got %d", len(result.Data))
	}
}

func TestListJobs_Pagination(t *testing.T) {
	token, _ := registerUser(t, "Pag Jobs", "pagjobs@example.com", "password123", "Pag Org")

	for i := 0; i < 5; i++ {
		createJob(t, token, map[string]interface{}{
			"url": "https://example.com/job/pag" + intToStr(uint(i)),
		})
	}

	resp := doRequest(t, "GET", "/api/jobs?page=1&limit=2", nil, token)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}

	var result struct {
		Data       []interface{} `json:"data"`
		Total      int           `json:"total"`
		Page       int           `json:"page"`
		Limit      int           `json:"limit"`
		TotalPages int           `json:"totalPages"`
	}
	parseBody(t, resp, &result)

	if len(result.Data) != 2 {
		t.Errorf("expected 2 jobs on page 1, got %d", len(result.Data))
	}
	if result.Total != 5 {
		t.Errorf("expected total 5, got %d", result.Total)
	}
	if result.TotalPages != 3 {
		t.Errorf("expected 3 total pages, got %d", result.TotalPages)
	}
}

func TestUpdateJobStatus_ToMatched(t *testing.T) {
	token, _ := registerUser(t, "Match Job", "matchjob@example.com", "password123", "Match Org")

	jobID := createJob(t, token, map[string]interface{}{
		"url": "https://example.com/job/tomatch",
	})

	resp := doRequest(t, "PATCH", "/api/jobs/"+intToStr(jobID), map[string]string{
		"status": "matched",
	}, token)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}

	resp = doRequest(t, "GET", "/api/jobs", nil, token)
	var result struct {
		Data []struct {
			ID     uint   `json:"id"`
			Status string `json:"status"`
		} `json:"data"`
	}
	parseBody(t, resp, &result)

	for _, j := range result.Data {
		if j.ID == jobID && j.Status != "matched" {
			t.Errorf("expected status 'matched', got '%s'", j.Status)
		}
	}
}

func TestGetJob_NotFound(t *testing.T) {
	token, _ := registerUser(t, "Get Job", "getjob@example.com", "password123", "Get Org")

	resp := doRequest(t, "GET", "/api/jobs/99999", nil, token)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 Not Found, got %d", resp.StatusCode)
	}
}

func intToStr(n uint) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}
