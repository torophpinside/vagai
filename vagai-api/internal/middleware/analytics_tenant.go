package middleware

import (
	"os"
	"strings"

	"github.com/anomalyco/vagai-api/internal/models"
	"github.com/gin-gonic/gin"
)

// IsAdminEmail verifica se o e-mail pertence à lista ADMIN_EMAILS.
// Usado no registro/login para promover automaticamente o usuário a admin.
func IsAdminEmail(email string) bool {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return false
	}
	for _, raw := range strings.Split(os.Getenv("ADMIN_EMAILS"), ",") {
		if strings.ToLower(strings.TrimSpace(raw)) == email {
			return true
		}
	}
	return false
}

// ResolveAnalyticsUserID resolve o user_id alvo da query de analytics.
// Admin: retorna o requested — 0 significa "todos os usuários" (dados
// globais) e qualquer user_id permite drill-down. Usuário comum: sempre os
// próprios dados, ignorando o requested (FR-021).
func ResolveAnalyticsUserID(c *gin.Context, requested uint) (uint, bool) {
	if c.GetString("user_role") != string(models.RoleAdmin) {
		return c.GetUint("user_id"), true
	}
	return requested, true
}
