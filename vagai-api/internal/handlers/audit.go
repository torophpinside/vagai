package handlers

import (
	"encoding/json"
	"log"

	"github.com/anomalyco/vagai-api/internal/models"
	"github.com/gin-gonic/gin"
)

// logAudit grava uma ação de gestão em AuditLog (FR-009). Best-effort: falha
// de auditoria é logada e não aborta a operação principal — disponibilidade
// de moderação > rastreio perfeito (research.md R5).
func logAudit(c *gin.Context, actorID, userID uint, action string, resourceID uint, details any) {
	encoded := "{}"
	if details != nil {
		if raw, err := json.Marshal(details); err == nil {
			encoded = string(raw)
		}
	}
	entry := models.AuditLog{
		UserID:     userID,
		ActorID:    actorID,
		Action:     action,
		Resource:   "user",
		ResourceID: resourceID,
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
		Details:    encoded,
	}
	if err := DB.Create(&entry).Error; err != nil {
		log.Printf("audit: falha ao gravar %s (user=%d): %v", action, userID, err)
	}
}

// actorIDFrom extrai o ID do administrador autenticado das claims do contexto.
func actorIDFrom(c *gin.Context) uint {
	if v, ok := c.Get("user_id"); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
}

// abortError responde com o status e a mensagem padronizados em português.
func abortError(c *gin.Context, status int, msg string) {
	c.JSON(status, gin.H{"error": msg})
}
