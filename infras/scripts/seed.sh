#!/bin/bash
# Seed 100 users into the Rice Marketplace database
# Usage: bash infras/scripts/seed.sh

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ENV_FILE="$SCRIPT_DIR/../.env"

# Default values (matching docker-compose.yml)
DB_HOST="${POSTGRES_HOST:-localhost}"
DB_PORT="${POSTGRES_PORT:-5432}"
DB_NAME="${POSTGRES_DB:-rice_marketplace}"
DB_USER="${POSTGRES_USER:-rice_user}"
DB_PASS="${POSTGRES_PASSWORD:-rice_secret_dev}"

# Load .env if exists
if [ -f "$ENV_FILE" ]; then
  export $(grep -v '^#' "$ENV_FILE" | xargs)
  DB_HOST="${POSTGRES_HOST:-$DB_HOST}"
  DB_PORT="${POSTGRES_PORT:-$DB_PORT}"
  DB_NAME="${POSTGRES_DB:-$DB_NAME}"
  DB_USER="${POSTGRES_USER:-$DB_USER}"
  DB_PASS="${POSTGRES_PASSWORD:-$DB_PASS}"
fi

echo "Seeding database: $DB_NAME on $DB_HOST:$DB_PORT..."

PGPASSWORD="$DB_PASS" psql \
  -h "$DB_HOST" \
  -p "$DB_PORT" \
  -U "$DB_USER" \
  -d "$DB_NAME" \
  -f "$SCRIPT_DIR/seed.sql"

if [ $? -eq 0 ]; then
  echo "Seed completed successfully!"
else
  echo "Seed failed!" >&2
  exit 1
fi
