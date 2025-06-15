package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/JMKobayashi/Rate-Limiter-GO/internal/entity"
	"github.com/go-redis/redis/v8"
)

type RedisRateLimiterRepository struct {
	client *redis.Client
}

func NewRedisRateLimiterRepository(host, port, password string, db int) *RedisRateLimiterRepository {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       db,
	})

	return &RedisRateLimiterRepository{
		client: client,
	}
}

func (r *RedisRateLimiterRepository) Save(ctx context.Context, limiter *entity.RateLimiter) error {
	key := fmt.Sprintf("rate_limit:%s", limiter.IP)
	if limiter.Token != "" {
		key = fmt.Sprintf("rate_limit:%s", limiter.Token)
	}

	data, err := json.Marshal(limiter)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, data, time.Second).Err()
}

func (r *RedisRateLimiterRepository) Get(ctx context.Context, key string) (*entity.RateLimiter, error) {
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var limiter entity.RateLimiter
	if err := json.Unmarshal(data, &limiter); err != nil {
		return nil, err
	}

	return &limiter, nil
}

func (r *RedisRateLimiterRepository) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}
