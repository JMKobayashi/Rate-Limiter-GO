package strategy

import "errors"

var (
	// ErrInvalidRepositoryType é retornado quando o tipo de repositório é inválido
	ErrInvalidRepositoryType = errors.New("tipo de repositório inválido")
)
