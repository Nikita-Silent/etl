# Quick Start

Минимальная проверка текущего ETL-стека.

## Перед стартом

- PostgreSQL в этом репозитории внешний, его нет в `docker-compose.yml`.
- Нужны заполненные `DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` в `.env`.

## За 5 минут

```bash
cp env.example .env
# отредактируйте .env

docker compose up -d
docker compose ps

curl http://localhost:${SERVER_PORT:-8080}/api/health

curl -X POST http://localhost:${SERVER_PORT:-8080}/api/load \
  -H 'Content-Type: application/json' \
  -d '{"date": "2024-12-18"}'

docker compose logs -f webhook-server
```

Ожидаемый health-ответ:

```json
{
  "status": "healthy",
  "timestamp": "2026-04-14T12:00:00Z",
  "service": "frontol-etl-webhook",
  "checks": {
    "database": {"status": "healthy", "latency_ms": 12},
    "ftp": {"status": "healthy", "latency_ms": 8},
    "queues": {
      "load_queue_size": 0,
      "download_queue_size": 0,
      "total_queue_size": 0,
      "active_operations": 0,
      "is_shutting_down": false
    }
  },
  "response_time_ms": 20
}
```

Если одна из зависимостей недоступна, `status` будет `degraded`, а HTTP-код `503`.

## Полезные команды

```bash
# повторно применить миграции вручную
docker compose run --rm migrate

# ручной ETL через CLI
docker compose run --rm clear-requests
sleep 60
docker compose run --rm loader ./frontol-loader 2024-12-18

# тест парсера
docker compose run --rm parser-test ./parser-test /app/data/response.txt
```

## Куда смотреть дальше

- [DOCKER_COMPOSE_GUIDE.md](DOCKER_COMPOSE_GUIDE.md)
- [docs/infrastructure/API.md](docs/infrastructure/API.md)
- [docs/CONFIGURATION.md](docs/CONFIGURATION.md)
- [MAKEFILE_REFERENCE.md](MAKEFILE_REFERENCE.md)
