-- ============================================================
-- VagAI Metabase - Views de Demanda de Tecnologias
-- ============================================================
-- Métricas de mercado para orientar quais tecnologias estudar,
-- com base nas vagas coletadas e no gap com o currículo.
--
-- Aplicar manualmente (init só roda no primeiro boot):
--   docker cp metabase/setup/04-create-skill-views.sql vagai-mysql:/tmp/
--   docker exec vagai-mysql sh -c 'mysql -u vagai -pvagai vagai < /tmp/04-create-skill-views.sql'
-- ============================================================

USE vagai;

-- -----------------------------------------------------------
-- Tabela de referência: padrões regex por tecnologia.
-- Adicione novas linhas para ampliar a cobertura sem tocar nas views.
-- Padrões em minúsculas, compatíveis com REGEXP do MySQL 8 (ICU).
-- -----------------------------------------------------------
CREATE TABLE IF NOT EXISTS tech_keywords (
    id INT AUTO_INCREMENT PRIMARY KEY,
    tecnologia VARCHAR(50) NOT NULL UNIQUE,
    padrao VARCHAR(255) NOT NULL
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT IGNORE INTO tech_keywords (tecnologia, padrao) VALUES
    ('Go',              '(\\bgolang\\b|\\bgo\\b)'),
    ('Python',          '\\bpython\\b'),
    ('JavaScript',      '(\\bjavascript\\b|\\bjs\\b)'),
    ('TypeScript',      '\\btypescript\\b'),
    ('Java',            '\\bjava(?!script)\\b'),
    ('Kotlin',          '\\bkotlin\\b'),
    ('Swift',           '\\bswift\\b'),
    ('Rust',            '\\brust\\b'),
    ('PHP',             '\\bphp\\b'),
    ('Ruby',            '\\bruby\\b'),
    ('C#',              '\\bc#'),
    ('C++',             '\\bc\\+\\+'),
    ('.NET',            '(\\bdotnet\\b|\\b\\.net\\b)'),
    ('React',           '\\breact(\\.?js)?\\b'),
    ('Next.js',         '\\bnext\\.?js\\b'),
    ('Vue',             '\\bvue(\\.?js)?\\b'),
    ('Angular',         '\\bangular\\b'),
    ('Svelte',          '\\bsvelte\\b'),
    ('Node.js',         '(\\bnode\\.?js\\b|\\bnodejs\\b|\\bnode\\b)'),
    ('Django',          '\\bdjango\\b'),
    ('Flask',           '\\bflask\\b'),
    ('FastAPI',         '\\bfastapi\\b'),
    ('Spring',          '\\bspring( boot)?\\b'),
    ('Laravel',         '\\blaravel\\b'),
    ('Rails',           '\\brails\\b'),
    ('Docker',          '\\bdocker\\b'),
    ('Kubernetes',      '(\\bkubernetes\\b|\\bk8s\\b)'),
    ('Terraform',       '\\bterraform\\b'),
    ('Ansible',         '\\bansible\\b'),
    ('Linux',           '\\blinux\\b'),
    ('AWS',             '(\\baws\\b|amazon web services)'),
    ('Azure',           '\\bazure\\b'),
    ('GCP',             '(\\bgcp\\b|google cloud)'),
    ('SQL',             '\\bsql\\b'),
    ('PostgreSQL',      '(postgres|postgresql)'),
    ('MySQL',           '\\bmysql\\b'),
    ('MongoDB',         '\\bmongo\\w*\\b'),
    ('Redis',           '\\bredis\\b'),
    ('Elasticsearch',   '(elasticsearch|opensearch)'),
    ('Kafka',           '\\bkafka\\b'),
    ('RabbitMQ',        '(rabbitmq|\\brmq\\b)'),
    ('GraphQL',         '\\bgraphql\\b'),
    ('gRPC',            '\\bgrpc\\b'),
    ('CI/CD',           'ci/cd'),
    ('Jenkins',         '\\bjenkins\\b'),
    ('GitHub Actions',  'github actions'),
    ('Machine Learning','(machine learning|aprendizado de maquina|aprendizado de máquina)'),
    ('LLM / IA Gen',    '(\\bllms?\\b|generative ai|ia generativa|intelig[eé]ncia artificial)'),
    ('Spark',           '\\bspark\\b'),
    ('Airflow',         '\\bairflow\\b'),
    ('DevOps',          '\\bdevops\\b'),
    ('SRE',             '\\bsre\\b'),
    ('Scrum/Kanban',    '(\\bscrum\\b|\\bkanban\\b)'),
    ('Flutter',         '\\bflutter\\b'),
    ('React Native',    'react native'),
    ('Android',         '\\bandroid\\b'),
    ('iOS',             '\\bios\\b'),
    ('TDD',             '(\\btdd\\b|test.driven)'),
    ('Microserviços',   '(microservices|microsservi|microservi)'),
    ('Segurança',       '(cybersecurity|seguran[çc]a|\\bappsec\\b|penetration test)'),
    ('Data Engineering','(data engineer|engenharia de dados)')
;

-- -----------------------------------------------------------
-- 1. Demanda total: quantas vagas citam cada tecnologia
-- -----------------------------------------------------------
CREATE OR REPLACE VIEW `view_tech_demandas` AS
SELECT
    k.tecnologia,
    COUNT(DISTINCT j.url) AS total_vagas,
    ROUND(COUNT(DISTINCT j.url) * 100.0 /
        (SELECT COUNT(DISTINCT url) FROM jobs), 1) AS pct_das_vagas
FROM jobs j
JOIN tech_keywords k
    ON REGEXP_LIKE(CONCAT_WS(' ', LOWER(j.title), LOWER(j.description)), k.padrao)
GROUP BY k.tecnologia;

-- -----------------------------------------------------------
-- 2. Tendência mensal: demanda por mês de coleta
-- -----------------------------------------------------------
CREATE OR REPLACE VIEW `view_tech_demandas_mes` AS
SELECT
    DATE_FORMAT(j.collected_at, '%Y-%m') AS mes,
    k.tecnologia,
    COUNT(DISTINCT j.url) AS total_vagas
FROM jobs j
JOIN tech_keywords k
    ON REGEXP_LIKE(CONCAT_WS(' ', LOWER(j.title), LOWER(j.description)), k.padrao)
GROUP BY mes, k.tecnologia;

-- -----------------------------------------------------------
-- 3. Demanda recente: últimos 30 dias (momento do mercado)
-- -----------------------------------------------------------
CREATE OR REPLACE VIEW `view_tech_demandas_30d` AS
SELECT
    k.tecnologia,
    COUNT(DISTINCT j.url) AS total_vagas_30d,
    ROUND(COUNT(DISTINCT j.url) * 100.0 /
        (SELECT COUNT(DISTINCT url) FROM jobs
         WHERE collected_at >= NOW() - INTERVAL 30 DAY), 1) AS pct_das_vagas_30d
FROM jobs j
JOIN tech_keywords k
    ON REGEXP_LIKE(CONCAT_WS(' ', LOWER(j.title), LOWER(j.description)), k.padrao)
WHERE j.collected_at >= NOW() - INTERVAL 30 DAY
GROUP BY k.tecnologia;

-- -----------------------------------------------------------
-- 4. Gap currículo x mercado: o que estudar primeiro
--    prioridade = demanda da tech; presente no currículo
--    entra com peso reduzido (atualizar em vez de aprender)
-- -----------------------------------------------------------
CREATE OR REPLACE VIEW `view_skill_gap` AS
SELECT
    k.tecnologia,
    COALESCE(m.total_vagas, 0) AS vagas_exigindo,
    ROUND(COALESCE(pct.pct_das_vagas, 0), 1) AS pct_ultimos_90d,
    COALESCE(r.na_curriculo, FALSE) AS ja_no_curriculo,
    CASE
        WHEN COALESCE(r.na_curriculo, FALSE) THEN 'manter/atualizar'
        ELSE 'aprender'
    END AS recomendacao,
    ROUND(COALESCE(m.total_vagas, 0) *
        IF(COALESCE(r.na_curriculo, FALSE), 0.25, 1)) AS prioridade_estudo
FROM tech_keywords k
LEFT JOIN (
    SELECT k2.tecnologia, COUNT(DISTINCT j.url) AS total_vagas
    FROM jobs j
    JOIN tech_keywords k2
        ON REGEXP_LIKE(CONCAT_WS(' ', LOWER(j.title), LOWER(j.description)), k2.padrao)
    GROUP BY k2.tecnologia
) m ON m.tecnologia = k.tecnologia
LEFT JOIN (
    SELECT k2.tecnologia, COUNT(DISTINCT j.url) AS total_vagas,
           ROUND(COUNT(DISTINCT j.url) * 100.0 /
               (SELECT COUNT(DISTINCT url) FROM jobs
                WHERE collected_at >= NOW() - INTERVAL 90 DAY), 1) AS pct_das_vagas
    FROM jobs j
    JOIN tech_keywords k2
        ON REGEXP_LIKE(CONCAT_WS(' ', LOWER(j.title), LOWER(j.description)), k2.padrao)
    WHERE j.collected_at >= NOW() - INTERVAL 90 DAY
    GROUP BY k2.tecnologia
) pct ON pct.tecnologia = k.tecnologia
LEFT JOIN (
    SELECT k2.tecnologia, TRUE AS na_curriculo
    FROM resumes res
    JOIN tech_keywords k2
        ON REGEXP_LIKE(
            LOWER(CONCAT_WS(' ', COALESCE(res.content, ''), COALESCE(CAST(res.data AS CHAR), ''))),
            k2.padrao)
    WHERE res.id = (
        SELECT id FROM resumes ORDER BY uploaded_at DESC LIMIT 1
    )
    GROUP BY k2.tecnologia
) r ON r.tecnologia = k.tecnologia
WHERE COALESCE(m.total_vagas, 0) > 0
ORDER BY prioridade_estudo DESC, vagas_exigindo DESC;
