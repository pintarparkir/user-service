#!/usr/bin/env bash
# Apply (or roll back) DB migrations using golang-migrate.
# Usage: ./scripts/migrate.sh up | down | force <version>

set -euo pipefail

DB_URL="${DB_URL:-postgres://postgres:postgres@localhost:5432/parkirpintar?sslmode=disable}"
MIGRATIONS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../data/migrations" && pwd 2>/dev/null || true)"

# Fall back to applying init.sql if no migrations dir exists.
if [[ -z "$MIGRATIONS_DIR" || ! -d "$MIGRATIONS_DIR" ]]; then
  echo "→ no migrations dir; applying data/init.sql + data/seed.sql via psql"
  psql "$DB_URL" -f data/init.sql
  psql "$DB_URL" -f data/seed.sql
  exit 0
fi

migrate -database "$DB_URL" -path "$MIGRATIONS_DIR" "$@"
