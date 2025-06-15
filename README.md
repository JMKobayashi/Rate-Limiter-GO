# Rate Limiter em Go

Este projeto implementa um rate limiter em Go que pode ser configurado para limitar requisições baseado em IP ou token de acesso.

## Funcionalidades

- Limitação de requisições por IP
- Limitação de requisições por token de acesso
- Configuração via variáveis de ambiente
- Persistência em Redis
- Estratégia de armazenamento extensível
- Middleware para servidor web

## Requisitos

- Go 1.16+
- Docker e Docker Compose
- Redis

## Configuração

As configurações podem ser feitas através de variáveis de ambiente ou arquivo `.env`:

```env
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Configurações do Rate Limiter
RATE_LIMIT_IP=10
RATE_LIMIT_TOKEN=100
BLOCK_DURATION=300  # em segundos
```

## Executando o Projeto

1. Clone o repositório
2. Execute `docker-compose up -d` para iniciar o Redis
3. Execute `go run main.go` para iniciar o servidor

O servidor estará disponível na porta 8080.

## Testes

Execute os testes com:

```bash
go test ./...
```

## Estrutura do Projeto

```
.
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── limiter/
│   │   ├── limiter.go
│   │   └── strategy/
│   │       ├── redis.go
│   │       └── interface.go
│   └── middleware/
│       └── rate_limiter.go
├── pkg/
│   └── config/
│       └── config.go
├── docker-compose.yml
├── Dockerfile
└── README.md
```
