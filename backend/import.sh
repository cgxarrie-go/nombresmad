#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"
export PGHOST="${PGHOST:-localhost}"
export PGPORT="${PGPORT:-5432}"
export PGUSER="${PGUSER:-postgres}"
export PGPASSWORD="${PGPASSWORD:-postgres}"
export PGDATABASE="${PGDATABASE:-nombresmad}"
echo "Applying migrations..."
psql -f migrations.sql || true
echo "Calling backend /api/migrate"
curl -X POST http://localhost:8080/api/migrate || true
echo "Calling backend /api/import"
curl -X POST http://localhost:8080/api/import || true
echo "Import attempts completed. Check server logs or responses above."
