# 🛠️ Технический стек

## 📋 Содержание

1. [Языки и фреймворки](#языки-и-фреймворки)
2. [База данных](#база-данных)
3. [Библиотеки и пакеты](#библиотеки-и-пакеты)
4. [Инструменты разработки](#инструменты-разработки)
5. [Инфраструктура](#инфраструктура)
6. [Почему именно эти технологии?](#почему-именно-эти-технологии)

---

## 💻 Языки и фреймворки

### Go 1.24+

**Основной язык разработки**

- **Версия:** 1.24 или выше
- **Преимущества для ETL:**
  - Высокая производительность
  - Встроенная поддержка concurrency (горутины)
  - Простая работа с сетью и файлами
  - Быстрая компиляция
  - Минимальные зависимости

**Ключевые возможности Go, используемые в проекте:**

```go
// 1. Горутины для асинхронной обработки
go s.runETLPipeline(requestID, date)

// 2. Channels для коммуникации
results := make(chan *Result, 10)

// 3. Context для управления жизненным циклом
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
defer cancel()

// 4. Defer для гарантированного выполнения
defer file.Close()
defer tx.Rollback(ctx)

// 5. Error wrapping
return fmt.Errorf("failed to parse file: %w", err)
```

---

## 🗄️ База данных

### PostgreSQL 17+

**Реляционная СУБД для хранения транзакций**

- **Версия:** 17 или выше
- **Особенности использования:**
  - 23 таблицы (2 справочника + 21 таблица транзакций)
  - Составные первичные ключи для идемпотентности
  - Индексы для оптимизации запросов
  - Транзакционная обработка (ACID)

**Ключевые возможности PostgreSQL:**

```sql
-- 1. ON CONFLICT для upsert операций
INSERT INTO transactions (...)
ON CONFLICT (id, folder) DO UPDATE SET ...

-- 2. JSONB для хранения динамических данных
reserved_fields JSONB

-- 3. Partial indexes для оптимизации
CREATE INDEX idx_sales ON transactions (date) WHERE operation_type = 0

-- 4. Транзакции для атомарности
BEGIN;
INSERT INTO ...;
INSERT INTO ...;
COMMIT;
```

---

## 📦 Библиотеки и пакеты

### Core Libraries (стандартная библиотека Go)

| Пакет | Назначение |
|-------|------------|
| `context` | Управление жизненным циклом операций |
| `encoding/json` | Работа с JSON |
| `net/http` | HTTP server и client |
| `os`, `io`, `bufio` | Работа с файлами |
| `time` | Работа с датой и временем |
| `github.com/rs/zerolog` | Основной backend structured logging (JSON/console) |
| `log/slog` | Фолбэк backend (временный, через feature-flag `LOG_BACKEND`) |

### Third-Party Libraries

#### 1. Database - pgx/v5

**Нативный драйвер PostgreSQL для Go**

```
github.com/jackc/pgx/v5
github.com/jackc/pgx/v5/pgxpool
```

**Почему pgx, а не database/sql или GORM:**

| Критерий | pgx ✅ | database/sql | GORM |
|----------|--------|--------------|------|
| Batch insert | Максимальная скорость | Медленнее | Overhead ORM |
| Memory | Минимальный | Средний | Высокий (рефлексия) |
| ETL операции | Идеален ⭐ | Подходит | Избыточен |
| Контроль SQL | Полный | Полный | Абстрагирован |
| PostgreSQL features | Полная поддержка | Базовая | Ограниченная |

**Пример использования:**

```go
// Connection pool
config, _ := pgxpool.ParseConfig(dsn)
config.MaxConns = 10
config.MinConns = 2
pool, _ := pgxpool.NewWithConfig(ctx, config)

// Batch operations
batch := &pgx.Batch{}
batch.Queue("INSERT INTO ...", args...)
results := pool.SendBatch(ctx, batch)
```

#### 2. FTP - jlaffaye/ftp

**FTP клиент для Go**

```
github.com/jlaffaye/ftp
```

**Возможности:**
- Подключение к FTP серверу
- Список файлов в директории
- Download/Upload файлов
- Создание/удаление директорий
- Passive mode support

**Пример:**

```go
// FTP_PORT берется из .env
conn, _ := ftp.Dial("ftp.example.com:"+os.Getenv("FTP_PORT"), ftp.DialWithTimeout(5*time.Second))
conn.Login(user, password)
files, _ := conn.List("/response")
```

#### 3. Migrations - golang-migrate

**Управление миграциями базы данных**

```
github.com/golang-migrate/migrate/v4
```

**Особенности:**
- Embedded migrations (встроенные в бинарник)
- Up/Down миграции
- Version tracking
- Dirty state handling

**Структура:**

```
pkg/migrate/migrations/
├── 000001_init_schema.up.sql
├── 000001_init_schema.down.sql
├── 000002_add_missing_constraints.up.sql
├── 000002_add_missing_constraints.down.sql
├── 000003_add_missing_tx_constraints.up.sql
├── 000003_add_missing_tx_constraints.down.sql
├── 000004_add_etl_file_load_state.up.sql
├── 000004_add_etl_file_load_state.down.sql
├── 000005_add_etl_operation_runs.up.sql
└── 000005_add_etl_operation_runs.down.sql
```

#### 4. Configuration - godotenv

**Загрузка переменных окружения из .env файла**

```
github.com/joho/godotenv
```

**Использование:**

```go
// Опциональная загрузка .env
if err := godotenv.Load(); err != nil {
    // .env файл не найден, используем переменные окружения
}

dbHost := os.Getenv("DB_HOST")
```

#### 5. API Documentation - scalar-go

**Интерактивная документация API (Scalar)**

```
github.com/bdpiprava/scalar-go
```

**Возможности:**
- Автоматическая генерация из OpenAPI спецификации
- Интерактивный интерфейс
- Try-it-out функциональность
- Красивый современный UI

---

## 🔧 Инструменты разработки

### Линтер - golangci-lint (v2.8.0)

**Статический анализатор кода**

```yaml
# .golangci.yml
run:
  timeout: 5m

linters:
  enable:
    - errcheck      # Проверка обработки ошибок
    - gosimple      # Упрощение кода
    - govet         # Проверка Go конструкций
    - ineffassign   # Неиспользуемые присваивания
    - staticcheck   # Статический анализ
    - unused        # Неиспользуемый код
    - gofmt         # Форматирование
    - goimports     # Импорты
```

**Команды:**

```bash
# Установка (Go 1.24+)
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.8.0

# Запуск линтера
golangci-lint run

# Автофикс некоторых проблем
golangci-lint run --fix
```

### Форматирование - gofmt

**Стандартное форматирование Go кода**

```bash
# Форматировать все файлы
go fmt ./...

# Или через goimports (форматирование + импорты)
goimports -w .
```

### Тестирование

**Встроенный фреймворк Go для тестирования**

```bash
# Unit тесты
go test ./...

# С verbose
go test -v ./...

# С покрытием
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Race detector
go test -race ./...

# Benchmarks
go test -bench=. ./...
```

**Пример теста:**

```go
func TestParseDate(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    time.Time
        wantErr bool
    }{
        {"valid date", "01.12.2024", time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC), false},
        {"invalid format", "2024-12-01", time.Time{}, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := parseDate(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("parseDate() error = %v, wantErr %v", err, tt.wantErr)
            }
            if !got.Equal(tt.want) {
                t.Errorf("parseDate() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Build Tool - Makefile

**Автоматизация сборки и развертывания**

```makefile
# Основные команды
.PHONY: build test lint fmt

build:
    go build -o bin/webhook-server ./cmd/webhook-server
    go build -o bin/loader ./cmd/loader

test:
    go test -v ./...

lint:
    golangci-lint run

fmt:
    go fmt ./...
```

---

## 🐳 Инфраструктура

### Docker

**Контейнеризация приложений**

**Dockerfile (multi-stage build):**

```dockerfile
# Этап 1: Сборка
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/webhook-server ./cmd/webhook-server

# Этап 2: Production образ
FROM alpine:3.19
COPY --from=builder /app/webhook-server /webhook-server
CMD ["/webhook-server"]
```

**Преимущества multi-stage:**
- ✅ Минимальный размер финального образа
- ✅ Нет build dependencies в production
- ✅ Безопасность (меньше attack surface)

### Docker Compose

**Оркестрация нескольких сервисов**

```yaml
version: '3.8'

services:
  # Init контейнер для миграций
  migrate:
    image: migrate/migrate
    command: ["-path=/migrations", "-database", "${DB_DSN}", "up"]

  # FTP сервер (для тестирования)
  ftp-server:
    image: fauria/vsftpd
    ports:
      - "${FTP_PORT}:${FTP_PORT}"

  # Webhook server (постоянный)
  webhook-server:
    build: .
    ports:
      - "${SERVER_PORT}:${SERVER_PORT}"
    environment:
      - DB_HOST=${DB_HOST}
      - FTP_HOST=ftp-server
    depends_on:
      - migrate
```

**Преимущества:**
- ✅ Автоматическое применение миграций
- ✅ Простой запуск всей инфраструктуры
- ✅ Изоляция сервисов

### FTP Server (для тестирования)

**fauria/vsftpd**

```yaml
ftp-server:
  image: fauria/vsftpd
  environment:
    - FTP_USER=frontol
    - FTP_PASS=frontol123
    - PASV_ADDRESS=127.0.0.1
    - PASV_MIN_PORT=21100
    - PASV_MAX_PORT=21110
  ports:
    - "${FTP_PORT}:${FTP_PORT}"
    - "${PASV_MIN_PORT}-${PASV_MAX_PORT}:${PASV_MIN_PORT}-${PASV_MAX_PORT}"
  volumes:
    - ftp-data:/home/vsftpd
```

---

## 🎨 Архитектурные паттерны

### Clean Architecture

**Разделение на слои:**

```
cmd/          - Presentation Layer (тонкий слой)
pkg/pipeline  - Business Logic Layer
pkg/repository - Data Access Layer
pkg/db, pkg/ftp - Infrastructure Layer
```

### Dependency Injection

**Инъекция зависимостей через конструкторы:**

```go
type Pipeline struct {
    cfg       *models.Config
    ftpClient *ftp.Client
    dbPool    *db.Pool
    logger    *logger.Logger
}

func NewPipeline(
    cfg *models.Config,
    ftpClient *ftp.Client,
    dbPool *db.Pool,
    logger *logger.Logger,
) *Pipeline {
    return &Pipeline{
        cfg:       cfg,
        ftpClient: ftpClient,
        dbPool:    dbPool,
        logger:    logger,
    }
}
```

### Repository Pattern

**Изоляция логики доступа к данным:**

```go
type Repository interface {
    LoadTransactionRegistrations(ctx context.Context, data []models.TransactionRegistration) error
    LoadBonusTransactions(ctx context.Context, data []models.BonusTransaction) error
}
```

---

## 🤔 Почему именно эти технологии?

### Go вместо Python/Java/Node.js

| Критерий | Go ✅ | Python | Java | Node.js |
|----------|------|--------|------|---------|
| Производительность | Высокая | Низкая | Высокая | Средняя |
| Memory footprint | Низкий | Средний | Высокий | Средний |
| Concurrency | Горутины ⭐ | GIL ограничение | Threads | Event loop |
| Deployment | Один бинарник ⭐ | Dependencies | JVM + JAR | Node + modules |
| Типизация | Статическая ⭐ | Динамическая | Статическая | Динамическая |
| Простота | Простой | Простой | Сложный | Средний |

**Вывод:** Go идеален для ETL задач с высокой производительностью и простым deployment.

### pgx вместо GORM/database/sql

| Критерий | pgx ✅ | GORM | database/sql |
|----------|--------|------|--------------|
| Batch insert | Нативная поддержка ⭐ | Через Raw SQL | Через циклы |
| PostgreSQL features | Полная поддержка ⭐ | Ограничена | Базовая |
| Производительность | Максимальная ⭐ | Overhead ORM | Хорошая |
| Connection pool | pgxpool ⭐ | gorm.DB | sql.DB |
| Типы данных | PostgreSQL types | Generic | Generic |

**Вывод:** pgx обеспечивает максимальную производительность для batch ETL операций.

### PostgreSQL вместо MySQL/MongoDB

| Критерий | PostgreSQL ✅ | MySQL | MongoDB |
|----------|--------------|-------|---------|
| ACID | Полная ⭐ | Полная | Eventual |
| JSON | JSONB ⭐ | JSON | Native |
| Индексы | Разнообразные ⭐ | Базовые | Хорошие |
| Analytical queries | Отличные ⭐ | Хорошие | Слабые |
| Транзакции | Полные ⭐ | Полные | Ограниченные |

**Вывод:** PostgreSQL лучше подходит для аналитических запросов и сложных транзакций.

### Docker Compose вместо Kubernetes

| Критерий | Docker Compose ✅ | Kubernetes |
|----------|-------------------|------------|
| Сложность | Низкая ⭐ | Высокая |
| Deployment | docker-compose up ⭐ | Множество манифестов |
| Масштабируемость | Ограниченная | Высокая |
| Подходит для | Small/Medium проектов ⭐ | Large проектов |

**Вывод:** Docker Compose достаточен для текущего масштаба проекта. Kubernetes можно добавить позже при необходимости.

---

## 📊 Метрики производительности

### Типичные показатели

**Парсинг:**
- ~10 μs на транзакцию (benchmark)
- ~10,000 транзакций в секунду

**Загрузка в БД:**
- ~1,000 строк в batch
- ~50,000 транзакций в секунду (с batch insert)

**Полный ETL цикл:**
- ~1-2 минуты для 10,000 транзакций
- ~10-15 минут для 100,000 транзакций

**Memory:**
- Webhook server: ~30-50 MB
- Loader: ~50-100 MB (зависит от размера файлов)

---

## 🔄 Обновление зависимостей

```bash
# Обновить все зависимости
go get -u ./...
go mod tidy

# Обновить конкретную зависимость
go get -u github.com/jackc/pgx/v5

# Проверить устаревшие зависимости
go list -u -m all
```

---

## 📚 См. также

- [ARCHITECTURE.md](ARCHITECTURE.md) - Архитектура проекта
- [../coding/CODING_RULES.md](../coding/CODING_RULES.md) - Правила написания кода
- [../test/TESTING.md](../test/TESTING.md) - Руководство по тестированию
- [Go Official Documentation](https://go.dev/doc/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Docker Documentation](https://docs.docker.com/)

---

**Последнее обновление:** 2026-01-03
