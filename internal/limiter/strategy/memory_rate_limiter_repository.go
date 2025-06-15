package strategy

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/JMKobayashi/Rate-Limiter-GO/internal/entity"
	"github.com/JMKobayashi/Rate-Limiter-GO/internal/repository"
)

// MemoryRateLimiterRepository implementa o repositório de rate limiter usando memória
type MemoryRateLimiterRepository struct {
	limiters map[string]*entity.RateLimiter
	mu       sync.RWMutex
}

// NewMemoryRateLimiterRepository cria um novo repositório em memória
func NewMemoryRateLimiterRepository() repository.RateLimiterRepository {
	return &MemoryRateLimiterRepository{
		limiters: make(map[string]*entity.RateLimiter),
	}
}

// Save salva um rate limiter na memória
func (r *MemoryRateLimiterRepository) Save(ctx context.Context, limiter *entity.RateLimiter) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := r.getKey(limiter)
	r.limiters[key] = limiter
	return nil
}

// Get recupera um rate limiter da memória
func (r *MemoryRateLimiterRepository) Get(ctx context.Context, key string) (*entity.RateLimiter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	limiter, exists := r.limiters[key]
	if !exists {
		return nil, nil
	}

	// Verifica se o bloqueio expirou
	if limiter.Blocked && time.Now().After(limiter.BlockedUntil) {
		limiter.Blocked = false
		limiter.BlockedUntil = time.Time{}
	}

	return limiter, nil
}

// Delete remove um rate limiter da memória
func (r *MemoryRateLimiterRepository) Delete(ctx context.Context, key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.limiters, key)
	return nil
}

// getKey retorna a chave para um rate limiter
func (r *MemoryRateLimiterRepository) getKey(limiter *entity.RateLimiter) string {
	if limiter.IP != "" {
		return fmt.Sprintf("rate_limiter:ip:%s", limiter.IP)
	}
	return fmt.Sprintf("rate_limiter:token:%s", limiter.Token)
}
