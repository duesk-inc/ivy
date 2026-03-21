#!/bin/sh
set -e

echo "Running database migrations..."
migrate -path /app/migrations -database "$DATABASE_URL" up

echo "Starting Ivy API server..."
exec /usr/local/bin/ivy-api
