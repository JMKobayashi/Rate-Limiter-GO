package strategy

// RepositoryType define o tipo de repositório a ser usado
type RepositoryType string

const (
	// RedisRepository é o tipo para usar o Redis como persistência
	RedisRepository RepositoryType = "redis"
	// MemoryRepository é o tipo para usar memória como persistência
	MemoryRepository RepositoryType = "memory"
)
