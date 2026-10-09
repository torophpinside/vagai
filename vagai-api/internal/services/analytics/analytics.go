// Package analytics implementa as queries de leitura de analytics do VagAI.
//
// Padrões do pacote:
//   - Toda query é escopada por organization_id (Constitution III).
//   - Resultados são cacheados em Redis por 5 minutos (quando disponível).
//   - Queries rodam com timeout de 5s (SC-020).
//   - Ausência de dados retorna nil + empty-state com mensagem de orientação.
package analytics

import (
	"context"
	"sync"

	"gorm.io/gorm"
)

// Service concentra as queries de analytics. Crie uma instância por request
// com o DB já escopado pelo middleware de tenant.
type Service struct {
	db    *gorm.DB
	cache *cache
}

// New cria um serviço de analytics apontando para o banco informado.
func New(db *gorm.DB) *Service {
	return &Service{db: db, cache: defaultCache()}
}

// DB expõe o banco do serviço (uso interno dos subpacotes de query).
func (s *Service) DB() *gorm.DB { return s.db }

var (
	cacheOnce sync.Once
	cacheInst *cache
)

// defaultCache inicializa o cache Redis uma única vez. Sem REDIS_URL ou com
// Redis indisponível o cache fica desabilitado (miss permanente, sem erro).
func defaultCache() *cache {
	cacheOnce.Do(func() {
		cacheInst = newCache()
	})
	return cacheInst
}

// resetCacheForTest permite reinicializar o cache em testes.
func resetCacheForTest() {
	cacheOnce = sync.Once{}
	cacheInst = nil
}

// ctxDB aplica o timeout de 5s e devolve o DB com o contexto vinculado.
func (s *Service) ctxDB(ctx context.Context) (*gorm.DB, context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	return s.db.WithContext(ctx), ctx, cancel
}
