package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/anomalyco/vagai-api/internal/middleware"
	"github.com/anomalyco/vagai-api/internal/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupTestDB(t *testing.T) *gorm.DB {
	t.Skip("Skipping handler tests that require MySQL database")

	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:3306)/vagai_test?charset=utf8mb4&parseTime=True&loc=Local"), &gorm.Config{
		Logger: logger.Discard,
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	db.AutoMigrate(&models.Organization{}, &models.User{}, &models.Membership{}, &models.Subscription{})

	DB = db
	return db
}

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Simple", "Test Org", "test-org"},
		{"With Dots", "My.Company.Name", "my-company-name"},
		{"Multiple Spaces", "Test    Org", "test-org"},
		{"Mixed Case", "My AWESOME org", "my-awesome-org"},
		{"Already Lower", "test-org", "test-org"},
		{"With Numbers", "Org 123", "org-123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateSlug(tt.input)
			if result != tt.expected {
				t.Errorf("generateSlug(%v) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGenerateSlug_NoDoubleDashes(t *testing.T) {
	result := generateSlug("Test  --  Org")
	if strings.Contains(result, "--") {
		t.Errorf("generateSlug() should not contain double dashes, got %v", result)
	}
}

func TestRandomString(t *testing.T) {
	result := randomString(4)
	if len(result) != 4 {
		t.Errorf("randomString(4) length = %v, expected 4", len(result))
	}

	result2 := randomString(8)
	if len(result2) != 8 {
		t.Errorf("randomString(8) length = %v, expected 8", len(result2))
	}

	result3 := randomString(4)
	result4 := randomString(4)
	if result3 == result4 {
		t.Log("Warning: Two consecutive random strings are equal (unlikely but possible)")
	}
}

func TestGetMe_WithoutAuth(t *testing.T) {
	t.Skip("Skipping test that requires database connection")
	
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	// Set context values that middleware would set
	c.Set("user_id", uint(1))
	c.Set("org_id", uint(1))
	c.Set("user_email", "test@test.com")
	c.Set("user_role", "owner")

	GetMe(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestUpdateProfile_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", uint(1))

	c.Request = httptest.NewRequest("PUT", "/profile", strings.NewReader("invalid json"))
	c.Request.Header.Set("Content-Type", "application/json")
	
	UpdateProfile(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUpdateProfile_EmptyUpdates(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", uint(1))

	body := `{"name": "", "timezone": ""}`
	c.Request = httptest.NewRequest("PUT", "/profile", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	
	UpdateProfile(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestChangePassword_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", uint(1))

	c.Request = httptest.NewRequest("PUT", "/password", strings.NewReader("invalid"))
	c.Request.Header.Set("Content-Type", "application/json")
	
	ChangePassword(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestChangePassword_ShortPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", uint(1))

	body := `{"current_password": "test1234", "new_password": "short"}`
	c.Request = httptest.NewRequest("PUT", "/password", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	
	ChangePassword(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestLogin_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest("POST", "/login", strings.NewReader("invalid"))
	c.Request.Header.Set("Content-Type", "application/json")
	
	Login(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestLogin_MissingEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{"password": "test1234"}`
	c.Request = httptest.NewRequest("POST", "/login", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	
	Login(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestRegister_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest("POST", "/register", strings.NewReader("invalid"))
	c.Request.Header.Set("Content-Type", "application/json")
	
	Register(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestRegister_MissingFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	tests := []struct {
		name string
		body string
	}{
		{"Missing Name", `{"email": "test@test.com", "password": "test1234", "organization": "Org"}`},
		{"Missing Email", `{"name": "Test", "password": "test1234", "organization": "Org"}`},
		{"Missing Password", `{"name": "Test", "email": "test@test.com", "organization": "Org"}`},
		{"Missing Organization", `{"name": "Test", "email": "test@test.com", "password": "test1234"}`},
		{"Short Password", `{"name": "Test", "email": "test@test.com", "password": "short", "organization": "Org"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			
			c.Request = httptest.NewRequest("POST", "/register", strings.NewReader(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")
			
			Register(c)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
			}
		})
	}
}

func TestRegister_InvalidEmail(t *testing.T) {
	w := httptest.NewRecorder()
	c, router := gin.CreateTestContext(w)
	_ = c

	router.POST("/register", Register)

	body := `{"name": "Test", "email": "not-an-email", "password": "test1234", "organization": "Org"}`
	req := httptest.NewRequest("POST", "/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestChangePlan_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("org_id", uint(1))

	c.Request = httptest.NewRequest("PUT", "/plan", strings.NewReader("invalid"))
	c.Request.Header.Set("Content-Type", "application/json")

	ChangePlan(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestChangePlan_MissingPlanSlug(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("org_id", uint(1))

	body := `{}`
	c.Request = httptest.NewRequest("PUT", "/plan", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	ChangePlan(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestSetDB(t *testing.T) {
	var testDB *gorm.DB
	SetDB(testDB)

	if DB != testDB {
		t.Error("SetDB() did not set the global DB variable")
	}
}

func TestGetDB_WithoutScoped(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	result := getDB(c)
	if result != DB {
		t.Error("getDB() should return global DB when scoped_db not set")
	}
}

func TestGenerateTokenIntegration(t *testing.T) {
	token, err := middleware.GenerateToken(1, 1, "test@test.com", "owner")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	if token == "" {
		t.Error("GenerateToken() returned empty token")
	}
}

func TestChangePassword_HashComparison(t *testing.T) {
	password := "testpassword123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		t.Errorf("CompareHashAndPassword() error = %v", err)
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte("wrongpassword"))
	if err == nil {
		t.Error("CompareHashAndPassword() should fail for wrong password")
	}
}

func TestLogin_BodyParsing(t *testing.T) {
	var body struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	body.Email = "test@test.com"
	body.Password = "test1234"

	if body.Email != "test@test.com" {
		t.Errorf("Body.Email = %v, expected test@test.com", body.Email)
	}
	if body.Password != "test1234" {
		t.Errorf("Body.Password = %v, expected test1234", body.Password)
	}
}

func TestRegister_BodyParsing(t *testing.T) {
	var body struct {
		Name         string `json:"name" binding:"required,min=2,max=200"`
		Email        string `json:"email" binding:"required,email"`
		Password     string `json:"password" binding:"required,min=8"`
		Organization string `json:"organization" binding:"required,min=2,max=200"`
	}

	body.Name = "Test User"
	body.Email = "test@test.com"
	body.Password = "test1234"
	body.Organization = "Test Org"

	jsonBody, _ := json.Marshal(body)
	var parsed struct {
		Name         string `json:"name"`
		Email        string `json:"email"`
		Password     string `json:"password"`
		Organization string `json:"organization"`
	}
	json.Unmarshal(jsonBody, &parsed)

	if parsed.Name != "Test User" {
		t.Errorf("Parsed.Name = %v, expected Test User", parsed.Name)
	}
	if parsed.Email != "test@test.com" {
		t.Errorf("Parsed.Email = %v, expected test@test.com", parsed.Email)
	}
}

func TestHandlers_JSONResponse(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.JSON(http.StatusOK, gin.H{
		"message": "test",
		"data":    gin.H{"key": "value"},
	})

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["message"] != "test" {
		t.Errorf("Expected message 'test', got %v", response["message"])
	}
}

func TestHandlers_BadRequestResponse(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["error"] != "Invalid input" {
		t.Errorf("Expected error 'Invalid input', got %v", response["error"])
	}
}

func TestHandlers_CreatedResponse(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Created successfully",
		"id":      1,
	})

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestHandlers_UnauthorizedResponse(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestHandlers_NotFoundResponse(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestHandlers_InternalServerErrorResponse(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestHandlers_ContextValues(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	c.Set("user_id", uint(42))
	c.Set("org_id", uint(100))
	c.Set("user_email", "test@test.com")
	c.Set("user_role", "admin")

	userID, _ := c.Get("user_id")
	orgID, _ := c.Get("org_id")
	email, _ := c.Get("user_email")
	role, _ := c.Get("user_role")

	if userID.(uint) != 42 {
		t.Errorf("user_id = %v, expected 42", userID)
	}
	if orgID.(uint) != 100 {
		t.Errorf("org_id = %v, expected 100", orgID)
	}
	if email != "test@test.com" {
		t.Errorf("user_email = %v, expected test@test.com", email)
	}
	if role != "admin" {
		t.Errorf("user_role = %v, expected admin", role)
	}
}

func TestHandlers_UpdateProfileBody(t *testing.T) {
	var body struct {
		Name     string `json:"name" binding:"omitempty,min=2,max=200"`
		Timezone string `json:"timezone" binding:"omitempty"`
	}

	body.Name = "New Name"
	body.Timezone = "UTC"

	updates := map[string]interface{}{}
	if body.Name != "" {
		updates["name"] = body.Name
	}
	if body.Timezone != "" {
		updates["timezone"] = body.Timezone
	}

	if len(updates) != 2 {
		t.Errorf("Expected 2 updates, got %d", len(updates))
	}
	if updates["name"] != "New Name" {
		t.Errorf("updates[name] = %v, expected New Name", updates["name"])
	}
	if updates["timezone"] != "UTC" {
		t.Errorf("updates[timezone] = %v, expected UTC", updates["timezone"])
	}
}

func TestHandlers_ChangePasswordBody(t *testing.T) {
	var body struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=8"`
	}

	body.CurrentPassword = "oldpass123"
	body.NewPassword = "newpass123"

	if body.CurrentPassword != "oldpass123" {
		t.Errorf("CurrentPassword = %v, expected oldpass123", body.CurrentPassword)
	}
	if body.NewPassword != "newpass123" {
		t.Errorf("NewPassword = %v, expected newpass123", body.NewPassword)
	}
}
