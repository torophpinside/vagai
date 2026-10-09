#!/usr/bin/env bash
# VagAI database migrations runner (Docker).
#
# Executa as migrations via `docker exec` no container MySQL do compose,
# então não é preciso ter o cliente mysql instalado no host.
#
# Uso:
#   ./migrate.sh                        aplica as migrations pendentes de migrations/
#   ./migrate.sh --dry-run              lista o que seria aplicado, sem executar
#   ./migrate.sh --force                reaplica todas as migrations
#   ./migrate.sh migrations/001_x.sql   aplica apenas arquivos específicos
#
# Container MySQL:
#   VAGAI_MYSQL_CONTAINER (env) > `docker compose ps -q mysql` > vagai-mysql
#
# Credenciais (nesta ordem de precedência: flag > variavel de ambiente > env do
# container > default). Variáveis lidas do .env se existir:
#   DB_USER (vagai) DB_PASSWORD DB_NAME (vagai)
#   MYSQL_USER / MYSQL_PASSWORD / MYSQL_DATABASE (do container) são usadas como fallback.
#
# Flags:
#   -u, --user USER   -p, --password PASS   -d, --database DB
#   --force  --dry-run  --check  --help
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MIGRATIONS_DIR="${MIGRATIONS_DIR:-$PROJECT_ROOT/migrations}"

usage() {
  awk 'NR==1{next} /^#/{sub(/^# ?/,""); print; next} {exit}' "$0"
  exit "${1:-0}"
}

# Carrega .env (sem sobrescrever variáveis já exportadas).
if [[ -f "$PROJECT_ROOT/.env" ]]; then
  set -a
  # shellcheck disable=SC1091
  source "$PROJECT_ROOT/.env"
  set +a
fi

# Overrides por flag (vazio = não informado).
FLAG_USER=""
FLAG_PASSWORD=""
FLAG_DATABASE=""
FORCE=0
DRY_RUN=0
CHECK=0
FILES=()

while [[ $# -gt 0 ]]; do
  case "$1" in
    -u|--user)     FLAG_USER="$2"; shift 2 ;;
    -p|--password) FLAG_PASSWORD="$2"; shift 2 ;;
    -d|--database) FLAG_DATABASE="$2"; shift 2 ;;
    --force)       FORCE=1; shift ;;
    --dry-run)     DRY_RUN=1; shift ;;
    --check)       CHECK=1; shift ;;
    -h|--help)     usage 0 ;;
    *)             FILES+=("$1"); shift ;;
  esac
done

if ! command -v docker >/dev/null 2>&1; then
  echo "docker não encontrado no PATH." >&2
  exit 1
fi

if [[ ! -d "$MIGRATIONS_DIR" ]]; then
  echo "Diretório de migrations não encontrado: $MIGRATIONS_DIR" >&2
  exit 1
fi

# --- Resolve o container MySQL -------------------------------------------------
CONTAINER="${VAGAI_MYSQL_CONTAINER:-}"
if [[ -z "$CONTAINER" ]] && docker compose version >/dev/null 2>&1; then
  CONTAINER="$(cd "$PROJECT_ROOT" && docker compose ps -q mysql 2>/dev/null || true)"
fi
if [[ -z "$CONTAINER" ]]; then
  if docker inspect -f '{{.State.Running}}' vagai-mysql >/dev/null 2>&1; then
    CONTAINER="vagai-mysql"
  fi
fi
if [[ -z "$CONTAINER" ]]; then
  echo "Container MySQL não encontrado/iniciado." >&2
  echo "Suba com: docker compose up -d mysql" >&2
  echo "Ou defina VAGAI_MYSQL_CONTAINER=<nome>" >&2
  exit 1
fi
if [[ "$(docker inspect -f '{{.State.Running}}' "$CONTAINER" 2>/dev/null)" != "true" ]]; then
  echo "Container '$CONTAINER' não está em execução." >&2
  exit 1
fi

container_env() {
  docker exec "$CONTAINER" printenv "$1" 2>/dev/null || true
}

# --- Resolve credenciais (flag > env > env do container > default) -------------
DB_USER="${FLAG_USER:-${DB_USER:-$(container_env MYSQL_USER)}}"
DB_USER="${DB_USER:-vagai}"

DB_PASSWORD="${FLAG_PASSWORD:-${DB_PASSWORD:-${MYSQL_PASSWORD:-$(container_env MYSQL_PASSWORD)}}}"
DB_PASSWORD="${DB_PASSWORD:-${MYSQL_ROOT_PASSWORD:-$(container_env MYSQL_ROOT_PASSWORD)}}"

DB_NAME="${FLAG_DATABASE:-${DB_NAME:-$(container_env MYSQL_DATABASE)}}"
DB_NAME="${DB_NAME:-vagai}"

# --- Lista de arquivos ---------------------------------------------------------
if [[ ${#FILES[@]} -eq 0 ]]; then
  shopt -s nullglob
  map=("$MIGRATIONS_DIR"/*.sql)
  shopt -u nullglob
  if [[ ${#map[@]} -eq 0 ]]; then
    echo "Nenhum arquivo .sql encontrado em $MIGRATIONS_DIR" >&2
    exit 1
  fi
  while IFS= read -r f; do FILES+=("$f"); done < <(printf '%s\n' "${map[@]}" | sort)
fi

# --- Helper: mysql dentro do container -----------------------------------------
run_mysql() {
  docker exec -i -e MYSQL_PWD="$DB_PASSWORD" "$CONTAINER" \
    mysql --user="$DB_USER" "$DB_NAME" "$@"
}

already_applied() {
  local count
  count="$(docker exec -i -e MYSQL_PWD="${DB_PASSWORD:-}" "$CONTAINER" mysql \
    --user="$DB_USER" --batch --skip-column-names "$DB_NAME" \
    -e "SELECT COUNT(*) FROM schema_migrations WHERE filename = '$1';" 2>/dev/null || true)"
  [[ "$count" == "1" ]]
}

record_applied() {
  docker exec -i -e MYSQL_PWD="$DB_PASSWORD" "$CONTAINER" \
    mysql --user="$DB_USER" "$DB_NAME" \
    -e "INSERT IGNORE INTO schema_migrations (filename) VALUES ('$1');"
}

if [[ "$CHECK" -eq 1 ]]; then
  echo "Testando conexão em $CONTAINER (db=$DB_NAME, user=$DB_USER)..."
  if run_mysql -e "SELECT 1 AS ok;" >/dev/null 2>&1; then
    echo "OK: conexão bem-sucedida."
    exit 0
  fi
  echo "FALHA na conexão. Verifique usuário/senha/banco." >&2
  run_mysql -e "SELECT 1;" >&2 || true
  exit 1
fi

if [[ "$DRY_RUN" -eq 0 ]]; then
  run_mysql <<'SQL'
CREATE TABLE IF NOT EXISTS schema_migrations (
  filename   VARCHAR(255) NOT NULL,
  applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (filename)
) ENGINE=InnoDB;
SQL
fi

echo "VagAI migrations (docker)"
echo "  container: $CONTAINER"
echo "  database:  $DB_NAME"
echo "  user:      $DB_USER"
echo "  dir:       $MIGRATIONS_DIR"
echo

applied=0
skipped=0

for f in "${FILES[@]}"; do
  name="$(basename "$f")"

  if [[ "$FORCE" -eq 0 ]] && already_applied "$name"; then
    echo "  = $name (já aplicada)"
    skipped=$((skipped + 1))
    continue
  fi

  if [[ "$DRY_RUN" -eq 1 ]]; then
    echo "  > $name (dry-run)"
    continue
  fi

  echo "  > $name ..."
  if ! run_mysql < "$f"; then
    echo "FALHA ao aplicar $name. Abortando." >&2
    exit 1
  fi
  record_applied "$name"
  applied=$((applied + 1))
  echo "    ok"
done

echo
echo "Aplicadas: $applied | Puladas: $skipped"
