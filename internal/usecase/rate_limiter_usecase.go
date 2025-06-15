package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/JMKobayashi/Rate-Limiter-GO/internal/entity"
	"github.com/JMKobayashi/Rate-Limiter-GO/internal/repository"
)

// RateLimiterUseCaseInterface define a interface para o caso de uso do limitador de taxa.
type RateLimiterUseCaseInterface interface {
	IsAllowed(ctx context.Context, identifier string, isToken bool) (bool, error)
}

// RateLimiterUseCase implementa a lógica do caso de uso do limitador de taxa.
type RateLimiterUseCase struct {
	repository     repository.RateLimiterRepository // Repositório para persistência do estado do limitador.
	rateLimitIP    int                              // Limite de requisições por IP.
	rateLimitToken int                              // Limite de requisições por Token.
	blockDuration  int                              // Duração do bloqueio em segundos.
}

// NewRateLimiterUseCase cria uma nova instância de RateLimiterUseCase.
// Parâmetros:
//
//	repository: Uma implementação da interface RateLimiterRepository para persistir os dados.
//	rateLimitIP: O número máximo de requisições permitidas para um IP antes de ser bloqueado.
//	rateLimitToken: O número máximo de requisições permitidas para um Token antes de ser bloqueado.
//	blockDuration: A duração do bloqueio em segundos.
func NewRateLimiterUseCase(repository repository.RateLimiterRepository, rateLimitIP, rateLimitToken, blockDuration int) RateLimiterUseCaseInterface {
	return &RateLimiterUseCase{
		repository:     repository,
		rateLimitIP:    rateLimitIP,
		rateLimitToken: rateLimitToken,
		blockDuration:  blockDuration,
	}
}

// IsAllowed verifica se uma requisição é permitida com base no identificador (IP ou Token) e nos limites de taxa definidos.
// Retorna 'true' se a requisição for permitida, 'false' caso contrário, e um erro se houver falha na operação.
func (uc *RateLimiterUseCase) IsAllowed(ctx context.Context, identifier string, isToken bool) (bool, error) {
	// Constrói uma chave única para o identificador (IP ou Token) a ser usada no repositório.
	key := fmt.Sprintf("rate_limit:%s", identifier)

	// Tenta obter o estado atual do limitador do repositório.
	// Isso é fundamental para manter o histórico de requisições e o estado de bloqueio.
	limiter, err := uc.repository.Get(ctx, key)
	if err != nil {
		// Se houver um erro ao buscar o limitador, retorna imediatamente a falha.
		return false, fmt.Errorf("falha ao obter limitador do repositório para '%s': %w", identifier, err)
	}

	// Se o limitador não foi encontrado no repositório (primeira requisição para este identificador),
	// cria uma nova instância de RateLimiter.
	if limiter == nil {
		limiter = entity.NewRateLimiter(identifier, "")
		if isToken {
			// Se a requisição é baseada em token, associa o token ao limitador.
			limiter.Token = identifier
		}
	}

	// Verifica se o limitador já está bloqueado devido a requisições anteriores.
	// Se estiver bloqueado, a requisição atual não é permitida.
	if limiter.IsBlocked() {
		// Não é necessário salvar novamente aqui, pois o estado de bloqueio já está persistido
		// ou o tempo de bloqueio ainda não expirou.
		return false, nil // Requisição negada: identificador está bloqueado.
	}

	// Determina qual limite de taxa aplicar: para IP ou para Token.
	currentLimit := uc.rateLimitIP
	if isToken {
		currentLimit = uc.rateLimitToken
	}

	// Incrementa o contador de requisições para o identificador.
	limiter.Requests++
	// Atualiza o timestamp da última requisição para este identificador.
	limiter.LastRequest = time.Now()

	var allowed bool // Variável para armazenar o resultado da verificação de permissão.

	// Verifica se o número de requisições excede o limite permitido.
	if limiter.Requests > int64(currentLimit) {
		// Se o limite foi excedido, bloqueia o identificador.
		// O método Block() do entity.RateLimiter deve calcular e definir o `BlockedUntil`.
		limiter.Block(time.Duration(uc.blockDuration) * time.Second)
		allowed = false // A requisição não é permitida.
	} else {
		allowed = true // A requisição é permitida.
	}

	// Salva o estado atualizado do limitador no repositório.
	// Este passo é CRÍTICO para persistir:
	// 1. O contador de `Requests` (para que ele continue acumulando em chamadas futuras).
	// 2. O `LastRequest` (para lógica de tempo).
	// 3. O estado de `BlockedUntil` (para que o bloqueio seja efetivo e persistente).
	saveErr := uc.repository.Save(ctx, limiter)
	if saveErr != nil {
		// Se houver um erro ao salvar o estado, retorna 'false' e o erro.
		// Isso é importante porque o estado não foi persistido, o que pode levar
		// a inconsistências futuras.
		return false, fmt.Errorf("falha ao salvar o estado do limitador para '%s': %w", identifier, saveErr)
	}

	// Retorna o resultado da verificação de permissão e nenhum erro,
	// indicando que a operação foi concluída e o estado persistido com sucesso.
	return allowed, nil
}
