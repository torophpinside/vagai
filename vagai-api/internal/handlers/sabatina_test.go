package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCreateSabatinaQuestions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/sabatinas/:id/questions", CreateSabatinaQuestions)

	body := ` [{"text":"Descreva um problema que você resolveu."}] `
	req := httptest.NewRequest(http.MethodPost, "/api/sabatinas/00000000-0000-0000-0000-000000000000/questions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "questions")
}
