package e2e

import (
	"net/http"
	"testing"
)

func addSite(t *testing.T, token string, overrides map[string]interface{}) uint {
	t.Helper()

	body := map[string]interface{}{
		"name": "Test Site",
		"url":  "https://testsite.com/jobs",
	}
	for k, v := range overrides {
		body[k] = v
	}

	resp := doRequest(t, "POST", "/api/sites", body, token)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("add site failed: status=%d", resp.StatusCode)
	}

	var result struct {
		ID uint `json:"id"`
	}
	parseBody(t, resp, &result)
	return result.ID
}

func TestAddSite_Success(t *testing.T) {
	token, _ := registerUser(t, "Site Creator", "sitecreator@example.com", "password123", "Site Org")

	id := addSite(t, token, nil)
	if id == 0 {
		t.Error("expected non-zero site ID")
	}
}

func TestAddSite_WithoutToken(t *testing.T) {
	resp := doRequest(t, "POST", "/api/sites", map[string]string{
		"name": "No Auth Site",
		"url":  "https://noauth.com",
	}, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d", resp.StatusCode)
	}
}

func TestListSites_Success(t *testing.T) {
	token, _ := registerUser(t, "List Sites", "listsites@example.com", "password123", "ListSites Org")

	addSite(t, token, map[string]interface{}{"name": "Site A", "url": "https://a.com"})
	addSite(t, token, map[string]interface{}{"name": "Site B", "url": "https://b.com"})

	resp := doRequest(t, "GET", "/api/sites", nil, token)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}

	var sites []struct {
		ID     uint   `json:"id"`
		Name   string `json:"name"`
		URL    string `json:"url"`
		Active bool   `json:"active"`
	}
	parseBody(t, resp, &sites)

	if len(sites) != 2 {
		t.Errorf("expected 2 sites, got %d", len(sites))
	}

	if !sites[0].Active {
		t.Error("expected new site to be active by default")
	}
}

func TestToggleSiteActive(t *testing.T) {
	token, _ := registerUser(t, "Toggle Site", "togglesite@example.com", "password123", "Toggle Org")

	siteID := addSite(t, token, map[string]interface{}{
		"name": "Toggleable",
		"url":  "https://toggle.com",
	})

	resp := doRequest(t, "PATCH", "/api/sites/"+intToStr(siteID), map[string]bool{
		"active": false,
	}, token)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK on deactivate, got %d", resp.StatusCode)
	}

	var result struct {
		Active bool `json:"active"`
	}
	parseBody(t, resp, &result)
	if result.Active {
		t.Error("expected site to be inactive after toggle")
	}

	resp = doRequest(t, "PATCH", "/api/sites/"+intToStr(siteID), map[string]bool{
		"active": true,
	}, token)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK on reactivate, got %d", resp.StatusCode)
	}

	parseBody(t, resp, &result)
	if !result.Active {
		t.Error("expected site to be active after reactivate")
	}
}

func TestDeleteSite(t *testing.T) {
	token, _ := registerUser(t, "Delete Site", "deletesite@example.com", "password123", "Delete Org")

	siteID := addSite(t, token, map[string]interface{}{
		"name": "To Delete",
		"url":  "https://todelete.com",
	})

	resp := doRequest(t, "DELETE", "/api/sites/"+intToStr(siteID), nil, token)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK on delete, got %d", resp.StatusCode)
	}

	resp = doRequest(t, "GET", "/api/sites", nil, token)
	var sites []interface{}
	parseBody(t, resp, &sites)
	if len(sites) != 0 {
		t.Errorf("expected 0 sites after delete, got %d", len(sites))
	}
}

func TestUpdateSite_NotFound(t *testing.T) {
	token, _ := registerUser(t, "NotFound Site", "notfoundsite@example.com", "password123", "NF Org")

	resp := doRequest(t, "PATCH", "/api/sites/99999", map[string]bool{
		"active": false,
	}, token)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 Not Found, got %d", resp.StatusCode)
	}
}
