package main

import (
	"log"

	"github.com/JMKobayashi/Rate-Limiter-GO/internal/infra/repository"
	"github.com/JMKobayashi/Rate-Limiter-GO/internal/middleware"
	"github.com/JMKobayashi/Rate-Limiter-GO/internal/usecase"
	"github.com/JMKobayashi/Rate-Limiter-GO/pkg/config"
	"github.com/gin-gonic/gin"
)

func main() {
	// Carrega as configurações
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Erro ao carregar configurações:", err)
	}

	// Inicializa o repositório Redis
	redisRepo := repository.NewRedisRateLimiterRepository(
		cfg.RedisHost,
		cfg.RedisPort,
		cfg.RedisPassword,
		cfg.RedisDB,
	)

	// Inicializa o caso de uso
	rateLimiterUseCase := usecase.NewRateLimiterUseCase(
		redisRepo,
		cfg.RateLimitIP,
		cfg.RateLimitToken,
		cfg.BlockDuration,
	)

	// Configura o servidor Gin
	r := gin.Default()

	// Adiciona o middleware de rate limiting
	r.Use(middleware.RateLimiter(rateLimiterUseCase))

	// Rota de exemplo
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello, World!",
		})
	})

	// Inicia o servidor
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Erro ao iniciar o servidor:", err)
	}
}
