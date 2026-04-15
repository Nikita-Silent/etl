# Docker Compose Guide

Актуальное руководство по запуску ETL через `docker-compose.yml`.

## Что есть в compose

Постоянные сервисы:

- `ftp-server`
- `webhook-server`

One-shot сервисы:

- `migrate`
- `ftp-structure-init`

CLI сервисы для ручного запуска:

- `loader`
- `parser-test`
- `send-request`
- `clear-requests`
- `ftp-check`

Важно: PostgreSQL внешний. Сервиса `postgres` в compose нет.

## Базовый запуск

```bash
cp env.example .env
# заполните DB_* и при необходимости FTP_*

docker-compose build
docker-compose up -d
docker-compose ps
docker-compose logs -f webhook-server
```

Миграции применяются через сервис `migrate`, который стартует как dependency. Повторно прогнать их вручную можно так:

```bash
docker-compose run --rm migrate
```

## Проверка здоровья сервиса

```bash
curl http://localhost:${SERVER_PORT:-8080}/api/health
```

Пример успешного ответа:

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

Если `database` или `ftp` unhealthy, endpoint вернет `503` и `status: degraded`.

## Запуск ETL

Через webhook:

```bash
curl -X POST http://localhost:${SERVER_PORT:-8080}/api/load \
  -H 'Content-Type: application/json' \
  -d '{}'

curl -X POST http://localhost:${SERVER_PORT:-8080}/api/load \
  -H 'Content-Type: application/json' \
  -d '{"date": "2024-12-18"}'
```

Ручной CLI pipeline:

```bash
docker-compose run --rm clear-requests
docker-compose run --rm send-request
sleep 60
docker-compose run --rm loader ./frontol-loader 2024-12-18
```

## Полезные команды

```bash
# логи
docker-compose logs -f webhook-server
docker-compose logs -f ftp-server

# shell
docker-compose exec webhook-server sh

# parser smoke test
docker-compose run --rm parser-test ./parser-test /app/data/response.txt

# FTP diagnostics
docker-compose run --rm ftp-check

# остановка
docker-compose down
```

## Override-файлы

```bash
# development overrides
docker-compose -f docker-compose.yml -f docker-compose.override.yml up -d

# production overrides
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

## См. также

- [QUICKSTART.md](QUICKSTART.md)
- [docs/infrastructure/API.md](docs/infrastructure/API.md)
- [docs/CONFIGURATION.md](docs/CONFIGURATION.md)
- [MAKEFILE_REFERENCE.md](MAKEFILE_REFERENCE.md)
