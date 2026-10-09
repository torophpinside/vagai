package analytics

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// cacheTTL é o TTL de cache das respostas de analytics (research.md #5).
const cacheTTL = 5 * time.Minute

// keyPrefix namespaceia as chaves de analytics no Redis.
const keyPrefix = "analytics:v1:"

// cache é um wrapper tolerante a falhas sobre o Redis. Sem REDIS_URL, com
// URL inválida ou com Redis fora do ar, o cache fica desabilitado e as
// queries seguem direto para o banco — nunca derruba o endpoint.
type cache struct {
	client *redis.Client
}

var (
	rawClient  *redis.Client
	clientOnce sync.Once
)

// newCache cria (uma única vez) o client Redis a partir de REDIS_URL.
func newCache() *cache {
	clientOnce.Do(func() {
		addr := strings.TrimSpace(os.Getenv("REDIS_URL"))
		if addr == "" {
			return
		}
		opts, err := redis.ParseURL(addr)
		if err != nil {
			return
		}
		client := redis.NewClient(opts)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := client.Ping(ctx).Err(); err != nil {
			_ = client.Close()
			return
		}
		rawClient = client
	})
	return &cache{client: rawClient}
}

// get deserializa a chave para dest. Retorna false em miss ou erro.
func (c *cache) get(ctx context.Context, key string, dest any) bool {
	if c == nil || c.client == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	raw, err := c.client.Get(ctx, keyPrefix+key).Bytes()
	if err != nil {
		return false
	}
	return json.Unmarshal(raw, dest) == nil
}

// set serializa value com TTL padrão. Erros são silenciosos (best-effort).
func (c *cache) set(ctx context.Context, key string, value any, ttl time.Duration) {
	if c == nil || c.client == nil {
		return
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	_ = c.client.Set(ctx, keyPrefix+key, raw, ttl).Err()
}

// cachedValue retorna o valor cacheado em key ou, em miss, o resultado de
// load e o grava por cacheTTL. Qualquer falha de cache cai no load.
func cachedValue[T any](ctx context.Context, s *Service, key string, load func(context.Context) (T, error)) (T, error) {
	var cached T
	if s.cache.get(ctx, key, &cached) {
		return cached, nil
	}
	value, err := load(ctx)
	if err != nil {
		var zero T
		return zero, err
	}
	s.cache.set(ctx, key, value, cacheTTL)
	return value, nil
}
