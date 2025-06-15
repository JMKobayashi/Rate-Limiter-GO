package middleware

import (
	"net/http"

	"github.com/JMKobayashi/Rate-Limiter-GO/internal/usecase"
	"github.com/gin-gonic/gin"
)

// RateLimiterMiddleware é um middleware para limitar requisições
type RateLimiterMiddleware struct {
	rateLimiterUseCase usecase.RateLimiterUseCaseInterface
}

// NewRateLimiterMiddleware cria um novo middleware de rate limiter
func NewRateLimiterMiddleware(rateLimiterUseCase usecase.RateLimiterUseCaseInterface) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{
		rateLimiterUseCase: rateLimiterUseCase,
	}
}

func RateLimiter(useCase usecase.RateLimiterUseCaseInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Verifica o token primeiro
		token := c.GetHeader("API_KEY")
		if token != "" {
			allowed, err := useCase.IsAllowed(c.Request.Context(), token, true)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				c.Abort()
				return
			}
			if !allowed {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error": "you have reached the maximum number of requests or actions allowed within a certain time frame",
				})
				c.Abort()
				return
			}
			c.Next()
			return
		}

		// Se não tem token, verifica o IP
		ip := c.ClientIP()
		allowed, err := useCase.IsAllowed(c.Request.Context(), ip, false)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "you have reached the maximum number of requests or actions allowed within a certain time frame",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
