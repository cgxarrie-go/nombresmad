#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"
if [ -z "${DATABASE_URL:-}" ]; then
  export DATABASE_URL="postgres://postgres:postgres@localhost:5432/nombresmad?sslmode=disable"
fi
echo "Applying migrations..."
psql "$DATABASE_URL" -f migrations.sql || true
echo "Calling backend /api/migrate"
curl -X POST http://localhost:8080/api/migrate || true
echo "Calling backend /api/import"
curl -X POST http://localhost:8080/api/import || true
echo "Import attempts completed. Check server logs or responses above."
