# VagAI

Sistema de busca automatizada de vagas com matching inteligente baseado em currículos.

## O que faz

O VagAI monitora sites de emprego, extrai vagas e as compara com seus currículos usando IA. O resultado é um painel com matches ranqueados por compatibilidade.

## Arquitetura

Três agentes especializados:

- **Crawler Agent** (`vagai-cli/`) — navega sites, extrai links e descrições de vagas
- **Matcher Agent** (`vagai-cli/`) — compara vagas com currículos usando similaridade textual e LM Studio
- **API Agent** (`vagai-api/`) — expõe dados via REST, gerencia autenticação e multi-tenancy

## Tech Stack

| Componente | Tecnologia |
|------------|-----------|
| API | Go 1.24 + Gin + GORM |
| CLI/Agents | Go 1.21 + Cobra + goquery |
| Frontend | Vue 3 + Vite + TailwindCSS |
| Banco | MySQL 8 + Redis 7 |
| IA | LM Studio (local) |
| Infra | Docker + docker-compose |

## Estrutura

```
vagai/
├── vagai-cli/          # CLI Go com agentes
│   ├── cmd/            # Comandos: crawl, match, sites, schedule
│   └── internal/agents/  # Crawler, Matcher, Registry
│
├── vagai-api/          # API REST
│   ├── internal/
│   │   ├── handlers/   # Endpoints HTTP
│   │   ├── services/   # Lógica de negócio + IA
│   │   └── repository/ # Acesso ao banco
│   └── uploads/        # Currículos (PDF/DOCX)
│
├── vagai-web/          # Frontend Vue 3
│   └── src/
│       ├── pages/      # Dashboard, Vagas, Matches
│       └── components/ # Gráficos, tabelas, filtros
│
├── docker-compose.yml  # Orquestração completa
└── docs/SPECS.md       # Especificação detalhada
```

## Começar

### Pré-requisitos

- Docker + docker-compose
- LM Studio rodando localmente (para matching com IA)
- Go 1.24+ (se for rodar sem Docker)

### Subir o ambiente

```bash
docker-compose up -d
```

Isso sobe MySQL, Redis, API, Web e o agendador de tarefas.

### Adicionar sites

Pela interface Web ou via API, adicione a URL do site desejado — o VagAI descobre automaticamente os seletores CSS com IA.

### Adicionar currículos

Envie via API/Web em PDF, DOCX ou TXT.

### Rodar busca manual

```bash
cd vagai-cli
go run main.go crawl --site remoteok
go run main.go match
```

## Funcionalidades

- **Crawling multi-site** — RemoteOK, WeWorkRemotely, LinkedIn
- **Matching com IA** — LM Studio para análise semântica de compatibilidade
- **Fallback tradicional** — similaridade textual quando IA não está disponível
- **Dashboard** — gráficos de vagas por tecnologia, empresa, tempo
- **Agendamento** — coleta automática via cron
- **Multi-tenancy** — isolamento de dados por usuário/empresa
- **Planos** — Free e Pro com limites diferenciados

## Status

Sistema funcional para busca, matching, autenticação multi-tenancy e visualização via dashboard.

## Testes

### E2E (API)

Testes de integração ponta a ponta em `vagai-api/internal/tests/e2e/`.

**Pré-requisito:** Docker rodando.

**Rodar todos:**
```bash
cd vagai-api
go test -v -count=1 -timeout 10m ./internal/tests/e2e/
```

**Rodar um específico:**
```bash
go test -v -count=1 -timeout 10m -run TestRegister_Success ./internal/tests/e2e/
```

**Como funciona:**
- `testcontainers-go` sobe um MySQL 8.0 limpo antes dos testes e o destrói ao final
- `TestMain` executa migrações (`AutoMigrate`), seed dos planos, e inicia um servidor HTTP (`httptest`) com as mesmas rotas da API
- Cada teste cria usuário real (org + subscription Free), obtém JWT e faz chamadas HTTP
- Helpers: `registerUser(t, name, email, password, org)`, `doRequest(t, method, path, body, token)`, `parseBody(t, resp, &dest)`
- 31 testes cobrindo: auth (registro, login, duplicata, /me, health), vagas (CRUD, duplicidade, paginação, marcar match), sites (CRUD, toggle ativo), matches (listagem, filtros)

## Próximos passos

- **UI/UX** — refinamento visual da plataforma (animações, responsividade, temas)
- **Pagamentos** — integração com gateway (Stripe) para upgrade de planos
- **Organização** — gestão de membros (convite, papéis), auditoria de ações, logs de atividade
- **Notificações** — alertas por email quando novos matches de alto score forem encontrados
- **API Keys** — gerenciamento de chaves de API para integração externa
- **Webhooks** — disparar eventos para serviços externos (novas vagas, matches, etc.)
- **Testes E2E (frontend)** — Playwright para fluxos críticos (login, registro, navegação)

## Docs

Especificação completa: [`docs/SPECS.md`](docs/SPECS.md)
