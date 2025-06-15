package entity

import (
	"errors"
	"net"
	"time"
)

var (
	ErrInvalidIP    = errors.New("IP inválido")
	ErrInvalidToken = errors.New("invalid token")
)

// RateLimiter representa um limitador de taxa
type RateLimiter struct {
	IP           string
	Token        string
	Requests     int64
	LastRequest  time.Time
	Blocked      bool
	BlockedUntil time.Time
}

// NewRateLimiter cria um novo limitador de taxa
func NewRateLimiter(ip, token string) (*RateLimiter, error) {
	if ip == "" && token == "" {
		return nil, errors.New("IP ou Token deve ser fornecido")
	}

	if ip != "" && token != "" {
		return nil, errors.New("Apenas IP ou Token deve ser fornecido, não ambos")
	}

	if ip != "" {
		if err := validateIP(ip); err != nil {
			return nil, err
		}
	}

	return &RateLimiter{
		IP:           ip,
		Token:        token,
		Requests:     0,
		LastRequest:  time.Now(),
		Blocked:      false,
		BlockedUntil: time.Time{},
	}, nil
}

// IncrementRequests incrementa o contador de requisições
func (r *RateLimiter) IncrementRequests() {
	r.Requests++
}

// Block bloqueia o limitador por um determinado tempo
func (r *RateLimiter) Block(duration time.Duration) {
	r.Blocked = true
	r.BlockedUntil = time.Now().Add(duration)
}

// IsBlocked verifica se o limitador está bloqueado
func (r *RateLimiter) IsBlocked() bool {
	if !r.Blocked {
		return false
	}

	if r.BlockedUntil.IsZero() {
		return false
	}

	if time.Now().After(r.BlockedUntil) {
		r.Blocked = false
		r.BlockedUntil = time.Time{}
		return false
	}

	return true
}

// Reset reseta o limitador
func (r *RateLimiter) Reset() {
	r.Requests = 0
	r.LastRequest = time.Time{}
	r.Blocked = false
	r.BlockedUntil = time.Time{}
}

// UpdateLastRequest atualiza o timestamp da última requisição
func (r *RateLimiter) UpdateLastRequest() {
	r.LastRequest = time.Now()
}

// validateIP valida se o IP é válido
func validateIP(ip string) error {
	if ip == "" {
		return errors.New("IP inválido")
	}

	if net.ParseIP(ip) == nil {
		return errors.New("IP inválido")
	}

	return nil
}

func (r *RateLimiter) Unblock() {
	r.Blocked = false
	r.BlockedUntil = time.Time{}
}
