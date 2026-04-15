# Makefile for Frontol ETL Docker deployment

.PHONY: help build up down logs clean test dev prod

# Default target
help:
	@echo "Frontol ETL Docker Commands:"
	@echo ""
	@echo "🚀 ETL Pipeline:"
	@echo "  make etl                  - Run full ETL for today"
	@echo "  make etl-date DATE=...    - Run full ETL for specific date"
	@echo "  make etl-webhook          - Trigger ETL via webhook (today)"
	@echo "  make etl-webhook-date     - Trigger ETL via webhook (specific date)"
	@echo ""
	@echo "🐳 Docker Services:"
	@echo "  make dev                  - Start in development mode"
	@echo "  make prod                 - Start in production mode"
	@echo "  make build                - Build all Docker images"
	@echo "  make up                   - Start all services"
	@echo "  make down                 - Stop all services"
	@echo "  make restart              - Restart services"
	@echo "  make status               - Show service status"
	@echo ""
	@echo "📋 Logs & Monitoring:"
	@echo "  make logs                 - View all logs"
	@echo "  make logs-webhook         - View webhook server logs"
	@echo "  make logs-ftp             - View FTP server logs"
	@echo "  make health               - Check service health"
	@echo ""
	@echo "🔧 Manual Operations:"
	@echo "  make loader               - Run loader manually"
	@echo "  make loader-date DATE=... - Run loader for specific date"
	@echo "  make send-request         - Send request files"
	@echo "  make clear-requests       - Clear request/response folders"
	@echo ""
	@echo "🗄️ Database:"
	@echo "  make migrate-up           - Apply migrations"
	@echo "  make migrate-version      - Show migration version"
	@echo "  make backup-db            - Backup database"
	@echo ""
	@echo "🧪 Testing:"
	@echo "  make check                - Run fmt + lint + tests"
	@echo "  make ci                   - Full CI pipeline"
	@echo "  make test-go              - Run Go tests"
	@echo "  make test-coverage        - Tests with coverage"
	@echo "  make test-ftp-structure   - Test FTP structure init container"
	@echo ""
	@echo "🛠️ Development:"
	@echo "  make build-local          - Build binaries locally"
	@echo "  make clean-local          - Clean local binaries"
	@echo "  make shell                - Open shell in container"
	@echo ""

# Build images
build:
	docker-compose build

# Start all services
up:
	docker-compose up -d

# Stop all services
down:
	docker-compose down

# View logs
logs:
	docker-compose logs -f

# View logs for specific service
logs-webhook:
	docker-compose logs -f webhook-server

logs-db:
	@echo "PostgreSQL is external; there are no docker-compose database logs in this repository"

logs-ftp:
	docker-compose logs -f ftp-server

# Clean up
clean:
	docker-compose down -v --rmi all

# Run tests
test:
	docker-compose run --rm parser-test ./parser-test /app/data/response.txt

# Development mode
dev:
	docker-compose -f docker-compose.yml -f docker-compose.override.yml up -d

# Production mode
prod:
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# Health check
health:
	@PORT=$${SERVER_PORT:-8080}; curl -fsS http://localhost:$$PORT/api/health

# Open shell in webhook container
shell:
	docker-compose exec webhook-server sh

# ==========================================
# ETL Operations via Docker Compose
# ==========================================

# Run full ETL pipeline for today
etl:
	@echo "🚀 Running full ETL pipeline for today..."
	@echo "Step 1/3: Clearing FTP folders..."
	docker-compose run --rm clear-requests
	@echo "Step 2/3: Waiting for responses (1 minute)..."
	@sleep 60
	@echo "Step 3/3: Loading data..."
	docker-compose run --rm loader
	@echo "✅ ETL pipeline completed!"

# Run full ETL pipeline for specific date
etl-date:
	@if [ -z "$(DATE)" ]; then echo "Usage: make etl-date DATE=YYYY-MM-DD"; exit 1; fi
	@echo "🚀 Running full ETL pipeline for date: $(DATE)..."
	@echo "Step 1/3: Clearing FTP folders..."
	docker-compose run --rm clear-requests
	@echo "Step 2/3: Waiting for responses (1 minute)..."
	@sleep 60
	@echo "Step 3/3: Loading data..."
	docker-compose run --rm loader ./frontol-loader $(DATE)
	@echo "✅ ETL pipeline completed!"

# Trigger ETL via webhook for today
etl-webhook:
	@echo "🔔 Triggering ETL via webhook for today..."
	curl -X POST http://localhost:8080/api/load \
		-H 'Content-Type: application/json' \
		-d '{}' \
		-w '\n'
	@echo "✅ ETL triggered! Check logs: make logs-webhook"

# Trigger ETL via webhook for specific date
etl-webhook-date:
	@if [ -z "$(DATE)" ]; then echo "Usage: make etl-webhook-date DATE=YYYY-MM-DD"; exit 1; fi
	@echo "🔔 Triggering ETL via webhook for date: $(DATE)..."
	curl -X POST http://localhost:8080/api/load \
		-H 'Content-Type: application/json' \
		-d '{"date": "$(DATE)"}' \
		-w '\n'
	@echo "✅ ETL triggered! Check logs: make logs-webhook"

# Run loader only (manual)
loader:
	docker-compose run --rm loader

# Run loader for specific date (manual)
loader-date:
	@if [ -z "$(DATE)" ]; then echo "Usage: make loader-date DATE=YYYY-MM-DD"; exit 1; fi
	docker-compose run --rm loader ./frontol-loader $(DATE)

# Send requests to kassas
send-request:
	docker-compose run --rm send-request

# Clear request/response folders
clear-requests:
	docker-compose run --rm clear-requests

# Clear database (delete all transaction data)
clear-db:
	@echo "⚠️  WARNING: This will delete all transaction data from the database!"
	@read -p "Are you sure? Type 'yes' to confirm: " confirm && [ "$$confirm" = "yes" ] || exit 1
	docker-compose run --rm webhook-server ./clear-db --confirm

# Clear database using SQL script
clear-db-sql:
	@echo "⚠️  WARNING: This will delete all transaction data from the database!"
	@read -p "Are you sure? Type 'yes' to confirm: " confirm && [ "$$confirm" = "yes" ] || exit 1
	docker-compose exec -T webhook-server sh -c "psql -h \$$DB_HOST -U \$$DB_USER -d \$$DB_NAME -f /app/scripts/clear-database.sql" || \
	docker-compose run --rm webhook-server sh -c "psql -h \$$DB_HOST -U \$$DB_USER -d \$$DB_NAME < /app/scripts/clear-database.sql"

# Initialize database
init-db:
	docker-compose run --rm migrate

# Show status
status:
	docker-compose ps

# Restart services
restart:
	docker-compose restart

# Update and restart
update:
	docker-compose pull
	docker-compose up -d

# Backup database
backup-db:
	@if [ -z "$$DB_HOST" ] || [ -z "$$DB_USER" ] || [ -z "$$DB_PASSWORD" ]; then echo "Set DB_HOST, DB_USER, and DB_PASSWORD in your environment or .env"; exit 1; fi
	PGPASSWORD="$$DB_PASSWORD" pg_dump -h "$$DB_HOST" -p "$${DB_PORT:-5432}" -U "$$DB_USER" "$${DB_NAME:-kassa_db}" > backup_$(shell date +%Y%m%d_%H%M%S).sql

# Restore database
restore-db:
	@if [ -z "$(FILE)" ]; then echo "Usage: make restore-db FILE=backup.sql"; exit 1; fi
	@if [ -z "$$DB_HOST" ] || [ -z "$$DB_USER" ] || [ -z "$$DB_PASSWORD" ]; then echo "Set DB_HOST, DB_USER, and DB_PASSWORD in your environment or .env"; exit 1; fi
	PGPASSWORD="$$DB_PASSWORD" psql -h "$$DB_HOST" -p "$${DB_PORT:-5432}" -U "$$DB_USER" "$${DB_NAME:-kassa_db}" < "$(FILE)"

# Show resource usage
stats:
	docker stats

# Build and push to registry (requires REGISTRY variable)
push:
	docker-compose build
	docker tag parcer_webhook-server $(REGISTRY)/frontol-etl:latest
	docker push $(REGISTRY)/frontol-etl:latest

# Run webhook server locally (for development)
run-local:
	go run ./cmd/webhook-server

# Run loader locally (for development)
run-loader-local:
	go run ./cmd/loader

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Run tests
test-go:
	go test ./...

# Run focused reliability regression suite
test-reliability:
	go test ./pkg/config ./pkg/models ./pkg/parser ./pkg/pipeline ./pkg/repository ./cmd/webhook-server ./tests/e2e

# Run tests with verbose output
test-verbose:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Test FTP structure init container
test-ftp-structure:
	@echo "Running FTP structure integration tests..."
	@./tests/integration/ftp-structure-test.sh
	@echo "Coverage report generated: coverage.html"

# Run tests with race detection
test-race:
	go test -race ./...

# Run race detection on critical reliability packages
test-race-critical:
	go test -race ./pkg/config ./pkg/parser ./pkg/pipeline ./pkg/repository ./cmd/webhook-server

# Run benchmarks
test-bench:
	go test -bench=. -benchmem ./...

# Run integration tests (requires running services)
test-integration:
	INTEGRATION_TEST=true go test -tags=integration -v ./tests/integration/...

# Run all tests
test-all: test-go test-race test-bench

# Quick check (fmt + lint + test)
check: fmt lint test-go
	@echo "All checks passed!"

# CI pipeline
ci: fmt lint test-reliability test-race-critical
	@echo "CI pipeline completed!"

# Build binaries locally
build-local:
	go build -o webhook-server ./cmd/webhook-server
	go build -o frontol-loader ./cmd/loader
	go build -o frontol-loader-local ./cmd/loader-local
	go build -o parser-test ./cmd/parser-test
	go build -o send-request ./cmd/send-request
	go build -o clear-requests ./cmd/clear-requests
	go build -o migrate ./cmd/migrate

# Clean local binaries
clean-local:
	rm -f webhook-server frontol-loader frontol-loader-local parser-test send-request clear-requests migrate

# ==========================================
# Database Migrations (golang-migrate)
# ==========================================

# Run all pending migrations
migrate-up:
	go run ./cmd/migrate up

# Rollback all migrations
migrate-down:
	go run ./cmd/migrate down

# Run N migrations (usage: make migrate-step N=1)
migrate-step:
	go run ./cmd/migrate step $(N)

# Show current migration version
migrate-version:
	go run ./cmd/migrate version

# Force migration version (usage: make migrate-force V=1)
migrate-force:
	go run ./cmd/migrate force $(V)

# Drop all tables (DANGEROUS!)
migrate-drop:
	go run ./cmd/migrate drop

# Create new migration (usage: make migrate-create NAME=add_users_table)
migrate-create:
	@if [ -z "$(NAME)" ]; then echo "Usage: make migrate-create NAME=migration_name"; exit 1; fi
	@VERSION=$$(ls -1 pkg/migrate/migrations/*.up.sql 2>/dev/null | wc -l | xargs printf "%06d"); \
	NEXT_VERSION=$$(printf "%06d" $$((10#$$VERSION + 1))); \
	touch pkg/migrate/migrations/$${NEXT_VERSION}_$(NAME).up.sql; \
	touch pkg/migrate/migrations/$${NEXT_VERSION}_$(NAME).down.sql; \
	echo "Created migrations:"; \
	echo "  pkg/migrate/migrations/$${NEXT_VERSION}_$(NAME).up.sql"; \
	echo "  pkg/migrate/migrations/$${NEXT_VERSION}_$(NAME).down.sql"

# Setup development environment
setup-dev:
	cp env.example .env
	@echo "Please edit .env file with your configuration"
	@echo "Then run: make dev"

# Quick start
quick-start: setup-dev build up
	@echo "Services started! Webhook server available at http://localhost:8080"
	@echo "Run 'make logs' to view logs"
	@echo "Run 'make health' to check service health"
