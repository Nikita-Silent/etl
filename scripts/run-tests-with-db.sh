#!/bin/bash
# Скрипт для запуска тестов с Docker контейнером PostgreSQL

set -e

echo "🚀 Starting PostgreSQL test container..."
docker compose -f docker-compose.test.yml up -d postgres-test

echo "⏳ Waiting for PostgreSQL to be ready..."
timeout 60 bash -c 'until docker exec frontol-postgres-test pg_isready -U frontol_user -d kassa_db_test; do sleep 1; done' || {
    echo "❌ PostgreSQL failed to start"
    docker compose -f docker-compose.test.yml logs postgres-test
    exit 1
}

echo "✅ PostgreSQL is ready!"

# Применяем миграции
echo "📦 Applying database migrations..."
export TEST_DB_HOST=localhost
export TEST_DB_PORT=5433
export TEST_DB_USER=frontol_user
export TEST_DB_PASSWORD=test_password
export TEST_DB_NAME=kassa_db_test
export DB_HOST=localhost
export DB_PORT=5433
export DB_USER=frontol_user
export DB_PASSWORD=test_password
export DB_NAME=kassa_db_test
export DB_SSLMODE=disable

go run ./cmd/migrate/main.go up || echo "⚠️  Migrations may already be applied"

# Запускаем тесты
echo "🧪 Running integration tests..."
export SKIP_INTEGRATION_TESTS=false
go test -v -tags=integration ./tests/integration/...

# Останавливаем контейнер
echo "🛑 Stopping PostgreSQL test container..."
docker compose -f docker-compose.test.yml down -v

echo "✅ Tests completed!"

