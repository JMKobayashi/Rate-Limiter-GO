package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JMKobayashi/Rate-Limiter-GO/internal/usecase"
	"github.com/gin-gonic/gin"
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

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
