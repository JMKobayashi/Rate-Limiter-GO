package strategy

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisStrategy struct {
	client *redis.Client
}

func NewRedisStrategy(host, port, password string, db int) *RedisStrategy {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       db,
	})

	return &RedisStrategy{
		client: client,
	}
}

func (r *RedisStrategy) Increment(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

func (r *RedisStrategy) Get(ctx context.Context, key string) (int64, error) {
	return r.client.Get(ctx, key).Int64()
}

func (r *RedisStrategy) Set(ctx context.Context, key string, value int64, expiration int) error {
	return r.client.Set(ctx, key, value, time.Duration(expiration)*time.Second).Err()
}

func (r *RedisStrategy) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}
