package e2e

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/anomalyco/vagai-api/internal/middleware"
	"github.com/anomalyco/vagai-api/internal/models"
)

// promoteAdmin registra um usuário comum e o promove a admin, devolvendo um
// token com role=admin (o token de registro ainda carrega role=user).
func promoteAdmin(t *testing.T, email string) string {
	t.Helper()
	_, userID := registerUser(t, "Admin Planos", email, "senhaSegura123", "")
	if err := testDB.Model(&models.User{}).Where("id = ?", userID).Update("role", models.RoleAdmin).Error; err != nil {
		t.Fatalf("promover admin: %v", err)
	}
	token, err := middleware.GenerateToken(userID, email, string(models.RoleAdmin))
	if err != nil {
		t.Fatalf("gerar token admin: %v", err)
	}
	return token
}

func uniqueSuffix() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}

func TestPlansAdmin(t *testing.T) {
	sfx := uniqueSuffix()
	userToken, _ := registerUser(t, "Comum Planos", "comum-"+sfx+"@example.com", "senhaSegura123", "")
	adminToken := promoteAdmin(t, "admin-"+sfx+"@example.com")

	t.Run("nao-admin recebe 403 em todas as rotas", func(t *testing.T) {
		cases := []struct {
			method string
			path   string
			body   interface{}
		}{
			{"GET", "/api/admin/plans", nil},
			{"POST", "/api/admin/plans", map[string]interface{}{"name": "X", "slug": "x-" + sfx, "type": "free"}},
			{"PATCH", "/api/admin/plans/1", map[string]interface{}{"name": "Y"}},
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

	t.Run("admin lista planos", func(t *testing.T) {
		resp := doRequest(t, "GET", "/api/admin/plans", nil, adminToken)
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			t.Fatalf("esperado 200, obtido %d", resp.StatusCode)
		}
		var plans []models.Plan
		parseBody(t, resp, &plans)
		if len(plans) == 0 {
			t.Fatal("esperado lista não vazia")
		}
	})

	freeSlug := "gratis-teste-" + sfx
	paidSlug := "pago-teste-" + sfx

	t.Run("cria plano gratuito valido", func(t *testing.T) {
		body := map[string]interface{}{
			"name": "Gratuito Teste", "slug": freeSlug, "type": "free",
			"price_monthly": 0, "price_yearly": 0, "features": []string{"tudo basico"},
		}
		resp := doRequest(t, "POST", "/api/admin/plans", body, adminToken)
		if resp.StatusCode != http.StatusCreated {
			resp.Body.Close()
			t.Fatalf("esperado 201, obtido %d", resp.StatusCode)
		}
		var plan models.Plan
		parseBody(t, resp, &plan)
		if plan.Slug != freeSlug || plan.Type != models.PlanTypeFree || !plan.Active || plan.IsDefault {
			t.Fatalf("plano inesperado: %+v", plan)
		}
	})

	t.Run("cria plano pago valido", func(t *testing.T) {
		body := map[string]interface{}{
			"name": "Pago Teste", "slug": paidSlug, "type": "paid", "price_monthly": 1900,
		}
		resp := doRequest(t, "POST", "/api/admin/plans", body, adminToken)
		if resp.StatusCode != http.StatusCreated {
			resp.Body.Close()
			t.Fatalf("esperado 201, obtido %d", resp.StatusCode)
		}
		var plan models.Plan
		parseBody(t, resp, &plan)
		if plan.Type != models.PlanTypePaid || plan.PriceMonthly != 1900 {
			t.Fatalf("plano inesperado: %+v", plan)
		}
	})

	t.Run("opcoes de IA e preparacao", func(t *testing.T) {
		slug := "opcoes-" + sfx
		body := map[string]interface{}{
			"name": "Opcoes", "slug": slug, "type": "free",
			"ai_matching": true, "interview_prep": true,
		}
		resp := doRequest(t, "POST", "/api/admin/plans", body, adminToken)
		if resp.StatusCode != http.StatusCreated {
			resp.Body.Close()
			t.Fatalf("criar: esperado 201, obtido %d", resp.StatusCode)
		}
		var plan models.Plan
		parseBody(t, resp, &plan)
		if !plan.AIMatching || !plan.InterviewPrep {
			t.Fatalf("criar: esperado ambas opções ativas, obtido ai=%v prep=%v", plan.AIMatching, plan.InterviewPrep)
		}

		// PATCH parcial: desliga somente a IA; preparação deve permanecer ativa.
		id := planIDBySlug(t, slug)
		resp = doRequest(t, "PATCH", fmt.Sprintf("/api/admin/plans/%d", id),
			map[string]interface{}{"ai_matching": false}, adminToken)
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			t.Fatalf("editar: esperado 200, obtido %d", resp.StatusCode)
		}
		var updated models.Plan
		parseBody(t, resp, &updated)
		if updated.AIMatching || !updated.InterviewPrep {
			t.Fatalf("editar: esperado ai=false prep=true, obtido ai=%v prep=%v", updated.AIMatching, updated.InterviewPrep)
		}
	})

	t.Run("validacoes de preco", func(t *testing.T) {
		cases := []struct {
			name string
			body map[string]interface{}
		}{
			{"pago sem preco", map[string]interface{}{"name": "Pago Zero", "slug": "pago-zero-" + sfx, "type": "paid"}},
			{"gratuito com preco", map[string]interface{}{"name": "Gratis Caro", "slug": "gratis-caro-" + sfx, "type": "free", "price_monthly": 100}},
			{"tipo invalido", map[string]interface{}{"name": "Estranho", "slug": "estranho-" + sfx, "type": "weird"}},
		}
		for _, tc := range cases {
			resp := doRequest(t, "POST", "/api/admin/plans", tc.body, adminToken)
			if resp.StatusCode != http.StatusBadRequest {
				resp.Body.Close()
				t.Fatalf("%s: esperado 400, obtido %d", tc.name, resp.StatusCode)
			}
			resp.Body.Close()
		}
	})

	t.Run("slug duplicado retorna 409", func(t *testing.T) {
		body := map[string]interface{}{"name": "Duplicado", "slug": freeSlug, "type": "free"}
		resp := doRequest(t, "POST", "/api/admin/plans", body, adminToken)
		if resp.StatusCode != http.StatusConflict {
			resp.Body.Close()
			t.Fatalf("esperado 409, obtido %d", resp.StatusCode)
		}
		resp.Body.Close()
	})

	t.Run("slug imutavel na edicao", func(t *testing.T) {
		id := planIDBySlug(t, paidSlug)
		body := map[string]interface{}{"slug": "novo-slug-" + sfx}
		resp := doRequest(t, "PATCH", fmt.Sprintf("/api/admin/plans/%d", id), body, adminToken)
		if resp.StatusCode != http.StatusBadRequest {
			resp.Body.Close()
			t.Fatalf("esperado 400, obtido %d", resp.StatusCode)
		}
		resp.Body.Close()
	})

	t.Run("regras do plano padrao", func(t *testing.T) {
		var def models.Plan
		if err := testDB.Where("is_default = ?", true).First(&def).Error; err != nil {
			t.Fatalf("nenhum plano padrão semeado: %v", err)
		}

		// Desativar o default → 409
		resp := doRequest(t, "PATCH", fmt.Sprintf("/api/admin/plans/%d", def.ID),
			map[string]interface{}{"active": false}, adminToken)
		if resp.StatusCode != http.StatusConflict {
			resp.Body.Close()
			t.Fatalf("desativar default: esperado 409, obtido %d", resp.StatusCode)
		}
		resp.Body.Close()

		// Trocar tipo do default para pago → 409
		resp = doRequest(t, "PATCH", fmt.Sprintf("/api/admin/plans/%d", def.ID),
			map[string]interface{}{"type": "paid", "price_monthly": 500}, adminToken)
		if resp.StatusCode != http.StatusConflict {
			resp.Body.Close()
			t.Fatalf("tipo do default: esperado 409, obtido %d", resp.StatusCode)
		}
		resp.Body.Close()

		// Marcar plano pago como default → 409
		paidID := planIDBySlug(t, paidSlug)
		resp = doRequest(t, "PATCH", fmt.Sprintf("/api/admin/plans/%d", paidID),
			map[string]interface{}{"is_default": true}, adminToken)
		if resp.StatusCode != http.StatusConflict {
			resp.Body.Close()
			t.Fatalf("default pago: esperado 409, obtido %d", resp.StatusCode)
		}
		resp.Body.Close()

		// Marcar plano inativo como default → 409
		inactiveSlug := "inativo-default-" + sfx
		resp = doRequest(t, "POST", "/api/admin/plans",
			map[string]interface{}{"name": "Inativo", "slug": inactiveSlug, "type": "free"}, adminToken)
		if resp.StatusCode != http.StatusCreated {
			resp.Body.Close()
			t.Fatalf("criar inativo: esperado 201, obtido %d", resp.StatusCode)
		}
		resp.Body.Close()
		inactiveID := planIDBySlug(t, inactiveSlug)
		resp = doRequest(t, "PATCH", fmt.Sprintf("/api/admin/plans/%d", inactiveID),
			map[string]interface{}{"active": false}, adminToken)
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			t.Fatalf("desativar não-default: esperado 200, obtido %d", resp.StatusCode)
		}
		resp.Body.Close()
		resp = doRequest(t, "PATCH", fmt.Sprintf("/api/admin/plans/%d", inactiveID),
			map[string]interface{}{"is_default": true}, adminToken)
		if resp.StatusCode != http.StatusConflict {
			resp.Body.Close()
			t.Fatalf("default inativo: esperado 409, obtido %d", resp.StatusCode)
		}
		resp.Body.Close()
	})

	t.Run("cadastro novo recebe o plano padrao", func(t *testing.T) {
		var def models.Plan
		if err := testDB.Where("is_default = ?", true).First(&def).Error; err != nil {
			t.Fatalf("nenhum plano padrão: %v", err)
		}
		email := "novo-" + sfx + "@example.com"
		resp := doRequest(t, "POST", "/api/auth/register",
			map[string]string{"name": "Novo Usuario", "email": email, "password": "senhaSegura123"}, "")
		if resp.StatusCode != http.StatusCreated {
			resp.Body.Close()
			t.Fatalf("registro: esperado 201, obtido %d", resp.StatusCode)
		}
		var result struct {
			User struct {
				PlanSlug string `json:"plan_slug"`
			} `json:"user"`
		}
		parseBody(t, resp, &result)
		if result.User.PlanSlug != def.Slug {
			t.Fatalf("plan_slug=%q, esperado %q", result.User.PlanSlug, def.Slug)
		}
	})
}

func TestPlansBackfill(t *testing.T) {
	sfx := uniqueSuffix()
	legacyFree := "legacy-gratis-" + sfx
	legacyPaid := "legacy-pago-" + sfx
	legacyMismatch := "legacy-mismatch-" + sfx

	insertLegacyPlan(t, legacyFree, "", 0, 0)
	insertLegacyPlan(t, legacyPaid, "", 2900, 29000)
	// Simula a armadilha do AutoMigrate: coluna nova preenchida com o default
	// 'free' numa linha cujo preço é pago. O backfill deve corrigir para 'paid'.
	insertLegacyPlan(t, legacyMismatch, "free", 990, 0)

	// Zera o default para exercitar a marcação automática do backfill.
	if err := testDB.Model(&models.Plan{}).Where("is_default = ?", true).
		Update("is_default", false).Error; err != nil {
		t.Fatalf("limpar defaults: %v", err)
	}

	models.BackfillPlans(testDB)
	models.BackfillPlans(testDB) // idempotência
	var free models.Plan
	if err := testDB.Where("slug = ?", legacyFree).First(&free).Error; err != nil {
		t.Fatalf("buscar legado grátis: %v", err)
	}
	if free.Type != models.PlanTypeFree || !free.Active {
		t.Fatalf("backfill grátis: type=%q active=%v", free.Type, free.Active)
	}

	var paid models.Plan
	if err := testDB.Where("slug = ?", legacyPaid).First(&paid).Error; err != nil {
		t.Fatalf("buscar legado pago: %v", err)
	}
	if paid.Type != models.PlanTypePaid || !paid.Active {
		t.Fatalf("backfill pago: type=%q active=%v", paid.Type, paid.Active)
	}

	var mismatch models.Plan
	if err := testDB.Where("slug = ?", legacyMismatch).First(&mismatch).Error; err != nil {
		t.Fatalf("buscar legado divergente: %v", err)
	}
	if mismatch.Type != models.PlanTypePaid {
		t.Fatalf("backfill deveria corrigir type divergente do preço, obtido %q", mismatch.Type)
	}

	var defaults int64
	testDB.Model(&models.Plan{}).Where("is_default = ?", true).Count(&defaults)
	if defaults != 1 {
		t.Fatalf("esperado exatamente 1 default, obtido %d", defaults)
	}
	var defaultPlan models.Plan
	testDB.Where("is_default = ?", true).First(&defaultPlan)
	if defaultPlan.Type != models.PlanTypeFree {
		t.Fatalf("default deveria ser gratuito, obtido %q", defaultPlan.Type)
	}

	// Limpeza: remove as linhas legadas e restaura o free semeado como default.
	testDB.Where("slug IN ?", []string{legacyFree, legacyPaid, legacyMismatch}).Delete(&models.Plan{})
	testDB.Model(&models.Plan{}).Where("is_default = ?", true).Update("is_default", false)
	models.BackfillPlans(testDB)
}

func insertLegacyPlan(t *testing.T, slug, planType string, monthly, yearly int) {
	t.Helper()
	err := testDB.Exec(`INSERT INTO plans
		(name, slug, type, price_monthly, price_yearly, max_jobs, max_resume_analyses, max_sites, max_crawls_per_day, features, active, is_default, created_at)
		VALUES (?, ?, ?, ?, ?, 100, 1, 5, 10, '[]', 1, 0, NOW())`,
		"Legacy", slug, planType, monthly, yearly).Error
	if err != nil {
		t.Fatalf("inserir plano legado %q: %v", slug, err)
	}
}

func planIDBySlug(t *testing.T, slug string) uint {
	t.Helper()
	var plan models.Plan
	if err := testDB.Where("slug = ?", slug).First(&plan).Error; err != nil {
		t.Fatalf("plano %q não encontrado: %v", slug, err)
	}
	return plan.ID
}
