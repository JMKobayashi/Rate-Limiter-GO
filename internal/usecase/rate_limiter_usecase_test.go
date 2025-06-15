package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/JMKobayashi/Rate-Limiter-GO/internal/entity"
)

type MockRepository struct {
	limiters map[string]*entity.RateLimiter
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		limiters: make(map[string]*entity.RateLimiter),
	}
}

func (m *MockRepository) Save(ctx context.Context, limiter *entity.RateLimiter) error {
	key := limiter.IP
	if limiter.Token != "" {
		key = limiter.Token
	}
	m.limiters[key] = limiter
	return nil
}

func (m *MockRepository) Get(ctx context.Context, key string) (*entity.RateLimiter, error) {
	return m.limiters[key], nil
}

func (m *MockRepository) Delete(ctx context.Context, key string) error {
	delete(m.limiters, key)
	return nil
}

func TestRateLimiterUseCase_IsAllowed(t *testing.T) {
	tests := []struct {
		name           string
		identifier     string
		isToken        bool
		rateLimitIP    int
		rateLimitToken int
		requests       int64
		expected       bool
	}{
		{
			name:           "IP within limit",
			identifier:     "192.168.1.1",
			isToken:        false,
			rateLimitIP:    10,
			rateLimitToken: 100,
			requests:       5,
			expected:       true,
		},
		{
			name:           "IP exceeding limit",
			identifier:     "192.168.1.1",
			isToken:        false,
			rateLimitIP:    10,
			rateLimitToken: 100,
			requests:       11,
			expected:       false,
		},
		{
			name:           "Token within limit",
			identifier:     "test-token",
			isToken:        true,
			rateLimitIP:    10,
			rateLimitToken: 100,
			requests:       50,
			expected:       true,
		},
		{
			name:           "Token exceeding limit",
			identifier:     "test-token",
			isToken:        true,
			rateLimitIP:    10,
			rateLimitToken: 100,
			requests:       101,
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockRepository()
			useCase := NewRateLimiterUseCase(repo, tt.rateLimitIP, tt.rateLimitToken, 300)

			// Setup initial state
			limiter := entity.NewRateLimiter(tt.identifier, "")
			if tt.isToken {
				limiter.Token = tt.identifier
			}
			limiter.Requests = tt.requests
			limiter.LastRequest = time.Now()
			repo.Save(context.Background(), limiter)

			// Test
			allowed, err := useCase.IsAllowed(context.Background(), tt.identifier, tt.isToken)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if allowed != tt.expected {
				t.Errorf("Expected IsAllowed() to be %v, got %v", tt.expected, allowed)
			}

			// Verify state after request
			updatedLimiter, _ := repo.Get(context.Background(), tt.identifier)
			if updatedLimiter != nil {
				if tt.expected && updatedLimiter.Blocked {
					t.Error("Expected limiter to not be blocked")
				}
				if !tt.expected && !updatedLimiter.Blocked {
					t.Error("Expected limiter to be blocked")
				}
			}
		})
	}
}

func TestRateLimiterUseCase_BlockDuration(t *testing.T) {
	repo := NewMockRepository()
	blockDuration := 1 // 1 second
	useCase := NewRateLimiterUseCase(repo, 10, 100, blockDuration)

	// Test blocking
	allowed, err := useCase.IsAllowed(context.Background(), "192.168.1.1", false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !allowed {
		t.Error("Expected first request to be allowed")
	}

	// Make requests to exceed limit
	for i := 0; i < 10; i++ {
		useCase.IsAllowed(context.Background(), "192.168.1.1", false)
	}

	// Should be blocked
	allowed, err = useCase.IsAllowed(context.Background(), "192.168.1.1", false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if allowed {
		t.Error("Expected to be blocked after exceeding limit")
	}

	// Verify blocked state
	limiter, _ := repo.Get(context.Background(), "192.168.1.1")
	if limiter != nil && !limiter.Blocked {
		t.Error("Expected limiter to be blocked")
	}

	// Wait for block duration
	time.Sleep(time.Duration(blockDuration+1) * time.Second)

	// Should be allowed again
	allowed, err = useCase.IsAllowed(context.Background(), "192.168.1.1", false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !allowed {
		t.Error("Expected to be allowed again after block duration")
	}

	// Verify unblocked state
	limiter, _ = repo.Get(context.Background(), "192.168.1.1")
	if limiter != nil && limiter.Blocked {
		t.Error("Expected limiter to be unblocked")
	}
}
