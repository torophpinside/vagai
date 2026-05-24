package e2e

import (
	"net/http"
	"testing"
)

func TestRegister_Success(t *testing.T) {
	token, userID := registerUser(t, "Test User", "test@example.com", "password123", "Test Org")

	if token == "" {
		t.Error("expected non-empty token")
	}
	if userID == 0 {
		t.Error("expected non-zero user ID")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	registerUser(t, "User One", "dupe@example.com", "password123", "Org One")

	resp := doRequest(t, "POST", "/api/auth/register", map[string]string{
		"name":         "User Two",
		"email":        "dupe@example.com",
		"password":     "password123",
		"organization": "Org Two",
	}, "")
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("expected 409 Conflict, got %d", resp.StatusCode)
	}
}

func TestRegister_ValidationErrors(t *testing.T) {
	tests := []struct {
		name       string
		body       map[string]string
		expectCode int
	}{
		{"empty name", map[string]string{"name": "", "email": "a@b.com", "password": "12345678", "organization": "Org"}, http.StatusBadRequest},
		{"invalid email", map[string]string{"name": "User", "email": "notanemail", "password": "12345678", "organization": "Org"}, http.StatusBadRequest},
		{"short password", map[string]string{"name": "User", "email": "a@b.com", "password": "123", "organization": "Org"}, http.StatusBadRequest},
		{"short org", map[string]string{"name": "User", "email": "a@b.com", "password": "12345678", "organization": "A"}, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := doRequest(t, "POST", "/api/auth/register", tt.body, "")
			if resp.StatusCode != tt.expectCode {
				t.Errorf("expected %d, got %d", tt.expectCode, resp.StatusCode)
			}
		})
	}
}

func TestLogin_Success(t *testing.T) {
	password := "password123"
	registerUser(t, "Login User", "login@example.com", password, "Login Org")

	resp := doRequest(t, "POST", "/api/auth/login", map[string]string{
		"email":    "login@example.com",
		"password": password,
	}, "")
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}

	var result struct {
		Token string `json:"token"`
	}
	parseBody(t, resp, &result)

	if result.Token == "" {
		t.Error("expected non-empty token on login")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	registerUser(t, "Wrong Pass", "wrongpass@example.com", "correctpw123", "Org")

	resp := doRequest(t, "POST", "/api/auth/login", map[string]string{
		"email":    "wrongpass@example.com",
		"password": "wrongpassword",
	}, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d", resp.StatusCode)
	}
}

func TestLogin_NonExistentEmail(t *testing.T) {
	resp := doRequest(t, "POST", "/api/auth/login", map[string]string{
		"email":    "nobody@example.com",
		"password": "password123",
	}, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d", resp.StatusCode)
	}
}

func TestGetMe_WithValidToken(t *testing.T) {
	token, _ := registerUser(t, "Me User", "me@example.com", "password123", "Me Org")

	resp := doRequest(t, "GET", "/api/me", nil, token)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}

	var result struct {
		User struct {
			ID    uint   `json:"id"`
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"user"`
		Organization struct {
			ID   uint   `json:"id"`
			Name string `json:"name"`
			Plan string `json:"plan"`
		} `json:"organization"`
	}
	parseBody(t, resp, &result)

	if result.User.Name != "Me User" {
		t.Errorf("expected name 'Me User', got '%s'", result.User.Name)
	}
	if result.Organization.Plan != "free" {
		t.Errorf("expected plan 'free', got '%s'", result.Organization.Plan)
	}
}

func TestGetMe_WithoutToken(t *testing.T) {
	resp := doRequest(t, "GET", "/api/me", nil, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d", resp.StatusCode)
	}
}

func TestGetMe_WithInvalidToken(t *testing.T) {
	resp := doRequest(t, "GET", "/api/me", nil, "Bearer invalidtoken123")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d", resp.StatusCode)
	}
}

func TestHealth(t *testing.T) {
	resp := doRequest(t, "GET", "/health", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}

	var result map[string]string
	parseBody(t, resp, &result)
	if result["status"] != "ok" {
		t.Errorf("expected status 'ok', got '%s'", result["status"])
	}
}
