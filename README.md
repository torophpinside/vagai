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
job-hunter/
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

### Configurar sites monitorados

Edite `vagai-cli/configs/sites.yaml`:

```yaml
sites:
  - name: remoteok
    url: https://remoteok.com
    enabled: true
  - name: weworkremotely
    url: https://weworkremotely.com
    enabled: true
```

### Adicionar currículos

Envie via API ou coloque os arquivos em `vagai-api/uploads/` (PDF ou DOCX).

### Rodar busca manual

```bash
cd vagai-cli
go run main.go crawl --site remoteok
go run main.go match --resume curriculum.pdf
```

## Funcionalidades

- **Crawling multi-site** — RemoteOK, WeWorkRemotely, LinkedIn
- **Matching com IA** — LM Studio para análise semântica de compatibilidade
- **Fallback tradicional** — similaridade textual quando IA não está disponível
- **Dashboard** — gráficos de vagas por tecnologia, empresa, tempo
- **Agendamento** — coleta automática via cron
- **Multi-tenancy** — isolamento de dados por usuário/empresa
- **Planos** — Free, Pro, Enterprise com limites diferenciados

## Status

9 Sprints concluídos. O sistema está funcional para busca, matching e visualização via dashboard.

## Docs

Especificação completa: [`docs/SPECS.md`](docs/SPECS.md)
