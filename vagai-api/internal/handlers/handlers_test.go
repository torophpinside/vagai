package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestListJobs_WithoutDB(t *testing.T) {
	t.Skip("Skipping test that requires database connection")
	
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest("GET", "/jobs", nil)
	ListJobs(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestListJobs_WithStatusFilter(t *testing.T) {
	t.Skip("Skipping test that requires database connection")
}

func TestListJobs_WithSiteFilter(t *testing.T) {
	t.Skip("Skipping test that requires database connection")
}

func TestGetJob_InvalidID(t *testing.T) {
	t.Skip("Skipping test that requires database connection")
}

func TestGetJob_NonExistent(t *testing.T) {
	t.Skip("Skipping test that requires database connection")
}

func TestUpdateJobStatus_InvalidBody(t *testing.T) {
	t.Skip("Skipping test that requires database connection")
}

func TestUpdateJobStatus_NonExistent(t *testing.T) {
	t.Skip("Skipping test that requires database connection")
}

func TestListMatches_WithThreshold(t *testing.T) {
	t.Skip("Skipping test that requires database connection")
}

func TestListMatches_WithSort(t *testing.T) {
	t.Skip("Skipping test that requires database connection")
}

func TestListMatches_WithAppliedFilter(t *testing.T) {
	t.Skip("Skipping test that requires database connection")
}

func TestListMatches_WithSiteFilter(t *testing.T) {
	t.Skip("Skipping test that requires database connection")
}

func TestUpdateMatch_InvalidBody(t *testing.T) {
	t.Skip("Skipping test that requires database connection")
}

func TestUpdateMatch_NonExistent(t *testing.T) {
	t.Skip("Skipping test that requires database connection")
}

func TestDeleteMatch_NonExistent(t *testing.T) {
	t.Skip("Skipping test that requires database connection")
}

func TestGetStats_WithoutDB(t *testing.T) {
	t.Skip("Skipping test that requires database connection")
}

func TestListSites_WithoutDB(t *testing.T) {
	t.Skip("Skipping test that requires database connection")
}

func TestDeleteSite_NonExistent(t *testing.T) {
	t.Skip("Skipping test that requires database connection")
}

func TestAddSite_InvalidBody(t *testing.T) {
	w := httptest.NewRecorder()
	c, router := gin.CreateTestContext(w)
	_ = c

	router.POST("/sites", AddSite)

	req := httptest.NewRequest("POST", "/sites", strings.NewReader("invalid"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestListResumes_WithoutDB(t *testing.T) {
	t.Skip("Skipping test that requires database connection")
}

func TestUploadResume_MissingFile(t *testing.T) {
	w := httptest.NewRecorder()
	c, router := gin.CreateTestContext(w)
	_ = c

	router.POST("/resumes/upload", UploadResume)

	req := httptest.NewRequest("POST", "/resumes/upload", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAnalyzeResume_MissingFile(t *testing.T) {
	w := httptest.NewRecorder()
	c, router := gin.CreateTestContext(w)
	_ = c

	router.POST("/resumes/analyze", AnalyzeResume)

	req := httptest.NewRequest("POST", "/resumes/analyze", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestQueryParameters(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/test?status=new&site=1&threshold=5&sort=desc&applied=true", nil)

	if c.Query("status") != "new" {
		t.Errorf("Query status = %v, expected new", c.Query("status"))
	}
	if c.Query("site") != "1" {
		t.Errorf("Query site = %v, expected 1", c.Query("site"))
	}
	if c.DefaultQuery("threshold", "1") != "5" {
		t.Errorf("Query threshold = %v, expected 5", c.DefaultQuery("threshold", "1"))
	}
	if c.DefaultQuery("sort", "desc") != "desc" {
		t.Errorf("Query sort = %v, expected desc", c.DefaultQuery("sort", "desc"))
	}
	if c.DefaultQuery("applied", "false") != "true" {
		t.Errorf("Query applied = %v, expected true", c.DefaultQuery("applied", "false"))
	}
}

func TestParamParsing(t *testing.T) {
	tests := []struct {
		name     string
		param    string
		expected uint
		valid    bool
	}{
		{"Valid ID", "123", 123, true},
		{"Zero", "0", 0, true},
		{"Large ID", "999999", 999999, true},
		{"Invalid", "abc", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Params = gin.Params{{Key: "id", Value: tt.param}}

			id := c.Param("id")
			if id != tt.param {
				t.Errorf("Param id = %v, expected %v", id, tt.param)
			}
		})
	}
}

func TestShouldBindJSON(t *testing.T) {
	var body struct {
		Status string `json:"status"`
	}

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("PUT", "/test", strings.NewReader(`{"status": "matched"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	err := c.ShouldBindJSON(&body)
	if err != nil {
		t.Fatalf("ShouldBindJSON() error = %v", err)
	}
	if body.Status != "matched" {
		t.Errorf("body.Status = %v, expected matched", body.Status)
	}
}

func TestShouldBindJSON_Invalid(t *testing.T) {
	var body struct {
		Status string `json:"status"`
	}

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("PUT", "/test", strings.NewReader(`invalid json`))
	c.Request.Header.Set("Content-Type", "application/json")

	err := c.ShouldBindJSON(&body)
	if err == nil {
		t.Error("ShouldBindJSON() should return error for invalid JSON")
	}
}

func TestHandler_ContextSetGet(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	c.Set("scoped_db", "test_db")
	c.Set("user_id", uint(1))
	c.Set("org_id", uint(10))

	scopedDB, exists := c.Get("scoped_db")
	if !exists {
		t.Fatal("scoped_db not found")
	}
	if scopedDB != "test_db" {
		t.Errorf("scoped_db = %v, expected test_db", scopedDB)
	}

	userID, _ := c.Get("user_id")
	if userID.(uint) != 1 {
		t.Errorf("user_id = %v, expected 1", userID)
	}
}

func TestHandler_FormFile(t *testing.T) {
	w := httptest.NewRecorder()
	c, router := gin.CreateTestContext(w)
	_ = c

	router.POST("/upload", func(c *gin.Context) {
		_, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Arquivo não enviado"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("POST", "/upload", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandler_MultipartRequest(t *testing.T) {
	w := httptest.NewRecorder()

	body := "--boundary\r\n" +
		"Content-Disposition: form-data; name=\"file\"; filename=\"test.txt\"\r\n" +
		"Content-Type: text/plain\r\n\r\n" +
		"test content\r\n" +
		"--boundary--\r\n"

	req := httptest.NewRequest("POST", "/upload", strings.NewReader(body))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=boundary")

	c, router := gin.CreateTestContext(w)
	_ = c

	router.POST("/upload", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"filename": file.Filename})
	})

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}
