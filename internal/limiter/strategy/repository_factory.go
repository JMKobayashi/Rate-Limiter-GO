package strategy

import (
	"fmt"
	"sync"

	"github.com/JMKobayashi/Rate-Limiter-GO/internal/repository"
	"github.com/redis/go-redis/v9"
)

// RepositoryFactory é uma fábrica para criar repositórios
type RepositoryFactory struct {
	repositories map[RepositoryType]repository.RateLimiterRepository
	mu           sync.RWMutex
}

// NewRepositoryFactory cria uma nova instância da fábrica de repositórios
func NewRepositoryFactory() *RepositoryFactory {
	return &RepositoryFactory{
		repositories: make(map[RepositoryType]repository.RateLimiterRepository),
	}
}

// NewRepository cria um novo repositório
func NewRepository(repoType RepositoryType, config map[string]interface{}) (repository.RateLimiterRepository, error) {
	switch repoType {
	case MemoryRepository:
		return NewMemoryRateLimiterRepository(), nil
	case RedisRepository:
		client, ok := config["client"].(*redis.Client)
		if !ok {
			return nil, fmt.Errorf("cliente Redis não fornecido")
		}
		return NewRedisRateLimiterRepository(client), nil
	default:
		return nil, ErrInvalidRepositoryType
	}
}

// GetRepository retorna uma instância do repositório
func (f *RepositoryFactory) GetRepository(repoType RepositoryType) (repository.RateLimiterRepository, error) {
	f.mu.RLock()
	repo, exists := f.repositories[repoType]
	f.mu.RUnlock()

	if exists {
		return repo, nil
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	// Verifica novamente após adquirir o lock
	repo, exists = f.repositories[repoType]
	if exists {
		return repo, nil
	}

	// Cria uma nova instância
	config := make(map[string]interface{})
	repo, err := NewRepository(repoType, config)
	if err != nil {
		return nil, err
	}

	f.repositories[repoType] = repo
	return repo, nil
}

// ... existing code ...
