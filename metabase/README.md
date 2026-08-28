# VagAI Metabase - BI Dashboards

## Visão Geral

Este diretório contém a configuração do **Metabase** para o projeto VagAI, com **30 relatórios** organizados em **7 dashboards**.

## Acesso

| Item | Valor |
|------|-------|
| URL | http://localhost:3001 |
| Email | admin@vagai.com |
| Senha | vagai-admin-2024 |

## Dashboards

### 1 - Pipeline de Vagas (5 relatórios)
- Funil de vagas por status
- Vagas coletadas por dia
- Top empresas com mais vagas
- Vagas por site de origem
- Lag de coleta (publicação vs coleta)

### 2 - Qualidade dos Matches (5 relatórios)
- Distribuição de scores de similaridade
- Score médio por currículo
- Taxa de aplicação
- Keywords mais frequentes
- Matches por tier de plano

### 3 - Receita & Assinaturas (5 relatórios)
- MRR (Monthly Recurring Revenue)
- Conversão trial → paid
- Churn rate mensal
- Assinaturas com cancelamento programado
- Próximos vencimentos (30 dias)

### 4 - Crescimento de Usuários (4 relatórios)
- Novas organizações por semana
- Usuários verificados vs não verificados
- Distribuição de planos
- Membros por organização

### 5 - Saúde do Sistema (4 relatórios)
- Atividade dos agents (crawler/matcher)
- Estatísticas de crawl
- Taxa de sucesso dos schedules
- Duração de execução

### 6 - Insights de Resumes (4 relatórios)
- Forças mais comuns
- Fraquezas mais frequentes
- Sugestões da IA
- Análises por período

### 7 - Performance dos Sites (3 relatórios)
- Sites ativos vs inativos
- Último crawl por site
- Eficiência (vagas/dia)

### 8 - Mercado de Tecnologia (4 relatórios)
- O que estudar primeiro (gap currículo x mercado)
- Top tecnologias demandadas (total histórico)
- Demanda últimos 30 dias
- Tendência mensal por tecnologia

## Arquivos

```
metabase/
├── README.md                    # Este arquivo
└── setup/
    ├── 00-setup-metabase.sh     # Setup automático
    ├── 01-create-cards.sql      # Cards via SQL
    ├── 01-create-views.sql      # Views MySQL (executado automaticamente no primeiro boot)
    ├── 02-create-cards-api.sh   # Script manual para criar cards via API
    ├── 03-init-mysql.sh         # Script de inicialização
    ├── 04-backfill-urls.sql     # Backfill de URLs
    └── 05-create-skill-cards-api.sh  # Dashboard 8: Mercado de Tecnologia
```

## Setup Manual

Se os cards não foram criados automaticamente:

```bash
# 1. Acesse o Metabase e complete o setup wizard
# 2. Execute o script de criação de cards
bash metabase/setup/02-create-cards-api.sh
```

## Views MySQL

As 21 views são criadas automaticamente quando o MySQL inicia:

| View | Dashboard |
|------|-----------|
| `view_vagas_funil` | Pipeline de Vagas |
| `view_vagas_por_dia` | Pipeline de Vagas |
| `view_top_empresas` | Pipeline de Vagas |
| `view_vagas_por_site` | Pipeline de Vagas |
| `view_lag_coleta` | Pipeline de Vagas |
| `view_score_distribuicao` | Qualidade dos Matches |
| `view_score_por_curriculo` | Qualidade dos Matches |
| `view_taxa_aplicacao` | Qualidade dos Matches |
| `view_keywords_frequentes` | Qualidade dos Matches |
| `view_matches_por_plano` | Qualidade dos Matches |
| `view_mrr` | Receita & Assinaturas |
| `view_conversao_trial` | Receita & Assinaturas |
| `view_churn_mensal` | Receita & Assinaturas |
| `view_churn_risco` | Receita & Assinaturas |
| `view_proximos_vencimentos` | Receita & Assinaturas |
| `view_orgs_por_semana` | Crescimento de Usuários |
| `view_usuarios_verificacao` | Crescimento de Usuários |
| `view_distribuicao_planos` | Crescimento de Usuários |
| `view_membros_por_org` | Crescimento de Usuários |
| `view_agentes_por_dia` | Saúde do Sistema |
| `view_crawl_stats` | Saúde do Sistema |
| `view_schedule_sucesso` | Saúde do Sistema |
| `view_duracao_execucao` | Saúde do Sistema |
| `view_forcas_comuns` | Insights de Resumes |
| `view_fraquezas_frequentes` | Insights de Resumes |
| `view_sugestoes_ia` | Insights de Resumes |
| `view_analises_por_periodo` | Insights de Resumes |
| `view_status_sites` | Performance dos Sites |
| `view_ultimo_crawl` | Performance dos Sites |
| `view_eficiencia_sites` | Performance dos Sites |

### Mercado de Tecnologia (aplicação manual)

As views abaixo usam a tabela `tech_keywords` (regex por tecnologia — adicione
linhas para ampliar a cobertura) e **não** são criadas pelo init automático:

```bash
docker cp metabase/setup/04-create-skill-views.sql vagai-mysql:/tmp/
docker exec vagai-mysql sh -c 'mysql -u vagai -pvagai vagai < /tmp/04-create-skill-views.sql'
bash metabase/setup/05-create-skill-cards-api.sh   # cards + dashboard 8
```

| View | Descrição |
|------|-----------|
| `view_tech_demandas` | Vagas citando cada tecnologia (% do total, dedup por URL) |
| `view_tech_demandas_mes` | Tendência mensal de demanda |
| `view_tech_demandas_30d` | Demanda nos últimos 30 dias |
| `view_skill_gap` | Gap entre mercado e currículo mais recente + prioridade de estudo |

Notas: matching por regex com fronteira de palavra (evita falso positivo tipo
"responsável" → SP); Java usa negative lookahead para não casar com JavaScript;
vagas duplicadas entre organizações contam uma vez (`COUNT(DISTINCT url)`).

## Comandos Úteis

```bash
# Iniciar apenas o Metabase
docker compose up metabase -d

# Ver logs do Metabase
docker logs -f vagai-metabase

# Acessar MySQL para verificar views
docker exec -it vagai-mysql mysql -u vagai -pvagai vagai

# Listar views
docker exec -it vagai-mysql mysql -u vagai -pvagai vagai -e "SHOW FULL TABLES WHERE Table_type = 'VIEW';"
```

## Notas Técnicas

- Preços estão em **centavos** (dividir por 100 para R$)
- Colunas JSON usam `JSON_TABLE()` do MySQL 8
- Filtre `deleted_at IS NULL` para registros ativos
- Metabase suporta JSON nativamente
