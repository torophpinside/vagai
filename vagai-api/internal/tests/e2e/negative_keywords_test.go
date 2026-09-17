package e2e

import (
	"net/http"
	"testing"
)

func TestNegativeKeywords_DefaultEmpty(t *testing.T) {
	token, _ := registerUser(t, "NK Default", "nkdefault@example.com", "password123", "NK Default Org")

	resp := doRequest(t, "GET", "/api/negative-keywords", nil, token)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}

	var keywords []string
	parseBody(t, resp, &keywords)
	if len(keywords) != 0 {
		t.Errorf("expected no keywords by default, got %v", keywords)
	}
}

func TestNegativeKeywords_UpdateAndRead(t *testing.T) {
	token, _ := registerUser(t, "NK Update", "nkupdate@example.com", "password123", "NK Update Org")

	resp := doRequest(t, "PUT", "/api/negative-keywords", map[string][]string{
		"keywords": {"freela", "PJ", "bilingüe"},
	}, token)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK on update, got %d", resp.StatusCode)
	}

	resp = doRequest(t, "GET", "/api/negative-keywords", nil, token)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK on read, got %d", resp.StatusCode)
	}

	var keywords []string
	parseBody(t, resp, &keywords)
	if len(keywords) != 3 {
		t.Fatalf("expected 3 keywords, got %v", keywords)
	}

	want := map[string]bool{"freela": true, "PJ": true, "bilingüe": true}
	for _, kw := range keywords {
		if !want[kw] {
			t.Errorf("unexpected keyword %q in %v", kw, keywords)
		}
	}
}

func TestNegativeKeywords_Unauthorized(t *testing.T) {
	resp := doRequest(t, "GET", "/api/negative-keywords", nil, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d", resp.StatusCode)
	}
}
