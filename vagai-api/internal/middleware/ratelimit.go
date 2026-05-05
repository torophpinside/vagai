package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimiter struct {
	clients map[string]*clientLimit
	mu      sync.Mutex
}

type clientLimit struct {
	count    int
	resetAt  time.Time
}

var limiter = &rateLimiter{
	clients: make(map[string]*clientLimit),
}

func RateLimit(maxRequests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		limiter.mu.Lock()
		client, exists := limiter.clients[ip]
		if !exists || time.Now().After(client.resetAt) {
			client = &clientLimit{
				count:   1,
				resetAt: time.Now().Add(window),
			}
			limiter.clients[ip] = client
		} else {
			client.count++
		}
		limiter.mu.Unlock()

		if client.count > maxRequests {
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", client.resetAt.Format(time.RFC3339))
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Limite de requisições excedido"})
			c.Abort()
			return
		}

		c.Header("X-RateLimit-Remaining", string(rune(maxRequests-client.count)))
		c.Next()
	}
}
