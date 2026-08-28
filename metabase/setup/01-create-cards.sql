-- ============================================================
-- VagAI Metabase Reports - MySQL Views for Metabase Cards
-- ============================================================
-- These views simplify Metabase card creation by pre-joining
-- and pre-aggregating data from the VagAI schema.
-- ============================================================

-- -----------------------------------------------------------
-- DASHBOARD 1: Pipeline de Vagas
-- -----------------------------------------------------------

-- 1.1 Funil de vagas por status
CREATE OR REPLACE VIEW `view_vagas_funil` AS
SELECT
    status,
    COUNT(*) AS total,
    ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER(), 2) AS percentual
FROM jobs
WHERE deleted_at IS NULL
GROUP BY status;

-- 1.2 Vagas coletadas por dia
CREATE OR REPLACE VIEW `view_vagas_por_dia` AS
SELECT
    DATE(collected_at) AS data_coleta,
    COUNT(*) AS total_vagas
FROM jobs
WHERE deleted_at IS NULL
GROUP BY DATE(collected_at)
ORDER BY data_coleta DESC;

-- 1.3 Top 10 empresas com mais vagas
CREATE OR REPLACE VIEW `view_top_empresas` AS
SELECT
    company AS empresa,
    COUNT(*) AS total_vagas,
    MIN(collected_at) AS primeira_vaga,
    MAX(collected_at) AS ultima_vaga
FROM jobs
WHERE deleted_at IS NULL AND company IS NOT NULL AND company != ''
GROUP BY company
ORDER BY total_vagas DESC
LIMIT 50;

-- 1.4 Vagas por site de origem
CREATE OR REPLACE VIEW `view_vagas_por_site` AS
SELECT
    s.name AS site,
    s.url AS url_site,
    COUNT(j.id) AS total_vagas,
    MAX(j.collected_at) AS ultimo_coleta
FROM jobs j
LEFT JOIN sites s ON j.site_id = s.id
WHERE j.deleted_at IS NULL
GROUP BY s.id, s.name, s.url
ORDER BY total_vagas DESC;

-- 1.5 Lag de coleta (dias entre posted_date e collected_at)
CREATE OR REPLACE VIEW `view_lag_coleta` AS
SELECT
    j.id,
    j.title AS vaga,
    j.company AS empresa,
    DATE(j.posted_date) AS data_publicacao,
    DATE(j.collected_at) AS data_coleta,
    DATEDIFF(j.collected_at, j.posted_date) AS dias_lag
FROM jobs j
WHERE j.deleted_at IS NULL
  AND j.posted_date IS NOT NULL
  AND j.collected_at IS NOT NULL;


-- -----------------------------------------------------------
-- DASHBOARD 2: Qualidade dos Matches
-- -----------------------------------------------------------

-- 2.1 Distribuição de scores (faixas)
CREATE OR REPLACE VIEW `view_score_distribuicao` AS
SELECT
    CASE
        WHEN m.similarity_score >= 80 THEN '80-100 (Excelente)'
        WHEN m.similarity_score >= 60 THEN '60-79 (Bom)'
        WHEN m.similarity_score >= 40 THEN '40-59 (Medio)'
        WHEN m.similarity_score >= 20 THEN '20-39 (Baixo)'
        ELSE '0-19 (Muito Baixo)'
    END AS faixa_score,
    COUNT(*) AS total_matches,
    ROUND(AVG(m.similarity_score), 2) AS score_medio,
    MIN(m.similarity_score) AS score_minimo,
    MAX(m.similarity_score) AS score_maximo
FROM matches m
GROUP BY faixa_score
ORDER BY score_minimo DESC;

-- 2.2 Score medio por curriculo
CREATE OR REPLACE VIEW `view_score_por_curriculo` AS
SELECT
    r.name AS curriculo,
    r.id AS curriculo_id,
    COUNT(m.id) AS total_matches,
    ROUND(AVG(m.similarity_score), 2) AS score_medio,
    ROUND(MAX(m.similarity_score), 2) AS melhor_score,
    SUM(CASE WHEN m.applied = 1 THEN 1 ELSE 0 END) AS total_aplicados
FROM matches m
JOIN resumes r ON m.resume_id = r.id
GROUP BY r.id, r.name
ORDER BY score_medio DESC;

-- 2.3 Taxa de aplicacao
CREATE OR REPLACE VIEW `view_taxa_aplicacao` AS
SELECT
    COUNT(*) AS total_matches,
    SUM(CASE WHEN applied = 1 THEN 1 ELSE 0 END) AS aplicados,
    SUM(CASE WHEN applied = 0 OR applied IS NULL THEN 1 ELSE 0 END) AS nao_aplicados,
    ROUND(SUM(CASE WHEN applied = 1 THEN 1 ELSE 0 END) * 100.0 / COUNT(*), 2) AS taxa_aplicacao_pct
FROM matches;

-- 2.4 Keywords mais frequentes nos matches
CREATE OR REPLACE VIEW `view_keywords_frequentes` AS
SELECT
    keyword,
    COUNT(*) AS frequencia,
    ROUND(AVG(m.similarity_score), 2) AS score_medio
FROM matches m,
JSON_TABLE(
    m.keywords_matched,
    '$[*]' COLUMNS (keyword VARCHAR(100) PATH '$')
) AS jt
WHERE m.keywords_matched IS NOT NULL
  AND JSON_VALID(m.keywords_matched)
GROUP BY keyword
ORDER BY frequencia DESC
LIMIT 50;

-- 2.5 Match count por tier de plano
CREATE OR REPLACE VIEW `view_matches_por_plano` AS
SELECT
    p.name AS plano,
    p.slug AS plano_slug,
    COUNT(m.id) AS total_matches,
    ROUND(AVG(m.similarity_score), 2) AS score_medio,
    SUM(CASE WHEN m.applied = 1 THEN 1 ELSE 0 END) AS aplicados
FROM matches m
JOIN organizations o ON m.organization_id = o.id
JOIN subscriptions s ON s.organization_id = o.id
JOIN plans p ON s.plan_id = p.id
WHERE o.deleted_at IS NULL AND s.deleted_at IS NULL
GROUP BY p.id, p.name, p.slug;


-- -----------------------------------------------------------
-- DASHBOARD 3: Receita & Assinaturas
-- -----------------------------------------------------------

-- 3.1 MRR (Monthly Recurring Revenue)
CREATE OR REPLACE VIEW `view_mrr` AS
SELECT
    p.name AS plano,
    COUNT(s.id) AS assinaturas_ativas,
    p.price_monthly AS preco_centavos,
    ROUND(p.price_monthly / 100.0, 2) AS preco_reais,
    ROUND(p.price_monthly / 100.0 * COUNT(s.id), 2) AS mrr_total_reais
FROM subscriptions s
JOIN plans p ON s.plan_id = p.id
WHERE s.status = 'active'
GROUP BY p.id, p.name, p.price_monthly;

-- 3.2 Conversao trial para paid
CREATE OR REPLACE VIEW `view_conversao_trial` AS
SELECT
    CASE
        WHEN status = 'trial' THEN 'Trial'
        WHEN status = 'active' THEN 'Ativo (Pago)'
        WHEN status = 'past_due' THEN 'Pagamento Pendente'
        WHEN status = 'canceled' THEN 'Cancelado'
        WHEN status = 'expired' THEN 'Expirado'
        ELSE status
    END AS status_assinatura,
    COUNT(*) AS total,
    ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER(), 2) AS percentual
FROM subscriptions
WHERE deleted_at IS NULL
GROUP BY status
ORDER BY FIELD(status, 'active', 'trial', 'past_due', 'canceled', 'expired');

-- 3.3 Churn rate por mes
CREATE OR REPLACE VIEW `view_churn_mensal` AS
SELECT
    DATE_FORMAT(created_at, '%Y-%m') AS mes,
    COUNT(*) AS total_assinaturas,
    SUM(CASE WHEN status IN ('canceled', 'expired') THEN 1 ELSE 0 END) AS canceladas,
    ROUND(SUM(CASE WHEN status IN ('canceled', 'expired') THEN 1 ELSE 0 END) * 100.0 / COUNT(*), 2) AS churn_rate_pct
FROM subscriptions
WHERE deleted_at IS NULL
GROUP BY DATE_FORMAT(created_at, '%Y-%m')
ORDER BY mes DESC;

-- 3.4 Assinaturas com cancelamento programado (risco)
CREATE OR REPLACE VIEW `view_churn_risco` AS
SELECT
    o.name AS organizacao,
    p.name AS plano,
    s.current_period_end AS vencimento,
    DATEDIFF(s.current_period_end, NOW()) AS dias_restantes
FROM subscriptions s
JOIN organizations o ON s.organization_id = o.id
JOIN plans p ON s.plan_id = p.id
WHERE s.cancel_at_period_end = 1
  AND s.status = 'active'
  AND s.deleted_at IS NULL
ORDER BY s.current_period_end ASC;

-- 3.5 Proximos vencimentos (30 dias)
CREATE OR REPLACE VIEW `view_proximos_vencimentos` AS
SELECT
    o.name AS organizacao,
    p.name AS plano,
    s.status,
    s.current_period_start AS inicio_periodo,
    s.current_period_end AS fim_periodo,
    DATEDIFF(s.current_period_end, NOW()) AS dias_ate_vencer
FROM subscriptions s
JOIN organizations o ON s.organization_id = o.id
JOIN plans p ON s.plan_id = p.id
WHERE s.status IN ('active', 'trial')
  AND s.current_period_end IS NOT NULL
  AND s.current_period_end >= NOW()
  AND s.current_period_end <= DATE_ADD(NOW(), INTERVAL 30 DAY)
  AND s.deleted_at IS NULL
ORDER BY s.current_period_end ASC;


-- -----------------------------------------------------------
-- DASHBOARD 4: Crescimento de Usuarios
-- -----------------------------------------------------------

-- 4.1 Novas organizacoes por semana
CREATE OR REPLACE VIEW `view_orgs_por_semana` AS
SELECT
    YEARWEEK(created_at, 1) AS semana,
    MIN(DATE(created_at)) AS inicio_semana,
    COUNT(*) AS novas_orgs
FROM organizations
WHERE deleted_at IS NULL
GROUP BY YEARWEEK(created_at, 1)
ORDER BY semana DESC;

-- 4.2 Usuarios verificados vs nao verificados
CREATE OR REPLACE VIEW `view_usuarios_verificacao` AS
SELECT
    CASE
        WHEN email_verified_at IS NOT NULL THEN 'Verificado'
        ELSE 'Nao Verificado'
    END AS status_verificacao,
    COUNT(*) AS total_usuarios
FROM users
WHERE deleted_at IS NULL
GROUP BY status_verificacao;

-- 4.3 Distribuicao de planos
CREATE OR REPLACE VIEW `view_distribuicao_planos` AS
SELECT
    o.plan AS plano,
    COUNT(DISTINCT o.id) AS total_organizacoes,
    COUNT(DISTINCT m.user_id) AS total_usuarios
FROM organizations o
LEFT JOIN memberships m ON o.id = m.organization_id
WHERE o.deleted_at IS NULL
GROUP BY o.plan
ORDER BY total_organizacoes DESC;

-- 4.4 Membros por organizacao
CREATE OR REPLACE VIEW `view_membros_por_org` AS
SELECT
    o.name AS organizacao,
    o.plan AS plano,
    COUNT(m.id) AS total_membros
FROM organizations o
LEFT JOIN memberships m ON o.id = m.organization_id
WHERE o.deleted_at IS NULL
GROUP BY o.id, o.name, o.plan
ORDER BY total_membros DESC;


-- -----------------------------------------------------------
-- DASHBOARD 5: Saude do Sistema
-- -----------------------------------------------------------

-- 5.1 Atividade dos agents por dia
CREATE OR REPLACE VIEW `view_agentes_por_dia` AS
SELECT
    DATE(created_at) AS data,
    agent_name AS agente,
    action AS acao,
    COUNT(*) AS total_eventos
FROM agent_logs
GROUP BY DATE(created_at), agent_name, action
ORDER BY data DESC;

-- 5.2 Jobs encontrados por crawl (via agent_logs)
CREATE OR REPLACE VIEW `view_crawl_stats` AS
SELECT
    DATE(al.created_at) AS data_crawl,
    JSON_UNQUOTE(JSON_EXTRACT(al.details, '$.site')) AS site,
    CAST(JSON_EXTRACT(al.details, '$.jobs') AS UNSIGNED) AS vagas_encontradas,
    al.action AS acao
FROM agent_logs al
WHERE al.agent_name = 'crawler'
  AND al.action = 'crawl_completed'
  AND al.details IS NOT NULL
ORDER BY data_crawl DESC;

-- 5.3 Taxa de sucesso dos schedules
CREATE OR REPLACE VIEW `view_schedule_sucesso` AS
SELECT
    s.name AS schedule_name,
    COUNT(sl.id) AS total_execucoes,
    SUM(CASE WHEN sl.status = 'success' THEN 1 ELSE 0 END) AS sucessos,
    SUM(CASE WHEN sl.status = 'failed' THEN 1 ELSE 0 END) AS falhas,
    ROUND(SUM(CASE WHEN sl.status = 'success' THEN 1 ELSE 0 END) * 100.0 / COUNT(*), 2) AS taxa_sucesso_pct
FROM schedules s
LEFT JOIN schedule_logs sl ON s.id = sl.schedule_id
GROUP BY s.id, s.name
ORDER BY total_execucoes DESC;

-- 5.4 Duracao media de execucao
CREATE OR REPLACE VIEW `view_duracao_execucao` AS
SELECT
    s.name AS schedule_name,
    COUNT(sl.id) AS total_execucoes,
    ROUND(AVG(TIMESTAMPDIFF(SECOND, sl.started_at, sl.finished_at)), 2) AS duracao_media_seg,
    MIN(TIMESTAMPDIFF(SECOND, sl.started_at, sl.finished_at)) AS duracao_min_seg,
    MAX(TIMESTAMPDIFF(SECOND, sl.started_at, sl.finished_at)) AS duracao_max_seg
FROM schedules s
JOIN schedule_logs sl ON s.id = sl.schedule_id
WHERE sl.finished_at IS NOT NULL
GROUP BY s.id, s.name
ORDER BY duracao_media_seg DESC;


-- -----------------------------------------------------------
-- DASHBOARD 6: Insights de Resumes
-- -----------------------------------------------------------

-- 6.1 Forcas mais comuns
CREATE OR REPLACE VIEW `view_forcas_comuns` AS
SELECT
    forca,
    COUNT(*) AS frequencia
FROM resume_analyses,
JSON_TABLE(
    strengths,
    '$[*]' COLUMNS (forca VARCHAR(255) PATH '$')
) AS jt
WHERE strengths IS NOT NULL
  AND JSON_VALID(strengths)
GROUP BY forca
ORDER BY frequencia DESC
LIMIT 30;

-- 6.2 Fraquezas mais frequentes
CREATE OR REPLACE VIEW `view_fraquezas_frequentes` AS
SELECT
    fraqueza,
    COUNT(*) AS frequencia
FROM resume_analyses,
JSON_TABLE(
    weaknesses,
    '$[*]' COLUMNS (fraqueza VARCHAR(255) PATH '$')
) AS jt
WHERE weaknesses IS NOT NULL
  AND JSON_VALID(weaknesses)
GROUP BY fraqueza
ORDER BY frequencia DESC
LIMIT 30;

-- 6.3 Sugestoes da IA
CREATE OR REPLACE VIEW `view_sugestoes_ia` AS
SELECT
    sugestao,
    COUNT(*) AS frequencia
FROM resume_analyses,
JSON_TABLE(
    suggestions,
    '$[*]' COLUMNS (sugestao VARCHAR(255) PATH '$')
) AS jt
WHERE suggestions IS NOT NULL
  AND JSON_VALID(suggestions)
GROUP BY sugestao
ORDER BY frequencia DESC
LIMIT 30;

-- 6.4 Analises por periodo
CREATE OR REPLACE VIEW `view_analises_por_periodo` AS
SELECT
    DATE(created_at) AS data_analise,
    COUNT(*) AS total_analises,
    COUNT(DISTINCT resume_id) AS curriculos_analisados
FROM resume_analyses
GROUP BY DATE(created_at)
ORDER BY data_analise DESC;


-- -----------------------------------------------------------
-- DASHBOARD 7: Performance dos Sites
-- -----------------------------------------------------------

-- 7.1 Status dos sites (ativo/inativo)
CREATE OR REPLACE VIEW `view_status_sites` AS
SELECT
    CASE WHEN active = 1 THEN 'Ativo' ELSE 'Inativo' END AS status_site,
    COUNT(*) AS total_sites
FROM sites
GROUP BY active;

-- 7.2 Ultimo crawl por site
CREATE OR REPLACE VIEW `view_ultimo_crawl` AS
SELECT
    s.name AS site,
    s.url AS url_site,
    s.active AS ativo,
    s.last_crawl AS ultimo_crawl,
    CASE
        WHEN s.last_crawl IS NULL THEN 'Nunca crawlado'
        WHEN s.last_crawl < DATE_SUB(NOW(), INTERVAL 7 DAY) THEN 'Desatualizado (>7d)'
        WHEN s.last_crawl < DATE_SUB(NOW(), INTERVAL 1 DAY) THEN 'Lag leve (1-7d)'
        ELSE 'Atualizado (<1d)'
    END AS status_freshness,
    DATEDIFF(NOW(), s.last_crawl) AS dias_desde_ultimo_crawl
FROM sites s
ORDER BY s.last_crawl DESC;

-- 7.3 Jobs por site (eficiencia)
CREATE OR REPLACE VIEW `view_eficiencia_sites` AS
SELECT
    s.name AS site,
    s.url AS url_site,
    COUNT(j.id) AS total_jobs,
    MIN(j.collected_at) AS primeira_coleta,
    MAX(j.collected_at) AS ultima_coleta,
    COUNT(DISTINCT DATE(j.collected_at)) AS dias_ativos,
    CASE
        WHEN COUNT(DISTINCT DATE(j.collected_at)) > 0
        THEN ROUND(COUNT(j.id) * 1.0 / COUNT(DISTINCT DATE(j.collected_at)), 1)
        ELSE 0
    END AS jobs_por_dia
FROM sites s
LEFT JOIN jobs j ON s.id = j.site_id AND j.deleted_at IS NULL
GROUP BY s.id, s.name, s.url
ORDER BY total_jobs DESC;
