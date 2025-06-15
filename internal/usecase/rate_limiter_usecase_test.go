package usecase

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/JMKobayashi/Rate-Limiter-GO/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockRateLimiterRepository é uma implementação mock do RateLimiterRepository
// para ser usada em testes unitários.
type MockRateLimiterRepository struct {
	mu       sync.Mutex                     // Mutex para garantir concorrência segura ao acessar o mapa
	limiters map[string]*entity.RateLimiter // Um mapa em memória para simular o armazenamento
	err      error                          // Um erro opcional para simular falhas no repositório
}

// NewMockRateLimiterRepository cria uma nova instância do mock do repositório.
func NewMockRateLimiterRepository() *MockRateLimiterRepository {
	return &MockRateLimiterRepository{
		limiters: make(map[string]*entity.RateLimiter),
	}
}

// Get simula a obtenção de um RateLimiter do "armazenamento".
// Se um erro for configurado no mock, ele será retornado.
func (m *MockRateLimiterRepository) Get(ctx context.Context, key string) (*entity.RateLimiter, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.err != nil {
		return nil, m.err
	}
	if limiter, exists := m.limiters[key]; exists {
		// Verifica se o bloqueio expirou
		if limiter.Blocked && !limiter.BlockedUntil.IsZero() && time.Now().After(limiter.BlockedUntil) {
			limiter.Blocked = false
			limiter.BlockedUntil = time.Time{}
			limiter.Requests = 0
			m.limiters[key] = limiter
		}
		return limiter, nil
	}
	return nil, nil
}

// Save simula o salvamento de um RateLimiter no "armazenamento".
// Se um erro for configurado no mock, ele será retornado.
func (m *MockRateLimiterRepository) Save(ctx context.Context, limiter *entity.RateLimiter) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.err != nil {
		return m.err
	}
	key := "rate_limit:" + limiter.IP
	if limiter.Token != "" {
		key = "rate_limit:" + limiter.Token
	}
	m.limiters[key] = limiter
	return nil
}

// Delete simula a remoção de um RateLimiter do "armazenamento".
func (m *MockRateLimiterRepository) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.limiters, key)
	return nil
}

// SetError permite configurar um erro para ser retornado pelas operações do mock.
func (m *MockRateLimiterRepository) SetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.err = err
}

// TestRateLimiterUseCase_IsAllowed testa o método IsAllowed do RateLimiterUseCase.
func TestRateLimiterUseCase_IsAllowed(t *testing.T) {
	tests := []struct {
		name           string
		identifier     string
		isToken        bool
		initialLimiter *entity.RateLimiter
		expected       bool
		config         struct {
			rateLimitIP        int
			rateLimitToken     int
			blockDurationIP    int
			blockDurationToken int
			enableIPLimiter    bool
			enableTokenLimiter bool
		}
	}{
		{
			name:       "IP dentro do limite",
			identifier: "192.168.1.1",
			isToken:    false,
			initialLimiter: &entity.RateLimiter{
				IP:           "192.168.1.1",
				Requests:     5,
				LastRequest:  time.Now(),
				Blocked:      false,
				BlockedUntil: time.Time{},
			},
			expected: true,
			config: struct {
				rateLimitIP        int
				rateLimitToken     int
				blockDurationIP    int
				blockDurationToken int
				enableIPLimiter    bool
				enableTokenLimiter bool
			}{
				rateLimitIP:        10,
				rateLimitToken:     100,
				blockDurationIP:    300,
				blockDurationToken: 600,
				enableIPLimiter:    true,
				enableTokenLimiter: true,
			},
		},
		{
			name:       "IP excede limite",
			identifier: "192.168.1.1",
			isToken:    false,
			initialLimiter: &entity.RateLimiter{
				IP:           "192.168.1.1",
				Requests:     10,
				LastRequest:  time.Now(),
				Blocked:      false,
				BlockedUntil: time.Time{},
			},
			expected: false,
			config: struct {
				rateLimitIP        int
				rateLimitToken     int
				blockDurationIP    int
				blockDurationToken int
				enableIPLimiter    bool
				enableTokenLimiter bool
			}{
				rateLimitIP:        10,
				rateLimitToken:     100,
				blockDurationIP:    300,
				blockDurationToken: 600,
				enableIPLimiter:    true,
				enableTokenLimiter: true,
			},
		},
		{
			name:       "Token dentro do limite",
			identifier: "test-token",
			isToken:    true,
			initialLimiter: &entity.RateLimiter{
				Token:        "test-token",
				Requests:     50,
				LastRequest:  time.Now(),
				Blocked:      false,
				BlockedUntil: time.Time{},
			},
			expected: true,
			config: struct {
				rateLimitIP        int
				rateLimitToken     int
				blockDurationIP    int
				blockDurationToken int
				enableIPLimiter    bool
				enableTokenLimiter bool
			}{
				rateLimitIP:        10,
				rateLimitToken:     100,
				blockDurationIP:    300,
				blockDurationToken: 600,
				enableIPLimiter:    true,
				enableTokenLimiter: true,
			},
		},
		{
			name:       "Token excede limite",
			identifier: "test-token",
			isToken:    true,
			initialLimiter: &entity.RateLimiter{
				Token:        "test-token",
				Requests:     100,
				LastRequest:  time.Now(),
				Blocked:      false,
				BlockedUntil: time.Time{},
			},
			expected: false,
			config: struct {
				rateLimitIP        int
				rateLimitToken     int
				blockDurationIP    int
				blockDurationToken int
				enableIPLimiter    bool
				enableTokenLimiter bool
			}{
				rateLimitIP:        10,
				rateLimitToken:     100,
				blockDurationIP:    300,
				blockDurationToken: 600,
				enableIPLimiter:    true,
				enableTokenLimiter: true,
			},
		},
		{
			name:       "IP limiter desabilitado",
			identifier: "192.168.1.1",
			isToken:    false,
			initialLimiter: &entity.RateLimiter{
				IP:           "192.168.1.1",
				Requests:     100,
				LastRequest:  time.Now(),
				Blocked:      false,
				BlockedUntil: time.Time{},
			},
			expected: true,
			config: struct {
				rateLimitIP        int
				rateLimitToken     int
				blockDurationIP    int
				blockDurationToken int
				enableIPLimiter    bool
				enableTokenLimiter bool
			}{
				rateLimitIP:        10,
				rateLimitToken:     100,
				blockDurationIP:    300,
				blockDurationToken: 600,
				enableIPLimiter:    false,
				enableTokenLimiter: true,
			},
		},
		{
			name:       "Token limiter desabilitado",
			identifier: "test-token",
			isToken:    true,
			initialLimiter: &entity.RateLimiter{
				Token:        "test-token",
				Requests:     1000,
				LastRequest:  time.Now(),
				Blocked:      false,
				BlockedUntil: time.Time{},
			},
			expected: true,
			config: struct {
				rateLimitIP        int
				rateLimitToken     int
				blockDurationIP    int
				blockDurationToken int
				enableIPLimiter    bool
				enableTokenLimiter bool
			}{
				rateLimitIP:        10,
				rateLimitToken:     100,
				blockDurationIP:    300,
				blockDurationToken: 600,
				enableIPLimiter:    true,
				enableTokenLimiter: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockRateLimiterRepository()
			useCase := NewRateLimiterUseCase(
				repo,
				tt.config.rateLimitIP,
				tt.config.rateLimitToken,
				tt.config.blockDurationIP,
				tt.config.blockDurationToken,
				tt.config.enableIPLimiter,
				tt.config.enableTokenLimiter,
			)

			if tt.initialLimiter != nil {
				err := repo.Save(context.Background(), tt.initialLimiter)
				require.NoError(t, err)
			}

			allowed, err := useCase.IsAllowed(context.Background(), tt.identifier, tt.isToken)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, allowed)
		})
	}
}

// TestRateLimiterUseCase_IsAllowed_RepositoryError testa o tratamento de erro do repositório.
func TestRateLimiterUseCase_IsAllowed_RepositoryError(t *testing.T) {
	repo := NewMockRateLimiterRepository()
	useCase := NewRateLimiterUseCase(repo, 10, 100, 300, 300, true, true)

	// Test Get error
	repo.err = errors.New("erro simulado do repositório")
	allowed, err := useCase.IsAllowed(context.Background(), "test_id", false)
	assert.Error(t, err)
	assert.False(t, allowed)
	assert.Contains(t, err.Error(), "erro simulado do repositório")

	// Test Save error
	repo.err = errors.New("erro simulado ao salvar")
	allowed, err = useCase.IsAllowed(context.Background(), "block_id", false)
	assert.Error(t, err)
	assert.False(t, allowed)
	assert.Contains(t, err.Error(), "erro simulado ao salvar")
}

// TestRateLimiterUseCase_BlockExpiration testa se o bloqueio expira corretamente.
func TestRateLimiterUseCase_BlockExpiration(t *testing.T) {
	tests := []struct {
		name          string
		identifier    string
		isToken       bool
		blockDuration int
		waitTime      time.Duration
		expectedAfter bool
	}{
		{
			name:          "IP block expira",
			identifier:    "192.168.1.1",
			isToken:       false,
			blockDuration: 1,
			waitTime:      time.Second * 2,
			expectedAfter: true,
		},
		{
			name:          "Token block expira",
			identifier:    "test-token",
			isToken:       true,
			blockDuration: 1,
			waitTime:      time.Second * 2,
			expectedAfter: true,
		},
		{
			name:          "IP block não expira",
			identifier:    "192.168.1.1",
			isToken:       false,
			blockDuration: 3,
			waitTime:      time.Second * 2,
			expectedAfter: false,
		},
		{
			name:          "Token block não expira",
			identifier:    "test-token",
			isToken:       true,
			blockDuration: 3,
			waitTime:      time.Second * 2,
			expectedAfter: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockRateLimiterRepository()
			useCase := NewRateLimiterUseCase(
				repo,
				10,
				100,
				tt.blockDuration,
				tt.blockDuration,
				true,
				true,
			)

			// Criar um limitador bloqueado
			limiter := &entity.RateLimiter{
				IP:           tt.identifier,
				Token:        "",
				Requests:     10,
				LastRequest:  time.Now(),
				Blocked:      true,
				BlockedUntil: time.Now().Add(time.Duration(tt.blockDuration) * time.Second),
			}

			if tt.isToken {
				limiter.IP = ""
				limiter.Token = tt.identifier
			}

			// Salvar o limitador
			err := repo.Save(context.Background(), limiter)
			require.NoError(t, err)

			// Verificar se está bloqueado
			allowed, err := useCase.IsAllowed(context.Background(), tt.identifier, tt.isToken)
			require.NoError(t, err)
			assert.False(t, allowed)

			// Esperar o tempo de bloqueio
			time.Sleep(tt.waitTime)

			// Verificar se o bloqueio expirou
			allowed, err = useCase.IsAllowed(context.Background(), tt.identifier, tt.isToken)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedAfter, allowed)

			// Verificar o estado do limitador
			got, err := repo.Get(context.Background(), "rate_limit:"+tt.identifier)
			require.NoError(t, err)
			require.NotNil(t, got)

			if tt.expectedAfter {
				assert.False(t, got.Blocked)
				assert.True(t, got.BlockedUntil.IsZero())
				assert.Equal(t, int64(1), got.Requests)
			} else {
				assert.True(t, got.Blocked)
				assert.True(t, got.BlockedUntil.After(time.Now()))
			}
		})
	}
}

func TestRateLimiterUseCase_ConcurrentRequests(t *testing.T) {
	repo := NewMockRateLimiterRepository()
	useCase := NewRateLimiterUseCase(
		repo,
		10,
		100,
		300,
		600,
		true,
		true,
	)

	// Teste concorrente para IP
	t.Run("Concurrent IP requests", func(t *testing.T) {
		var wg sync.WaitGroup
		ip := "192.168.1.1"
		requests := 20
		successCount := 0
		var mu sync.Mutex

		for i := 0; i < requests; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				allowed, err := useCase.IsAllowed(context.Background(), ip, false)
				require.NoError(t, err)
				if allowed {
					mu.Lock()
					successCount++
					mu.Unlock()
				}
			}()
		}

		wg.Wait()
		assert.LessOrEqual(t, successCount, 10, "Número de requisições bem-sucedidas deve ser menor ou igual ao limite")
	})

	// Teste concorrente para Token
	t.Run("Concurrent Token requests", func(t *testing.T) {
		var wg sync.WaitGroup
		token := "test-token"
		requests := 120
		successCount := 0
		var mu sync.Mutex

		for i := 0; i < requests; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				allowed, err := useCase.IsAllowed(context.Background(), token, true)
				require.NoError(t, err)
				if allowed {
					mu.Lock()
					successCount++
					mu.Unlock()
				}
			}()
		}

		wg.Wait()
		assert.LessOrEqual(t, successCount, 100, "Número de requisições bem-sucedidas deve ser menor ou igual ao limite")
	})
}

func TestRateLimiterUseCase_RepositoryErrors(t *testing.T) {
	repo := NewMockRateLimiterRepository()
	useCase := NewRateLimiterUseCase(
		repo,
		10,
		100,
		300,
		600,
		true,
		true,
	)

	t.Run("Get error", func(t *testing.T) {
		repo.SetError(errors.New("erro simulado do repositório"))
		allowed, err := useCase.IsAllowed(context.Background(), "test", false)
		assert.Error(t, err)
		assert.False(t, allowed)
		assert.Contains(t, err.Error(), "erro simulado do repositório")
	})

	t.Run("Save error", func(t *testing.T) {
		repo.SetError(nil) // Limpa o erro anterior
		allowed, err := useCase.IsAllowed(context.Background(), "test", false)
		require.NoError(t, err)
		assert.True(t, allowed)

		repo.SetError(errors.New("erro simulado ao salvar"))
		allowed, err = useCase.IsAllowed(context.Background(), "test", false)
		assert.Error(t, err)
		assert.False(t, allowed)
		assert.Contains(t, err.Error(), "erro simulado ao salvar")
	})
}
