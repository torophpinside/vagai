package middleware

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func TenantScope(db *gorm.DB) func() *gorm.DB {
	return func() *gorm.DB {
		return db
	}
}

func ScopedDB(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		orgID, exists := c.Get("org_id")
		if !exists {
			c.Next()
			return
		}

		// Salvar o orgID no contexto para os handlers usarem explicitamente
		c.Set("org_id", orgID)
		c.Next()
	}
}

func GetScopedDB(c *gin.Context) *gorm.DB {
	if scoped, exists := c.Get("scoped_db"); exists {
		return scoped.(*gorm.DB)
	}
	if db, exists := c.Get("db"); exists {
		return db.(*gorm.DB)
	}
	return nil
}
