# Frontol 6 ETL Loader

[![CI](https://github.com/user/go-frontol-loader/workflows/CI/badge.svg)](https://github.com/user/go-frontol-loader/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/user/go-frontol-loader)](https://goreportcard.com/report/github.com/user/go-frontol-loader)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.24-blue)](https://go.dev/)

ETL pipeline for processing Frontol 6 export files and loading them into PostgreSQL.

## ✨ Features

- **FTP Integration**: Downloads files from FTP server and processes them
- **Modular Architecture**: Clean separation via `cmd/` and `pkg/`
- **Idempotent Loads**: `INSERT ... ON CONFLICT DO UPDATE`
- **Webhook Trigger**: Asynchronous ETL via HTTP
- **Database Migrations**: Managed by `cmd/migrate`
- **Structured Logging**: `slog` logging with context
- **Docker-first**: Compose-based dev and test flows

## 📚 Documentation

- **Docs index:** `docs/README.md`
- **Frontol 6 spec (source of truth):** `docs/Frontol_6_Integration.pdf`
- **Frontol 6 summary:** `docs/frontol_6_integration.md`
- **OpenAPI spec:** `api/openapi.yaml`
- **Testing guide:** `docs/test/TESTING.md`
- **Coding rules:** `docs/coding/CODING_RULES.md`

## 🚀 Quick Start (Docker)

```bash
# 1. Setup env
cp env.example .env

# 2. Start services (migrations run automatically)
docker compose up -d

# 3. Health check
curl http://localhost:${SERVER_PORT}/api/health

# 4. Trigger ETL
curl -X POST http://localhost:${SERVER_PORT}/api/load \
  -H 'Content-Type: application/json' \
  -d '{"date": "2024-12-18"}'
```

## 🧱 Project Structure

```
cmd/                 # CLI apps and services
pkg/                 # Core pipeline logic
tests/               # Integration tests
docs/                # Project documentation
scripts/             # Utility scripts
```

## 🔧 Build (Local)

```bash
make build-local
```

## 🧪 Tests

```bash
make test-go
make test-integration
```

## 🔍 See Also

- `docs/infrastructure/BUSINESS_LOGIC.md` - ETL pipeline details
- `docs/database/DDL_SPEC.md` - DB schema source of truth
