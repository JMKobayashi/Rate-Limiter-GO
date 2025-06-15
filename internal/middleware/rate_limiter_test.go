package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JMKobayashi/Rate-Limiter-GO/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type MockUseCase struct {
	allowed bool
	err     error
}

func (m *MockUseCase) IsAllowed(ctx context.Context, identifier string, isToken bool) (bool, error) {
	return m.allowed, m.err
}

func TestRateLimiter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		useCase        usecase.RateLimiterUseCaseInterface
		headers        map[string]string
		expectedStatus int
	}{
		{
			name: "Allowed request with token",
			useCase: &MockUseCase{
				allowed: true,
				err:     nil,
			},
			headers: map[string]string{
				"API_KEY": "test-token",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Blocked request with token",
			useCase: &MockUseCase{
				allowed: false,
				err:     nil,
			},
			headers: map[string]string{
				"API_KEY": "test-token",
			},
			expectedStatus: http.StatusTooManyRequests,
		},
		{
			name: "Allowed request without token",
			useCase: &MockUseCase{
				allowed: true,
				err:     nil,
			},
			headers:        map[string]string{},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Blocked request without token",
			useCase: &MockUseCase{
				allowed: false,
				err:     nil,
			},
			headers:        map[string]string{},
			expectedStatus: http.StatusTooManyRequests,
		},
		{
			name: "Error from use case",
			useCase: &MockUseCase{
				allowed: false,
				err:     fmt.Errorf("test error"),
			},
			headers:        map[string]string{},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(RateLimiter(tt.useCase))
			router.GET("/", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest("GET", "/", nil)
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// MockRateLimiterUseCase é uma implementação mock do RateLimiterUseCaseInterface
type MockRateLimiterUseCase struct {
	allowed bool
	err     error
}

func (m *MockRateLimiterUseCase) IsAllowed(ctx context.Context, identifier string, isToken bool) (bool, error) {
	return m.allowed, m.err
}

func TestRateLimiterMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		ip             string
		token          string
		mockAllowed    bool
		mockError      error
		expectedStatus int
	}{
		{
			name:           "Requisição permitida com IP",
			ip:             "192.168.1.1",
			token:          "",
			mockAllowed:    true,
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Requisição bloqueada com IP",
			ip:             "192.168.1.1",
			token:          "",
			mockAllowed:    false,
			mockError:      nil,
			expectedStatus: http.StatusTooManyRequests,
		},
		{
			name:           "Requisição permitida com Token",
			ip:             "192.168.1.1",
			token:          "test-token",
			mockAllowed:    true,
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Requisição bloqueada com Token",
			ip:             "192.168.1.1",
			token:          "test-token",
			mockAllowed:    false,
			mockError:      nil,
			expectedStatus: http.StatusTooManyRequests,
		},
		{
			name:           "Erro no rate limiter",
			ip:             "192.168.1.1",
			token:          "",
			mockAllowed:    false,
			mockError:      assert.AnError,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Criar mock do use case
			mockUseCase := &MockRateLimiterUseCase{
				allowed: tt.mockAllowed,
				err:     tt.mockError,
			}

			// Criar router do Gin
			router := gin.New()
			router.Use(RateLimiter(mockUseCase))
			router.GET("/", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			// Criar request de teste
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.ip
			if tt.token != "" {
				req.Header.Set("API_KEY", tt.token)
			}

			// Criar response recorder
			rr := httptest.NewRecorder()

			// Executar request
			router.ServeHTTP(rr, req)

			// Verificar status code
			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestRateLimiterMiddleware_ConcurrentRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Criar mock do use case que alterna entre permitir e bloquear
	mockUseCase := &MockUseCase{
		allowed: true,
		err:     nil,
	}

	router := gin.New()
	router.Use(RateLimiter(mockUseCase))
	router.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Criar request de teste
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1"

	// Criar response recorder
	rr := httptest.NewRecorder()

	// Executar request
	router.ServeHTTP(rr, req)

	// Verificar status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Simular bloqueio
	mockUseCase.allowed = false

	// Executar request novamente
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Verificar status code
	assert.Equal(t, http.StatusTooManyRequests, rr.Code)
}

func TestRateLimiterMiddleware_ContextTimeout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Criar mock do use case que demora para responder
	mockUseCase := &MockUseCase{
		allowed: true,
		err:     nil,
	}

	router := gin.New()
	router.Use(RateLimiter(mockUseCase))
	router.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Criar request de teste com contexto com timeout
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1"
	ctx, cancel := context.WithTimeout(req.Context(), time.Millisecond*100)
	defer cancel()
	req = req.WithContext(ctx)

	// Criar response recorder
	rr := httptest.NewRecorder()

	// Executar request
	router.ServeHTTP(rr, req)

	// Verificar status code
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRateLimiterMiddleware_InvalidIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Criar mock do use case
	mockUseCase := &MockUseCase{
		allowed: true,
		err:     nil,
	}

	router := gin.New()
	router.Use(RateLimiter(mockUseCase))
	router.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Criar request de teste com IP inválido
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "invalid-ip"

	// Criar response recorder
	rr := httptest.NewRecorder()

	// Executar request
	router.ServeHTTP(rr, req)

	// Verificar status code
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestRateLimiterMiddleware_NoIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Criar mock do use case
	mockUseCase := &MockUseCase{
		allowed: true,
		err:     nil,
	}

	router := gin.New()
	router.Use(RateLimiter(mockUseCase))
	router.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Criar request de teste sem IP
	req := httptest.NewRequest("GET", "/", nil)

	// Criar response recorder
	rr := httptest.NewRecorder()

	// Executar request
	router.ServeHTTP(rr, req)

	// Verificar status code
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
