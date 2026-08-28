#!/bin/bash
# ============================================================
# VagAI Metabase - Manual Card Creation Script
# ============================================================
# Run this if the automatic setup didn't create cards.
# Requires: curl, jq (optional)
#
# Usage:
#   1. Start Metabase: docker compose up metabase -d
#   2. Complete initial setup in browser (http://localhost:3001)
#   3. Run this script: bash metabase/setup/02-create-cards-api.sh
# ============================================================

set -e

METABASE_URL="http://localhost:3001"
EMAIL="torophpinside@vagai.ia"
PASSWORD="admin1234"

echo "============================================"
echo "  VagAI Metabase - Card Creation"
echo "============================================"

# -----------------------------------------------------------
# Authenticate
# -----------------------------------------------------------
echo ""
echo "[1/3] Authenticating..."

SESSION=$(curl -s -X POST "$METABASE_URL/api/session" \
  -H "Content-Type: application/json" \
  -d "{\"username\": \"$EMAIL\", \"password\": \"$PASSWORD\"}" \
  2>/dev/null)

SESSION_TOKEN=$(echo "$SESSION" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

if [ -z "$SESSION_TOKEN" ]; then
    echo "  ERROR: Could not authenticate. Check credentials."
    echo "  Make sure you completed the Metabase setup wizard first."
    exit 1
fi

echo "  Authenticated successfully"

# -----------------------------------------------------------
# Get database ID
# -----------------------------------------------------------
echo ""
echo "[2/3] Getting database ID..."

DB_INFO=$(curl -s "$METABASE_URL/api/database" \
  -H "X-Metabase-Session: $SESSION_TOKEN" 2>/dev/null)

DB_ID=$(echo "$DB_INFO" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)

if [ -z "$DB_ID" ]; then
    echo "  ERROR: No database found. Please add the MySQL database first."
    exit 1
fi

echo "  Database ID: $DB_ID"

# -----------------------------------------------------------
# Create cards
# -----------------------------------------------------------
echo ""
echo "[3/3] Creating 30 report cards..."

create_card() {
    local name="$1"
    local desc="$2"
    local sql="$3"
    local display="${4:-table}"

    # Escape SQL for JSON
    local escaped_sql=$(echo "$sql" | tr '\n' ' ' | sed 's/"/\\"/g')

    RESPONSE=$(curl -s -X POST "$METABASE_URL/api/card" \
      -H "Content-Type: application/json" \
      -H "X-Metabase-Session: $SESSION_TOKEN" \
      -d "{
        \"name\": \"$name\",
        \"description\": \"$desc\",
        \"display\": \"$display\",
        \"dataset_query\": {
          \"type\": \"native\",
          \"native\": {
            \"query\": \"$escaped_sql\",
            \"template_tags\": {}
          },
          \"database\": $DB_ID
        },
        \"visualization_settings\": {}
      }" 2>/dev/null)

    CARD_ID=$(echo "$RESPONSE" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
    if [ -n "$CARD_ID" ]; then
        echo "  ✓ $name (ID: $CARD_ID)"
    else
        echo "  ✗ $name - FAILED"
    fi
    echo "$CARD_ID"
}

# DASHBOARD 1: Pipeline de Vagas
echo ""
echo "  --- Dashboard 1: Pipeline de Vagas ---"
create_card "1.1 Funil de Vagas" \
    "Distribuicao de vagas por status" \
    "SELECT status, COUNT(*) AS total, ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER(), 2) AS percentual FROM jobs WHERE deleted_at IS NULL GROUP BY status" \
    "funnel"

create_card "1.2 Vagas por Dia" \
    "Volume diario de vagas coletadas" \
    "SELECT DATE(collected_at) AS data_coleta, COUNT(*) AS total_vagas FROM jobs WHERE deleted_at IS NULL GROUP BY DATE(collected_at) ORDER BY data_coleta" \
    "line"

create_card "1.3 Top Empresas" \
    "Empresas com mais vagas" \
    "SELECT company AS empresa, COUNT(*) AS total_vagas FROM jobs WHERE deleted_at IS NULL AND company IS NOT NULL AND company != '' GROUP BY company ORDER BY total_vagas DESC LIMIT 15" \
    "bar"

create_card "1.4 Vagas por Site" \
    "Distribuicao por site de origem" \
    "SELECT s.name AS site, COUNT(j.id) AS total_vagas FROM jobs j LEFT JOIN sites s ON j.site_id = s.id WHERE j.deleted_at IS NULL GROUP BY s.id, s.name ORDER BY total_vagas DESC" \
    "pie"

create_card "1.5 Lag de Coleta" \
    "Dias entre publicacao e coleta" \
    "SELECT DATEDIFF(j.collected_at, j.posted_date) AS dias_lag, COUNT(*) AS total FROM jobs j WHERE j.deleted_at IS NULL AND j.posted_date IS NOT NULL GROUP BY dias_lag ORDER BY dias_lag" \
    "bar"

# DASHBOARD 2: Qualidade dos Matches
echo ""
echo "  --- Dashboard 2: Qualidade dos Matches ---"
create_card "2.1 Distribuicao de Scores" \
    "Faixas de similaridade" \
    "SELECT CASE WHEN similarity_score >= 80 THEN '80-100' WHEN similarity_score >= 60 THEN '60-79' WHEN similarity_score >= 40 THEN '40-59' WHEN similarity_score >= 20 THEN '20-39' ELSE '0-19' END AS faixa, COUNT(*) AS total, ROUND(AVG(similarity_score), 2) AS score_medio FROM matches GROUP BY faixa ORDER BY score_medio DESC" \
    "bar"

create_card "2.2 Score por Curriculo" \
    "Desempenho por curriculo" \
    "SELECT r.name AS curriculo, COUNT(m.id) AS matches, ROUND(AVG(m.similarity_score), 2) AS score_medio, ROUND(MAX(m.similarity_score), 2) AS melhor_score FROM matches m JOIN resumes r ON m.resume_id = r.id GROUP BY r.id, r.name ORDER BY score_medio DESC" \
    "bar"

create_card "2.3 Taxa de Aplicacao" \
    "Percentual de aplicacoes" \
    "SELECT COUNT(*) AS total, SUM(CASE WHEN applied = 1 THEN 1 ELSE 0 END) AS aplicados, ROUND(SUM(CASE WHEN applied = 1 THEN 1 ELSE 0 END) * 100.0 / COUNT(*), 2) AS taxa_pct FROM matches" \
    "gauge"

create_card "2.4 Keywords Frequentes" \
    "Palavras-chave nos matches" \
    "SELECT keyword, COUNT(*) AS frequencia FROM matches m, JSON_TABLE(m.keywords_matched, '\$[*]' COLUMNS (keyword VARCHAR(100) PATH '\$')) AS jt WHERE m.keywords_matched IS NOT NULL AND JSON_VALID(m.keywords_matched) GROUP BY keyword ORDER BY frequencia DESC LIMIT 20" \
    "bar"

create_card "2.5 Matches por Plano" \
    "Matches por tier" \
    "SELECT p.name AS plano, COUNT(m.id) AS total_matches, ROUND(AVG(m.similarity_score), 2) AS score_medio FROM matches m JOIN organizations o ON m.organization_id = o.id JOIN subscriptions s ON s.organization_id = o.id JOIN plans p ON s.plan_id = p.id WHERE o.deleted_at IS NULL AND s.deleted_at IS NULL GROUP BY p.id, p.name" \
    "bar"

# DASHBOARD 3: Receita & Assinaturas
echo ""
echo "  --- Dashboard 3: Receita & Assinaturas ---"
create_card "3.1 MRR" \
    "Monthly Recurring Revenue" \
    "SELECT p.name AS plano, COUNT(s.id) AS assinaturas, ROUND(p.price_monthly / 100.0, 2) AS preco_rs, ROUND(p.price_monthly / 100.0 * COUNT(s.id), 2) AS mrr_rs FROM subscriptions s JOIN plans p ON s.plan_id = p.id WHERE s.status = 'active' GROUP BY p.id, p.name, p.price_monthly" \
    "table"

create_card "3.2 Conversao Trial" \
    "Distribuicao de status" \
    "SELECT CASE WHEN status = 'trial' THEN 'Trial' WHEN status = 'active' THEN 'Ativo' WHEN status = 'past_due' THEN 'Pendente' WHEN status = 'canceled' THEN 'Cancelado' WHEN status = 'expired' THEN 'Expirado' ELSE status END AS status_ass, COUNT(*) AS total FROM subscriptions WHERE deleted_at IS NULL GROUP BY status ORDER BY FIELD(status, 'active', 'trial', 'past_due', 'canceled', 'expired')" \
    "pie"

create_card "3.3 Churn Mensal" \
    "Taxa de cancelamento" \
    "SELECT DATE_FORMAT(created_at, '%Y-%m') AS mes, COUNT(*) AS total, SUM(CASE WHEN status IN ('canceled', 'expired') THEN 1 ELSE 0 END) AS canceladas, ROUND(SUM(CASE WHEN status IN ('canceled', 'expired') THEN 1 ELSE 0 END) * 100.0 / COUNT(*), 2) AS churn_pct FROM subscriptions WHERE deleted_at IS NULL GROUP BY mes ORDER BY mes" \
    "line"

create_card "3.4 Churn Risco" \
    "Cancelamentos programados" \
    "SELECT o.name AS org, p.name AS plano, s.current_period_end AS vencimento, DATEDIFF(s.current_period_end, NOW()) AS dias_restantes FROM subscriptions s JOIN organizations o ON s.organization_id = o.id JOIN plans p ON s.plan_id = p.id WHERE s.cancel_at_period_end = 1 AND s.status = 'active' AND s.deleted_at IS NULL ORDER BY s.current_period_end" \
    "table"

create_card "3.5 Proximos Vencimentos" \
    "Vencimentos em 30 dias" \
    "SELECT o.name AS org, p.name AS plano, s.current_period_end AS vencimento, DATEDIFF(s.current_period_end, NOW()) AS dias_ate_vencer FROM subscriptions s JOIN organizations o ON s.organization_id = o.id JOIN plans p ON s.plan_id = p.id WHERE s.status IN ('active', 'trial') AND s.current_period_end IS NOT NULL AND s.current_period_end >= NOW() AND s.current_period_end <= DATE_ADD(NOW(), INTERVAL 30 DAY) AND s.deleted_at IS NULL ORDER BY s.current_period_end" \
    "table"

# DASHBOARD 4: Crescimento de Usuarios
echo ""
echo "  --- Dashboard 4: Crescimento de Usuarios ---"
create_card "4.1 Novas Orgs/Semana" \
    "Crescimento semanal" \
    "SELECT MIN(DATE(created_at)) AS semana, COUNT(*) AS novas_orgs FROM organizations WHERE deleted_at IS NULL GROUP BY YEARWEEK(created_at, 1) ORDER BY semana" \
    "line"

create_card "4.2 Verificacao Email" \
    "Usuarios verificados" \
    "SELECT CASE WHEN email_verified_at IS NOT NULL THEN 'Verificado' ELSE 'Nao Verificado' END AS status, COUNT(*) AS total FROM users WHERE deleted_at IS NULL GROUP BY status" \
    "pie"

create_card "4.3 Planos" \
    "Distribuicao de planos" \
    "SELECT o.plan AS plano, COUNT(DISTINCT o.id) AS orgs, COUNT(DISTINCT m.user_id) AS usuarios FROM organizations o LEFT JOIN memberships m ON o.id = m.organization_id WHERE o.deleted_at IS NULL GROUP BY o.plan ORDER BY orgs DESC" \
    "bar"

create_card "4.4 Membros/Org" \
    "Membros por organizacao" \
    "SELECT o.name AS org, o.plan AS plano, COUNT(m.id) AS membros FROM organizations o LEFT JOIN memberships m ON o.id = m.organization_id WHERE o.deleted_at IS NULL GROUP BY o.id, o.name, o.plan ORDER BY membros DESC LIMIT 20" \
    "bar"

# DASHBOARD 5: Saude do Sistema
echo ""
echo "  --- Dashboard 5: Saude do Sistema ---"
create_card "5.1 Atividade Agents" \
    "Eventos por dia" \
    "SELECT DATE(created_at) AS data, agent_name AS agente, COUNT(*) AS eventos FROM agent_logs GROUP BY data, agente ORDER BY data DESC LIMIT 30" \
    "line"

create_card "5.2 Crawl Stats" \
    "Vagas por crawl" \
    "SELECT DATE(created_at) AS data, JSON_UNQUOTE(JSON_EXTRACT(details, '\$.site')) AS site, CAST(JSON_EXTRACT(details, '\$.jobs') AS UNSIGNED) AS vagas FROM agent_logs WHERE agent_name = 'crawler' AND action = 'crawl_completed' AND details IS NOT NULL ORDER BY data DESC LIMIT 30" \
    "bar"

create_card "5.3 Sucesso Schedules" \
    "Taxa de sucesso" \
    "SELECT s.name AS schedule, COUNT(sl.id) AS execucoes, ROUND(SUM(CASE WHEN sl.status = 'success' THEN 1 ELSE 0 END) * 100.0 / COUNT(*), 2) AS taxa_sucesso FROM schedules s LEFT JOIN schedule_logs sl ON s.id = sl.schedule_id GROUP BY s.id, s.name" \
    "gauge"

create_card "5.4 Duracao Execucao" \
    "Tempo medio" \
    "SELECT s.name AS schedule, COUNT(sl.id) AS execucoes, ROUND(AVG(TIMESTAMPDIFF(SECOND, sl.started_at, sl.finished_at)), 2) AS duracao_media_seg FROM schedules s JOIN schedule_logs sl ON s.id = sl.schedule_id WHERE sl.finished_at IS NOT NULL GROUP BY s.id, s.name" \
    "bar"

# DASHBOARD 6: Insights de Resumes
echo ""
echo "  --- Dashboard 6: Insights de Resumes ---"
create_card "6.1 Forcas Comuns" \
    "Habilidades fortes" \
    "SELECT forca, COUNT(*) AS frequencia FROM resume_analyses, JSON_TABLE(strengths, '\$[*]' COLUMNS (forca VARCHAR(255) PATH '\$')) AS jt WHERE strengths IS NOT NULL AND JSON_VALID(strengths) GROUP BY forca ORDER BY frequencia DESC LIMIT 15" \
    "bar"

create_card "6.2 Fraquezas Frequentes" \
    "Areas de melhoria" \
    "SELECT fraqueza, COUNT(*) AS frequencia FROM resume_analyses, JSON_TABLE(weaknesses, '\$[*]' COLUMNS (fraqueza VARCHAR(255) PATH '\$')) AS jt WHERE weaknesses IS NOT NULL AND JSON_VALID(weaknesses) GROUP BY fraqueza ORDER BY frequencia DESC LIMIT 15" \
    "bar"

create_card "6.3 Sugestoes IA" \
    "Recomendacoes" \
    "SELECT sugestao, COUNT(*) AS frequencia FROM resume_analyses, JSON_TABLE(suggestions, '\$[*]' COLUMNS (sugestao VARCHAR(255) PATH '\$')) AS jt WHERE suggestions IS NOT NULL AND JSON_VALID(suggestions) GROUP BY sugestao ORDER BY frequencia DESC LIMIT 15" \
    "bar"

create_card "6.4 Analises/Periodo" \
    "Volume de analises" \
    "SELECT DATE(created_at) AS data, COUNT(*) AS analises, COUNT(DISTINCT resume_id) AS curriculos FROM resume_analyses GROUP BY data ORDER BY data" \
    "line"

# DASHBOARD 7: Performance dos Sites
echo ""
echo "  --- Dashboard 7: Performance dos Sites ---"
create_card "7.1 Status Sites" \
    "Ativos vs inativos" \
    "SELECT CASE WHEN active = 1 THEN 'Ativo' ELSE 'Inativo' END AS status, COUNT(*) AS total FROM sites GROUP BY active" \
    "pie"

create_card "7.2 Ultimo Crawl" \
    "Atualizacao dos crawls" \
    "SELECT name AS site, url, last_crawl AS ultimo_crawl, CASE WHEN last_crawl IS NULL THEN 'Nunca' WHEN last_crawl < DATE_SUB(NOW(), INTERVAL 7 DAY) THEN 'Desatualizado' ELSE 'Atualizado' END AS status_freshness FROM sites ORDER BY last_crawl DESC" \
    "table"

create_card "7.3 Eficiencia Sites" \
    "Vagas por site" \
    "SELECT s.name AS site, COUNT(j.id) AS total_jobs, COUNT(DISTINCT DATE(j.collected_at)) AS dias_ativos, CASE WHEN COUNT(DISTINCT DATE(j.collected_at)) > 0 THEN ROUND(COUNT(j.id) * 1.0 / COUNT(DISTINCT DATE(j.collected_at)), 1) ELSE 0 END AS jobs_por_dia FROM sites s LEFT JOIN jobs j ON s.id = j.site_id AND j.deleted_at IS NULL GROUP BY s.id, s.name ORDER BY total_jobs DESC" \
    "bar"

# -----------------------------------------------------------
# Create Dashboards and assign cards
# -----------------------------------------------------------
echo ""
echo "Creating dashboards and assigning cards..."

create_dashboard() {
    local name="$1"
    local desc="$2"

    RESPONSE=$(curl -s -X POST "$METABASE_URL/api/dashboard" \
      -H "Content-Type: application/json" \
      -H "X-Metabase-Session: $SESSION_TOKEN" \
      -d "{\"name\": \"$name\", \"description\": \"$desc\"}" 2>/dev/null)

    DASH_ID=$(echo "$RESPONSE" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
    echo "  Dashboard: $name (ID: $DASH_ID)"
    echo "$DASH_ID"
}

create_dashboard "1 - Pipeline de Vagas" "Analise de coleta e status das vagas"
create_dashboard "2 - Qualidade dos Matches" "Analise de similaridade e matches"
create_dashboard "3 - Receita & Assinaturas" "MRR, churn e conversao"
create_dashboard "4 - Crescimento de Usuarios" "Evolucao de usuarios e organizacoes"
create_dashboard "5 - Saude do Sistema" "Agents, crawls e schedules"
create_dashboard "6 - Insights de Resumes" "Analise de curriculos e habilidades"
create_dashboard "7 - Performance dos Sites" "Saude e eficiencia dos sites"

echo ""
echo "============================================"
echo "  Setup complete!"
echo "  Access: $METABASE_URL"
echo "============================================"
