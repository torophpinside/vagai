package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestRateLimit_AllowRequest(t *testing.T) {
	limiter.clients = make(map[string]*clientLimit)

	w := httptest.NewRecorder()
	_, router := gin.CreateTestContext(w)
	router.Use(RateLimit(5, time.Minute))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestRateLimit_ExceedLimit(t *testing.T) {
	limiter.clients = make(map[string]*clientLimit)

	handler := RateLimit(2, time.Minute)

	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Request.RemoteAddr = "10.0.0.1:1234"

		handler(c)

		if i < 2 && w.Code != http.StatusOK {
			t.Errorf("Request %d: Expected status %d, got %d", i+1, http.StatusOK, w.Code)
		}
		if i == 2 && w.Code != http.StatusTooManyRequests {
			t.Errorf("Request %d: Expected status %d, got %d", i+1, http.StatusTooManyRequests, w.Code)
		}
	}

	if w := httptest.NewRecorder(); true {
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Request.RemoteAddr = "10.0.0.1:1234"
		handler(c)

		remaining := w.Header().Get("X-RateLimit-Remaining")
		if remaining != "0" {
			t.Errorf("Expected X-RateLimit-Remaining = 0, got %v", remaining)
		}

		resetHeader := w.Header().Get("X-RateLimit-Reset")
		if resetHeader == "" {
			t.Error("Expected X-RateLimit-Reset header")
		}

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		if _, ok := response["error"]; !ok {
			t.Error("Expected error message in response")
		}
	}
}

func TestRateLimit_DifferentIPs(t *testing.T) {
	limiter.clients = make(map[string]*clientLimit)

	handler := RateLimit(1, time.Minute)

	w1 := httptest.NewRecorder()
	c1, _ := gin.CreateTestContext(w1)
	c1.Request = httptest.NewRequest("GET", "/test", nil)
	c1.Request.RemoteAddr = "192.168.1.1:1234"

	handler(c1)

	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = httptest.NewRequest("GET", "/test", nil)
	c2.Request.RemoteAddr = "192.168.1.2:1234"

	handler(c2)

	if w2.Code != http.StatusOK {
		t.Errorf("Expected status %d for different IP, got %d", http.StatusOK, w2.Code)
	}
}

func TestRateLimit_WindowReset(t *testing.T) {
	limiter.clients = make(map[string]*clientLimit)

	ip := "10.0.0.2"
	limiter.clients[ip] = &clientLimit{
		count:   5,
		resetAt: time.Now().Add(-1 * time.Second),
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.RemoteAddr = ip + ":1234"

	handler := RateLimit(5, time.Minute)
	handler(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d after window reset, got %d", http.StatusOK, w.Code)
	}
}

func TestRateLimit_CustomWindow(t *testing.T) {
	limiter.clients = make(map[string]*clientLimit)

	w := httptest.NewRecorder()
	_, router := gin.CreateTestContext(w)
	router.Use(RateLimit(10, time.Second))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "10.0.0.3:1234"
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestRateLimit_ConcurrentRequests(t *testing.T) {
	limiter.clients = make(map[string]*clientLimit)

	handler := RateLimit(3, time.Minute)

	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Request.RemoteAddr = "10.0.0.4:1234"
		handler(c)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: Expected status %d, got %d", i+1, http.StatusOK, w.Code)
		}
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.RemoteAddr = "10.0.0.4:1234"
	handler(c)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status %d on 4th request, got %d", http.StatusTooManyRequests, w.Code)
	}
}

func TestRateLimit_RemainingHeader(t *testing.T) {
	limiter.clients = make(map[string]*clientLimit)

	w := httptest.NewRecorder()
	_, router := gin.CreateTestContext(w)
	router.Use(RateLimit(5, time.Minute))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "10.0.0.5:1234"
	router.ServeHTTP(w, req)

	remaining := w.Header().Get("X-RateLimit-Remaining")
	if remaining == "" {
		t.Error("Expected X-RateLimit-Remaining header to be set")
	}
}
