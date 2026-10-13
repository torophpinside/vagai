package middleware

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/anomalyco/vagai-api/internal/models"
	"github.com/anomalyco/vagai-api/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token de autenticação não fornecido"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Formato de token inválido. Use: Bearer <token>"})
			c.Abort()
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(getJWTSecret()), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido ou expirado"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*Claims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Claims inválidos"})
			c.Abort()
			return
		}

		// Sessão aberta de conta suspensa/excluída é bloqueada na hora
		// (spec 013, FR-010/SC-003): marcador em memória/Redis gravado pelo
		// handler de suspensão/exclusão.
		if services.IsSuspended(claims.UserID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Conta suspensa"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Next()
	}
}

// RequireAdmin bloqueia qualquer requisição cujo papel não seja admin. Deve
// rodar após JWTAuth (que popula "user_role" a partir das claims do token).
// Protege as rotas /api/admin/* (gestão de planos — spec 012).
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("user_role")
		if r, ok := role.(string); !ok || r != string(models.RoleAdmin) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Acesso restrito a administradores"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func GenerateToken(userID uint, email, role string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(getJWTSecret()))
}

func GenerateRefreshToken(userID uint) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatUint(uint64(userID), 10),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(getJWTSecret() + "_refresh"))
}

// getJWTSecret retorna o segredo JWT configurado via variável de ambiente.
// Não existe mais fallback hardcoded: o servidor só inicia se JWT_SECRET estiver
// definido (validação feita no main), então retornar string vazia aqui só ocorre
// em executáveis auxiliares/contextos de teste.
func getJWTSecret() string {
	return os.Getenv("JWT_SECRET")
}
