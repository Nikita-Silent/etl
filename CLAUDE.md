# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Frontol 6 ETL Loader - A Go-based ETL pipeline for processing Frontol 6 export files from FTP and loading them into PostgreSQL. The system supports asynchronous webhook triggers and handles 20+ transaction types with idempotent upserts.

## Documentation

Comprehensive documentation is available in the `docs/` directory. Refer to these files when working on specific areas:

**Start Here:**
- **docs/README.md** - Documentation navigation hub. Use this as the entry point to find the right documentation for your task.

**Architecture & Design:**
- **docs/ARCHITECTURE.md** - System architecture, design patterns, component interactions, and data flows. Read this when understanding how components work together or planning architectural changes.
- **docs/TECH_STACK.md** - Technologies, libraries, and their justifications. Consult when adding new dependencies or understanding technology choices (e.g., why pgx instead of GORM).

**Development:**
- **docs/CODING_RULES.md** - Code style, security practices, performance patterns, and testing guidelines. Review before writing code to ensure consistency with project standards.
- **docs/TESTING.md** - Complete testing guide including unit tests, integration tests, benchmarks, and E2E testing. Use when writing tests or debugging test failures.
- **docs/BUSINESS_LOGIC.md** - ETL pipeline business logic, file parsing details, and data transformation rules. Essential for understanding or modifying the ETL process.

**Database:**
- **docs/DATABASE.md** - Database schema, tables, indexes, migrations, and query examples. Reference when working with database schema, writing queries, or understanding data relationships.

**API & Integration:**
- **docs/API.md** - HTTP webhook API, CLI interfaces, and usage examples. Use when working with API endpoints or integrating external systems.
- **docs/CONFIGURATION.md** - All environment variables with descriptions and examples. Consult when adding new config options or troubleshooting configuration issues.

**Operations:**
- **docs/DEPLOYMENT.md** - Production deployment guides (Docker Compose, systemd), monitoring, and backup strategies. Read before deploying to production or setting up new environments.
- **docs/TROUBLESHOOTING.md** - Common problems and solutions for Docker, database, FTP, and ETL issues. Start here when debugging problems.

**Planning:**
- **docs/ROADMAP.md** - Project roadmap, planned features, and current limitations. Review when planning new features or understanding project direction.

## Common Commands

### Development
```bash
# Build all binaries locally
go build -o webhook-server ./cmd/webhook-server/main.go
go build -o frontol-loader ./cmd/loader/main.go
go build -o migrate ./cmd/migrate/main.go

# Or use Makefile
make build-local

# Run webhook server locally
export $(cat configs/development.env | xargs)
./webhook-server
```

### Testing
```bash
# Unit tests only (default - no external dependencies)
go test -v -race ./pkg/...

# Integration tests (requires PostgreSQL)
go test -v -tags=integration ./tests/integration/...

# With coverage
go test -coverprofile=coverage.out -covermode=atomic ./pkg/...

# Quick check (format, lint, test)
make check

# Full CI pipeline
make ci
```

### Linting
```bash
# Run golangci-lint (version: latest)
golangci-lint run --timeout=5m

# Or via Makefile
make lint
```

### Database Migrations
```bash
# Apply all pending migrations
go run ./cmd/migrate/main.go up

# Show current version
go run ./cmd/migrate/main.go version

# Create new migration
make migrate-create NAME=add_feature

# Or manually (auto-increments version number)
# Creates: pkg/migrate/migrations/NNNNNN_add_feature.{up,down}.sql
```

### Running ETL
```bash
# Via webhook (asynchronous)
curl -X POST http://localhost:8080/api/load \
  -H 'Content-Type: application/json' \
  -d '{"date": "2024-12-18"}'

# Via CLI (synchronous)
./frontol-loader 2024-12-18

# Or use Makefile
make etl-webhook-date DATE=2024-12-18
```

### Docker Operations
```bash
# Start all services (includes auto-migrations)
docker-compose up -d

# View logs
docker-compose logs -f webhook-server

# Run ETL via Docker
make etl-date DATE=2024-12-18
```

## High-Level Architecture

### ETL Pipeline Flow (pkg/pipeline)

The core `pipeline.Run()` function orchestrates a 4-step sequence:

1. **Clear Request Folders** - Removes old request.txt files from all kassa folders
2. **Send Requests** - Generates date-specific request.txt files (`DATE_FROM=YYYY-MM-DD;DATE_TO=YYYY-MM-DD`) and uploads to FTP
3. **Wait for Responses** - Configurable delay (default 1 minute) for Frontol systems to generate export files
4. **Process Files** - Downloads, parses, and loads files in parallel using goroutines with mutex-protected FTP operations

Key characteristics:
- **Idempotent**: Files marked as `.processed` after success; skipped on re-runs
- **Parallel Processing**: Goroutines + WaitGroup for concurrent file handling
- **Unique Paths**: Files stored as `{LocalDir}/{KassaCode}/{FolderName}/{filename}` to avoid conflicts
- **Atomic Transactions**: Each file loads in a single database transaction (all-or-nothing)

### Package Architecture

```
cmd/
├── webhook-server/    # HTTP server with request queue system
├── loader/           # CLI wrapper for pipeline.Run()
├── migrate/          # Database migration tool
└── [other tools]

pkg/
├── pipeline/         # ETL orchestration (Run function)
├── ftp/             # FTP operations (ListFiles, Download, MarkAsProcessed)
├── parser/          # Frontol file parsing (dispatches by transaction type)
├── repository/      # Data loading with upserts
├── db/              # PostgreSQL connection pool (pgx/v5)
├── config/          # Configuration loading and validation
└── models/          # Transaction data structures
```

### Webhook Server Architecture

The webhook server (`cmd/webhook-server`) implements a **request queue system** with these features:

- **Multiple Operation Types**: Separate queues for "load" (ETL) and "download" (data export)
- **Sequential Processing**: Each queue has 1 worker goroutine (prevents parallel ETL runs)
- **Immediate Response**: Returns `202 Accepted` with request_id before processing
- **Queue Capacity**: 100 items per operation type
- **Bearer Token Auth**: Optional authentication for API endpoints (can be disabled)

API endpoints:
- `POST /api/load` - Trigger ETL (protected)
- `GET /api/files?source_folder=XX&date=YYYY-MM-DD` - Download data (protected)
- `GET /api/queue/status` - Queue metrics (protected)
- `GET /api/health` - Health check (public)
- `GET /api/docs` - Interactive Scalar API documentation (public)

### Database Design

**Composite Primary Key Pattern**:
All transaction tables use `(transaction_id_unique, source_folder)` as the composite primary key. This enables:
- Multiple kassas to have transactions with same IDs
- Idempotent upserts using `ON CONFLICT ... DO UPDATE`
- Safe parallel processing from different source folders

**Connection Pool** (pkg/db):
- pgx/v5 with connection pooling (10 max, 2 min connections)
- Transactions use `READ COMMITTED` isolation
- Retry logic with exponential backoff for deadlocks/serialization failures
- Automatic Windows-1251 → UTF-8 encoding conversion

**Migration System**:
- golang-migrate/migrate library
- Migrations in `pkg/migrate/migrations/NNNNNN_name.{up,down}.sql`
- Init container runs migrations automatically on deployment

### Interface-Based Design

Key interfaces enable testing with mocks:

**DatabasePool** (`pkg/db/interfaces.go`):
- Implementation: `db.Pool` (wraps pgxpool.Pool)
- Mock: `db.MockPool` (in `pkg/db/mocks.go`)
- Methods: BeginTx, Query, LoadData, Load{TransactionType}

**FTPClient** (`pkg/ftp/interfaces.go`):
- Implementation: `ftp.Client` (wraps jlaffaye/ftp)
- Mock: `ftp.MockClient` (in `pkg/ftp/mocks.go`)
- Methods: ListFiles, DownloadFile, MarkFileAsProcessed, SendRequestsToAllKassas

### Testing Strategy

**Unit Tests** (`pkg/*_test.go`):
- Use mocks for external dependencies (database, FTP)
- Focus on business logic, interface contracts, edge cases
- Run by default in CI (no external services required)

**Integration Tests** (`tests/integration/`):
- Build tag: `//go:build integration`
- Require real PostgreSQL (started via docker-compose.test.yml)
- Test end-to-end pipelines with actual data loading
- Gated by `SKIP_INTEGRATION_TESTS` environment variable

### Concurrency Patterns

**Pipeline Processing**:
- Goroutines for parallel file downloads/parsing
- `sync.Mutex` protects FTP client (connections not concurrent-safe)
- `sync.WaitGroup` for goroutine coordination

**Webhook Server**:
- Sequential operation queues prevent race conditions
- ETL runs processed one at a time per operation type
- Report sending protected by `sync.Once` flag

### Configuration

Environment variables loaded from `.env` + system environment:

Required:
- `DB_PASSWORD`, `FTP_USER`, `FTP_PASSWORD`

Important:
- `KASSA_STRUCTURE` - Format: `"001:req1,resp1;002:req2,resp2"` (kassa_code:request_folder,response_folder)
- `WEBHOOK_BEARER_TOKEN` - Optional API authentication (empty = disabled)
- `WEBHOOK_REPORT_URL` - Optional webhook for ETL completion reports
- `WEBHOOK_TIMEOUT_MINUTES` - Report timeout (0 = wait forever)

### File Processing Details

**Frontol File Structure**:
1. **Header** (3 lines):
   - Line 1: Processing flag (`#` = not processed, `@` or `1` = processed)
   - Line 2: Database ID
   - Line 3: Report number
2. **Transaction Lines**: Semicolon-separated fields with transaction type in 4th position

**Transaction Dispatching**:
The parser reads transaction type (field 4) and dispatches to specific parsers for 20+ types:
- Transaction Registration (1, 2, 4, 6, etc.)
- Bonus/Discount transactions (9, 10, 15, 17)
- Bill operations (18, 19)
- Employee operations (20-23)
- And more...

Each transaction type has a dedicated PostgreSQL table and loader method.

### Issue Tracking

This project uses [Beads (bd)](https://github.com/steveyegge/beads) for issue tracking:

```bash
# Load workflow context before starting work
bd prime

# Find available work
bd ready

# Create new issue
bd create --title="..." --type=task

# Claim work
bd update <id> --status=in_progress

# Mark complete
bd close <id>
```

**Important**: Track ALL work in bd (never use markdown TODOs or comment-based task lists). Git hooks auto-sync on commit/merge.

### CI/CD Pipeline

GitHub Actions workflow (`.github/workflows/ci.yml`) runs:
1. **Test Job**: go vet, unit tests with race detection, integration tests with PostgreSQL
2. **Lint Job**: golangci-lint with custom configuration
3. **Build Job**: Compiles all 12 binaries (webhook-server, loader, migrate, check-missing, clear-db, ftp-check, ftp-server, restore-raw-data, etc.)
4. **Docker Job**: Validates Docker image build and docker-compose config

Integration tests use `docker-compose.test.yml` with PostgreSQL on port 5433.
