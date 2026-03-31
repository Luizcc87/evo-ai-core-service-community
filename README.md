# EvoAI Core Service

Microserviço de IA integrado ao Evolution que fornece funcionalidades avançadas de inteligência artificial, incluindo ferramentas personalizadas, agentes inteligentes e gerenciamento de chaves API.

## 🚀 Quick Start

```bash
# Clone o repositório
git clone <repository-url>
cd evo-ai-core-service

# Configure as variáveis de ambiente
cp .env.template .env
# Edite o .env com suas configurações

# Instale dependências e inicie o serviço
make start
```

## 📋 Requisitos

- **Go** 1.24+
- **PostgreSQL** 12+
- **Evolution** em execução
- **golang-migrate** (instalado automaticamente via `make install`)
- **swag** (instalado automaticamente via `make install`)

## ⚙️ Configuração

### 1. Banco de Dados

O serviço utiliza o mesmo banco de dados do Evolution para integração perfeita:

```bash
# No .env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=root
DB_NAME=evo_chat  # Mesmo banco do Evolution
DB_SSLMODE=disable
```

### 2. Integração Evolution

```bash
# URL base do Evolution (sem barra no final)
EVOLUTION_BASE_URL=http://localhost:3000

# JWT Secret deve coincidir com o SECRET_KEY_BASE do Evolution
JWT_SECRET_KEY=your_evolution_secret_key_base_here
```

### 3. Porta do Serviço

```bash
PORT=5555
```

## 🛠️ Comandos Disponíveis

### Quick Start
```bash
make start          # Instala tudo e inicia em modo desenvolvimento
make help           # Mostra todos os comandos disponíveis
```

### Desenvolvimento
```bash
make dev            # Inicia em modo desenvolvimento com race detection
make run            # Inicia em modo produção
make build          # Compila a aplicação
go run cmd/api/main.go -dev  # Execução direta com carregamento do .env
```

### Banco de Dados
```bash
make db-setup       # Cria banco e executa migrações
make db-reset       # Reseta banco completamente (com confirmação)
make migrate-up     # Executa migrações pendentes
make migrate-down   # Desfaz última migração
make migrate-create NAME=nome_da_migracao  # Cria nova migração
```

### Utilitários
```bash
make install        # Instala dependências e ferramentas
make swag           # Gera documentação Swagger
make tidy           # Organiza módulos Go
make clean          # Limpa arquivos gerados
```

## 🏗️ Arquitetura

### Estrutura do Projeto

```
evo-ai-core-service/
├── cmd/api/                    # Ponto de entrada e anotações Swagger
├── internal/                   # Código privado da aplicação
│   ├── config/                # Gerenciamento de configuração
│   ├── middleware/            # Middleware HTTP (auth, CORS, etc.)
│   ├── utils/                 # Pacotes utilitários (context, JWT, string, etc.)
│   └── infra/postgres/        # Conexão de banco e infraestrutura
├── pkg/                       # Módulos públicos seguindo domain-driven design
│   ├── agent/                # Gerenciamento de agentes IA com integração A2A
│   ├── api_key/              # Gerenciamento de chaves API para serviços IA
│   ├── custom_tool/          # Definições de ferramentas HTTP para agentes
│   ├── custom_mcp_server/    # Gerenciamento de servidores Model Context Protocol
│   ├── folder/               # Gerenciamento de organização e workspace
│   ├── folder_share/         # Compartilhamento de pastas e permissões
│   └── mcp_server/           # Registro e gerenciamento de servidores MCP
├── migrations/               # Migrações de schema do banco de dados
└── docs/                     # Documentação Swagger gerada
```

### Padrão de Módulos Domain-Driven

Cada domínio de negócio segue estrutura consistente:

- **Handler**: Manipulação de requisições HTTP, validação, formatação de resposta
- **Service**: Lógica de negócio, chamadas a serviços externos, validação
- **Repository**: Operações de banco de dados com GORM
- **Model**: Estruturas de dados com DTOs de requisição/resposta
- **Module**: Injeção de dependência e inicialização do módulo

## 🔒 Autenticação EvoAuth

O serviço utiliza **Bearer Token** para autenticação via tokens do EvoAuth, garantindo integração perfeita com o sistema existente.

### Como Funciona

1. **Token Validation**: Valida tokens via endpoint `/api/v1/me` do EvoAuth
2. **Bearer Token**: Utiliza apenas o header Authorization com Bearer token
3. **Account Context**: Extrai automaticamente informações do usuário e contas ativas
4. **Request Context**: Injeta user_id, email, name e dados completos no contexto da requisição

### Header Suportado

```http
# Formato Bearer Token (único suportado)
Authorization: Bearer sua_access_token_aqui
```

### Exemplo de Requisição

```bash
# Com Authorization Bearer
curl -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
     http://localhost:5555/api/v1/agents
```

### Middleware de Autenticação

O middleware (`internal/middleware/evo_auth.go`) implementa:

- ✅ **Validação automática** via API do EvoAuth
- ✅ **Extração de contexto** (user_id, email, name, user completo, accounts)
- ✅ **Bearer Token** único formato suportado
- ✅ **Logs detalhados** para debugging

### Estrutura de Resposta do EvoAuth

```json
{
  "user": {
    "id": "fded2c11-5023-4341-80fa-9e1143a8b944",
    "name": "Admin User",
    "email": "admin@evo-auth-service.com",
    "display_name": null,
    "availability": "online",
    "mfa_enabled": false,
    "confirmed": true
  },
  "accounts": [
    {
      "id": "14f7da56-d6b5-4441-bc6b-ee76296b1e99",
      "name": "Default Account",
      "status": "active",
      "locale": "en",
      "created_at": "2025-08-21T17:11:20.459Z",
      "updated_at": "2025-08-21T17:11:20.459Z"
    }
  ]
}
```

### Configuração Necessária

```env
# .env
EVOLUTION_BASE_URL=http://localhost:3000  # URL base do Evolution (sem barra no final)
JWT_SECRET_KEY=your_evolution_secret_key_base  # Deve coincidir com Evolution
```

## 📚 API Endpoints

### Agents (Agentes IA)
- `GET /api/v1/agents` - Listar agentes (suporte a skip/limit e page/pageSize)
- `POST /api/v1/agents` - Criar agente
- `PUT /api/v1/agents/:id` - Atualizar agente
- `DELETE /api/v1/agents/:id` - Deletar agente
- `GET /api/v1/agents/:id` - Obter agente por ID

### Custom Tools (Ferramentas Personalizadas)
- `GET /api/v1/custom-tools` - Listar ferramentas
- `POST /api/v1/custom-tools` - Criar ferramenta
- `PUT /api/v1/custom-tools/:id` - Atualizar ferramenta
- `DELETE /api/v1/custom-tools/:id` - Deletar ferramenta
- `GET /api/v1/custom-tools/:id/test` - Testar ferramenta

### API Keys (Chaves de API)
- `GET /api/v1/apikeys` - Listar chaves
- `POST /api/v1/apikeys` - Criar chave
- `PUT /api/v1/apikeys/:id` - Atualizar chave
- `DELETE /api/v1/apikeys/:id` - Deletar chave

### Folders (Pastas)
- `GET /api/v1/folders` - Listar pastas
- `GET /api/v1/folders/accessible-folders` - Listar pastas acessíveis (compatibilidade frontend)
- `POST /api/v1/folders` - Criar pasta
- `PUT /api/v1/folders/:id` - Atualizar pasta
- `DELETE /api/v1/folders/:id` - Deletar pasta

### System
- `GET /health` - Health check
- `GET /ready` - Readiness check
- `GET /swagger/*` - Documentação Swagger

## 📖 Documentação

### Swagger UI
Após iniciar o serviço, acesse:
- **Local**: http://localhost:5555/swagger/index.html
- **Produção**: https://your-domain.com:5555/swagger/index.html

### Arquivos de Documentação

- [`CLAUDE.md`](./CLAUDE.md) - Guia detalhado para desenvolvimento
- [`docs/MICROSERVICE_INTEGRATION.md`](./docs/MICROSERVICE_INTEGRATION.md) - Guia de integração para futuros microserviços

## 🗃️ Banco de Dados

### Arquitetura de Banco

- **Banco Compartilhado**: Utiliza o mesmo PostgreSQL do Evolution
- **Prefixo de Tabelas**: Todas as tabelas prefixadas com `evo_core_` para evitar conflitos
- **Chaves Estrangeiras**: Referencia `accounts(id)` do Evolution para multi-tenancy
- **Segurança de Migração**: Backup da tabela `schema_migrations` do Rails antes de executar migrações Go

### Tabelas Principais

- `evo_core_agents` - Agentes IA com diferentes tipos
- `evo_core_custom_tools` - Ferramentas personalizadas HTTP
- `evo_core_api_keys` - Armazenamento criptografado de credenciais
- `evo_core_folders` - Organização de workspace
- `evo_core_folder_shares` - Compartilhamento de pastas
- `evo_core_mcp_servers` - Servidores Model Context Protocol

## 🧬 Integração com Frontend

### CORS Configuration

O serviço suporta CORS para múltiplas origens:

```go
// Headers CORS para OAuth Bearer Token
"Access-Control-Allow-Headers": "Authorization, X-Account-Id, X-User-Id, Content-Type"
"Access-Control-Expose-Headers": "Authorization, X-Account-Id, X-User-Id"
```

### Frontend Integration

Exemplo de configuração no frontend (Vue.js/Axios):

```javascript
// agentApi.ts
const agentApi = axios.create({
  baseURL: `${import.meta.env.VITE_EVO_AI_CORE_SERVICE_URL}/api/v1`,
});

agentApi.interceptors.request.use(config => {
  const authHeaders = JSON.parse(localStorage.getItem('authHeaders'));
  config.headers = {
    ...config.headers,
    'Authorization': `Bearer ${oauthToken}`,
  };
  return config;
});
```

## 🔧 Desenvolvimento

### Adicionando um Novo Módulo

1. Crie a estrutura em `pkg/novo_modulo/`
2. Implemente Handler, Service, Repository e Model seguindo o padrão existente
3. Crie migrações em `migrations/`
4. Registre rotas no `main.go`
5. Atualize documentação Swagger com anotações

### Exemplo de Estrutura de Módulo

```go
// pkg/novo_modulo/module.go
type Module struct {
    Handler handler.NovoModuloHandler
    Service service.NovoModuloService  
    Repo    repository.NovoModuloRepository
}

func New(db *gorm.DB) *Module {
    repo := repository.NewNovoModuloRepository(db)
    svc := service.NewNovoModuloService(repo)
    handler := handler.NewNovoModuloHandler(svc)

    return &Module{
        Handler: handler,
        Service: svc,
        Repo:    repo,
    }
}
```

### Padrões de Validação

- Use `NullableUUID` para campos UUID opcionais que podem vir como string vazia
- Implemente validação de entrada com binding tags do Gin

## 🚢 Deploy

### Docker (Recomendado)

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o bin/app cmd/api/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/bin/app .
EXPOSE 5555
CMD ["./app"]
```

### Build Direto

```bash
# Build
make build

# Execute
./bin/app
```

## 📊 Monitoramento

### Health Checks

```bash
# Verificar se está funcionando
curl http://localhost:5555/health

# Verificar se está pronto para receber tráfego  
curl http://localhost:5555/ready
```

### Logs Estruturados

O serviço utiliza logs estruturados com Gin Logger:

- Health checks são ignorados nos logs para reduzir ruído
- Logs detalhados de autenticação em desenvolvimento
- Configuração personalizada para diferentes ambientes

## 🤝 Contribuição

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/AmazingFeature`)
3. Commit suas mudanças (`git commit -m 'Add some AmazingFeature'`)
4. Push para a branch (`git push origin feature/AmazingFeature`)
5. Abra um Pull Request

### Padrões de Código

- Siga os padrões Go com `gofmt`
- Use anotações Swagger para documentar endpoints
- Implemente testes unitários quando possível
- Mantenha consistência com a arquitetura existente

## 📄 Licença

Este projeto está sob a licença MIT. Veja o arquivo `LICENSE` para mais detalhes.

## 🐛 Reportar Problemas

Encontrou um bug ou tem uma sugestão? Abra uma issue no GitHub:

1. Descreva o problema detalhadamente
2. Inclua passos para reproduzir
3. Informe versão do Go e sistema operacional
4. Anexe logs se necessário

## 🆘 Suporte

- **Documentação**: Veja os arquivos em `/docs` e `CLAUDE.md`
- **Issues**: Use o GitHub Issues para bugs e solicitações
- **Discussões**: Use GitHub Discussions para dúvidas gerais

## 🗺️ Roadmap

- [ ] Integração com mais provedores de IA
- [ ] Sistema de plugins extensível
- [ ] Métricas e monitoramento avançado (OpenTelemetry)
- [ ] Cache distribuído para performance
- [ ] Testes automatizados abrangentes
- [ ] CI/CD pipeline completa

---

Desenvolvido com ❤️ para a comunidade Evolution