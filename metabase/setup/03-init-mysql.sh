#!/bin/bash
# ============================================================
# VagAI Metabase - Init Script
# ============================================================
# This script is executed by docker-entrypoint-initdb.d
# It creates all SQL views needed for Metabase dashboards.
# ============================================================

echo "============================================"
echo "  VagAI: Creating Metabase views..."
echo "============================================"

# Wait for MySQL to be fully ready
sleep 5

# Execute views SQL
if [ -f /docker-entrypoint-initdb.d/01-create-views.sql ]; then
    echo "Executing views SQL..."
    mysql -u vagai -pvagai vagai < /docker-entrypoint-initdb.d/01-create-views.sql
    echo "Views created successfully!"
else
    echo "WARNING: 01-create-views.sql not found"
fi

echo "============================================"
echo "  Metabase views setup complete"
echo "============================================"
