package strategy

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/JMKobayashi/Rate-Limiter-GO/internal/entity"
	"github.com/JMKobayashi/Rate-Limiter-GO/internal/repository"
	"github.com/redis/go-redis/v9"
)

// RedisRateLimiterRepository implementa o repositório de rate limiter usando Redis
type RedisRateLimiterRepository struct {
	client *redis.Client
}

// NewRedisRateLimiterRepository cria um novo repositório Redis
func NewRedisRateLimiterRepository(client *redis.Client) repository.RateLimiterRepository {
	return &RedisRateLimiterRepository{
		client: client,
	}
}

// Save salva um rate limiter no Redis
func (r *RedisRateLimiterRepository) Save(ctx context.Context, limiter *entity.RateLimiter) error {
	key := r.getKey(limiter)
	data, err := json.Marshal(limiter)
	if err != nil {
		return fmt.Errorf("erro ao serializar rate limiter: %v", err)
	}

	// Define o TTL baseado no estado do limiter
	var ttl time.Duration
	if limiter.Blocked {
		// Se está bloqueado, usa o tempo de bloqueio
		ttl = time.Until(limiter.BlockedUntil)
		if ttl < 0 {
			ttl = 0
		}
	} else {
		// Se não está bloqueado, usa 24 horas como TTL padrão
		ttl = 24 * time.Hour
	}

	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("erro ao salvar rate limiter no Redis: %v", err)
	}

	return nil
}

// Get recupera um rate limiter do Redis
func (r *RedisRateLimiterRepository) Get(ctx context.Context, key string) (*entity.RateLimiter, error) {
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao recuperar rate limiter do Redis: %v", err)
	}

	var limiter entity.RateLimiter
	if err := json.Unmarshal(data, &limiter); err != nil {
		return nil, fmt.Errorf("erro ao deserializar rate limiter: %v", err)
	}

	// Verifica se o bloqueio expirou
	if limiter.Blocked && time.Now().After(limiter.BlockedUntil) {
		limiter.Blocked = false
		limiter.BlockedUntil = time.Time{}
		limiter.Requests = 0
		err = r.Save(ctx, &limiter)
		if err != nil {
			return nil, err
		}
	}

	return &limiter, nil
}

// Delete remove um rate limiter do Redis
func (r *RedisRateLimiterRepository) Delete(ctx context.Context, key string) error {
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("erro ao deletar rate limiter do Redis: %v", err)
	}
	return nil
}

// getKey retorna a chave do Redis para um rate limiter
func (r *RedisRateLimiterRepository) getKey(limiter *entity.RateLimiter) string {
	if limiter.IP != "" {
		return fmt.Sprintf("rate_limiter:ip:%s", limiter.IP)
	}
	return fmt.Sprintf("rate_limiter:token:%s", limiter.Token)
}
