#!/bin/sh
set -eu

echo "Running database migrations..."
attempt=1
max_attempts=20

while [ "$attempt" -le "$max_attempts" ]; do
  if /app/migrate -path /app/migrations -database "$DATABASE_URL" up; then
    break
  fi
  echo "Migration attempt $attempt failed, retrying..."
  attempt=$((attempt + 1))
  sleep 2
done

if [ "$attempt" -gt "$max_attempts" ]; then
  echo "Migrations failed after retries"
  exit 1
fi

echo "Starting TaskFlow API..."
exec /app/main
