-- =============================================================================
-- VagAI — Migração SaaS 001: Organization/Membership/Subscription -> User
-- =============================================================================
--
-- Contexto
--   A plataforma deixou de ser multi-tenant por "organization" e passou a ter
--   apenas dois papéis em `users`: `user` (dono dos próprios dados) e `admin`
--   (gestão gerencial). Plano, assinatura Stripe, palavras-chave negativas e
--   trial agora são atributos diretos do usuário (1:1).
--
-- Estratégia de dados (Opção A)
--   O PRIMEIRO usuário de cada organização (preferindo role = 'owner', depois o
--   menor membership.id) torna-se dono de TODOS os dados daquela organização.
--
-- Como executar
--   1. PARE o serviço da API e do CLI (janela de manutenção).
--   2. Faça backup do banco.
--   3. Execute via ./migrate.sh (recomendado) ou:
--        mysql -u <user> -p <database> < migrations/001_saas_migration.sql
--   4. Suba o novo binário. O AutoMigrate do GORM apenas valida/ajusta o schema.
--
--   O script é idempotente: usa checagens em information_schema antes de
--   adicionar/remover colunas, tabelas e índices. Pode ser reexecutado com
--   segurança após uma execução interrompida.
--
-- ADMIN_EMAILS
--   Após a migração, promova os administradores da plataforma (um UPDATE por
--   e-mail), por exemplo:
--     UPDATE users SET role = 'admin' WHERE email IN ('admin@vagai.ia');
--   Em produção, considere manter essa lista na variável de ambiente
--   ADMIN_EMAILS do serviço e um job para sincronizar.
-- =============================================================================

SET FOREIGN_KEY_CHECKS = 0;

-- -----------------------------------------------------------------------------
-- Helpers idempotentes (DDL dinâmico)
-- -----------------------------------------------------------------------------
DROP PROCEDURE IF EXISTS vagai_add_column;
DROP PROCEDURE IF EXISTS vagai_drop_column;
DROP PROCEDURE IF EXISTS vagai_add_index;
DROP PROCEDURE IF EXISTS vagai_add_unique_index;
DROP PROCEDURE IF EXISTS vagai_drop_index;
DROP PROCEDURE IF EXISTS vagai_drop_fk;
DROP PROCEDURE IF EXISTS vagai_drop_table_if_exists;
DROP PROCEDURE IF EXISTS vagai_backfill_owner;

DELIMITER $$

CREATE PROCEDURE vagai_add_column(IN tbl VARCHAR(64), IN col VARCHAR(64), IN ddl VARCHAR(500))
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables
             WHERE table_schema = DATABASE() AND table_name = tbl) THEN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_schema = DATABASE() AND table_name = tbl AND column_name = col) THEN
      SET @s = CONCAT('ALTER TABLE `', tbl, '` ADD COLUMN ', ddl);
      PREPARE stmt FROM @s; EXECUTE stmt; DEALLOCATE PREPARE stmt;
    END IF;
  END IF;
END$$

CREATE PROCEDURE vagai_drop_column(IN tbl VARCHAR(64), IN col VARCHAR(64))
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns
             WHERE table_schema = DATABASE() AND table_name = tbl AND column_name = col) THEN
    SET @s = CONCAT('ALTER TABLE `', tbl, '` DROP COLUMN `', col, '`');
    PREPARE stmt FROM @s; EXECUTE stmt; DEALLOCATE PREPARE stmt;
  END IF;
END$$

CREATE PROCEDURE vagai_add_index(IN tbl VARCHAR(64), IN idx VARCHAR(64), IN cols VARCHAR(255))
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables
             WHERE table_schema = DATABASE() AND table_name = tbl) THEN
    IF NOT EXISTS (SELECT 1 FROM information_schema.statistics
                   WHERE table_schema = DATABASE() AND table_name = tbl AND index_name = idx) THEN
      SET @s = CONCAT('ALTER TABLE `', tbl, '` ADD INDEX `', idx, '` (', cols, ')');
      PREPARE stmt FROM @s; EXECUTE stmt; DEALLOCATE PREPARE stmt;
    END IF;
  END IF;
END$$

CREATE PROCEDURE vagai_add_unique_index(IN tbl VARCHAR(64), IN idx VARCHAR(64), IN cols VARCHAR(255))
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables
             WHERE table_schema = DATABASE() AND table_name = tbl) THEN
    IF NOT EXISTS (SELECT 1 FROM information_schema.statistics
                   WHERE table_schema = DATABASE() AND table_name = tbl AND index_name = idx) THEN
      SET @s = CONCAT('ALTER TABLE `', tbl, '` ADD UNIQUE INDEX `', idx, '` (', cols, ')');
      PREPARE stmt FROM @s; EXECUTE stmt; DEALLOCATE PREPARE stmt;
    END IF;
  END IF;
END$$

CREATE PROCEDURE vagai_drop_index(IN tbl VARCHAR(64), IN idx VARCHAR(64))
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.statistics
             WHERE table_schema = DATABASE() AND table_name = tbl AND index_name = idx) THEN
    SET @s = CONCAT('ALTER TABLE `', tbl, '` DROP INDEX `', idx, '`');
    PREPARE stmt FROM @s; EXECUTE stmt; DEALLOCATE PREPARE stmt;
  END IF;
END$$

CREATE PROCEDURE vagai_drop_fk(IN tbl VARCHAR(64), IN fk VARCHAR(64))
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.table_constraints
             WHERE table_schema = DATABASE() AND table_name = tbl
               AND constraint_name = fk AND constraint_type = 'FOREIGN KEY') THEN
    SET @s = CONCAT('ALTER TABLE `', tbl, '` DROP FOREIGN KEY `', fk, '`');
    PREPARE stmt FROM @s; EXECUTE stmt; DEALLOCATE PREPARE stmt;
  END IF;
END$$

CREATE PROCEDURE vagai_drop_table_if_exists(IN tbl VARCHAR(64))
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables
             WHERE table_schema = DATABASE() AND table_name = tbl) THEN
    SET @s = CONCAT('DROP TABLE `', tbl, '`');
    PREPARE stmt FROM @s; EXECUTE stmt; DEALLOCATE PREPARE stmt;
  END IF;
END$$

DELIMITER ;

-- -----------------------------------------------------------------------------
-- 1) Novas colunas de SaaS em `users`
-- -----------------------------------------------------------------------------
CALL vagai_add_column('users', 'role',                "role VARCHAR(20) NOT NULL DEFAULT 'user'");
CALL vagai_add_column('users', 'plan_slug',            "plan_slug VARCHAR(50) NOT NULL DEFAULT 'free'");
CALL vagai_add_column('users', 'negative_keywords',    "negative_keywords JSON NULL");
CALL vagai_add_column('users', 'stripe_customer_id',   "stripe_customer_id VARCHAR(255) NULL");
CALL vagai_add_column('users', 'subscription_status',  "subscription_status VARCHAR(20) NULL DEFAULT 'trial'");
CALL vagai_add_column('users', 'subscription_ends_at', "subscription_ends_at DATETIME(3) NULL");
CALL vagai_add_column('users', 'suspended',            "suspended TINYINT(1) NOT NULL DEFAULT 0");
CALL vagai_add_column('users', 'trial_ends_at',        "trial_ends_at DATETIME(3) NULL");
CALL vagai_add_column('users', 'email_verified_at',    "email_verified_at DATETIME(3) NULL");

CALL vagai_add_index('users', 'idx_users_role', '`role`');
CALL vagai_add_index('users', 'idx_users_stripe_customer_id', '`stripe_customer_id`');

-- Garante normalização JSON (evita NULL em leituras que fazem unmarshal).
UPDATE users SET negative_keywords = JSON_ARRAY() WHERE negative_keywords IS NULL;

-- -----------------------------------------------------------------------------
-- 2) Garante a coluna `user_id` em TODAS as tabelas de dados
--    Algumas tabelas legadas só possuíam `organization_id`; outras já receberam
--    `user_id` via AutoMigrate. As chamadas abaixo são idempotentes.
-- -----------------------------------------------------------------------------
CALL vagai_add_column('sites',                              'user_id', 'user_id BIGINT UNSIGNED NOT NULL DEFAULT 0');
CALL vagai_add_column('jobs',                               'user_id', 'user_id BIGINT UNSIGNED NOT NULL DEFAULT 0');
CALL vagai_add_column('resumes',                            'user_id', 'user_id BIGINT UNSIGNED NOT NULL DEFAULT 0');
CALL vagai_add_column('matches',                            'user_id', 'user_id BIGINT UNSIGNED NOT NULL DEFAULT 0');
CALL vagai_add_column('resume_analyses',                    'user_id', 'user_id BIGINT UNSIGNED NOT NULL DEFAULT 0');
CALL vagai_add_column('agent_logs',                         'user_id', 'user_id BIGINT UNSIGNED NOT NULL DEFAULT 0');
CALL vagai_add_column('audit_logs',                         'user_id', 'user_id BIGINT UNSIGNED NOT NULL DEFAULT 0');
CALL vagai_add_column('schedules',                          'user_id', 'user_id BIGINT UNSIGNED NOT NULL DEFAULT 0');
CALL vagai_add_column('interview_preparations',             'user_id', 'user_id BIGINT UNSIGNED NOT NULL DEFAULT 0');
CALL vagai_add_column('interview_questions',                'user_id', 'user_id BIGINT UNSIGNED NOT NULL DEFAULT 0');
CALL vagai_add_column('practice_entries',                   'user_id', 'user_id BIGINT UNSIGNED NOT NULL DEFAULT 0');
CALL vagai_add_column('preparation_verifications',          'user_id', 'user_id BIGINT UNSIGNED NOT NULL DEFAULT 0');
CALL vagai_add_column('ai_analysis_results',                'user_id', 'user_id BIGINT UNSIGNED NOT NULL DEFAULT 0');
CALL vagai_add_column('preparation_verification_histories', 'user_id', 'user_id BIGINT UNSIGNED NOT NULL DEFAULT 0');
CALL vagai_add_column('api_keys',                           'user_id', 'user_id BIGINT UNSIGNED NOT NULL DEFAULT 0');

CALL vagai_add_index('sites',                              'idx_sites_user_id',                              '`user_id`');
CALL vagai_add_index('jobs',                               'idx_jobs_user_id',                               '`user_id`');
CALL vagai_add_index('resumes',                            'idx_resumes_user_id',                            '`user_id`');
CALL vagai_add_index('matches',                            'idx_matches_user_id',                            '`user_id`');
CALL vagai_add_index('resume_analyses',                    'idx_resume_analyses_user_id',                    '`user_id`');
CALL vagai_add_index('agent_logs',                         'idx_agent_logs_user_id',                         '`user_id`');
CALL vagai_add_index('audit_logs',                         'idx_audit_logs_user_id',                         '`user_id`');
CALL vagai_add_index('schedules',                          'idx_schedules_user_id',                          '`user_id`');
CALL vagai_add_index('interview_preparations',             'idx_interview_preparations_user_id',             '`user_id`');
CALL vagai_add_index('interview_questions',                'idx_interview_questions_user_id',                '`user_id`');
CALL vagai_add_index('practice_entries',                   'idx_practice_entries_user_id',                   '`user_id`');
CALL vagai_add_index('preparation_verifications',          'idx_preparation_verifications_user_id',          '`user_id`');
CALL vagai_add_index('ai_analysis_results',                'idx_ai_analysis_results_user_id',                '`user_id`');
CALL vagai_add_index('preparation_verification_histories', 'idx_preparation_verification_histories_user_id', '`user_id`');
CALL vagai_add_index('api_keys',                           'idx_api_keys_user_id',                           '`user_id`');

-- -----------------------------------------------------------------------------
-- 3) Remove as views antigas que referenciam organization_id ANTES de dropar as
--    colunas. As views user-scoped são recriadas no boot (CreateAnalyticsViews).
-- -----------------------------------------------------------------------------
DROP VIEW IF EXISTS matching_distribution;
DROP VIEW IF EXISTS user_metrics_summary;
DROP VIEW IF EXISTS org_metrics_summary;
DROP VIEW IF EXISTS funnel_summary;

-- -----------------------------------------------------------------------------
-- 4) Backfill de posse e atributos (Opção A)
--    Executado apenas se a tabela legada `memberships` ainda existir — assim o
--    script é seguro para rodar mais de uma vez.
-- -----------------------------------------------------------------------------
DROP PROCEDURE IF EXISTS vagai_backfill_owner;
DELIMITER $$
CREATE PROCEDURE vagai_backfill_owner()
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables
             WHERE table_schema = DATABASE() AND table_name = 'memberships') THEN

  DROP TEMPORARY TABLE IF EXISTS tmp_org_owner;
  CREATE TEMPORARY TABLE tmp_org_owner (
    organization_id BIGINT UNSIGNED NOT NULL PRIMARY KEY,
    user_id         BIGINT UNSIGNED NOT NULL,
    INDEX (user_id)
  ) ENGINE = InnoDB;

  INSERT INTO tmp_org_owner (organization_id, user_id)
  SELECT organization_id, user_id
  FROM (
    SELECT m.organization_id,
           m.user_id,
           ROW_NUMBER() OVER (
             PARTITION BY m.organization_id
             ORDER BY (m.role = 'owner') DESC, m.id ASC
           ) AS rn
    FROM memberships m
  ) AS sub
  WHERE sub.rn = 1;

  -- Atributos SaaS do usuário dono (plano, Stripe, trial, keywords).
  UPDATE users u
    JOIN tmp_org_owner o      ON o.user_id = u.id
    LEFT JOIN organizations org ON org.id = o.organization_id
    LEFT JOIN (
      SELECT s.organization_id, s.status, s.current_period_end
      FROM subscriptions s
      JOIN (SELECT organization_id, MAX(id) AS id
              FROM subscriptions
             WHERE deleted_at IS NULL
             GROUP BY organization_id) x ON x.id = s.id
    ) sub ON sub.organization_id = o.organization_id
    SET
      u.plan_slug            = COALESCE(NULLIF(org.plan, ''), u.plan_slug, 'free'),
      u.stripe_customer_id   = COALESCE(NULLIF(org.stripe_customer_id, ''), u.stripe_customer_id),
      u.subscription_status  = COALESCE(NULLIF(sub.status, ''), u.subscription_status, 'trial'),
      u.subscription_ends_at = COALESCE(sub.current_period_end, u.subscription_ends_at),
      u.trial_ends_at        = COALESCE(u.trial_ends_at, org.trial_ends_at),
      u.negative_keywords    = CASE
                                 WHEN (u.negative_keywords IS NULL OR JSON_LENGTH(u.negative_keywords) = 0)
                                      AND org.negative_keywords IS NOT NULL
                                 THEN org.negative_keywords
                                 ELSE u.negative_keywords
                               END;

  -- Reatribuição de posse dos dados: organization_id -> user_id
  UPDATE sites s                              JOIN tmp_org_owner o ON o.organization_id = s.organization_id  SET s.user_id = o.user_id;
  UPDATE jobs j                               JOIN tmp_org_owner o ON o.organization_id = j.organization_id  SET j.user_id = o.user_id;
  UPDATE resumes r                            JOIN tmp_org_owner o ON o.organization_id = r.organization_id  SET r.user_id = o.user_id;
  UPDATE matches m                            JOIN tmp_org_owner o ON o.organization_id = m.organization_id  SET m.user_id = o.user_id;
  UPDATE resume_analyses ra                   JOIN tmp_org_owner o ON o.organization_id = ra.organization_id SET ra.user_id = o.user_id;
  UPDATE agent_logs al                        JOIN tmp_org_owner o ON o.organization_id = al.organization_id SET al.user_id = o.user_id;
  UPDATE audit_logs au                        JOIN tmp_org_owner o ON o.organization_id = au.organization_id SET au.user_id = o.user_id;
  UPDATE schedules sc                         JOIN tmp_org_owner o ON o.organization_id = sc.organization_id SET sc.user_id = o.user_id;
  UPDATE interview_preparations ip            JOIN tmp_org_owner o ON o.organization_id = ip.organization_id SET ip.user_id = o.user_id;
  UPDATE interview_questions iq               JOIN tmp_org_owner o ON o.organization_id = iq.organization_id SET iq.user_id = o.user_id;
  UPDATE practice_entries pe                  JOIN tmp_org_owner o ON o.organization_id = pe.organization_id SET pe.user_id = o.user_id;
  UPDATE preparation_verifications pv         JOIN tmp_org_owner o ON o.organization_id = pv.organization_id SET pv.user_id = o.user_id;
  UPDATE ai_analysis_results ar               JOIN tmp_org_owner o ON o.organization_id = ar.organization_id SET ar.user_id = o.user_id;
  UPDATE preparation_verification_histories pvh JOIN tmp_org_owner o ON o.organization_id = pvh.organization_id SET pvh.user_id = o.user_id;
  UPDATE api_keys ak                          JOIN tmp_org_owner o ON o.organization_id = ak.organization_id SET ak.user_id = o.user_id;

  DROP TEMPORARY TABLE IF EXISTS tmp_org_owner;

  END IF;
END$$
DELIMITER ;

CALL vagai_backfill_owner();
DROP PROCEDURE IF EXISTS vagai_backfill_owner;

-- Sem assinatura mapeada permanece em trial.
UPDATE users SET subscription_status = 'trial' WHERE subscription_status IS NULL;
UPDATE users SET negative_keywords = JSON_ARRAY() WHERE negative_keywords IS NULL;

-- -----------------------------------------------------------------------------
-- 5) Deduplicação e índices únicos novos
--    Remove colisões em (user_id, url) e (user_id, preparation_id) mantendo o
--    registro de menor id antes de (re)criar os índices únicos.
-- -----------------------------------------------------------------------------
DELETE j1 FROM jobs j1
JOIN jobs j2
  ON j1.user_id = j2.user_id AND j1.url = j2.url AND j1.id > j2.id;

DELETE pv1 FROM preparation_verifications pv1
JOIN preparation_verifications pv2
  ON pv1.user_id = pv2.user_id AND pv1.preparation_id = pv2.preparation_id AND pv1.id > pv2.id;

-- Remove índices antigos herdados do multi-tenant e recria com o novo escopo.
CALL vagai_drop_index('jobs', 'idx_job_url_org');
CALL vagai_drop_index('jobs', 'idx_job_url_user');
CALL vagai_add_unique_index('jobs', 'idx_job_url_user', '`user_id`, `url`');

CALL vagai_drop_index('preparation_verifications', 'idx_prep_verification');
CALL vagai_add_unique_index('preparation_verifications', 'idx_prep_verification', '`user_id`, `preparation_id`');

-- -----------------------------------------------------------------------------
-- 5a) Remove índices legados restantes que citam organization_id
-- -----------------------------------------------------------------------------
CALL vagai_drop_index('agent_logs',                         'idx_agent_logs_organization_id');
CALL vagai_drop_index('ai_analysis_results',                'idx_ai_analysis_results_organization_id');
CALL vagai_drop_index('ai_analysis_results',                'idx_ai_org_prep');
CALL vagai_drop_index('api_keys',                           'idx_api_keys_organization_id');
CALL vagai_drop_index('audit_logs',                         'idx_audit_logs_organization_id');
CALL vagai_drop_index('interview_preparations',             'idx_interview_org');
CALL vagai_drop_index('interview_preparations',             'idx_interview_preparations_organization_id');
CALL vagai_drop_index('interview_questions',                'idx_interview_questions_organization_id');
CALL vagai_drop_index('jobs',                               'idx_jobs_organization_id');
CALL vagai_drop_index('matches',                            'idx_match_organization');
CALL vagai_drop_index('matches',                            'idx_matches_organization_id');
CALL vagai_drop_index('practice_entries',                   'idx_practice_entries_organization_id');
CALL vagai_drop_index('preparation_verifications',          'idx_preparation_verifications_organization_id');
CALL vagai_drop_index('preparation_verification_histories', 'idx_preparation_verification_histories_organization_id');
CALL vagai_drop_index('preparation_verification_histories', 'idx_verification_org');
CALL vagai_drop_index('resume_analyses',                    'idx_resume_analyses_organization_id');
CALL vagai_drop_index('resumes',                            'idx_resume_organization');
CALL vagai_drop_index('resumes',                            'idx_resumes_organization_id');
CALL vagai_drop_index('schedules',                          'idx_schedules_organization_id');
CALL vagai_drop_index('sites',                              'idx_sites_organization_id');

-- -----------------------------------------------------------------------------
-- 5c) Remoção da FK/coluna legada de organização em `users`
--    É necessário dropar a FK explicitamente antes da coluna.
-- -----------------------------------------------------------------------------
CALL vagai_drop_fk('users', 'fk_users_organization');
CALL vagai_drop_index('users', 'fk_users_organization');
CALL vagai_drop_column('users', 'organization_id');

-- -----------------------------------------------------------------------------
-- 6) Remoção das colunas legadas organization_id
-- -----------------------------------------------------------------------------
CALL vagai_drop_column('sites', 'organization_id');
CALL vagai_drop_column('jobs', 'organization_id');
CALL vagai_drop_column('resumes', 'organization_id');
CALL vagai_drop_column('matches', 'organization_id');
CALL vagai_drop_column('resume_analyses', 'organization_id');
CALL vagai_drop_column('agent_logs', 'organization_id');
CALL vagai_drop_column('audit_logs', 'organization_id');
CALL vagai_drop_column('schedules', 'organization_id');
CALL vagai_drop_column('interview_preparations', 'organization_id');
CALL vagai_drop_column('interview_questions', 'organization_id');
CALL vagai_drop_column('practice_entries', 'organization_id');
CALL vagai_drop_column('preparation_verifications', 'organization_id');
CALL vagai_drop_column('ai_analysis_results', 'organization_id');
CALL vagai_drop_column('preparation_verification_histories', 'organization_id');
CALL vagai_drop_column('api_keys', 'organization_id');

-- -----------------------------------------------------------------------------
-- 7) Remoção do modelo multi-tenant
-- -----------------------------------------------------------------------------
CALL vagai_drop_table_if_exists('subscriptions');
CALL vagai_drop_table_if_exists('memberships');
CALL vagai_drop_table_if_exists('organizations');

-- -----------------------------------------------------------------------------
-- 8) Promoção de administradores (ajuste conforme ADMIN_EMAILS)
-- -----------------------------------------------------------------------------
-- UPDATE users SET role = 'admin' WHERE email IN ('admin@vagai.ia');

-- -----------------------------------------------------------------------------
-- 9) Limpeza dos helpers
-- -----------------------------------------------------------------------------
DROP PROCEDURE IF EXISTS vagai_add_column;
DROP PROCEDURE IF EXISTS vagai_drop_column;
DROP PROCEDURE IF EXISTS vagai_add_index;
DROP PROCEDURE IF EXISTS vagai_add_unique_index;
DROP PROCEDURE IF EXISTS vagai_drop_index;
DROP PROCEDURE IF EXISTS vagai_drop_fk;
DROP PROCEDURE IF EXISTS vagai_drop_table_if_exists;

SET FOREIGN_KEY_CHECKS = 1;
