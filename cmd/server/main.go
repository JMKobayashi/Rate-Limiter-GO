package main

import (
	"context"
	"log"
	"time"

	"github.com/JMKobayashi/Rate-Limiter-GO/internal/limiter/strategy"
	"github.com/JMKobayashi/Rate-Limiter-GO/internal/middleware"
	"github.com/JMKobayashi/Rate-Limiter-GO/internal/usecase"
	"github.com/JMKobayashi/Rate-Limiter-GO/pkg/config"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Carrega as configurações
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Erro ao carregar configurações:", err)
	}
	log.Printf("Configurações carregadas: Redis=%s:%s, DB=%d", cfg.RedisHost, cfg.RedisPort, cfg.RedisDB)

	// Configura o cliente Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisHost + ":" + cfg.RedisPort,
		Password: cfg.RedisPassword,
		DB:       int(cfg.RedisDB),
	})

	// Testa conexão com Redis
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Erro ao conectar com Redis: %v", err)
	}
	log.Println("Conexão com Redis estabelecida com sucesso")

	// Configura a estratégia Redis
	redisConfig := map[string]interface{}{
		"client": redisClient,
	}

	// Inicializa a estratégia Redis usando a factory
	redisStrategy, err := strategy.NewRepository(strategy.RedisRepository, redisConfig)
	if err != nil {
		log.Fatalf("Erro ao inicializar estratégia Redis: %v", err)
	}
	log.Println("Estratégia Redis inicializada com sucesso")

	// Inicializa o caso de uso
	rateLimiterUseCase := usecase.NewRateLimiterUseCase(
		redisStrategy,
		cfg.RateLimitIP,
		cfg.RateLimitToken,
		cfg.BlockDurationIP,
		cfg.BlockDurationToken,
		cfg.EnableIPLimiter,
		cfg.EnableTokenLimiter,
	)
	log.Printf("Rate Limiter configurado: IP=%d, Token=%d, BlockIP=%d, BlockToken=%d",
		cfg.RateLimitIP, cfg.RateLimitToken, cfg.BlockDurationIP, cfg.BlockDurationToken)

	// Configura o servidor Gin
	r := gin.Default()

	// Adiciona o middleware de rate limiting
	r.Use(middleware.RateLimiter(rateLimiterUseCase))
	log.Println("Middleware de Rate Limiting adicionado")

	// Rota de exemplo
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello, World!",
		})
	})

	// Inicia o servidor
	log.Println("Iniciando servidor na porta 8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
}
