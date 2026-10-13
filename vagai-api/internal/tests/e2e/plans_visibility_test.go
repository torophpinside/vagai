package e2e

import (
	"net/http"
	"testing"

	"github.com/anomalyco/vagai-api/internal/models"
)

func insertInactivePlan(t *testing.T, slug string) {
	t.Helper()
	err := testDB.Exec(`INSERT INTO plans
		(name, slug, type, price_monthly, price_yearly, max_jobs, max_resume_analyses, max_sites, max_crawls_per_day, features, active, is_default, created_at)
		VALUES (?, ?, 'free', 0, 0, 100, 1, 5, 10, '[]', 0, 0, NOW())`,
		"Inativo Visibilidade", slug).Error
	if err != nil {
		t.Fatalf("inserir plano inativo %q: %v", slug, err)
	}
}

func TestPlansVisibility(t *testing.T) {
	sfx := uniqueSuffix()
	inactiveSlug := "inativo-vis-" + sfx
	insertInactivePlan(t, inactiveSlug)
	t.Cleanup(func() {
		testDB.Where("slug = ?", inactiveSlug).Delete(&models.Plan{})
	})

	userToken, userID := registerUser(t, "Vitrine", "vitrine-"+sfx+"@example.com", "senhaSegura123", "")

	t.Run("usuario comum nao ve planos inativos", func(t *testing.T) {
		resp := doRequest(t, "GET", "/api/plans", nil, userToken)
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			t.Fatalf("esperado 200, obtido %d", resp.StatusCode)
		}
		var plans []models.Plan
		parseBody(t, resp, &plans)
		for _, p := range plans {
			if p.Slug == inactiveSlug {
				t.Fatalf("plano inativo %q não deveria aparecer para usuário comum", inactiveSlug)
			}
			if !p.Active {
				t.Fatalf("plano %q inativo apareceu na vitrine", p.Slug)
			}
		}
	})

	t.Run("usuario com plano inativo ve o proprio", func(t *testing.T) {
		if err := testDB.Model(&models.User{}).Where("id = ?", userID).
			Update("plan_slug", inactiveSlug).Error; err != nil {
			t.Fatalf("apontar usuário para plano inativo: %v", err)
		}
		resp := doRequest(t, "GET", "/api/plans", nil, userToken)
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			t.Fatalf("esperado 200, obtido %d", resp.StatusCode)
		}
		var plans []models.Plan
		parseBody(t, resp, &plans)
		found := false
		for _, p := range plans {
			if p.Slug == inactiveSlug {
				found = true
			}
		}
		if !found {
			t.Fatalf("usuário com plano inativo deveria ver %q", inactiveSlug)
		}
	})

	t.Run("mudar para plano inativo retorna 404", func(t *testing.T) {
		resp := doRequest(t, "POST", "/api/me/plan",
			map[string]string{"plan_slug": inactiveSlug}, userToken)
		if resp.StatusCode != http.StatusNotFound {
			resp.Body.Close()
			t.Fatalf("esperado 404, obtido %d", resp.StatusCode)
		}
		resp.Body.Close()
	})

	t.Run("paywall de plano pago permanece", func(t *testing.T) {
		resp := doRequest(t, "POST", "/api/me/plan",
			map[string]string{"plan_slug": "pro"}, userToken)
		if resp.StatusCode != http.StatusPaymentRequired {
			resp.Body.Close()
			t.Fatalf("esperado 402, obtido %d", resp.StatusCode)
		}
		resp.Body.Close()
	})
}
