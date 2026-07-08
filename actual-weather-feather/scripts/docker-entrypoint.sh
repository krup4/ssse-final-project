#!/bin/sh
set -e

# Wait for PostgreSQL to accept connections before applying migrations.
if [ -n "$DATABASE_URL" ]; then
  echo "Waiting for PostgreSQL at ${POSTGRES_HOST:-postgres}:${POSTGRES_PORT:-5432}..."
  until pg_isready -h "${POSTGRES_HOST:-postgres}" -p "${POSTGRES_PORT:-5432}" -U "${POSTGRES_USER:-postgres}" >/dev/null 2>&1; do
    sleep 1
  done

  echo "Applying database migrations..."
  python -m app.core.migrations
fi

# Hand control to the command supplied via docker run / compose (defaults to uvicorn).
exec "$@"
