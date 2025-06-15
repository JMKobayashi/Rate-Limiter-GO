package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	RedisHost          string
	RedisPort          string
	RedisPassword      string
	RedisDB            int
	RateLimitIP        int
	RateLimitToken     int
	BlockDurationIP    int
	BlockDurationToken int
	EnableIPLimiter    bool
	EnableTokenLimiter bool
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	config := &Config{
		RedisHost:          getEnv("REDIS_HOST", "localhost"),
		RedisPort:          getEnv("REDIS_PORT", "6379"),
		RedisPassword:      getEnv("REDIS_PASSWORD", ""),
		RedisDB:            getEnvAsInt("REDIS_DB", 0),
		RateLimitIP:        getEnvAsInt("RATE_LIMIT_IP", 10),
		RateLimitToken:     getEnvAsInt("RATE_LIMIT_TOKEN", 100),
		BlockDurationIP:    getEnvAsInt("BLOCK_DURATION_IP", 300),
		BlockDurationToken: getEnvAsInt("BLOCK_DURATION_TOKEN", 600),
		EnableIPLimiter:    getEnvAsBool("ENABLE_IP_LIMITER", true),
		EnableTokenLimiter: getEnvAsBool("ENABLE_TOKEN_LIMITER", true),
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvAsInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	return strings.ToLower(value) == "true"
}
