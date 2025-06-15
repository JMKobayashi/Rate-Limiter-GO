# Rate Limiter

Este é um rate limiter em Go que pode ser configurado para limitar o número máximo de requisições por segundo com base em um endereço IP específico ou em um token de acesso.

## Funcionalidades

- Limitação de requisições por IP
- Limitação de requisições por token de acesso
- Configuração via variáveis de ambiente
- Persistência em Redis (com possibilidade de trocar por outro mecanismo)
- Middleware para servidor web
- Resposta HTTP 429 quando o limite é excedido

## Requisitos

- Go 1.16 ou superior
- Docker e Docker Compose
- Redis

## Configuração

O rate limiter pode ser configurado através de variáveis de ambiente ou de um arquivo `.env` na pasta raiz:

```env
# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Rate Limiter
RATE_LIMIT_IP=10
RATE_LIMIT_TOKEN=100
BLOCK_DURATION_IP=300
BLOCK_DURATION_TOKEN=600
ENABLE_IP_LIMITER=true
ENABLE_TOKEN_LIMITER=true
```

### Variáveis de Ambiente

- `REDIS_HOST`: Host do Redis (padrão: localhost)
- `REDIS_PORT`: Porta do Redis (padrão: 6379)
- `REDIS_PASSWORD`: Senha do Redis (padrão: vazio)
- `REDIS_DB`: Banco de dados do Redis (padrão: 0)
- `RATE_LIMIT_IP`: Número máximo de requisições por segundo por IP (padrão: 10)
- `RATE_LIMIT_TOKEN`: Número máximo de requisições por segundo por token (padrão: 100)
- `BLOCK_DURATION_IP`: Tempo de bloqueio em segundos para IP (padrão: 300)
- `BLOCK_DURATION_TOKEN`: Tempo de bloqueio em segundos para token (padrão: 600)
- `ENABLE_IP_LIMITER`: Habilita/desabilita limitação por IP (padrão: true)
- `ENABLE_TOKEN_LIMITER`: Habilita/desabilita limitação por token (padrão: true)

## Executando com Docker

1. Clone o repositório:
```bash
git clone https://github.com/seu-usuario/rate-limiter.git
cd rate-limiter
```

2. Execute com Docker Compose:
```bash
docker-compose up -d
```

O servidor web estará disponível na porta 8080.

## Uso

### Limitação por IP

O rate limiter irá limitar o número de requisições por IP de acordo com a configuração `RATE_LIMIT_IP`. Se um IP exceder o limite, ele será bloqueado pelo tempo definido em `BLOCK_DURATION_IP`.

Exemplo com curl:
```bash
curl http://localhost:8080
```

### Limitação por Token

Para usar a limitação por token, inclua o header `API_KEY` na requisição:

```bash
curl -H "API_KEY: seu-token-aqui" http://localhost:8080
```

O rate limiter irá limitar o número de requisições por token de acordo com a configuração `RATE_LIMIT_TOKEN`. Se um token exceder o limite, ele será bloqueado pelo tempo definido em `BLOCK_DURATION_TOKEN`.

### Resposta

Quando o limite é excedido, o rate limiter retorna:

- Código HTTP: 429
- Mensagem:
```json
{
  "error": "you have reached the maximum number of requests or actions allowed within a certain time frame"
}
```

## Teste de Carga

O projeto inclui um teste de carga que pode ser usado para verificar o comportamento do rate limiter sob diferentes condições.

### Executando o Teste de Carga

1. Certifique-se de que o servidor está rodando:
```bash
docker-compose up -d
```

2. Execute o teste de carga:
```bash
go run cmd/loadtest/main.go
```

O teste de carga:
- Usa 50 goroutines concorrentes
- Roda por 30 segundos
- Testa tanto limitação por IP quanto por token
- Coleta métricas de performance

### Resultados do Teste de Carga

#### Teste por IP
```
Total de requisições: 14.151
Requisições bem-sucedidas: 72 (0.51%)
Requisições limitadas: 14.079 (99.49%)
Erros: 0 (0.00%)
Duração média das requisições bem-sucedidas: 1.04s
Requisições por segundo: 471.70
```

#### Teste por Token
```
Total de requisições: 14.250
Requisições bem-sucedidas: 267 (1.87%)
Requisições limitadas: 13.983 (98.13%)
Erros: 0 (0.00%)
Duração média das requisições bem-sucedidas: 245ms
Requisições por segundo: 475.00
```

### Análise dos Resultados

1. **Limitação por IP**:
   - A maioria das requisições (99.49%) foi limitada, o que é esperado devido ao limite de 10 requisições por IP
   - A duração média das requisições bem-sucedidas foi de 1.04s
   - A taxa de requisições foi de 471.70 req/s

2. **Limitação por Token**:
   - A maioria das requisições (98.13%) foi limitada, o que é esperado devido ao limite de 100 requisições por token
   - A duração média das requisições bem-sucedidas foi de 245ms
   - A taxa de requisições foi de 475.00 req/s

3. **Comparação**:
   - O token tem uma taxa de sucesso maior (1.87%) que o IP (0.51%) devido ao limite maior
   - A duração média é menor para token devido ao menor número de requisições limitadas
   - A taxa de requisições por segundo é similar em ambos os casos, mostrando estabilidade do sistema

## Estrutura do Projeto

```
.
├── cmd/
│   ├── server/
│   │   └── main.go
│   └── loadtest/
│       └── main.go
├── internal/
│   ├── entity/
│   │   └── rate_limiter.go
│   ├── limiter/
│   │   └── strategy/
│   │       ├── strategy.go
│   │       ├── redis_repository.go
│   │       └── memory_repository.go
│   ├── middleware/
│   │   └── rate_limiter.go
│   └── usecase/
│       └── rate_limiter_usecase.go
├── pkg/
│   └── config/
│       └── config.go
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
└── README.md
```

## Desenvolvimento

### Rodando Localmente

1. Instale as dependências:
```bash
go mod download
```

2. Configure o Redis:
```bash
docker run -d -p 6379:6379 redis:alpine
```

3. Execute o servidor:
```bash
go run cmd/server/main.go
```

### Testes

Para executar os testes:

```bash
go test ./... -v
```

### Adicionando Novas Estratégias de Persistência

1. Crie um novo arquivo em `internal/limiter/strategy/` (ex: `mongodb_repository.go`)
2. Implemente a interface `RateLimiterRepository`
3. Adicione o novo tipo na factory em `strategy.go`

Exemplo:
```go
// Em strategy.go
const (
    MongoDBRepository RepositoryType = "mongodb"
)

// Na factory
case MongoDBRepository:
    return NewMongoDBRepository(config), nil
```

## Troubleshooting

### Redis não está acessível
- Verifique se o Redis está rodando: `docker ps`
- Verifique as configurações no `.env`
- Teste a conexão: `redis-cli ping`

### Porta 8080 em uso
- Mude a porta no `docker-compose.yml`
- Ou mate o processo usando a porta: `sudo lsof -i :8080`

### Variáveis de ambiente incorretas
- Verifique o arquivo `.env`
- Use `docker-compose config` para validar as configurações
