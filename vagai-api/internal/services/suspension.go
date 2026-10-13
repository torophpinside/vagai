// Package services fornece utilitários transversais da API. O marcador de
// suspensão bloqueia sessões já abertas de contas suspensas/excluídas sem
// hit por request no MySQL (research.md R4, FR-010/SC-003).
package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	suspensionKeyPrefix = "susp:"
	suspensionTTL       = 25 * time.Hour
)

var (
	localMu       sync.RWMutex
	localMarkers  = map[uint]time.Time{}
	redisClient   *redis.Client
	redisInitOnce sync.Once
)

// initRedis cria (uma única vez) o client Redis a partir de REDIS_URL.
// Tolerante a falhas: sem URL, URL inválida ou Redis fora do ar, apenas o
// armazenamento em processo é usado (instância única do compose).
func initRedis() *redis.Client {
	redisInitOnce.Do(func() {
		addr := strings.TrimSpace(os.Getenv("REDIS_URL"))
		if addr == "" {
			return
		}
		opts, err := redis.ParseURL(addr)
		if err != nil {
			log.Printf("suspension: REDIS_URL inválida, usando armazenamento em processo: %v", err)
			return
		}
		client := redis.NewClient(opts)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := client.Ping(ctx).Err(); err != nil {
			_ = client.Close()
			log.Printf("suspension: Redis indisponível, usando armazenamento em processo: %v", err)
			return
		}
		redisClient = client
	})
	return redisClient
}

// SetSuspendedMarker marca o usuário como suspenso para bloqueio de sessão
// aberta (usado em suspend e delete).
func SetSuspendedMarker(userID uint) {
	setMarker(userID)
}

// ClearSuspendedMarker remove o bloqueio de sessão (restore e reactivate).
func ClearSuspendedMarker(userID uint) {
	clearMarker(userID)
}

// IsSuspended reporta se o usuário tem marcador de suspensão ativo.
func IsSuspended(userID uint) bool {
	return hasMarker(userID)
}

func setMarker(userID uint) {
	localMu.Lock()
	localMarkers[userID] = time.Now().Add(suspensionTTL)
	localMu.Unlock()

	if c := initRedis(); c != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		_ = c.Set(ctx, fmt.Sprintf("%s%d", suspensionKeyPrefix, userID), "1", suspensionTTL).Err()
	}
}

func clearMarker(userID uint) {
	localMu.Lock()
	delete(localMarkers, userID)
	localMu.Unlock()

	if c := initRedis(); c != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		_ = c.Del(ctx, fmt.Sprintf("%s%d", suspensionKeyPrefix, userID)).Err()
	}
}

func hasMarker(userID uint) bool {
	localMu.RLock()
	exp, ok := localMarkers[userID]
	localMu.RUnlock()
	if ok {
		if time.Now().Before(exp) {
			return true
		}
		localMu.Lock()
		delete(localMarkers, userID)
		localMu.Unlock()
	}

	if c := initRedis(); c != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		n, err := c.Exists(ctx, fmt.Sprintf("%s%d", suspensionKeyPrefix, userID)).Result()
		if err == nil && n > 0 {
			return true
		}
	}
	return false
}
