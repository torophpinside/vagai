#!/bin/bash
# ============================================================
# VagAI Metabase - Dashboard 8: Mercado de Tecnologia
# ============================================================
# Cria cards + dashboard com métricas de demanda de tecnologias
# e gap com o currículo (views de 04-create-skill-views.sql).
#
# Pré-requisitos:
#   1. Views aplicadas: docker cp metabase/setup/04-create-skill-views.sql vagai-mysql:/tmp/
#      docker exec vagai-mysql sh -c 'mysql -u vagai -pvagai vagai < /tmp/04-create-skill-views.sql'
#   2. Metabase no ar em http://localhost:3001 com setup completo
#
# Uso: bash metabase/setup/05-create-skill-cards-api.sh
# ============================================================

set -e

METABASE_URL="http://localhost:3001"
EMAIL="torophpinside@vagai.ia"
PASSWORD="admin1234"

echo "============================================"
echo "  VagAI Metabase - Mercado de Tecnologia"
echo "============================================"

echo ""
echo "[1/3] Authenticating..."

SESSION=$(curl -s -X POST "$METABASE_URL/api/session" \
  -H "Content-Type: application/json" \
  -d "{\"username\": \"$EMAIL\", \"password\": \"$PASSWORD\"}" \
  2>/dev/null)

SESSION_TOKEN=$(echo "$SESSION" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

if [ -z "$SESSION_TOKEN" ]; then
    echo "  ERROR: Could not authenticate. Check credentials."
    exit 1
fi

echo "  Authenticated successfully"

echo ""
echo "[2/3] Getting database ID..."

DB_INFO=$(curl -s "$METABASE_URL/api/database" \
  -H "X-Metabase-Session: $SESSION_TOKEN" 2>/dev/null)

DB_ID=$(echo "$DB_INFO" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)

if [ -z "$DB_ID" ]; then
    echo "  ERROR: No database found."
    exit 1
fi

echo "  Database ID: $DB_ID"

echo ""
echo "[3/3] Creating cards..."

extract_id() {
    # Extrai o id raiz do objeto JSON retornado pela API
    echo "$1" | python3 -c 'import json,sys; print(json.load(sys.stdin)["id"])' 2>/dev/null
}

create_card() {
    local name="$1"
    local desc="$2"
    local sql="$3"
    local display="${4:-table}"

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

    CARD_ID=$(extract_id "$RESPONSE")
    if [ -n "$CARD_ID" ]; then
        echo "  OK $name (ID: $CARD_ID)" >&2
    else
        echo "  FAIL $name: $(echo "$RESPONSE" | head -c 200)" >&2
    fi
    echo "$CARD_ID"
}

CARD_GAP=$(create_card "8.1 O Que Estudar Primeiro" \
    "Gap entre demandas das vagas e o curriculo mais recente" \
    "SELECT tecnologia, vagas_exigindo, pct_ultimos_90d, recomendacao, prioridade_estudo FROM view_skill_gap ORDER BY prioridade_estudo DESC LIMIT 25" \
    "table")

CARD_TOP=$(create_card "8.2 Top Tecnologias Demandadas" \
    "Tecnologias mais citadas nas vagas coletadas" \
    "SELECT tecnologia AS Tec, total_vagas AS Vagas, pct_das_vagas AS Percentual FROM view_tech_demandas ORDER BY total_vagas DESC LIMIT 20" \
    "bar")

CARD_30D=$(create_card "8.3 Demanda Ultimos 30 Dias" \
    "Momento atual do mercado" \
    "SELECT tecnologia AS Tec, total_vagas_30d AS Vagas, pct_das_vagas_30d AS Percentual FROM view_tech_demandas_30d ORDER BY total_vagas_30d DESC LIMIT 20" \
    "bar")

CARD_MES=$(create_card "8.4 Tendencia Mensal" \
    "Evolucao da demanda por mes" \
    "SELECT mes, tecnologia AS Tec, total_vagas AS Vagas FROM view_tech_demandas_mes ORDER BY mes, Vagas DESC" \
    "line")

echo ""
echo "Creating dashboard..."

RESPONSE=$(curl -s -X POST "$METABASE_URL/api/dashboard" \
  -H "Content-Type: application/json" \
  -H "X-Metabase-Session: $SESSION_TOKEN" \
  -d '{"name": "8 - Mercado de Tecnologia", "description": "Quais tecnologias estudar com base nas vagas"}' 2>/dev/null)

DASH_ID=$(extract_id "$RESPONSE")
echo "  Dashboard: 8 - Mercado de Tecnologia (ID: $DASH_ID)"

if [ -n "$DASH_ID" ]; then
    # API moderna: PUT /api/dashboard/:id com dashcards de IDs negativos
    curl -s -X PUT "$METABASE_URL/api/dashboard/$DASH_ID" \
      -H "Content-Type: application/json" \
      -H "X-Metabase-Session: $SESSION_TOKEN" \
      -d "{
        \"dashcards\": [
          {\"id\": -1, \"card_id\": $CARD_GAP, \"row\": 0,  \"col\": 0, \"size_x\": 18, \"size_y\": 8},
          {\"id\": -2, \"card_id\": $CARD_TOP, \"row\": 8,  \"col\": 0, \"size_x\": 9,  \"size_y\": 6},
          {\"id\": -3, \"card_id\": $CARD_30D, \"row\": 8,  \"col\": 9, \"size_x\": 9,  \"size_y\": 6},
          {\"id\": -4, \"card_id\": $CARD_MES, \"row\": 14, \"col\": 0, \"size_x\": 18, \"size_y\": 8}
        ]
      }" > /dev/null 2>&1 || true
    echo "  Cards adicionados ao dashboard"
fi

echo ""
echo "============================================"
echo "  Setup complete!"
echo "  Access: $METABASE_URL"
echo "============================================"
