package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestGenerateToken(t *testing.T) {
	token, err := GenerateToken(1, 10, "test@example.com", "owner")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	if token == "" {
		t.Error("GenerateToken() returned empty token")
	}

	claims := &Claims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(getJWTSecret()), nil
	})
	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}
	if !parsedToken.Valid {
		t.Error("Token is not valid")
	}
	if claims.UserID != 1 {
		t.Errorf("Claims.UserID = %v, expected 1", claims.UserID)
	}
	if claims.OrganizationID != 10 {
		t.Errorf("Claims.OrganizationID = %v, expected 10", claims.OrganizationID)
	}
	if claims.Email != "test@example.com" {
		t.Errorf("Claims.Email = %v, expected test@example.com", claims.Email)
	}
	if claims.Role != "owner" {
		t.Errorf("Claims.Role = %v, expected owner", claims.Role)
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	token, err := GenerateRefreshToken(1, 10)
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}
	if token == "" {
		t.Error("GenerateRefreshToken() returned empty token")
	}

	claims := &jwt.RegisteredClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(getJWTSecret() + "_refresh"), nil
	})
	if err != nil {
		t.Fatalf("Failed to parse refresh token: %v", err)
	}
	if !parsedToken.Valid {
		t.Error("Refresh token is not valid")
	}
}

func TestJWTAuth_MissingToken(t *testing.T) {
	w := httptest.NewRecorder()
	_, router := gin.CreateTestContext(w)
	router.Use(JWTAuth())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	if _, ok := response["error"]; !ok {
		t.Error("Expected error message in response")
	}
}

func TestJWTAuth_InvalidFormat(t *testing.T) {
	w := httptest.NewRecorder()
	_, router := gin.CreateTestContext(w)
	router.Use(JWTAuth())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "InvalidToken123")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestJWTAuth_InvalidToken(t *testing.T) {
	w := httptest.NewRecorder()
	_, router := gin.CreateTestContext(w)
	router.Use(JWTAuth())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestJWTAuth_ValidToken(t *testing.T) {
	token, err := GenerateToken(1, 10, "test@example.com", "owner")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	w := httptest.NewRecorder()
	_, router := gin.CreateTestContext(w)
	router.Use(JWTAuth())
	router.GET("/test", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		orgID, _ := c.Get("org_id")
		email, _ := c.Get("user_email")
		role, _ := c.Get("user_role")

		c.JSON(http.StatusOK, gin.H{
			"user_id": userID,
			"org_id":  orgID,
			"email":   email,
			"role":    role,
		})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["user_id"].(float64) != 1 {
		t.Errorf("Expected user_id 1, got %v", response["user_id"])
	}
	if response["org_id"].(float64) != 10 {
		t.Errorf("Expected org_id 10, got %v", response["org_id"])
	}
	if response["email"] != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %v", response["email"])
	}
	if response["role"] != "owner" {
		t.Errorf("Expected role owner, got %v", response["role"])
	}
}

func TestJWTAuth_ExpiredToken(t *testing.T) {
	claims := Claims{
		UserID:         1,
		OrganizationID: 10,
		Email:          "test@example.com",
		Role:           "owner",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(getJWTSecret()))

	w := httptest.NewRecorder()
	_, router := gin.CreateTestContext(w)
	router.Use(JWTAuth())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d for expired token, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestGetJWTSecret_Env(t *testing.T) {
	os.Setenv("JWT_SECRET", "custom-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	secret := getJWTSecret()
	if secret != "custom-secret-key" {
		t.Errorf("getJWTSecret() = %v, expected custom-secret-key", secret)
	}
}

func TestGetJWTSecret_Default(t *testing.T) {
	os.Unsetenv("JWT_SECRET")

	secret := getJWTSecret()
	expected := "vagai-super-secret-key-change-in-production"
	if secret != expected {
		t.Errorf("getJWTSecret() = %v, expected %v", secret, expected)
	}
}

func TestClaimsStruct(t *testing.T) {
	claims := Claims{
		UserID:         42,
		OrganizationID: 100,
		Email:          "claims@test.com",
		Role:           "admin",
	}

	if claims.UserID != 42 {
		t.Errorf("Claims.UserID = %v, expected 42", claims.UserID)
	}
	if claims.OrganizationID != 100 {
		t.Errorf("Claims.OrganizationID = %v, expected 100", claims.OrganizationID)
	}
	if claims.Email != "claims@test.com" {
		t.Errorf("Claims.Email = %v, expected claims@test.com", claims.Email)
	}
	if claims.Role != "admin" {
		t.Errorf("Claims.Role = %v, expected admin", claims.Role)
	}
}
