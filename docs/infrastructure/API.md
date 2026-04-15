# 🔌 API и Интерфейсы

## 📋 Содержание

1. [Обзор](#обзор)
2. [Webhook HTTP API](#webhook-http-api)
3. [CLI Интерфейсы](#cli-интерфейсы)
4. [Конфигурация](#конфигурация)
5. [Примеры использования](#примеры-использования)

---

## 🎯 Обзор

Frontol ETL предоставляет два основных способа взаимодействия:

1. **HTTP API (Webhook Server)** - для автоматизированного запуска через веб-запросы
2. **CLI (Command Line Interface)** - для ручного управления и отладки

---

## 🌐 Webhook HTTP API

### Общая информация

**Base URL:** `http://localhost:$SERVER_PORT` (по умолчанию)
**Protocol:** HTTP/1.1
**Content-Type:** `application/json`
**Authentication:** Bearer Token (через `WEBHOOK_BEARER_TOKEN`)
Все endpoints из OpenAPI, кроме `/api/health`, требуют Bearer авторизации.
**Источник истины:** `api/openapi.yaml` (также доступен по `GET /api/openapi.yaml`)

### Endpoints

---

#### 1. POST /api/load

Запуск ETL pipeline для указанной даты (асинхронно).

**Request:**
```http
POST /api/load HTTP/1.1
Host: localhost:$SERVER_PORT
Content-Type: application/json
Authorization: Bearer <token>  # Если WEBHOOK_BEARER_TOKEN установлен

{
  "date": "2024-12-18"  # Опционально, по умолчанию сегодня
}
```

**Request Schema:**
```json
{
  "date": "string (YYYY-MM-DD) | optional"
}
```

**Response (202 Accepted):**
```json
{
  "status": "queued",
  "date": "2024-12-18",
  "message": "Request added to queue",
  "request_id": "req_1234567890"
}
```

**Response Codes:**
- `202 Accepted` - Запрос принят, обработка запущена в фоне
- `400 Bad Request` - Неверный формат даты
- `401 Unauthorized` - Неверный Bearer token
- `503 Service Unavailable` - Очередь переполнена

**Response Headers:**
- `X-Request-ID` - HTTP request identifier
- `X-Operation-ID` - идентификатор жизненного цикла ETL-операции для трассировки логов

**Примеры:**

```bash
# ETL для сегодняшней даты
curl -X POST http://localhost:$SERVER_PORT/api/load \
  -H 'Content-Type: application/json' \
  -d '{}'

# ETL для конкретной даты
curl -X POST http://localhost:$SERVER_PORT/api/load \
  -H 'Content-Type: application/json' \
  -d '{"date": "2024-12-18"}'

# С Bearer токеном
curl -X POST http://localhost:$SERVER_PORT/api/load \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer your_secret_token' \
  -d '{"date": "2024-12-18"}'
```

---

#### 2. GET /api/health

Проверка состояния сервера (health check).

**Request:**
```http
GET /api/health HTTP/1.1
Host: localhost:$SERVER_PORT
```

**Response (200 OK / 503 Service Unavailable):**
```json
{
  "status": "healthy",
  "timestamp": "2024-12-18T12:00:00Z",
  "service": "frontol-etl-webhook",
  "checks": {
    "database": {
      "status": "healthy",
      "latency_ms": 12
    },
    "ftp": {
      "status": "healthy",
      "latency_ms": 8
    },
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

**Response Codes:**
- `200 OK` - сервер и его зависимости healthy
- `503 Service Unavailable` - хотя бы одна из зависимостей (`database` или `ftp`) unhealthy, общий статус будет `degraded`

**Примеры:**

```bash
# Простая проверка
curl http://localhost:$SERVER_PORT/api/health

# Проверка с jq для парсинга JSON
curl http://localhost:$SERVER_PORT/api/health | jq .

# Мониторинг (каждые 30 секунд)
watch -n 30 'curl -s http://localhost:$SERVER_PORT/api/health | jq .'
```

---

#### 3. GET /api/docs

Интерактивная документация API (Scalar).

**Request:**
```http
GET /api/docs HTTP/1.1
Host: localhost:$SERVER_PORT
```

**Response:** HTML страница с интерактивной документацией

**Особенности:**
- Интерактивный интерфейс для тестирования API
- Автоматическая генерация из OpenAPI спецификации
- Try-it-out функциональность
- Подробное описание всех endpoints

**Доступ:**

Откройте в браузере: [http://localhost:$SERVER_PORT/api/docs](http://localhost:$SERVER_PORT/api/docs)

---

#### 4. GET /api/openapi.yaml

OpenAPI спецификация в YAML. Используется как источник истины для HTTP API.
Доступна без авторизации.

```http
GET /api/openapi.yaml HTTP/1.1
Host: localhost:$SERVER_PORT
```

---

#### 5. GET /api/files

Выгрузка данных из БД в файл по параметрам `source_folder` и `date`.
Подробная схема запроса и ответов — в `api/openapi.yaml`.
> Endpoint синхронный: ответ стримится сразу в рамках HTTP запроса и не ставится в in-memory очередь.

---

#### 6. GET /api/queue/status

Статус очереди обработки запросов.
Подробная схема ответа — в `api/openapi.yaml`.

**Поля ответа (фактическая реализация):**
- `queue_provider` — `memory`
- `total_queue_size`, `load_queue_size`, `download_queue_size`, `active_operations` — состояние in-memory очередей.
- `is_shutting_down` — сервер находится в процессе graceful shutdown.

---

#### 7. GET /api/kassas

Список доступных касс (source_folder).
Подробная схема ответа — в `api/openapi.yaml`.

### Асинхронная обработка

После получения `202 Accepted`, запрос попадает во внутреннюю in-memory очередь `load`, а ETL выполняется отдельным queue worker.

**Поток выполнения:**

```
POST /api/load
    ▼
HTTP 202 Accepted (немедленный ответ)
    ▼
ETL Pipeline (в фоне):
    1. Clear FTP folders
    2. Send requests
    3. Wait (configurable delay)
    4. Download files
    5. Parse transactions
    6. Load to database
    7. Send report (если WEBHOOK_REPORT_URL настроен)
```

**Мониторинг выполнения:**

```bash
# Просмотр логов в реальном времени
docker-compose logs -f webhook-server

# Поиск по operation_id
docker-compose logs webhook-server | grep 'operation_id="op_1234567890"'

# Поиск по request_id
docker-compose logs webhook-server | grep "req_1234567890"
```

---

### Отчеты (Webhook Reports)

Если настроен `WEBHOOK_REPORT_URL`, сервер отправит POST запрос с отчетом после завершения ETL.
Формат отчета соответствует схеме `WebhookReport` в `api/openapi.yaml`.

Фактические статусы ETL после reliability refactor:
- `completed` — все этапы завершились без операционных ошибок
- `partial` — данные загружены, но были recoverable ошибки/предупреждения
- `failed` — pipeline не завершился успешно

Дополнительные диагностические поля отчета:
- `error_breakdown`
- `error_samples`
- `files_recovered`
- `kassa_details`

---

### Безопасность

#### Bearer Token Authentication

Настройте в `.env`:
```bash
WEBHOOK_BEARER_TOKEN=your_secret_token_here
```

Отправляйте токен в заголовке:
```bash
curl -X POST http://localhost:$SERVER_PORT/api/load \
  -H 'Authorization: Bearer your_secret_token_here' \
  -H 'Content-Type: application/json' \
  -d '{"date": "2024-12-18"}'
```

#### CORS

По умолчанию CORS отключен. Для включения добавьте middleware в `cmd/webhook-server/main.go`.

---

## 💻 CLI Интерфейсы

### 1. Loader - Основной загрузчик

**Назначение:** Загрузка данных с FTP в PostgreSQL

**Использование:**
```bash
# Загрузка для сегодняшней даты
./frontol-loader

# Загрузка для конкретной даты
./frontol-loader 2024-12-18

# Через Docker Compose
docker-compose run --rm loader
docker-compose run --rm loader ./frontol-loader 2024-12-18
```

**Аргументы:**
- `[date]` - Дата в формате YYYY-MM-DD (опционально, по умолчанию сегодня)

**Переменные окружения:**
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` - Подключение к БД
- `FTP_HOST`, `FTP_PORT`, `FTP_USER`, `FTP_PASSWORD` - Подключение к FTP
- `KASSA_STRUCTURE` - Структура касс
- `LOG_LEVEL` - Уровень логирования

**Пример вывода:**
```
2024-12-18T12:00:00Z INFO ETL started date=2024-12-18
2024-12-18T12:00:05Z INFO Cleared FTP folders
2024-12-18T12:00:10Z INFO Sent requests kassa_count=3
2024-12-18T12:01:10Z INFO Downloaded files count=10
2024-12-18T12:01:15Z INFO Parsed transactions count=1500
2024-12-18T12:01:20Z INFO Loaded to database rows=1500
2024-12-18T12:01:20Z INFO ETL completed duration=80s
```

---

### 2. Migrate - Миграции БД

**Назначение:** Управление схемой базы данных

**Использование:**
```bash
# Применить все миграции
./migrate up

# Откатить все миграции
./migrate down

# Применить N миграций
./migrate step 1
./migrate step -1  # Откат 1 миграции

# Показать версию
./migrate version

# Принудительно установить версию
./migrate force 3

# Через Makefile
make migrate-up
make migrate-down
make migrate-version
```

**Команды:**
- `up` - Применить все миграции
- `down` - Откатить все миграции
- `step N` - Применить/откатить N миграций
- `version` - Показать текущую версию
- `force V` - Принудительно установить версию V

---

### 3. Parser Test - Тестер парсера

**Назначение:** Тестирование парсера на файлах

**Использование:**
```bash
# Локально
./parser-test /path/to/file.txt

# Через Docker
docker-compose run --rm parser-test ./parser-test /app/data/response.txt
```

**Пример вывода:**
```
Parsing file: /app/data/response.txt
Header: DBID=12345, ReportNum=67890
Transactions parsed: 15 types
  - tx_item_registration_1_11: 100
  - tx_position_discount_15: 10
  - tx_bonus_accrual_9: 5
```

---

### 4. Send Request - Отправка запросов

**Назначение:** Отправка request.txt к кассам через FTP

**Использование:**
```bash
# Локально
./send-request

# Через Docker
docker-compose run --rm send-request
```

**Что делает:**
1. Генерирует `request.txt` для каждой кассы
2. Отправляет на FTP в папки `request/`
3. Логирует статус отправки

---

### 5. Clear Requests - Очистка папок

**Назначение:** Очистка папок request/response на FTP

**Использование:**
```bash
# Локально
./clear-requests

# Через Docker
docker-compose run --rm clear-requests
```

**Что делает:**
1. Удаляет все файлы из папок `request/`
2. Удаляет все файлы из папок `response/`
3. Логирует статус очистки

---

## ⚙️ Конфигурация

### Переменные окружения

Полный список переменных см. в файле [../CONFIGURATION.md](../CONFIGURATION.md)

**Основные параметры:**

#### Database
```bash
DB_HOST=postgres.example.com
DB_PORT=5432
DB_USER=frontol_user
DB_PASSWORD=secure_password
DB_NAME=kassa_db
DB_SSLMODE=disable  # или require для production
```

#### FTP
```bash
FTP_HOST=ftp-server
FTP_PORT=21
FTP_USER=frontol
FTP_PASSWORD=frontol123
FTP_REQUEST_DIR=/request
FTP_RESPONSE_DIR=/response
```

#### Webhook Server
```bash
SERVER_PORT=8080
WEBHOOK_BEARER_TOKEN=your_secret_token  # Опционально
WEBHOOK_REPORT_URL=https://monitoring.example.com/reports  # Опционально
```

#### Kassa Structure
```bash
KASSA_STRUCTURE=P13:P13;N22:N22_Inter,N22_FURN;SH54:SH54
```

Формат: `KASSA_CODE:FOLDER1,FOLDER2;KASSA_CODE2:FOLDER3`

#### Application
```bash
LOG_LEVEL=info  # debug, info, warn, error
BATCH_SIZE=1000
MAX_RETRIES=3
RETRY_DELAY_SECONDS=5
WAIT_DELAY_MINUTES=1
```

---

## 📝 Примеры использования

### Сценарий 1: Автоматизированный ETL через cron

```bash
#!/bin/bash
# etl-daily.sh

# ETL для вчерашнего дня
DATE=$(date -d "yesterday" +%Y-%m-%d)

curl -X POST http://localhost:$SERVER_PORT/api/load \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer your_token' \
  -d "{\"date\": \"$DATE\"}" \
  >> /var/log/etl-trigger.log 2>&1
```

**Crontab:**
```cron
# Запуск ETL каждый день в 2:00 AM
0 2 * * * /path/to/etl-daily.sh
```

---

### Сценарий 2: Интеграция с системой мониторинга

**Python скрипт:**
```python
import requests
import json
from datetime import datetime, timedelta

def trigger_etl(date_str):
    url = "http://localhost:$SERVER_PORT/api/load"
    headers = {
        "Content-Type": "application/json",
        "Authorization": "Bearer your_token"
    }
    payload = {"date": date_str}

    response = requests.post(url, headers=headers, json=payload)

    if response.status_code == 202:
        data = response.json()
        print(f"ETL started: {data['request_id']}")
        return data['request_id']
    else:
        print(f"Error: {response.status_code}")
        return None

# Запуск ETL для последних 7 дней
for i in range(7):
    date = datetime.now() - timedelta(days=i)
    date_str = date.strftime("%Y-%m-%d")
    trigger_etl(date_str)
```

---

### Сценарий 3: Ручной запуск с проверкой

```bash
#!/bin/bash
# manual-etl.sh

# 1. Проверка health
echo "Checking server health..."
HEALTH=$(curl -s http://localhost:$SERVER_PORT/api/health | jq -r .status)

if [ "$HEALTH" != "healthy" ]; then
    echo "Server is not healthy!"
    exit 1
fi

echo "Server is healthy ✓"

# 2. Запуск ETL
echo "Triggering ETL for date: $1"
RESPONSE=$(curl -s -X POST http://localhost:$SERVER_PORT/api/load \
  -H 'Content-Type: application/json' \
  -d "{\"date\": \"$1\"}")

REQUEST_ID=$(echo $RESPONSE | jq -r .request_id)
echo "ETL started with request_id: $REQUEST_ID"

# 3. Мониторинг логов
echo "Monitoring logs..."
docker-compose logs -f webhook-server | grep "$REQUEST_ID"
```

**Использование:**
```bash
./manual-etl.sh 2024-12-18
```

---

### Сценарий 4: Проверка и повторный запуск при ошибке

```bash
#!/bin/bash
# etl-with-retry.sh

DATE=$1
MAX_RETRIES=3
RETRY_DELAY=60  # секунды

for i in $(seq 1 $MAX_RETRIES); do
    echo "Attempt $i of $MAX_RETRIES..."

    RESPONSE=$(curl -s -X POST http://localhost:$SERVER_PORT/api/load \
      -H 'Content-Type: application/json' \
      -d "{\"date\": \"$DATE\"}")

    STATUS=$(echo $RESPONSE | jq -r .status)

    if [ "$STATUS" == "queued" ]; then
        echo "ETL started successfully ✓"
        exit 0
    fi

    echo "Failed, retrying in ${RETRY_DELAY}s..."
    sleep $RETRY_DELAY
done

echo "All retries failed ✗"
exit 1
```

---

## 🔍 Мониторинг и отладка

### Просмотр логов

```bash
# Все логи webhook-server
docker-compose logs webhook-server

# В реальном времени
docker-compose logs -f webhook-server

# Фильтрация по уровню
docker-compose logs webhook-server | grep ERROR
docker-compose logs webhook-server | grep "request_id=req_123"

# Последние 100 строк
docker-compose logs --tail 100 webhook-server
```

### Формат логов (JSON)

```json
{
  "time": "2024-12-18T12:00:00Z",
  "level": "INFO",
  "msg": "ETL started",
  "request_id": "req_1234567890",
  "date": "2024-12-18",
  "component": "pipeline"
}
```

### Health check для мониторинга

```bash
# Prometheus-style check
curl -f http://localhost:$SERVER_PORT/api/health || exit 1

# В Kubernetes liveness probe
livenessProbe:
  httpGet:
    path: /api/health
    port: ${SERVER_PORT}
  initialDelaySeconds: 10
  periodSeconds: 30
```

---

## 📚 См. также

- [DEPLOYMENT.md](DEPLOYMENT.md) - Подробное руководство по развертыванию webhook
- [../CONFIGURATION.md](../CONFIGURATION.md) - Переменные окружения
- [../../DOCKER_COMPOSE_GUIDE.md](../../DOCKER_COMPOSE_GUIDE.md) - Docker Compose
- [BUSINESS_LOGIC.md](BUSINESS_LOGIC.md) - Бизнес-логика ETL

---

**Последнее обновление:** 2026-01-03
