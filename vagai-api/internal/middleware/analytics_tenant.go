package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/anomalyco/vagai-api/internal/models"
	"github.com/gin-gonic/gin"
)

// RequireAnalyticsAccess aplica FR-020: acesso a analytics restrito a
// administradores da organização (owner/admin) e a super admins.
//
// Super admin: e-mails listados na env SUPER_ADMIN_EMAILS (separados por
// vírgula). Super admins podem consultar qualquer organização (inclusive
// todas, sem organization_id). Administradores de organização só consultam a
// própria organização.
func RequireAnalyticsAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("user_role")
		email := c.GetString("user_email")

		if IsSuperAdmin(email) {
			c.Set("analytics_super_admin", true)
			c.Next()
			return
		}

		if role != string(models.RoleOwner) && role != string(models.RoleAdmin) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Acesso a analytics restrito a administradores da organização",
			})
			return
		}

		c.Set("analytics_super_admin", false)
		c.Next()
	}
}

// IsSuperAdmin verifica se o e-mail pertence à lista SUPER_ADMIN_EMAILS.
func IsSuperAdmin(email string) bool {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return false
	}
	for _, raw := range strings.Split(os.Getenv("SUPER_ADMIN_EMAILS"), ",") {
		if strings.ToLower(strings.TrimSpace(raw)) == email {
			return true
		}
	}
	return false
}

// IsAnalyticsSuperAdmin reports se a request atual é de um super admin.
func IsAnalyticsSuperAdmin(c *gin.Context) bool {
	v, exists := c.Get("analytics_super_admin")
	if !exists {
		return IsSuperAdmin(c.GetString("user_email"))
	}
	super, _ := v.(bool)
	return super
}

// ResolveAnalyticsOrgID resolve o organization_id alvo da query de analytics.
//
//   - super admin: requested (0 = todas as organizações);
//   - admin da org: 0 → própria organização; ≠0 precisa ser a própria.
//
// Retorna (orgID, false) quando o acesso é negado (chamador deve responder 403).
func ResolveAnalyticsOrgID(c *gin.Context, requested uint) (uint, bool) {
	ownOrg := c.GetUint("org_id")

	if IsAnalyticsSuperAdmin(c) {
		return requested, true
	}

	if requested == 0 {
		return ownOrg, true
	}

	if requested != ownOrg {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Acesso negado à organização informada",
		})
		return 0, false
	}

	return ownOrg, true
}
