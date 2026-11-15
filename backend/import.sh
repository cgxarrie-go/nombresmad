#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"
export PGHOST="${PGHOST:-localhost}"
export PGPORT="${PGPORT:-5432}"
export PGUSER="${PGUSER:-postgres}"
export PGPASSWORD="${PGPASSWORD:-postgres}"
export PGDATABASE="${PGDATABASE:-nombresmad}"
echo "Applying migrations..."
psql -f migrations/migrations.sql || true
echo "To import items, sign in to the admin interface and select the JSON file."
