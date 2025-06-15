package repository

import (
	"context"

	"github.com/JMKobayashi/Rate-Limiter-GO/internal/entity"
)

type RateLimiterRepository interface {
	Save(ctx context.Context, limiter *entity.RateLimiter) error
	Get(ctx context.Context, key string) (*entity.RateLimiter, error)
	Delete(ctx context.Context, key string) error
}
