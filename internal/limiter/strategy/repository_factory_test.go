package strategy

import (
	"context"
	"testing"

	"github.com/JMKobayashi/Rate-Limiter-GO/internal/entity"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRepository(t *testing.T) {
	tests := []struct {
		name     string
		repoType RepositoryType
		config   map[string]interface{}
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "Memory repository",
			repoType: MemoryRepository,
			config:   map[string]interface{}{},
			wantErr:  false,
		},
		{
			name:     "Redis repository",
			repoType: RedisRepository,
			config: map[string]interface{}{
				"client": redis.NewClient(&redis.Options{
					Addr: "localhost:6379",
				}),
			},
			wantErr: false,
		},
		{
			name:     "Invalid repository type",
			repoType: "invalid",
			config:   map[string]interface{}{},
			wantErr:  true,
			errMsg:   "tipo de repositório inválido",
		},
		{
			name:     "Redis without client",
			repoType: RedisRepository,
			config:   map[string]interface{}{},
			wantErr:  true,
			errMsg:   "cliente Redis não fornecido",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := NewRepository(tt.repoType, tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
				assert.Nil(t, repo)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, repo)

				// Testar operações básicas
				limiter, err := entity.NewRateLimiter("192.168.1.1", "")
				require.NoError(t, err)

				// Salvar
				err = repo.Save(context.Background(), limiter)
				require.NoError(t, err)

				// Buscar
				got, err := repo.Get(context.Background(), "rate_limit:192.168.1.1")
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, limiter.IP, got.IP)

				// Deletar
				err = repo.Delete(context.Background(), "rate_limit:192.168.1.1")
				require.NoError(t, err)

				// Verificar se foi deletado
				got, err = repo.Get(context.Background(), "rate_limit:192.168.1.1")
				require.NoError(t, err)
				assert.Nil(t, got)
			}
		})
	}
}

func TestRepositoryFactory_Singleton(t *testing.T) {
	factory := NewRepositoryFactory()

	// Criar primeiro repositório
	repo1, err := factory.GetRepository(MemoryRepository)
	require.NoError(t, err)
	require.NotNil(t, repo1)

	// Criar segundo repositório
	repo2, err := factory.GetRepository(MemoryRepository)
	require.NoError(t, err)
	require.NotNil(t, repo2)

	// Verificar se são a mesma instância
	assert.Equal(t, repo1, repo2)

	// Criar repositório Redis
	repo3, err := factory.GetRepository(RedisRepository)
	require.NoError(t, err)
	require.NotNil(t, repo3)

	// Verificar se é uma instância diferente
	assert.NotEqual(t, repo1, repo3)
}

func TestRepositoryFactory_ConcurrentAccess(t *testing.T) {
	factory := NewRepositoryFactory()

	// Criar repositório
	repo, err := factory.GetRepository(MemoryRepository)
	require.NoError(t, err)
	require.NotNil(t, repo)

	// Criar um limitador
	limiter, err := entity.NewRateLimiter("192.168.1.1", "")
	require.NoError(t, err)

	// Salvar o limitador
	err = repo.Save(context.Background(), limiter)
	require.NoError(t, err)

	// Simular acesso concorrente
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			got, err := repo.Get(context.Background(), "rate_limit:192.168.1.1")
			require.NoError(t, err)
			require.NotNil(t, got)
			done <- true
		}()
	}

	// Aguardar todas as goroutines terminarem
	for i := 0; i < 10; i++ {
		<-done
	}
}
