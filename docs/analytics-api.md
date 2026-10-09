# API de Analytics (Data Analytics & Database Insights)

Documentação dos endpoints de analytics do VagAI (`/api/v1/analytics/*`). Especificação funcional em `specs/010-data-analytics/spec.md`; contrato completo de payloads em `specs/010-data-analytics/contracts/api_contract.md`.

## Base Path
`/api/v1/analytics/`

## Autenticação e Controle de Acesso (FR-020)

- Todos os endpoints exigem JWT válido (`Authorization: Bearer <token>`).
- **Org admin** (`owner`/`admin`): acesso somente aos dados da própria organização. Passar `organization_id` de outra org (ou usar `/org/:id`) retorna `403`.
- **Super admin** (`SUPER_ADMIN_EMAILS`): acesso a qualquer organização e agregado global com `organization_id=0` em todos os endpoints (exceto `/org/:id`, que exige uma organização real).

## Erros

| Status | Tipo | Mensagem |
|--------|------|----------|
| 401 | Unauthorized | Token ausente/inválido |
| 403 | Forbidden | Acesso negado à organização solicitada |
| 404 | Not Found | Organização não encontrada (somente `/org/:id`) |
| 422 | Unprocessable Entity | Parâmetros de paginação inválidos; `period`/`min_frequency`/`organization_id` fora do formato |
| 500 | Internal Error | Analytics service unavailable |

## Endpoints

| Método | Rota | US / FR | Descrição |
|--------|------|---------|-----------|
| GET | `/matching-distribution` | US1 (FR-010) | Distribuição de scores, média/mediana, conversão match→candidatura |
| GET | `/resume-insights` | US2 (FR-011) | Insights de qualidade de currículos |
| GET | `/job-trends` | US3 (FR-012) | Tendências do mercado de vagas |
| GET | `/org/:organization_id` | US4 (FR-013) | Métricas agregadas por organização |
| GET | `/interview-prep` | US5 (FR-014) | Análise das preparações de entrevista |
| GET | `/ai-quality` | US6 (FR-015) | Qualidade das análises de IA |
| GET | `/time-to-hire` | US7 (FR-016) | Tempos de pipeline (match, análise, candidatura) |
| GET | `/skills-gap` | US8 (FR-017) | Gaps de skills candidato vs. mercado |
| GET | `/funnel` | US9 (FR-018) | Funil de usuários 4 estágios |
| GET | `/geographic` | US10 (FR-019) | Insights geográficos |

## Parâmetros comuns

| Parâmetro | Tipo | Obrigatório | Padrão | Descrição |
|-----------|------|-------------|--------|-----------|
| `organization_id` | uint | não* | org do token | Escopo do tenant; `0` = agregado global (super admin) |
| `page` | uint | não | 1 | Página (≥ 1; inválido → 422) |
| `page_size` | uint | não | 50 | Itens por página (1–100; inválido → 422) |

\* Para org admins o escopo padrão é a própria org do token; `organization_id` de outra org → 403.

### Parâmetros específicos

| Endpoint | Parâmetro | Valores |
|----------|-----------|---------|
| `/job-trends` | `period` | `week`, `month`, `quarter`, `year` (padrão `month`; inválido → 422) |
| `/skills-gap` | `min_frequency` | uint ≥ 1 (padrão 5; inválido → 422) |

## Estado vazio

Quando a organização não possui dados relevantes, o endpoint responde `200` com o shape padrão zerado e um bloco `empty_state`:

```json
{
  "empty_state": {
    "empty": true,
    "message": "Nenhum match registrado ainda",
    "guidance": "Execute o matching de vagas com seus currículos para começar a ver a distribuição de scores e taxas de conversão."
  }
}
```

## Exemplos

### GET /api/v1/analytics/matching-distribution

```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/analytics/matching-distribution?organization_id=17
```

```json
{
  "buckets": [
    {"bucket": "0.0-0.2", "count": 0, "percentage": 0},
    {"bucket": "0.2-0.4", "count": 1, "percentage": 16.7}
  ],
  "summary": {
    "average_score": 0.62,
    "median_score": 0.61,
    "total_matches": 6,
    "conversion_rate": 33.3
  },
  "top_job_titles": [
    {"title": "Backend Developer", "total_matches": 3, "applied": 1, "success_rate": 33.3}
  ],
  "pagination": {"page": 1, "page_size": 50, "total_pages": 1, "total_items": 6}
}
```

### GET /api/v1/analytics/org/:id

```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/analytics/org/17
```

```json
{
  "organization": {"id": 17, "name": "Analytics Corp", "plan": "free", "subscription_status": "active", "trial_to_paid": true},
  "metrics": {
    "jobs_posted": 5,
    "resumes_uploaded": 3,
    "matches_made": 6,
    "applications_made": 2,
    "conversion_rate": 33.3,
    "feature_utilization": {"matching": true, "ai_analysis": true, "interview_prep": true, "skill_gap": true}
  },
  "timing": {"average_time_to_match_days": 0, "average_time_to_analysis_days": 0}
}
```

### GET /api/v1/analytics/funnel

```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/analytics/funnel?organization_id=17
```

```json
{
  "funnel_stages": [
    {"stage": "signups", "count": 1, "conversion_rate": 100},
    {"stage": "uploaded_resumes", "count": 3, "conversion_rate": 300},
    {"stage": "matched_with_jobs", "count": 6, "conversion_rate": 600},
    {"stage": "applied_to_jobs", "count": 2, "conversion_rate": 200}
  ],
  "drop_off_points": ["applied_to_jobs", "matched_with_jobs"],
  "total_users": 1
}
```

## Comandos CLI equivalentes

Os mesmos relatórios estão disponíveis via CLI (`vagai analytics <subcomando>`), com auto-detecção de organização ou `--org <id>`:

| CLI | Endpoint equivalente |
|-----|----------------------|
| `vagai analytics matching-distribution` | GET `/matching-distribution` |
| `vagai analytics resume-insights` | GET `/resume-insights` |
| `vagai analytics job-trends` | GET `/job-trends` |
| `vagai analytics org-metrics <id>` | GET `/org/:id` |
| `vagai analytics interview-prep` | GET `/interview-prep` |
| `vagai analytics ai-quality` | GET `/ai-quality` |
| `vagai analytics time-to-hire` | GET `/time-to-hire` |
| `vagai analytics skills-gap` | GET `/skills-gap` |
| `vagai analytics funnel` | GET `/funnel` |
| `vagai analytics geographic` | GET `/geographic` |

## Referências

- API Contract: `specs/010-data-analytics/contracts/api_contract.md`
- Data Model: `specs/010-data-analytics/data-model.md`
- Feature Spec: `specs/010-data-analytics/spec.md`