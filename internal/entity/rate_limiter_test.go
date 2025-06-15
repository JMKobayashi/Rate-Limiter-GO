package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRateLimiter(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		token   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "IP válido",
			ip:      "192.168.1.1",
			token:   "",
			wantErr: false,
		},
		{
			name:    "Token válido",
			ip:      "",
			token:   "test-token",
			wantErr: false,
		},
		{
			name:    "IP e Token vazios",
			ip:      "",
			token:   "",
			wantErr: true,
			errMsg:  "IP ou Token deve ser fornecido",
		},
		{
			name:    "IP e Token preenchidos",
			ip:      "192.168.1.1",
			token:   "test-token",
			wantErr: true,
			errMsg:  "Apenas IP ou Token deve ser fornecido, não ambos",
		},
		{
			name:    "IP inválido",
			ip:      "invalid-ip",
			token:   "",
			wantErr: true,
			errMsg:  "IP inválido",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter, err := NewRateLimiter(tt.ip, tt.token)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
				assert.Nil(t, limiter)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, limiter)
				assert.Equal(t, tt.ip, limiter.IP)
				assert.Equal(t, tt.token, limiter.Token)
				assert.Equal(t, int64(0), limiter.Requests)
				assert.False(t, limiter.Blocked)
				assert.True(t, limiter.BlockedUntil.IsZero())
			}
		})
	}
}

func TestRateLimiter_IncrementRequests(t *testing.T) {
	limiter, err := NewRateLimiter("192.168.1.1", "")
	require.NoError(t, err)

	// Incrementar requisições
	limiter.IncrementRequests()
	assert.Equal(t, int64(1), limiter.Requests)

	limiter.IncrementRequests()
	assert.Equal(t, int64(2), limiter.Requests)
}

func TestRateLimiter_Block(t *testing.T) {
	limiter, err := NewRateLimiter("192.168.1.1", "")
	require.NoError(t, err)

	// Bloquear por 1 segundo
	blockDuration := time.Duration(1) * time.Second
	limiter.Block(blockDuration)

	// Verificar estado do bloqueio
	assert.True(t, limiter.Blocked)
	assert.True(t, limiter.BlockedUntil.After(time.Now()))
	assert.True(t, limiter.BlockedUntil.Before(time.Now().Add(time.Duration(blockDuration+1)*time.Second)))
}

func TestRateLimiter_IsBlocked(t *testing.T) {
	limiter, err := NewRateLimiter("192.168.1.1", "")
	require.NoError(t, err)

	// Verificar estado inicial
	assert.False(t, limiter.IsBlocked())

	// Bloquear
	limiter.Block(1)
	assert.True(t, limiter.IsBlocked())

	// Esperar o bloqueio expirar
	time.Sleep(time.Second * 2)
	assert.False(t, limiter.IsBlocked())
}

func TestRateLimiter_Reset(t *testing.T) {
	limiter, err := NewRateLimiter("192.168.1.1", "")
	require.NoError(t, err)

	// Configurar estado inicial
	limiter.IncrementRequests()
	limiter.Block(1)

	// Resetar
	limiter.Reset()

	// Verificar estado após reset
	assert.Equal(t, int64(0), limiter.Requests)
	assert.False(t, limiter.Blocked)
	assert.True(t, limiter.BlockedUntil.IsZero())
}

func TestRateLimiter_UpdateLastRequest(t *testing.T) {
	limiter, err := NewRateLimiter("192.168.1.1", "")
	require.NoError(t, err)

	// Atualizar última requisição
	now := time.Now()
	limiter.UpdateLastRequest()

	// Verificar se a última requisição foi atualizada
	assert.True(t, limiter.LastRequest.After(now) || limiter.LastRequest.Equal(now))
}

func TestRateLimiter_ValidateIP(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		wantErr bool
	}{
		{
			name:    "IP válido IPv4",
			ip:      "192.168.1.1",
			wantErr: false,
		},
		{
			name:    "IP válido IPv6",
			ip:      "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			wantErr: false,
		},
		{
			name:    "IP inválido",
			ip:      "invalid-ip",
			wantErr: true,
		},
		{
			name:    "IP vazio",
			ip:      "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateIP(tt.ip)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
