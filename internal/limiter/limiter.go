package limiter

import (
	"context"
	"fmt"

	"github.com/JMKobayashi/Rate-Limiter-GO/internal/limiter/strategy"
)

type RateLimiter struct {
	storage        strategy.StorageStrategy
	rateLimitIP    int
	rateLimitToken int
	blockDuration  int
}

func NewRateLimiter(storage strategy.StorageStrategy, rateLimitIP, rateLimitToken, blockDuration int) *RateLimiter {
	return &RateLimiter{
		storage:        storage,
		rateLimitIP:    rateLimitIP,
		rateLimitToken: rateLimitToken,
		blockDuration:  blockDuration,
	}
}

func (rl *RateLimiter) IsAllowed(ctx context.Context, identifier string, isToken bool) (bool, error) {
	key := fmt.Sprintf("rate_limit:%s", identifier)

	// Verifica se está bloqueado
	blocked, err := rl.storage.Get(ctx, fmt.Sprintf("blocked:%s", key))
	if err == nil && blocked > 0 {
		return false, nil
	}

	// Incrementa o contador
	count, err := rl.storage.Increment(ctx, key)
	if err != nil {
		return false, err
	}

	// Define o limite baseado no tipo de identificador
	limit := rl.rateLimitIP
	if isToken {
		limit = rl.rateLimitToken
	}

	// Se excedeu o limite, bloqueia
	if count > int64(limit) {
		err = rl.storage.Set(ctx, fmt.Sprintf("blocked:%s", key), 1, rl.blockDuration)
		if err != nil {
			return false, err
		}
		return false, nil
	}

	// Define expiração para o contador
	err = rl.storage.Set(ctx, key, count, 1)
	if err != nil {
		return false, err
	}

	return true, nil
}
