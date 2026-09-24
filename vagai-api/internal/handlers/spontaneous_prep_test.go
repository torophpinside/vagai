package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// doSpontaneous invoca o handler diretamente com um org_id no contexto.
// US2 agora persiste, então os casos de sucesso (geração + persistência) vivem
// nos testes de ponta a ponta (spontaneous_prep_e2e_test.go). Aqui ficam apenas
// os casos de validação pré-banco: erros que ocorrem antes de qualquer acesso
// ao banco (org ausente, JSON malformado, descrição vazia).
func doSpontaneous(t *testing.T, orgID uint, body string) *httptest.ResponseRecorder {
	t.Helper()
	t.Setenv("LMSTUDIO_URL", "http://127.0.0.1:59999")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("org_id", orgID)
	c.Request = httptest.NewRequest("POST", "/api/interview-prep/spontaneous", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	CreateSpontaneousPreparation(c)
	return w
}

const handlerSpontaneousDesc = "Vaga backend com Go, MySQL e Docker em produção."

// T005 [US1] Descrição ausente => 400 (antes de qualquer escrita).
func TestCreateSpontaneousPreparation_MissingDescription(t *testing.T) {
	w := doSpontaneous(t, 42, `{"title":"Backend Developer","company":"Tech Corp"}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d (body=%s)", w.Code, w.Body.String())
	}
	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil || body.Error == "" {
		t.Errorf("expected error message, got %s", w.Body.String())
	}
}

// T005 [US1] JSON malformado => 400.
func TestCreateSpontaneousPreparation_MalformedJSON(t *testing.T) {
	w := doSpontaneous(t, 42, `{"description":`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d (body=%s)", w.Code, w.Body.String())
	}
}

// T011 [US2] Claim de organização ausente => 401.
func TestCreateSpontaneousPreparation_MissingOrganizationClaim(t *testing.T) {
	w := doSpontaneous(t, 0, `{"description":"`+handlerSpontaneousDesc+`"}`)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d (body=%s)", w.Code, w.Body.String())
	}
}