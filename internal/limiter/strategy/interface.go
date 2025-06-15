package strategy

import (
	"context"

	"github.com/JMKobayashi/Rate-Limiter-GO/internal/entity"
)

// RateLimiterRepository define a interface para persistência do rate limiter
type RateLimiterRepository interface {
	Save(ctx context.Context, limiter *entity.RateLimiter) error
	Get(ctx context.Context, key string) (*entity.RateLimiter, error)
	Delete(ctx context.Context, key string) error
}

// StorageStrategy define a interface para estratégias de armazenamento
type StorageStrategy interface {
	// Increment incrementa o contador para uma chave específica
	Increment(ctx context.Context, key string) (int64, error)

	// Get retorna o valor atual do contador
	Get(ctx context.Context, key string) (int64, error)

	// Set define um valor para uma chave com expiração
	Set(ctx context.Context, key string, value int64, expiration int) error

	// Delete remove uma chave
	Delete(ctx context.Context, key string) error
}
