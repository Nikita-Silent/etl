# Руководство по проверке работоспособности

Данный документ описывает последовательность проверки всех компонентов проекта Frontol ETL.

Примечание: PostgreSQL в этом репозитории внешний. Команды с `psql` ниже предполагают локально доступный клиент `psql` и корректно заполненные `DB_*` переменные.

---

## 📋 Предварительные требования

### 1. Установленное ПО

```bash
# Проверка Go версии (должна быть >= 1.24)
go version

# Проверка Docker и Docker Compose
docker --version
docker-compose --version

# Проверка golangci-lint (для линтера)
golangci-lint --version
```

### 2. Конфигурация окружения

```bash
# Скопировать пример конфигурации
cp env.example .env

# Отредактировать .env (установить пароли)
nano .env
```

**Минимально необходимые параметры в `.env`:**
```env
DB_HOST=postgres.example.com
DB_USER=frontol_user
DB_PASSWORD=your_secure_password
FTP_USER=frontol
FTP_PASSWORD=your_ftp_password
KASSA_STRUCTURE=P13:P13;N22:N22_Inter,N22_FURN
```

---

## 🧪 1. Проверка компиляции

### Сборка всех бинарников

```bash
# Очистка старых бинарников
make clean-local

# Сборка всех бинарников
make build-local

# Проверка созданных файлов
ls -lh webhook-server frontol-loader frontol-loader-local parser-test send-request clear-requests migrate
```

**Ожидаемый результат:** Все 7 бинарников успешно скомпилированы.

---

## 🧪 2. Проверка кода

### Форматирование

```bash
# Проверка форматирования
make fmt
```

**Ожидаемый результат:** Нет изменений в файлах (код уже отформатирован).

### Линтер

```bash
# Запуск линтера
make lint
```

**Ожидаемый результат:** Нет ошибок или предупреждений.

---

## 🧪 3. Unit тесты HTTP API (httptest)

```bash
# Запуск unit тестов API
go test ./cmd/webhook-server
```

**Ожидаемый результат:** Все тесты проходят (статус `ok`).

### Unit тесты

```bash
# Запуск всех unit тестов
make test-go

# С подробным выводом
make test-verbose

# С покрытием кода
make test-coverage
# Откройте coverage.html в браузере

# С race detector
make test-race

# Focused reliability regression suite
make test-reliability

# Race detection for critical ETL packages
make test-race-critical

# Бенчмарки
make test-bench
```

**Ожидаемый результат:** Все тесты проходят успешно.

**Список протестированных пакетов:**
- ✅ pkg/config
- ✅ pkg/db
- ✅ pkg/ftp
- ✅ pkg/logger
- ✅ pkg/models
- ✅ pkg/parser
- ✅ pkg/server
- ✅ pkg/validation
- ✅ pkg/migrate

### Быстрая проверка

```bash
# Комбо: форматирование + линтер + тесты
make check

# CI gate: lint + focused regression + race on critical packages
make ci
```

---

## 🐳 3. Проверка Docker окружения

### Запуск сервисов

```bash
# Запуск в development режиме
make dev

# Проверка запущенных контейнеров
make status

# Просмотр логов
make logs

# Логи конкретного сервиса
make logs-webhook
make logs-db
make logs-ftp
```

**Ожидаемые сервисы:**
- ✅ `migrate` - one-shot миграции (обычно `exited 0`)
- ✅ `ftp-server` - FTP сервер
- ✅ `webhook-server` - HTTP webhook сервер

### Проверка доступности сервисов

```bash
# Проверка внешнего PostgreSQL
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "${DB_PORT:-5432}" -U "$DB_USER" "$DB_NAME" -c "SELECT 1;"

# Проверка webhook server
curl http://localhost:$SERVER_PORT/api/health

# Проверка FTP (если установлен ftp клиент)
ftp localhost $FTP_PORT
# Логин: frontol
# Пароль: из .env
```

---

## 🗄️ 4. Проверка миграций базы данных

### Применение миграций

```bash
# Показать текущую версию (должна быть пустая)
make migrate-version

# Применить все миграции
make migrate-up

# Проверить версию (должна быть 5)
make migrate-version
```

**Ожидаемый результат:**
```
Current version: 5
```

### Проверка структуры БД

```bash
# Подключиться к внешней БД
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "${DB_PORT:-5432}" -U "$DB_USER" "$DB_NAME"

# Проверить таблицы
\dt

# Выход
\q
```

### Тест отката миграций

```bash
# Откатить 1 миграцию
make migrate-step N=-1

# Проверить версию (должна быть 4)
make migrate-version

# Применить обратно
make migrate-step N=1

# Проверить версию (должна быть 5)
make migrate-version
```

---

## 🔌 5. Проверка HTTP Webhook Server

### Health Check

```bash
# Проверка здоровья сервера
curl http://localhost:$SERVER_PORT/api/health

# Ожидаемый ответ:
# {
#   "status": "healthy",
#   "timestamp": "2024-12-18T12:00:00Z",
#   "service": "frontol-etl-webhook",
#   "checks": { ... },
#   "response_time_ms": 20
# }
```

### Тестовый webhook запрос

```bash
# Отправка тестового webhook
curl -X POST http://localhost:$SERVER_PORT/api/load \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2024-12-01"
  }'

# Ожидаемый ответ:
# {
#   "status": "queued",
#   "date": "2024-12-01",
#   "message": "Request added to queue",
#   "request_id": "req_..."
# }
```

### Проверка логов webhook

```bash
# Смотрим логи обработки
make logs-webhook

# Должны увидеть structured logs:
# [WEBHOOK] INFO HTTP request method=POST path=/api/load status=202 ...
# [WEBHOOK] INFO ETL pipeline started date=2024-12-01 ...
```

---

## 📊 6. Проверка парсера

### Тест парсинга тестового файла

```bash
# Запуск парсера на тестовых данных
./parser-test data/response.txt

# Или через Docker
docker-compose run --rm parser-test ./parser-test /app/data/response.txt
```

**Ожидаемый результат:** Успешный парсинг транзакций из файла.

### Проверка функций парсера в Go

```go
// Создайте test_parser.go
package main

import (
    "fmt"
    "log"
    "github.com/user/go-frontol-loader/pkg/parser"
)

func main() {
    transactions, header, err := parser.ParseFile(
        "data/response.txt",
        "test_folder",
    )
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Header: DBID=%s, ReportNum=%s\n", 
        header.DBID, header.ReportNum)
    fmt.Printf("Transactions parsed: %d types\n", len(transactions))
}
```

```bash
go run test_parser.go
```

---

## 📁 7. Проверка FTP операций

### Проверка FTP клиента

```go
// Создайте test_ftp.go
package main

import (
    "fmt"
    "log"
    "github.com/user/go-frontol-loader/pkg/config"
    "github.com/user/go-frontol-loader/pkg/ftp"
)

func main() {
    cfg, _ := config.LoadConfig()
    
    client, err := ftp.NewClient(cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    files, err := client.ListFiles("/")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("FTP files: %d\n", len(files))
}
```

```bash
go run test_ftp.go
```

### Отправка запросов к кассам

```bash
# Отправить request.txt во все кассы
./send-request

# Или через Docker
docker-compose run --rm send-request
```

### Очистка папок

```bash
# Очистить папки request и response
./clear-requests

# Или через Docker
docker-compose run --rm clear-requests
```

---

## 🔄 8. Полный ETL пайплайн (E2E тест)

### Подготовка

```bash
# 1. Убедитесь что все сервисы запущены
make status

# 2. Примените миграции
make migrate-up

# 3. Очистите FTP папки
make clear-requests
```

### Запуск ETL

```bash
# Отправить request.txt к кассам
make send-request

# Подождать ответа от касс (зависит от настройки)
sleep 60

# Запустить loader
make loader

# Или для конкретной даты
make loader-date DATE=2024-12-01
```

### Проверка результатов

```bash
# Подключиться к внешней БД
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "${DB_PORT:-5432}" -U "$DB_USER" "$DB_NAME"

# Проверить загруженные данные
SELECT COUNT(*) FROM tx_item_registration_1_11;
SELECT COUNT(*) FROM tx_special_price_3;
SELECT COUNT(*) FROM tx_bonus_accrual_9;

# Проверить последние транзакции
SELECT 
    transaction_date,
    transaction_time,
    transaction_type,
    source_folder,
    COUNT(*)
FROM tx_item_registration_1_11
GROUP BY transaction_date, transaction_time, transaction_type, source_folder
ORDER BY transaction_date DESC, transaction_time DESC
LIMIT 10;
```

---

## 🧪 9. Интеграционные тесты

### Запуск интеграционных тестов

```bash
# Требуют запущенные сервисы (БД, FTP)
INTEGRATION_TEST=true make test-integration
```

**Что тестируется:**
- ✅ Подключение к PostgreSQL
- ✅ Подключение к FTP серверу
- ✅ Структура kassa folders
- ✅ Создание и загрузка request файлов

---

## 🎯 10. Проверка structured logging

### Создание тестового логгера

```go
// test_logger.go
package main

import (
    "context"
    "github.com/user/go-frontol-loader/pkg/logger"
)

func main() {
    // Text формат
    log := logger.New(logger.Config{
        Level:  "debug",
        Format: "text",
    })
    
    log.Info("Test message", "key", "value")
    log.Error("Error message", "error", "something went wrong")
    
    // JSON формат
    jsonLog := logger.New(logger.Config{
        Level:  "info",
        Format: "json",
    })
    
    ctx := context.Background()
    jsonLog.LogETLStart(ctx, "2024-12-01")
    jsonLog.LogETLEnd(ctx, "2024-12-01", 10, 1000, nil)
}
```

```bash
go run test_logger.go
```

---

## 🎨 11. Проверка валидации

### Тест валидатора

```go
// test_validator.go
package main

import (
    "fmt"
    "github.com/user/go-frontol-loader/pkg/validation"
)

func main() {
    dateValidator := validation.NewComposite(
        validation.Required("date"),
        validation.DateFormat("date", "2006-01-02"),
        validation.NotInFuture("date", "2006-01-02"),
    )

    if err := dateValidator.Validate("2024-12-18"); err != nil {
        fmt.Println("Validation error:", err)
        return
    }

    fmt.Println("✓ Validation passed")
    
    // Тест валидации даты
    dateValidator := validator.DateValidator{}
    if err := dateValidator.ValidateDate("2024-12-01"); err != nil {
        fmt.Printf("Date validation error: %v\n", err)
    } else {
        fmt.Println("✓ Date is valid")
    }
}
```

```bash
go run test_validator.go
```

---

## ✅ Чеклист проверки

### Базовые проверки

- [ ] Go версия >= 1.24
- [ ] Все бинарники компилируются без ошибок
- [ ] `make fmt` не вносит изменений
- [ ] `make lint` проходит без ошибок
- [ ] `make test-go` - все тесты проходят
- [ ] `make test-race` - нет race conditions

### Docker окружение

- [ ] `make dev` запускает все сервисы
- [ ] Внешний PostgreSQL доступен по `DB_HOST`/`DB_PORT`
- [ ] FTP сервер доступен на порту $FTP_PORT
- [ ] Webhook server доступен на порту $SERVER_PORT
- [ ] `curl http://localhost:$SERVER_PORT/api/health` возвращает `healthy`

### Миграции

- [ ] `make migrate-up` применяет 5 миграций
- [ ] `make migrate-version` показывает версию 5
- [ ] В БД создано 16+ таблиц
- [ ] `make migrate-down` откатывает миграции

### Функциональность

- [ ] Парсер обрабатывает тестовый файл
- [ ] FTP клиент подключается к серверу
- [ ] Webhook принимает POST запросы
- [ ] Логирование работает (text и json)
- [ ] Валидация конфигурации работает

### Производительность

- [ ] Бенчмарки парсера < 10 μs на транзакцию
- [ ] Бенчмарки логгера < 3 μs на запись
- [ ] Покрытие тестами > 70%

---

## 🐛 Решение проблем

### Проблема: Миграции не применяются

```bash
# Проверить подключение к БД
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "${DB_PORT:-5432}" -U "$DB_USER" "$DB_NAME" -c "SELECT 1;"

# В compose нет контейнера PostgreSQL: БД внешняя, проверьте доступность хоста и креды
make logs-db

# Принудительно установить версию (если dirty)
make migrate-force V=0
make migrate-up
```

### Проблема: Тесты падают

```bash
# Установить все зависимости
go mod download
go mod tidy

# Очистить кэш
go clean -testcache

# Запустить с verbose
go test -v ./...
```

### Проблема: Docker сервисы не запускаются

```bash
# Остановить и удалить все контейнеры
make down

# Пересоздать с нуля
make clean
make build
make up
```

### Проблема: FTP недоступен

```bash
# Проверить порт
netstat -an | grep $FTP_PORT

# Проверить логи FTP сервера
make logs-ftp

# Проверить .env настройки
cat .env | grep FTP
```

---

## 📚 Дополнительные команды

```bash
# Показать статус всех сервисов
make status

# Открыть shell в webhook контейнере
make shell

# Бэкап базы данных
make backup-db

# Восстановить базу данных
make restore-db FILE=backup_20241218_120000.sql

# Показать использование ресурсов
make stats

# CI pipeline (полная проверка)
make ci
```

---

## 🎯 Успешная проверка

Если все пункты чеклиста выполнены, проект готов к:
- ✅ Разработке новых функций
- ✅ Production деплою
- ✅ Code review
- ✅ CI/CD интеграции
