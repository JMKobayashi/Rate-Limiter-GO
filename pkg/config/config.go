package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	RedisHost      string
	RedisPort      string
	RedisPassword  string
	RedisDB        int
	RateLimitIP    int
	RateLimitToken int
	BlockDuration  int
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
	rateLimitIP, _ := strconv.Atoi(getEnv("RATE_LIMIT_IP", "10"))
	rateLimitToken, _ := strconv.Atoi(getEnv("RATE_LIMIT_TOKEN", "100"))
	blockDuration, _ := strconv.Atoi(getEnv("BLOCK_DURATION", "300"))

	return &Config{
		RedisHost:      getEnv("REDIS_HOST", "localhost"),
		RedisPort:      getEnv("REDIS_PORT", "6379"),
		RedisPassword:  getEnv("REDIS_PASSWORD", ""),
		RedisDB:        redisDB,
		RateLimitIP:    rateLimitIP,
		RateLimitToken: rateLimitToken,
		BlockDuration:  blockDuration,
	}, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
