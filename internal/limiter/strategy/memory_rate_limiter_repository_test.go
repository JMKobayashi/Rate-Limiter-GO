package strategy

import (
	"context"
	"testing"
	"time"

	"github.com/JMKobayashi/Rate-Limiter-GO/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryRateLimiterRepository(t *testing.T) {
	repo := NewMemoryRateLimiterRepository()

	t.Run("Save and Get", func(t *testing.T) {
		// Criar um limitador
		limiter, err := entity.NewRateLimiter("192.168.1.1", "")
		require.NoError(t, err)

		// Salvar o limitador
		err = repo.Save(context.Background(), limiter)
		require.NoError(t, err)

		// Buscar o limitador
		got, err := repo.Get(context.Background(), "rate_limit:192.168.1.1")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, limiter.IP, got.IP)
		assert.Equal(t, limiter.Token, got.Token)
		assert.Equal(t, limiter.Requests, got.Requests)
		assert.Equal(t, limiter.Blocked, got.Blocked)
		assert.True(t, limiter.BlockedUntil.Equal(got.BlockedUntil))
	})

	t.Run("Get non-existent", func(t *testing.T) {
		got, err := repo.Get(context.Background(), "rate_limit:non-existent")
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("Delete", func(t *testing.T) {
		// Criar um limitador
		limiter, err := entity.NewRateLimiter("192.168.1.1", "")
		require.NoError(t, err)

		// Salvar o limitador
		err = repo.Save(context.Background(), limiter)
		require.NoError(t, err)

		// Deletar o limitador
		err = repo.Delete(context.Background(), "rate_limit:192.168.1.1")
		require.NoError(t, err)

		// Verificar se foi deletado
		got, err := repo.Get(context.Background(), "rate_limit:192.168.1.1")
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("Update existing", func(t *testing.T) {
		// Criar um limitador
		limiter, err := entity.NewRateLimiter("192.168.1.1", "")
		require.NoError(t, err)

		// Salvar o limitador
		err = repo.Save(context.Background(), limiter)
		require.NoError(t, err)

		// Atualizar o limitador
		limiter.Requests = 10
		limiter.Blocked = true
		limiter.BlockedUntil = time.Now().Add(time.Hour)

		err = repo.Save(context.Background(), limiter)
		require.NoError(t, err)

		// Verificar se foi atualizado
		got, err := repo.Get(context.Background(), "rate_limit:192.168.1.1")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, int64(10), got.Requests)
		assert.True(t, got.Blocked)
		assert.True(t, got.BlockedUntil.After(time.Now()))
	})

	t.Run("Concurrent access", func(t *testing.T) {
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
	})

	t.Run("Block expiration", func(t *testing.T) {
		// Criar um limitador bloqueado
		limiter, err := entity.NewRateLimiter("192.168.1.1", "")
		require.NoError(t, err)
		limiter.Blocked = true
		limiter.BlockedUntil = time.Now().Add(time.Second)

		// Salvar o limitador
		err = repo.Save(context.Background(), limiter)
		require.NoError(t, err)

		// Verificar se está bloqueado
		got, err := repo.Get(context.Background(), "rate_limit:192.168.1.1")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.True(t, got.Blocked)

		// Esperar o bloqueio expirar
		time.Sleep(time.Second * 2)

		// Verificar se o bloqueio expirou
		got, err = repo.Get(context.Background(), "rate_limit:192.168.1.1")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.False(t, got.Blocked)
		assert.True(t, got.BlockedUntil.IsZero())
	})

	t.Run("Token rate limiter", func(t *testing.T) {
		// Criar um limitador com token
		limiter, err := entity.NewRateLimiter("", "test-token")
		require.NoError(t, err)

		// Salvar o limitador
		err = repo.Save(context.Background(), limiter)
		require.NoError(t, err)

		// Buscar o limitador
		got, err := repo.Get(context.Background(), "rate_limit:test-token")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, limiter.Token, got.Token)
		assert.Empty(t, got.IP)
	})
}
