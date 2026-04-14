# ⚙️ Конфигурация

## 📋 Содержание

1. [Обзор](#обзор)
2. [Переменные окружения](#переменные-окружения)
3. [Примеры конфигураций](#примеры-конфигураций)
4. [Безопасность](#безопасность)
5. [Валидация конфигурации](#валидация-конфигурации)

---

## 🎯 Обзор

Frontol ETL настраивается исключительно через **переменные окружения**.

**Преимущества:**
- 🔒 Пароли не хранятся в git
- 🔄 Легко менять для разных окружений
- 📝 Один источник правды
- 🐳 Совместимость с Docker и Kubernetes

---

## 📝 Переменные окружения

### Database (PostgreSQL)

| Переменная | Обязательно | По умолчанию | Описание |
|------------|-------------|--------------|----------|
| `DB_HOST` | ✅ Да | - | Хост PostgreSQL кластера |
| `DB_PORT` | ❌ Нет | `5432` | Порт PostgreSQL |
| `DB_CONNECT_TIMEOUT_SECONDS` | ❌ Нет | `5` | Таймаут подключения и ping PostgreSQL |
| `DB_USER` | ✅ Да | - | Пользователь БД |
| `DB_PASSWORD` | ✅ Да | - | **Пароль БД (изменить!)** |
| `DB_NAME` | ✅ Да | - | Имя базы данных |
| `DB_SSLMODE` | ❌ Нет | `disable` | SSL режим (`disable`, `require`, `verify-full`) |

**Пример:**

```bash
DB_HOST=postgres.example.com
DB_PORT=5432
DB_USER=frontol_user
DB_PASSWORD=secure_password_change_me
DB_NAME=kassa_db
DB_SSLMODE=disable  # require для production
```

Все numeric-переменные валидируются в startup. Значения вроде `DB_PORT=abc` или `BATCH_SIZE=oops` теперь приводят к fail-fast ошибке конфигурации, а не к тихому откату в default.

---

### FTP Server

| Переменная | Обязательно | По умолчанию | Описание |
|------------|-------------|--------------|----------|
| `PUBLICHOST` | ❌ Нет | `localhost` | Публичный hostname для FTP passive mode |
| `FTP_HOST` | ✅ Да | - | Hostname FTP сервера |
| `FTP_PORT` | ❌ Нет | `21` | Порт FTP сервера |
| `FTP_CONNECT_TIMEOUT_SECONDS` | ❌ Нет | `5` | Таймаут установления FTP соединения |
| `FTP_USER` | ✅ Да | - | Пользователь FTP |
| `FTP_PASSWORD` | ✅ Да | - | **Пароль FTP (изменить!)** |
| `FTP_REQUEST_DIR` | ❌ Нет | `/request` | Директория для request файлов |
| `FTP_RESPONSE_DIR` | ❌ Нет | `/response` | Директория для response файлов |
| `FTP_POOL_SIZE` | ❌ Нет | `5` | Размер пула FTP соединений |

**Пример:**

```bash
PUBLICHOST=ftp.example.com  # или внешний IP
FTP_HOST=ftp-server
FTP_PORT=21
FTP_USER=frontol
FTP_PASSWORD=frontol123  # изменить для production
FTP_REQUEST_DIR=/request
FTP_RESPONSE_DIR=/response
```

---

### Kassa Structure

| Переменная | Обязательно | По умолчанию | Описание |
|------------|-------------|--------------|----------|
| `KASSA_STRUCTURE` | ✅ Да | - | Структура касс (код:папки) |

**Формат:** `KASSA_CODE:FOLDER1,FOLDER2;KASSA_CODE2:FOLDER3`

**Пример:**

```bash
KASSA_STRUCTURE=P13:P13;N22:N22_Inter,N22_FURN;SH54:SH54
```

**Объяснение:**
- `P13` - код кассы, папка `P13`
- `N22` - код кассы, папки `N22_Inter` и `N22_FURN`
- `SH54` - код кассы, папка `SH54`

---

### Application

| Переменная | Обязательно | По умолчанию | Описание |
|------------|-------------|--------------|----------|
| `LOCAL_DIR` | ❌ Нет | `/tmp/frontol` | Локальная директория для файлов |
| `BATCH_SIZE` | ❌ Нет | `1000` | Размер batch для загрузки в БД |
| `MAX_RETRIES` | ❌ Нет | `3` | Максимум попыток при ошибках |
| `RETRY_DELAY_SECONDS` | ❌ Нет | `5` | Задержка между попытками (сек) |
| `WAIT_DELAY_MINUTES` | ❌ Нет | `1` | Задержка ожидания Frontol (мин) |
| `PIPELINE_LOAD_TIMEOUT_MINUTES` | ❌ Нет | `60` | Таймаут DB load/reconcile стадии одного файла |
| `CLI_RUN_TIMEOUT_MINUTES` | ❌ Нет | `30` | Внешний таймаут CLI entrypoints (`cmd/loader`, `cmd/loader-local`) |
| `OPERATION_STALE_TIMEOUT_MINUTES` | ❌ Нет | `120` | Через сколько незавершенная ETL-операция считается stale и помечается abandoned при следующем запуске |
| `WORKER_POOL_SIZE` | ❌ Нет | `10` | Размер worker pool для обработки файлов |
| `LOG_LEVEL` | ❌ Нет | `info` | Уровень логирования |
| `LOG_FORMAT` | ❌ Нет | `json` | Формат логов (`json`, `text`, `console`) |
| `LOG_BACKEND` | ❌ Нет | `zerolog` | Бэкенд логирования (`zerolog`, `slog`) |
| `LOG_SINK` | ❌ Нет | `stdout` | Куда писать логи (`stdout`, `loki`, `both`) |
| `LOKI_URL` | ❌ Нет | - | HTTP endpoint Loki push API (`/loki/api/v1/push`) |
| `LOKI_TENANT_ID` | ❌ Нет | - | Значение заголовка `X-Scope-OrgID` для multi-tenant Loki |
| `LOKI_BEARER_TOKEN` | ❌ Нет | - | Bearer token для прямой отправки логов в Loki |
| `LOKI_BATCH_WAIT_MS` | ❌ Нет | `1000` | Максимальное время накопления log batch перед отправкой |
| `LOKI_BATCH_SIZE` | ❌ Нет | `100` | Максимум log entries в одном push batch |
| `LOKI_TIMEOUT_SECONDS` | ❌ Нет | `5` | HTTP timeout отправки batch в Loki |
| `LOKI_LABELS` | ❌ Нет | `service=frontol-etl` | Статические labels в формате `key=value,key2=value2` |

**Допустимые значения `LOG_LEVEL`:**
- `debug` - Детальная информация для отладки
- `info` - Обычная информация (по умолчанию)
- `warn` - Предупреждения
- `error` - Только ошибки

**Пример:**

```bash
LOCAL_DIR=/data/frontol
BATCH_SIZE=5000
MAX_RETRIES=5
RETRY_DELAY_SECONDS=10
WAIT_DELAY_MINUTES=2
LOG_LEVEL=info
```

---

### Webhook Server

| Переменная | Обязательно | По умолчанию | Описание |
|------------|-------------|--------------|----------|
| `SERVER_PORT` | ❌ Нет | `8080` | Порт webhook сервера |
| `WEBHOOK_REPORT_URL` | ❌ Нет | - | URL для отправки отчетов (опционально) |
| `WEBHOOK_TIMEOUT_MINUTES` | ❌ Нет | `0` | Таймаут отправки webhook отчета (`0` = только после завершения pipeline) |
| `WEBHOOK_REPORT_HTTP_TIMEOUT_SECONDS` | ❌ Нет | `30` | Таймаут исходящего HTTP запроса с webhook отчетом |
| `WEBHOOK_REPORT_RESULT_WAIT_SECONDS` | ❌ Нет | `5` | Сколько ждать готовый отчет после завершения pipeline before warning |
| `WEBHOOK_BEARER_TOKEN` | ❌ Нет | - | Bearer token для авторизации (опционально) |
| `HTTP_READ_HEADER_TIMEOUT_SECONDS` | ❌ Нет | `5` | `http.Server` read header timeout |
| `HTTP_READ_TIMEOUT_SECONDS` | ❌ Нет | `15` | `http.Server` read timeout |
| `HTTP_WRITE_TIMEOUT_SECONDS` | ❌ Нет | `30` | `http.Server` write timeout |
| `HTTP_IDLE_TIMEOUT_SECONDS` | ❌ Нет | `60` | `http.Server` idle timeout |
| `SHUTDOWN_TIMEOUT_SECONDS` | ❌ Нет | `30` | Таймаут graceful shutdown для webhook server |

**Пример:**

```bash
SERVER_PORT=8080
WEBHOOK_REPORT_URL=https://monitoring.example.com/api/reports
WEBHOOK_TIMEOUT_MINUTES=1
WEBHOOK_BEARER_TOKEN=your_secret_token_here
SHUTDOWN_TIMEOUT_SECONDS=30
```

## Поведение валидации

- `KASSA_STRUCTURE` обязателен и больше не имеет fallback структуры по умолчанию.
- Пустые коды касс, пустые папки и битые группы в `KASSA_STRUCTURE` приводят к ошибке startup.
- Numeric-параметры (`DB_PORT`, `FTP_PORT`, `FTP_POOL_SIZE`, `BATCH_SIZE`, `MAX_RETRIES`, `WORKER_POOL_SIZE`, `SERVER_PORT`, `WEBHOOK_TIMEOUT_MINUTES`, `SHUTDOWN_TIMEOUT_SECONDS`, `PASV_*`) валидируются fail-fast.
- Runtime timeout-параметры (`DB_CONNECT_TIMEOUT_SECONDS`, `FTP_CONNECT_TIMEOUT_SECONDS`, `PIPELINE_LOAD_TIMEOUT_MINUTES`, `CLI_RUN_TIMEOUT_MINUTES`, `WEBHOOK_REPORT_HTTP_TIMEOUT_SECONDS`, `WEBHOOK_REPORT_RESULT_WAIT_SECONDS`, `HTTP_*_TIMEOUT_SECONDS`, `SHUTDOWN_TIMEOUT_SECONDS`) должны быть больше 0.
- Для Loki/Grafana используйте `LOG_FORMAT=json` и `LOG_BACKEND=zerolog`.

## Timeout Map

- `DB_CONNECT_TIMEOUT_SECONDS` - timeout на установление и первичный `Ping()` PostgreSQL.
- `FTP_CONNECT_TIMEOUT_SECONDS` - timeout на `ftp.Dial(...)`.
- `WAIT_DELAY_MINUTES` - бизнес-ожидание Frontol перед чтением `response` файлов.
- `PIPELINE_LOAD_TIMEOUT_MINUTES` - лимит на стадию `load/reconcile` одного файла.
- `CLI_RUN_TIMEOUT_MINUTES` - внешний timeout CLI запусков ETL.
- `OPERATION_STALE_TIMEOUT_MINUTES` - TTL для operation-level lifecycle registry; после рестарта старые `started/queued/processing/timeout_reported` операции будут помечены как `abandoned`.
- `WEBHOOK_TIMEOUT_MINUTES` - timeout SLA webhook-отчета как бизнес-события, не HTTP клиента.
- `WEBHOOK_REPORT_HTTP_TIMEOUT_SECONDS` - timeout исходящего HTTP запроса на `WEBHOOK_REPORT_URL`.
- `WEBHOOK_REPORT_RESULT_WAIT_SECONDS` - сколько ждать сформированный итог после завершения pipeline before logging timeout warning.
- `HTTP_READ_HEADER_TIMEOUT_SECONDS`, `HTTP_READ_TIMEOUT_SECONDS`, `HTTP_WRITE_TIMEOUT_SECONDS`, `HTTP_IDLE_TIMEOUT_SECONDS` - таймауты встроенного `http.Server`.
- `SHUTDOWN_TIMEOUT_SECONDS` - лимит graceful shutdown для `webhook-server`.

## Loki/Grafana Logging

После reliability/logging refactor приложение пишет Loki-friendly structured events в JSON при `LOG_FORMAT=json`.

Прямая отправка в Loki включается через `LOG_SINK=loki` или `LOG_SINK=both`.
Без этой настройки поведение остается прежним: логи идут только в `stdout`.

Ключевые типы событий:

- `api_request_received`
- `api_request_completed`
- `api_request_rejected`
- `download_export_summary`
- `etl_run_summary`
- `etl_kassa_summary`
- `auth_missing`, `auth_invalid_format`, `auth_invalid_token`

Полезные общие поля для Loki/Grafana:

- `event`
- `log_kind`
- `component`
- `operation_id`
- `request_id`
- `endpoint`
- `operation`
- `status` / `outcome`
- `duration_ms`
- `date`
- `source_folder`
- `client_ip`

`request_id` описывает HTTP request scope, а `operation_id` сопровождает бизнес-операцию от приема запроса до завершения загрузки или выгрузки.

---

## 📂 Примеры конфигураций

### Development

**`.env.development`:**

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=frontol
DB_PASSWORD=dev_password
DB_NAME=kassa_db
DB_SSLMODE=disable

# FTP
PUBLICHOST=localhost
FTP_HOST=localhost
FTP_PORT=21
FTP_USER=frontol
FTP_PASSWORD=frontol123
FTP_REQUEST_DIR=/request
FTP_RESPONSE_DIR=/response

# Kassa
KASSA_STRUCTURE=P13:P13;N22:N22_Inter,N22_FURN

# Application
LOCAL_DIR=/tmp/frontol
BATCH_SIZE=10
MAX_RETRIES=3
RETRY_DELAY_SECONDS=5
WAIT_DELAY_MINUTES=1
LOG_LEVEL=debug

# Webhook
SERVER_PORT=8080
WEBHOOK_REPORT_URL=
WEBHOOK_BEARER_TOKEN=
```

**Использование:**

```bash
cp .env.development .env
docker-compose up -d
```

---

### Production

**`.env.production`:**

```bash
# Database (внешний PostgreSQL кластер)
DB_HOST=postgres.prod.example.com
DB_PORT=5432
DB_USER=frontol_prod
DB_PASSWORD=Xy9$m2Kp3L#n  # Сложный пароль!
DB_NAME=kassa_production
DB_SSLMODE=require  # Обязательно SSL в production

# FTP (production FTP server)
PUBLICHOST=ftp.prod.example.com
FTP_HOST=ftp.prod.example.com
FTP_PORT=21
FTP_USER=frontol_prod
FTP_PASSWORD=SecureFtp2024!  # Сложный пароль!
FTP_REQUEST_DIR=/requests
FTP_RESPONSE_DIR=/responses

# Kassa (все кассы)
KASSA_STRUCTURE=P13:P13;N22:N22_Inter,N22_FURN;SH54:SH54;M17:M17

# Application
LOCAL_DIR=/data/frontol
BATCH_SIZE=5000
MAX_RETRIES=5
RETRY_DELAY_SECONDS=10
WAIT_DELAY_MINUTES=2
LOG_LEVEL=info

# Webhook
SERVER_PORT=8080
WEBHOOK_REPORT_URL=https://monitoring.example.com/etl/reports
WEBHOOK_BEARER_TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Использование:**

```bash
# На production сервере
cp .env.production .env
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

---

### Testing

**`.env.testing`:**

```bash
# Database (test DB)
DB_HOST=localhost
DB_PORT=5432
DB_USER=frontol_test
DB_PASSWORD=test_password
DB_NAME=kassa_test
DB_SSLMODE=disable

# FTP (local test server)
PUBLICHOST=localhost
FTP_HOST=localhost
FTP_PORT=21
FTP_USER=frontol
FTP_PASSWORD=frontol123
FTP_REQUEST_DIR=/request
FTP_RESPONSE_DIR=/response

# Kassa (minimal)
KASSA_STRUCTURE=P13:P13

# Application (fast for tests)
LOCAL_DIR=/tmp/frontol-test
BATCH_SIZE=10
MAX_RETRIES=1
RETRY_DELAY_SECONDS=1
WAIT_DELAY_MINUTES=0
LOG_LEVEL=debug

# Webhook
SERVER_PORT=8080
WEBHOOK_REPORT_URL=
WEBHOOK_BEARER_TOKEN=
```

---

## 🔐 Безопасность

### 1. Обязательно измените пароли

```bash
# ❌ НЕ ИСПОЛЬЗУЙТЕ в production:
DB_PASSWORD=password
FTP_PASSWORD=frontol123

# ✅ Используйте сложные пароли:
DB_PASSWORD=Xy9$m2Kp3L#n
FTP_PASSWORD=SecureFtp2024!
```

### 2. Генерация безопасных паролей

```bash
# Генерация случайного пароля (32 символа)
openssl rand -base64 32

# Генерация Bearer token
openssl rand -hex 32
```

### 3. SSL для базы данных в production

```bash
# Обязательно в production
DB_SSLMODE=require

# Или с проверкой сертификата
DB_SSLMODE=verify-full
```

### 4. Не коммитьте .env в git

```bash
# .gitignore (уже добавлено)
.env
.env.*
!env.example
```

### 5. Используйте Bearer token для webhook

```bash
# Генерация токена
WEBHOOK_BEARER_TOKEN=$(openssl rand -hex 32)
echo "WEBHOOK_BEARER_TOKEN=$WEBHOOK_BEARER_TOKEN" >> .env

# Использование
curl -X POST http://localhost:$SERVER_PORT/api/load \
  -H "Authorization: Bearer $WEBHOOK_BEARER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"date": "2024-12-18"}'
```

---

## ✅ Валидация конфигурации

### Автоматическая валидация

Приложение проверяет обязательные переменные при старте:

```go
func validateConfig(cfg *Config) error {
    if cfg.DBHost == "" {
        return errors.New("DB_HOST is required")
    }
    if cfg.DBPassword == "" {
        return errors.New("DB_PASSWORD is required")
    }
    // ... остальные проверки
    return nil
}
```

### Ручная проверка

```bash
# Проверка всех переменных
docker-compose config

# Проверка конкретной переменной
docker-compose config | grep DB_HOST

# Проверка в реальном времени (внутри контейнера)
docker-compose exec webhook-server env | grep DB_
```

### Синтаксис в docker-compose.yml

```yaml
# Формат
${VARIABLE_NAME:-default_value}

# Примеры:
DB_HOST: ${DB_HOST}  # Обязательная (no default)
DB_PORT: ${DB_PORT:-5432}  # Опциональная (default 5432)
```

---

## 🔍 Проверка конфигурации

### Скрипт проверки

**`check-config.sh`:**

```bash
#!/bin/bash

echo "=== Configuration Check ==="

# Обязательные переменные
REQUIRED_VARS=(
    "DB_HOST"
    "DB_USER"
    "DB_PASSWORD"
    "DB_NAME"
    "FTP_HOST"
    "FTP_USER"
    "FTP_PASSWORD"
    "KASSA_STRUCTURE"
)

# Проверка
MISSING=0
for var in "${REQUIRED_VARS[@]}"; do
    value=$(grep "^$var=" .env | cut -d'=' -f2)
    if [ -z "$value" ]; then
        echo "❌ $var is missing or empty"
        MISSING=$((MISSING+1))
    else
        echo "✅ $var is set"
    fi
done

if [ $MISSING -gt 0 ]; then
    echo ""
    echo "❌ $MISSING required variables are missing"
    exit 1
else
    echo ""
    echo "✅ All required variables are set"
    exit 0
fi
```

**Использование:**

```bash
chmod +x check-config.sh
./check-config.sh
```

---

## 🔧 Troubleshooting

### Переменная не подставляется

**Проблема:**
```
Warning: The "DB_HOST" variable is not set. Defaulting to a blank string.
```

**Решение:**

```bash
# Проверить .env
cat .env | grep DB_HOST

# Если отсутствует, добавить
echo "DB_HOST=postgres.example.com" >> .env

# Перезапустить
docker-compose down
docker-compose up -d
```

### Значение неправильное

**Проверка:**

```bash
# Что в .env
cat .env | grep DB_HOST

# Что использует docker compose
docker-compose config | grep DB_HOST

# В реальном времени в контейнере
docker-compose exec webhook-server env | grep DB_HOST
```

### Пароль с специальными символами

**Используйте кавычки:**

```bash
# ❌ Плохо:
DB_PASSWORD=pass$word

# ✅ Хорошо:
DB_PASSWORD="pass$word"
# или
DB_PASSWORD='pass$word'
```
---

**Последнее обновление:** 2026-01-03
