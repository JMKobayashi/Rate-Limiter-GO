package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/JMKobayashi/Rate-Limiter-GO/internal/entity"
	"github.com/JMKobayashi/Rate-Limiter-GO/internal/limiter/strategy"
)

type RateLimiterUseCaseInterface interface {
	IsAllowed(ctx context.Context, identifier string, isToken bool) (bool, error)
}

type RateLimiterUseCase struct {
	repository         strategy.RateLimiterRepository
	rateLimitIP        int
	rateLimitToken     int
	blockDurationIP    int
	blockDurationToken int
	enableIPLimiter    bool
	enableTokenLimiter bool
}

func NewRateLimiterUseCase(
	repository strategy.RateLimiterRepository,
	rateLimitIP,
	rateLimitToken,
	blockDurationIP,
	blockDurationToken int,
	enableIPLimiter,
	enableTokenLimiter bool,
) RateLimiterUseCaseInterface {
	return &RateLimiterUseCase{
		repository:         repository,
		rateLimitIP:        rateLimitIP,
		rateLimitToken:     rateLimitToken,
		blockDurationIP:    blockDurationIP,
		blockDurationToken: blockDurationToken,
		enableIPLimiter:    enableIPLimiter,
		enableTokenLimiter: enableTokenLimiter,
	}
}

func (uc *RateLimiterUseCase) IsAllowed(ctx context.Context, identifier string, isToken bool) (bool, error) {
	// Se a limitação por token estiver desabilitada e for um token, permite a requisição
	if isToken && !uc.enableTokenLimiter {
		return true, nil
	}

	// Se a limitação por IP estiver desabilitada e não for um token, permite a requisição
	if !isToken && !uc.enableIPLimiter {
		return true, nil
	}

	// Define a chave baseada no tipo (IP ou token)
	key := fmt.Sprintf("rate_limiter:%s:%s", map[bool]string{true: "token", false: "ip"}[isToken], identifier)

	limiter, err := uc.repository.Get(ctx, key)
	if err != nil {
		return false, err
	}

	// Se não existe um limiter, cria um novo
	if limiter == nil {
		if isToken {
			limiter = &entity.RateLimiter{
				Token:        identifier,
				Requests:     0,
				LastRequest:  time.Now(),
				Blocked:      false,
				BlockedUntil: time.Time{},
			}
		} else {
			limiter, err = entity.NewRateLimiter(identifier, "")
			if err != nil {
				return false, err
			}
		}
	}

	// Verifica se está bloqueado
	if limiter.IsBlocked() {
		return false, nil
	}

	// Define os limites baseados no tipo
	limit := uc.rateLimitIP
	blockDuration := uc.blockDurationIP
	if isToken {
		limit = uc.rateLimitToken
		blockDuration = uc.blockDurationToken
	}

	// Incrementa o contador de requisições
	limiter.Requests++
	limiter.LastRequest = time.Now()

	// Verifica se excedeu o limite
	if limiter.Requests > int64(limit) {
		limiter.Block(time.Duration(blockDuration) * time.Second)
		err = uc.repository.Save(ctx, limiter)
		if err != nil {
			return false, err
		}
		return false, nil
	}

	// Salva o estado atual
	err = uc.repository.Save(ctx, limiter)
	if err != nil {
		return false, err
	}

	return true, nil
}
