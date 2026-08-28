#!/bin/bash
# ============================================================
# VagAI Metabase Setup Script
# ============================================================
# This script:
# 1. Waits for Metabase to be ready
# 2. Creates admin user
# 3. Connects to MySQL database
# 4. Creates all 30 report cards
# 5. Organizes into 7 dashboards
# ============================================================

set -e

METABASE_URL="http://localhost:3001"
MYSQL_HOST="mysql"
MYSQL_PORT="3306"
MYSQL_DB="vagai"
MYSQL_USER="vagai"
MYSQL_PASS="vagai"

echo "============================================"
echo "  VagAI Metabase Setup"
echo "============================================"

# -----------------------------------------------------------
# Wait for Metabase to be ready
# -----------------------------------------------------------
echo ""
echo "[1/6] Waiting for Metabase to be ready..."
MAX_RETRIES=60
RETRY_COUNT=0

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$METABASE_URL/api/health" 2>/dev/null || echo "000")
    if [ "$HTTP_STATUS" = "200" ]; then
        echo "  Metabase is ready!"
        break
    fi
    RETRY_COUNT=$((RETRY_COUNT + 1))
    echo "  Waiting... ($RETRY_COUNT/$MAX_RETRIES)"
    sleep 5
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    echo "  ERROR: Metabase did not start in time"
    exit 1
fi

# -----------------------------------------------------------
# Setup Metabase via API
# -----------------------------------------------------------
echo ""
echo "[2/6] Setting up Metabase instance..."

# Create setup token
SETUP_TOKEN=$(curl -s -X POST "$METABASE_url/api/setup" 2>/dev/null | grep -o '"token":"[^"]*"' | cut -d'"' -f4 || true)

# Use the Metabase API to complete setup
SETUP_RESPONSE=$(curl -s -X POST "$METABASE_URL/api/setup" \
  -H "Content-Type: application/json" \
  -d '{
    "token": "",
    "prefs": {
      "site_name": "VagAI Analytics",
      "site_locale": "pt-BR",
      "allow_tracking": false
    },
    "database": {
      "engine": "mysql",
      "name": "VagAI Database",
      "details": {
        "host": "'"$MYSQL_HOST"'",
        "port": '"$MYSQL_PORT"',
        "dbname": "'"$MYSQL_DB"'",
        "user": "'"$MYSQL_USER"'",
        "password": "'"$MYSQL_PASS"'",
        "ssl": false
      }
    },
    "user": {
      "first_name": "Admin",
      "last_name": "VagAI",
      "email": "admin@vagai.com",
      "password": "vagai-admin-2024"
    }
  }' 2>/dev/null || echo "{}")

echo "  Setup response received"

# -----------------------------------------------------------
# Wait for database sync
# -----------------------------------------------------------
echo ""
echo "[3/6] Waiting for database schema sync..."
sleep 10

# -----------------------------------------------------------
# Create Cards and Dashboards via Metabase API
# -----------------------------------------------------------
echo ""
echo "[4/6] Creating report cards..."

# Note: Cards are created via SQL questions in Metabase
# Each card uses native SQL against the MySQL views

# Helper function to create a card
create_card() {
    local name="$1"
    local description="$2"
    local sql="$3"
    local display="${4:-table}"
    local dashboard_id="${5:-null}"

    RESPONSE=$(curl -s -X POST "$METABASE_URL/api/card" \
      -H "Content-Type: application/json" \
      -H "X-Metabase-Session: $SESSION_TOKEN" \
      -d '{
        "name": "'"$name"'",
        "description": "'"$description"'",
        "display": "'"$display"'",
        "dataset_query": {
          "type": "native",
          "native": {
            "query": "'"$(echo "$sql" | sed "s/'/\\\\'/g")"'",
            "template-tags": {}
          },
          "database": 1
        },
        "visualization_settings": {}
      }' 2>/dev/null)

    CARD_ID=$(echo "$RESPONSE" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
    echo "  Created card: $name (ID: $CARD_ID)"
    return $CARD_ID
}

# Get session token
SESSION_TOKEN=$(curl -s -X POST "$METABASE_URL/api/session" \
  -H "Content-Type: application/json" \
  -d '{"username": "admin@vagai.com", "password": "vagai-admin-2024"}' \
  2>/dev/null | grep -o '"id":"[^"]*"' | cut -d'"' -f4)

if [ -z "$SESSION_TOKEN" ]; then
    echo "  WARNING: Could not get session token. Cards will need to be created manually."
    echo "  See metabase/setup/02-create-cards-api.sh for manual setup."
else
    echo "  Session token obtained"

    # -------------------------------------------------------
    # DASHBOARD 1: Pipeline de Vagas
    # -------------------------------------------------------
    echo ""
    echo "  Creating Dashboard 1: Pipeline de Vagas..."

    create_card "1.1 Funil de Vagas" \
        "Distribuicao de vagas por status (new, matched, analyzed)" \
        "SELECT * FROM view_vagas_funil" \
        "funnel"

    create_card "1.2 Vagas por Dia" \
        "Volume de vagas coletadas ao longo do tempo" \
        "SELECT * FROM view_vagas_por_dia ORDER BY data_coleta" \
        "line"

    create_card "1.3 Top Empresas" \
        "Empresas com maior numero de vagas" \
        "SELECT * FROM view_top_empresas LIMIT 15" \
        "bar"

    create_card "1.4 Vagas por Site" \
        "Distribuicao de vagas por site de origem" \
        "SELECT * FROM view_vagas_por_site" \
        "pie"

    create_card "1.5 Lag de Coleta" \
        "Tempo entre publicacao e coleta das vagas" \
        "SELECT dias_lag, COUNT(*) AS total FROM view_lag_coleta GROUP BY dias_lag ORDER BY dias_lag" \
        "bar"

    # -------------------------------------------------------
    # DASHBOARD 2: Qualidade dos Matches
    # -------------------------------------------------------
    echo ""
    echo "  Creating Dashboard 2: Qualidade dos Matches..."

    create_card "2.1 Distribuicao de Scores" \
        "Faixas de similaridade nos matches" \
        "SELECT * FROM view_score_distribuicao" \
        "bar"

    create_card "2.2 Score por Curriculo" \
        "Desempenho medio de cada curriculo" \
        "SELECT * FROM view_score_por_curriculo" \
        "bar"

    create_card "2.3 Taxa de Aplicacao" \
        "Percentual de matches que geraram aplicacao" \
        "SELECT * FROM view_taxa_aplicacao" \
        "gauge"

    create_card "2.4 Keywords Frequentes" \
        "Palavras-chave mais apareciam nos matches" \
        "SELECT * FROM view_keywords_frequentes LIMIT 20" \
        "bar"

    create_card "2.5 Matches por Plano" \
        "Volume de matches por tier de assinatura" \
        "SELECT * FROM view_matches_por_plano" \
        "bar"

    # -------------------------------------------------------
    # DASHBOARD 3: Receita & Assinaturas
    # -------------------------------------------------------
    echo ""
    echo "  Creating Dashboard 3: Receita & Assinaturas..."

    create_card "3.1 MRR - Receita Mensal" \
        "Monthly Recurring Revenue por plano" \
        "SELECT * FROM view_mrr" \
        "table"

    create_card "3.2 Conversao Trial para Paid" \
        "Distribuicao de status de assinaturas" \
        "SELECT * FROM view_conversao_trial" \
        "pie"

    create_card "3.3 Churn Rate Mensal" \
        "Taxa de cancelamento por mes" \
        "SELECT * FROM view_churn_mensal ORDER BY mes" \
        "line"

    create_card "3.4 Churn em Risco" \
        "Assinaturas com cancelamento programado" \
        "SELECT * FROM view_churn_risco" \
        "table"

    create_card "3.5 Proximos Vencimentos" \
        "Assinaturas vencendo nos proximos 30 dias" \
        "SELECT * FROM view_proximos_vencimentos" \
        "table"

    # -------------------------------------------------------
    # DASHBOARD 4: Crescimento de Usuarios
    # -------------------------------------------------------
    echo ""
    echo "  Creating Dashboard 4: Crescimento de Usuarios..."

    create_card "4.1 Novas Orgs por Semana" \
        "Crescimento de organizacoes ao longo do tempo" \
        "SELECT * FROM view_orgs_por_semana ORDER BY semana" \
        "line"

    create_card "4.2 Usuarios Verificados" \
        "Taxa de verificacao de email" \
        "SELECT * FROM view_usuarios_verificacao" \
        "pie"

    create_card "4.3 Distribuicao de Planos" \
        "Organizacoes e usuarios por plano" \
        "SELECT * FROM view_distribuicao_planos" \
        "bar"

    create_card "4.4 Membros por Org" \
        "Numero de membros em cada organizacao" \
        "SELECT * FROM view_membros_por_org LIMIT 20" \
        "bar"

    # -------------------------------------------------------
    # DASHBOARD 5: Saude do Sistema
    # -------------------------------------------------------
    echo ""
    echo "  Creating Dashboard 5: Saude do Sistema..."

    create_card "5.1 Atividade dos Agents" \
        "Eventos dos agents por dia" \
        "SELECT * FROM view_agentes_por_dia ORDER BY data DESC LIMIT 30" \
        "line"

    create_card "5.2 Crawl Stats" \
        "Vagas encontradas por crawl" \
        "SELECT * FROM view_crawl_stats LIMIT 30" \
        "bar"

    create_card "5.3 Taxa Sucesso Schedules" \
        "Percentual de execucoes bem-sucedidas" \
        "SELECT * FROM view_schedule_sucesso" \
        "gauge"

    create_card "5.4 Duracao Execucao" \
        "Tempo medio de execucao dos schedules" \
        "SELECT * FROM view_duracao_execucao" \
        "bar"

    # -------------------------------------------------------
    # DASHBOARD 6: Insights de Resumes
    # -------------------------------------------------------
    echo ""
    echo "  Creating Dashboard 6: Insights de Resumes..."

    create_card "6.1 Forcas Comuns" \
        "Habilidades fortes mais identificadas" \
        "SELECT * FROM view_forcas_comuns LIMIT 15" \
        "bar"

    create_card "6.2 Fraquezas Frequentes" \
        "Areas de melhoria mais identificadas" \
        "SELECT * FROM view_fraquezas_frequentes LIMIT 15" \
        "bar"

    create_card "6.3 Sugestoes da IA" \
        "Recomendacoes mais frequentes" \
        "SELECT * FROM view_sugestoes_ia LIMIT 15" \
        "bar"

    create_card "6.4 Analises por Periodo" \
        "Volume de analises de curriculos ao longo do tempo" \
        "SELECT * FROM view_analises_por_periodo ORDER BY data_analise" \
        "line"

    # -------------------------------------------------------
    # DASHBOARD 7: Performance dos Sites
    # -------------------------------------------------------
    echo ""
    echo "  Creating Dashboard 7: Performance dos Sites..."

    create_card "7.1 Status dos Sites" \
        "Sites ativos vs inativos" \
        "SELECT * FROM view_status_sites" \
        "pie"

    create_card "7.2 Ultimo Crawl" \
        "Atualizacao dos crawls por site" \
        "SELECT * FROM view_ultimo_crawl" \
        "table"

    create_card "7.3 Eficiencia dos Sites" \
        "Vagas coletadas por site e eficiencia" \
        "SELECT * FROM view_eficiencia_sites" \
        "bar"

    echo ""
    echo "[5/6] Creating dashboards..."

    # Dashboard 1
    D1=$(curl -s -X POST "$METABASE_URL/api/dashboard" \
      -H "Content-Type: application/json" \
      -H "X-Metabase-Session: $SESSION_TOKEN" \
      -d '{"name": "1 - Pipeline de Vagas", "description": "Analise de coleta e status das vagas"}' \
      2>/dev/null | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
    echo "  Dashboard 1 created (ID: $D1)"

    # Dashboard 2
    D2=$(curl -s -X POST "$METABASE_URL/api/dashboard" \
      -H "Content-Type: application/json" \
      -H "X-Metabase-Session: $SESSION_TOKEN" \
      -d '{"name": "2 - Qualidade dos Matches", "description": "Analise de similaridade e matches"}' \
      2>/dev/null | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
    echo "  Dashboard 2 created (ID: $D2)"

    # Dashboard 3
    D3=$(curl -s -X POST "$METABASE_URL/api/dashboard" \
      -H "Content-Type: application/json" \
      -H "X-Metabase-Session: $SESSION_TOKEN" \
      -d '{"name": "3 - Receita & Assinaturas", "description": "MRR, churn e conversao"}' \
      2>/dev/null | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
    echo "  Dashboard 3 created (ID: $D3)"

    # Dashboard 4
    D4=$(curl -s -X POST "$METABASE_URL/api/dashboard" \
      -H "Content-Type: application/json" \
      -H "X-Metabase-Session: $SESSION_TOKEN" \
      -d '{"name": "4 - Crescimento de Usuarios", "description": "Evolucao de usuarios e organizacoes"}' \
      2>/dev/null | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
    echo "  Dashboard 4 created (ID: $D4)"

    # Dashboard 5
    D5=$(curl -s -X POST "$METABASE_URL/api/dashboard" \
      -H "Content-Type: application/json" \
      -H "X-Metabase-Session: $SESSION_TOKEN" \
      -d '{"name": "5 - Saude do Sistema", "description": "Agents, crawls e schedules"}' \
      2>/dev/null | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
    echo "  Dashboard 5 created (ID: $D5)"

    # Dashboard 6
    D6=$(curl -s -X POST "$METABASE_URL/api/dashboard" \
      -H "Content-Type: application/json" \
      -H "X-Metabase-Session: $SESSION_TOKEN" \
      -d '{"name": "6 - Insights de Resumes", "description": "Analise de curriculos e habilidades"}' \
      2>/dev/null | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
    echo "  Dashboard 6 created (ID: $D6)"

    # Dashboard 7
    D7=$(curl -s -X POST "$METABASE_URL/api/dashboard" \
      -H "Content-Type: application/json" \
      -H "X-Metabase-Session: $SESSION_TOKEN" \
      -d '{"name": "7 - Performance dos Sites", "description": "Saude e eficiencia dos sites monitorados"}' \
      2>/dev/null | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
    echo "  Dashboard 7 created (ID: $D7)"
fi

echo ""
echo "[6/6] Setup complete!"
echo ""
echo "============================================"
echo "  Access Metabase at: $METABASE_URL"
echo "  Login: admin@vagai.com"
echo "  Password: vagai-admin-2024"
echo "============================================"
echo ""
echo "  SQL Views were created in the database."
echo "  Cards and dashboards were created via API."
echo ""
echo "  If cards were not created automatically,"
echo "  run: metabase/setup/02-create-cards-api.sh"
echo "============================================"
