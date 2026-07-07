#!/bin/sh
set -e

if [ -n "$DATABASE_URL" ]; then
  echo "Applying database migrations..."
  /actual-weather --migrate
fi

echo "Starting actual-weather service..."
exec /usr/local/bin/actual-weather
