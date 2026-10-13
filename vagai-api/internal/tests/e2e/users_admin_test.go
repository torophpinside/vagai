package e2e

import (
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/anomalyco/vagai-api/internal/models"
)

// listUsersResponse espelha o envelope de ListUsersAdmin.
type listUsersResponse struct {
	Data []struct {
		ID        uint   `json:"id"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		PlanSlug  string `json:"plan_slug"`
		Suspended bool   `json:"suspended"`
	} `json:"data"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"totalPages"`
}

func TestUsersAdmin(t *testing.T) {
	sfx := uniqueSuffix()
	userToken, _ := registerUser(t, "Comum Gestao", "comum-gestao-"+sfx+"@example.com", "senhaSegura123", "")
	adminToken := promoteAdmin(t, "admin-gestao-"+sfx+"@example.com")

	t.Run("nao-admin recebe 403 em todas as rotas", func(t *testing.T) {
		cases := []struct {
			method string
			path   string
			body   interface{}
		}{
			{"GET", "/api/admin/users", nil},
			{"GET", "/api/admin/users/1", nil},
			{"PATCH", "/api/admin/users/1", map[string]interface{}{"name": "X"}},
			{"PATCH", "/api/admin/users/1/plan", map[string]interface{}{"plan_slug": "free"}},
			{"POST", "/api/admin/users/1/suspend", nil},
			{"POST", "/api/admin/users/1/restore", nil},
			{"DELETE", "/api/admin/users/1", nil},
			{"POST", "/api/admin/users/1/reactivate", nil},
		}
		for _, tc := range cases {
			resp := doRequest(t, tc.method, tc.path, tc.body, userToken)
			if resp.StatusCode != http.StatusForbidden {
				resp.Body.Close()
				t.Fatalf("%s %s: esperado 403, obtido %d", tc.method, tc.path, resp.StatusCode)
			}
			resp.Body.Close()
		}
	})

	// Alvo comum dedicado para os testes de listagem/perfil/edição.
	_, targetID := registerUser(t, "Alvo Listagem", "alvo-list-"+sfx+"@example.com", "senhaSegura123", "")

	t.Run("lista com envelope, busca, filtros e paginacao", func(t *testing.T) {
		resp := doRequest(t, "GET", "/api/admin/users?page=1&limit=5", nil, adminToken)
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			t.Fatalf("esperado 200, obtido %d", resp.StatusCode)
		}
		var page listUsersResponse
		parseBody(t, resp, &page)
		if page.Total <= 0 || page.Page != 1 || page.Limit != 5 || len(page.Data) == 0 {
			t.Fatalf("envelope inesperado: %+v", page)
		}
		for _, u := range page.Data {
			if u.ID == 0 || u.Email == "" {
				t.Fatalf("registro sem dados básicos: %+v", u)
			}
		}

		// Busca por e-mail parcial encontra o alvo.
		resp = doRequest(t, "GET", "/api/admin/users?search=alvo-list-"+sfx, nil, adminToken)
		var found listUsersResponse
		parseBody(t, resp, &found)
		if found.Total != 1 || found.Data[0].ID != targetID {
			t.Fatalf("busca inesperada: %+v", found)
		}

		// Admin nunca aparece na listagem (FR-002).
		for _, u := range found.Data {
			if u.Email == "admin-gestao-"+sfx+"@example.com" {
				t.Fatal("admin não deve aparecer na listagem")
			}
		}

		// Filtro suspenso: vazio antes da suspensão.
		resp = doRequest(t, "GET", "/api/admin/users?status=suspended&search=alvo-list-"+sfx, nil, adminToken)
		var susp listUsersResponse
		parseBody(t, resp, &susp)
		if susp.Total != 0 {
			t.Fatalf("esperado lista vazia de suspensos, obtido %d", susp.Total)
		}

		// Situação inválida → 422.
		resp = doRequest(t, "GET", "/api/admin/users?status=whatever", nil, adminToken)
		if resp.StatusCode != http.StatusUnprocessableEntity {
			resp.Body.Close()
			t.Fatalf("esperado 422 para status inválido, obtido %d", resp.StatusCode)
		}
		resp.Body.Close()
	})

	t.Run("perfil com contadores de atividade", func(t *testing.T) {
		// Conteúdo de atividade para o alvo.
		if err := testDB.Create(&models.Job{UserID: targetID, Title: "Dev Go", URL: "http://j/" + sfx}).Error; err != nil {
			t.Fatalf("criar job: %v", err)
		}
		if err := testDB.Create(&models.Resume{UserID: targetID, Name: "CV Teste", Data: "[]", UploadedAt: time.Now()}).Error; err != nil {
			t.Fatalf("criar resume: %v", err)
		}
		if err := testDB.Create(&models.ResumeAnalysis{UserID: targetID, FileName: "cv.pdf", FullAnalysis: "ok", Strengths: "[]", Weaknesses: "[]", Suggestions: "[]"}).Error; err != nil {
			t.Fatalf("criar resume analysis: %v", err)
		}
		job := models.Job{}
		testDB.Where("user_id = ?", targetID).First(&job)
		resume := models.Resume{}
		testDB.Where("user_id = ?", targetID).First(&resume)
		appliedAt := time.Now()
		if err := testDB.Create(&models.Match{UserID: targetID, JobID: job.ID, ResumeID: &resume.ID, SimilarityScore: 88, KeywordsMatched: "[]", Applied: true, AppliedAt: &appliedAt}).Error; err != nil {
			t.Fatalf("criar match: %v", err)
		}

		resp := doRequest(t, "GET", "/api/admin/users/"+strconv.FormatUint(uint64(targetID), 10), nil, adminToken)
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			t.Fatalf("esperado 200, obtido %d", resp.StatusCode)
		}
		var profile struct {
			ID       uint   `json:"id"`
			Email    string `json:"email"`
			Activity struct {
				Jobs    int64 `json:"jobs"`
				Resumes int64 `json:"resumes"`
				Matches int64 `json:"matches"`
				Applied int64 `json:"applied"`
			} `json:"activity"`
		}
		parseBody(t, resp, &profile)
		if profile.ID != targetID || profile.Activity.Jobs < 1 || profile.Activity.Resumes < 1 || profile.Activity.Matches < 1 || profile.Activity.Applied < 1 {
			t.Fatalf("perfil/atividade inesperados: %+v", profile)
		}
	})

	t.Run("guards: propria conta e contas admin", func(t *testing.T) {
		// Admin não gerencia a própria conta: descobre o próprio id via listagem? Não —
		// usa o id do admin criado no início (via DB).
		var adminUser models.User
		if err := testDB.Where("email = ?", "admin-gestao-"+sfx+"@example.com").First(&adminUser).Error; err != nil {
			t.Fatalf("carregar admin: %v", err)
		}
		path := "/api/admin/users/" + strconv.FormatUint(uint64(adminUser.ID), 10)
		resp := doRequest(t, "PATCH", path, map[string]interface{}{"name": "X"}, adminToken)
		if resp.StatusCode != http.StatusForbidden {
			resp.Body.Close()
			t.Fatalf("auto-gestão: esperado 403, obtido %d", resp.StatusCode)
		}
		resp.Body.Close()

		// Promove um segundo admin e tenta gerenciá-lo → 403.
		_, otherAdminID := registerUser(t, "Admin Dois", "admin-dois-"+sfx+"@example.com", "senhaSegura123", "")
		testDB.Model(&models.User{}).Where("id = ?", otherAdminID).Update("role", models.RoleAdmin)
		path = "/api/admin/users/" + strconv.FormatUint(uint64(otherAdminID), 10)
		resp = doRequest(t, "DELETE", path, nil, adminToken)
		if resp.StatusCode != http.StatusForbidden {
			resp.Body.Close()
			t.Fatalf("gerenciar admin: esperado 403, obtido %d", resp.StatusCode)
		}
		resp.Body.Close()
	})

	t.Run("edicao valida e validacoes", func(t *testing.T) {
		path := "/api/admin/users/" + strconv.FormatUint(uint64(targetID), 10)

		// Nome vazio → 422.
		resp := doRequest(t, "PATCH", path, map[string]interface{}{"name": "   "}, adminToken)
		if resp.StatusCode != http.StatusUnprocessableEntity {
			resp.Body.Close()
			t.Fatalf("nome vazio: esperado 422, obtido %d", resp.StatusCode)
		}
		resp.Body.Close()

		// E-mail duplicado (outro usuário) → 409.
		resp = doRequest(t, "PATCH", path, map[string]interface{}{"email": "comum-gestao-" + sfx + "@example.com"}, adminToken)
		if resp.StatusCode != http.StatusConflict {
			resp.Body.Close()
			t.Fatalf("e-mail duplicado: esperado 409, obtido %d", resp.StatusCode)
		}
		resp.Body.Close()

		// Edição válida → 200 e persiste.
		resp = doRequest(t, "PATCH", path, map[string]interface{}{"name": "Alvo Editado", "city": "Recife"}, adminToken)
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			t.Fatalf("edição: esperado 200, obtido %d", resp.StatusCode)
		}
		resp.Body.Close()
		var updated models.User
		testDB.Unscoped().First(&updated, targetID)
		if updated.Name != "Alvo Editado" || updated.City != "Recife" {
			t.Fatalf("edição não persistiu: %+v", updated)
		}
	})

	t.Run("atribuicao de plano nao toca assinatura", func(t *testing.T) {
		path := "/api/admin/users/" + strconv.FormatUint(uint64(targetID), 10) + "/plan"

		// Plano inexistente → 404.
		resp := doRequest(t, "PATCH", path, map[string]interface{}{"plan_slug": "nao-existe-" + sfx}, adminToken)
		if resp.StatusCode != http.StatusNotFound {
			resp.Body.Close()
			t.Fatalf("plano inexistente: esperado 404, obtido %d", resp.StatusCode)
		}
		resp.Body.Close()

		// Simula assinatura paga ativa e atribui outro plano.
		testDB.Model(&models.User{}).Where("id = ?", targetID).
			Updates(map[string]interface{}{"subscription_status": models.SubStatusActive, "plan_slug": "pro"})
		resp = doRequest(t, "PATCH", path, map[string]interface{}{"plan_slug": "free"}, adminToken)
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			t.Fatalf("atribuir plano: esperado 200, obtido %d", resp.StatusCode)
		}
		resp.Body.Close()
		var after models.User
		testDB.First(&after, targetID)
		if after.PlanSlug != "free" {
			t.Fatalf("plan_slug deveria ser free, obtido %s", after.PlanSlug)
		}
		if after.SubscriptionStatus != models.SubStatusActive {
			t.Fatalf("assinatura cobrada não pode mudar: %s", after.SubscriptionStatus)
		}
	})

	t.Run("suspensao bloqueia login e sessao aberta; restaura", func(t *testing.T) {
		path := "/api/admin/users/" + strconv.FormatUint(uint64(targetID), 10)
		resp := doRequest(t, "POST", path+"/suspend", nil, adminToken)
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			t.Fatalf("suspender: esperado 200, obtido %d", resp.StatusCode)
		}
		resp.Body.Close()

		// Novo login → 403 Conta suspensa.
		resp = doRequest(t, "POST", "/api/auth/login", map[string]interface{}{
			"email": "alvo-list-" + sfx + "@example.com", "password": "senhaSegura123",
		}, "")
		if resp.StatusCode != http.StatusForbidden {
			resp.Body.Close()
			t.Fatalf("login suspenso: esperado 403, obtido %d", resp.StatusCode)
		}
		resp.Body.Close()

		// Sessão aberta é coberta no subtest dedicado abaixo; aqui validamos
		// que a listagem reflete a suspensão.
		resp = doRequest(t, "GET", "/api/admin/users?status=suspended&search=alvo-list-"+sfx, nil, adminToken)
		var susp listUsersResponse
		parseBody(t, resp, &susp)
		if susp.Total != 1 {
			t.Fatalf("filtro suspenso deveria ter 1, obtido %d", susp.Total)
		}

		// Restaura → login volta a funcionar.
		resp = doRequest(t, "POST", path+"/restore", nil, adminToken)
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			t.Fatalf("restaurar: esperado 200, obtido %d", resp.StatusCode)
		}
		resp.Body.Close()
		resp = doRequest(t, "POST", "/api/auth/login", map[string]interface{}{
			"email": "alvo-list-" + sfx + "@example.com", "password": "senhaSegura123",
		}, "")
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			t.Fatalf("login após restore: esperado 200, obtido %d", resp.StatusCode)
		}
		resp.Body.Close()
	})

	t.Run("bloqueio de sessao aberta apos suspensao", func(t *testing.T) {
		// Usuário dedicado: login gera token, depois suspende e usa o token.
		email := "sessao-aberta-" + sfx + "@example.com"
		token, sessID := registerUser(t, "Sessão Aberta", email, "senhaSegura123", "")
		path := "/api/admin/users/" + strconv.FormatUint(uint64(sessID), 10)

		// Token funciona antes da suspensão.
		resp := doRequest(t, "GET", "/api/me", nil, token)
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			t.Fatalf("token antes da suspensão: esperado 200, obtido %d", resp.StatusCode)
		}
		resp.Body.Close()

		resp = doRequest(t, "POST", path+"/suspend", nil, adminToken)
		resp.Body.Close()

		// Mesmo token (24h de validade) agora → 403 Conta suspensa.
		resp = doRequest(t, "GET", "/api/me", nil, token)
		if resp.StatusCode != http.StatusForbidden {
			resp.Body.Close()
			t.Fatalf("sessão aberta suspensa: esperado 403, obtido %d", resp.StatusCode)
		}
		resp.Body.Close()

		resp = doRequest(t, "POST", path+"/restore", nil, adminToken)
		resp.Body.Close()
		resp = doRequest(t, "GET", "/api/me", nil, token)
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			t.Fatalf("token após restore: esperado 200, obtido %d", resp.StatusCode)
		}
		resp.Body.Close()
	})

	t.Run("exclusao em duas fases, reativacao e expurgo", func(t *testing.T) {
		email := "exclusao-" + sfx + "@example.com"
		_, delID := registerUser(t, "Conta Excluir", email, "senhaSegura123", "")
		path := "/api/admin/users/" + strconv.FormatUint(uint64(delID), 10)

		// Exclui com janela padrão.
		resp := doRequest(t, "DELETE", path, nil, adminToken)
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			t.Fatalf("excluir: esperado 200, obtido %d", resp.StatusCode)
		}
		resp.Body.Close()

		// Sai da listagem padrão, aparece no filtro excluída.
		resp = doRequest(t, "GET", "/api/admin/users?search=exclusao-"+sfx, nil, adminToken)
		var active listUsersResponse
		parseBody(t, resp, &active)
		if active.Total != 0 {
			t.Fatalf("conta excluída não deve aparecer na listagem padrão, obtido %d", active.Total)
		}
		resp = doRequest(t, "GET", "/api/admin/users?status=deleted&search=exclusao-"+sfx, nil, adminToken)
		var deleted listUsersResponse
		parseBody(t, resp, &deleted)
		if deleted.Total != 1 {
			t.Fatalf("filtro excluída deveria ter 1, obtido %d", deleted.Total)
		}

		// E-mail reservado: novo cadastro com o mesmo e-mail → 409.
		resp = doRequest(t, "POST", "/api/auth/register", map[string]interface{}{
			"name": "Outra", "email": email, "password": "senhaSegura123",
		}, "")
		if resp.StatusCode != http.StatusConflict {
			resp.Body.Close()
			t.Fatalf("e-mail reservado na janela: esperado 409, obtido %d", resp.StatusCode)
		}
		resp.Body.Close()

		// Login falha (soft-delete).
		resp = doRequest(t, "POST", "/api/auth/login", map[string]interface{}{
			"email": email, "password": "senhaSegura123",
		}, "")
		if resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			t.Fatal("login de conta excluída não pode succeeder")
		}
		resp.Body.Close()

		// Reativação devolve a conta intacta (SC-009).
		resp = doRequest(t, "POST", path+"/reactivate", nil, adminToken)
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			t.Fatalf("reativar: esperado 200, obtido %d", resp.StatusCode)
		}
		resp.Body.Close()
		var restored models.User
		testDB.Unscoped().First(&restored, delID)
		if restored.DeletedAt.Valid || restored.PurgeAfter != nil || restored.Suspended {
			t.Fatalf("reativação não limpou o estado: %+v", restored)
		}
		if restored.Email != email {
			t.Fatalf("reativação deve devolver e-mail intacto, obtido %s", restored.Email)
		}

		// Exclui de novo e expurga com retention=0.
		os.Setenv("USER_DELETE_RETENTION_DAYS", "0")
		defer os.Unsetenv("USER_DELETE_RETENTION_DAYS")
		resp = doRequest(t, "DELETE", path, nil, adminToken)
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			t.Fatalf("reexcluir: esperado 200, obtido %d", resp.StatusCode)
		}
		resp.Body.Close()

		if n := models.PurgeDeletedUsers(testDB, 0); n < 1 {
			t.Fatalf("expurgo deveria remover ao menos 1 conta, obtido %d", n)
		}
		var purged models.User
		if err := testDB.Unscoped().First(&purged, delID).Error; err == nil {
			t.Fatal("conta deveria ter sido removida pelo expurgo")
		}

		// Após expurgo o e-mail volta a ficar livre.
		resp = doRequest(t, "POST", "/api/auth/register", map[string]interface{}{
			"name": "Re-registro", "email": email, "password": "senhaSegura123",
		}, "")
		if resp.StatusCode != http.StatusCreated {
			resp.Body.Close()
			t.Fatalf("re-registro após expurgo: esperado 201, obtido %d", resp.StatusCode)
		}
		resp.Body.Close()
	})

	t.Run("reativacao apos janela expirada e recusada", func(t *testing.T) {
		email := "janela-expirada-" + sfx + "@example.com"
		_, lateID := registerUser(t, "Janela Expirada", email, "senhaSegura123", "")
		path := "/api/admin/users/" + strconv.FormatUint(uint64(lateID), 10)

		resp := doRequest(t, "DELETE", path, nil, adminToken)
		resp.Body.Close()

		// Força o fim da janela.
		past := time.Now().Add(-time.Minute)
		testDB.Unscoped().Model(&models.User{}).Where("id = ?", lateID).Update("purge_after", past)

		resp = doRequest(t, "POST", path+"/reactivate", nil, adminToken)
		if resp.StatusCode != http.StatusConflict {
			resp.Body.Close()
			t.Fatalf("reativação expirada: esperado 409, obtido %d", resp.StatusCode)
		}
		resp.Body.Close()

		// Limpa: expurga para não vazar dados entre testes.
		os.Setenv("USER_DELETE_RETENTION_DAYS", "0")
		models.PurgeDeletedUsers(testDB, 0)
		os.Unsetenv("USER_DELETE_RETENTION_DAYS")
	})

	t.Run("auditoria cobre todas as acoes de gestao", func(t *testing.T) {
		email := "auditoria-" + sfx + "@example.com"
		_, audID := registerUser(t, "Conta Auditoria", email, "senhaSegura123", "")
		path := "/api/admin/users/" + strconv.FormatUint(uint64(audID), 10)

		doRequest(t, "PATCH", path, map[string]interface{}{"name": "Audited"}, adminToken).Body.Close()
		doRequest(t, "PATCH", path+"/plan", map[string]interface{}{"plan_slug": "pro"}, adminToken).Body.Close()
		doRequest(t, "POST", path+"/suspend", nil, adminToken).Body.Close()
		doRequest(t, "POST", path+"/restore", nil, adminToken).Body.Close()
		doRequest(t, "DELETE", path, nil, adminToken).Body.Close()
		doRequest(t, "POST", path+"/reactivate", nil, adminToken).Body.Close()

		want := []string{"user.update", "user.plan_assign", "user.suspend", "user.restore", "user.delete", "user.reactivate"}
		for _, action := range want {
			var count int64
			testDB.Model(&models.AuditLog{}).Where("action = ? AND resource_id = ? AND actor_id > 0", action, audID).Count(&count)
			if count == 0 {
				t.Fatalf("ação %s sem registro de auditoria", action)
			}
		}
	})
}
