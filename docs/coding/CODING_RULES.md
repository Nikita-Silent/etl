# Правила написания кода для проекта Frontol ETL

**Версия Go:** 1.24+

Данный документ описывает стандарты и правила написания кода для обеспечения безопасности, производительности и единого стиля.

---

## 📁 Структура проекта

```
/
├── cmd/                    # Точки входа приложений
│   ├── loader/             # Основной ETL загрузчик
│   ├── webhook-server/     # HTTP сервер
│   └── ...
├── pkg/                    # Переиспользуемые пакеты
│   ├── config/             # Конфигурация
│   ├── db/                 # Работа с БД (pgx)
│   ├── ftp/                # FTP клиент
│   ├── logger/             # Structured logging (slog)
│   ├── models/             # Структуры данных
│   ├── parser/             # Парсинг файлов
│   ├── pipeline/           # ETL пайплайн
│   ├── server/             # HTTP сервер с graceful shutdown
│   └── validator/          # Валидация данных
├── tests/
│   ├── integration/        # Интеграционные тесты
│   └── testdata/           # Тестовые данные
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── .golangci.yml           # Конфигурация линтера
```

### Правила структуры:
- **cmd/** — только `main.go` с минимальной логикой (инициализация, запуск)
- **pkg/** — вся бизнес-логика, разделённая по доменам
- **tests/** — интеграционные тесты и тестовые данные
- Каждый пакет должен иметь одну чётко определённую ответственность
- Избегать циклических зависимостей между пакетами

### Почему pgx, а не GORM:

| Критерий | pgx ✅ | GORM |
|----------|--------|------|
| Batch insert | Максимальная скорость | Overhead ORM |
| Memory | Минимальный | Рефлексия |
| ETL операции | Идеален | Избыточен |
| Контроль SQL | Полный | Абстрагирован |

**Для ETL-проекта pgx оптимален.** GORM добавляет ~20-30% overhead без реальной пользы для массовых операций.

---

## 🔒 Безопасность

### 1. Конфиденциальные данные

```go
// ❌ ЗАПРЕЩЕНО: хардкод секретов
const DBPassword = "my_password"

// ✅ ПРАВИЛЬНО: из переменных окружения
password := os.Getenv("DB_PASSWORD")
if password == "" {
    return nil, fmt.Errorf("DB_PASSWORD is required")
}
```

**Обязательно:**
- Все пароли, токены, ключи — только через переменные окружения
- Использовать `.env` файлы только для локальной разработки
- `.env` добавлен в `.gitignore`
- В `env.example` — только примеры без реальных значений
- Никогда не логировать пароли и токены

### 2. Валидация входных данных

```go
// ✅ ПРАВИЛЬНО: валидация обязательных полей
func LoadConfig() (*models.Config, error) {
    config := &models.Config{...}
    
    if config.DBPassword == "" {
        return nil, fmt.Errorf("DB_PASSWORD is required")
    }
    if config.FTPUser == "" {
        return nil, fmt.Errorf("FTP_USER is required")
    }
    
    return config, nil
}
```

```go
// ✅ ПРАВИЛЬНО: валидация формата даты
date := req.Date
if date != "" {
    if _, err := time.Parse("2006-01-02", date); err != nil {
        http.Error(w, "Invalid date format. Expected YYYY-MM-DD", http.StatusBadRequest)
        return
    }
}
```

### 3. SQL безопасность

```go
// ❌ ЗАПРЕЩЕНО: конкатенация SQL запросов
query := "SELECT * FROM users WHERE id = " + userID

// ✅ ПРАВИЛЬНО: параметризованные запросы
query := "SELECT * FROM users WHERE id = $1"
rows, err := db.Query(ctx, query, userID)
```

### 4. Обработка файлов

```go
// ✅ ПРАВИЛЬНО: проверка и закрытие ресурсов
file, err := os.Open(filePath)
if err != nil {
    return nil, fmt.Errorf("failed to open file: %w", err)
}
defer file.Close()
```

### 5. HTTP безопасность

```go
// ✅ ПРАВИЛЬНО: проверка HTTP метода
if r.Method != http.MethodPost {
    http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    return
}

// ✅ ПРАВИЛЬНО: таймауты для HTTP клиентов
client := &http.Client{
    Timeout: 30 * time.Second,
}

// ✅ ПРАВИЛЬНО: закрытие тела ответа
resp, err := client.Do(req)
if err != nil {
    return err
}
defer resp.Body.Close()
```

---

## ⚡ Производительность

### 1. Connection Pooling

```go
// ✅ ПРАВИЛЬНО: использование пула соединений
func NewPool(cfg *models.Config) (*Pool, error) {
    config, err := pgxpool.ParseConfig(dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to parse database config: %w", err)
    }
    
    // Настройка пула
    config.MaxConns = 10
    config.MinConns = 2
    config.MaxConnLifetime = time.Hour
    config.MaxConnIdleTime = time.Minute * 30
    
    pool, err := pgxpool.NewWithConfig(context.Background(), config)
    // ...
}
```

### 2. Батчевая обработка

```go
// ✅ ПРАВИЛЬНО: пакетная вставка данных
func (p *Pool) LoadData(ctx context.Context, tableName string, columns []string, rows [][]interface{}) error {
    if len(rows) == 0 {
        return nil // Ранний выход при пустых данных
    }
    
    // Batch insert
    for _, row := range rows {
        _, err := p.Exec(ctx, query, row...)
        if err != nil {
            return fmt.Errorf("failed to insert data into %s: %w", tableName, err)
        }
    }
    return nil
}
```

### 3. Context и таймауты

```go
// ✅ ПРАВИЛЬНО: использование контекста с таймаутом
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
defer cancel()

result, err := pipeline.Run(ctx, s.logger, s.config, date)
```

```go
// ✅ ПРАВИЛЬНО: таймаут для подключения к БД
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

if err := pool.Ping(ctx); err != nil {
    pool.Close()
    return nil, fmt.Errorf("failed to ping database: %w", err)
}
```

### 4. Горутины

```go
// ✅ ПРАВИЛЬНО: асинхронная обработка
go s.runETLPipeline(requestID, date)
```

**Правила для горутин:**
- Всегда передавать context для возможности отмены
- Не создавать горутины в циклах без ограничения (использовать worker pool)
- Обеспечивать graceful shutdown

### 5. Буферизация

```go
// ✅ ПРАВИЛЬНО: буферизованное чтение файлов
scanner := bufio.NewScanner(file)
for scanner.Scan() {
    line := scanner.Text()
    // обработка строки
}
```

---

## 📝 Стиль кода

### 1. Именование

#### Пакеты
```go
// ✅ ПРАВИЛЬНО: короткие, в нижнем регистре, без подчёркиваний
package config
package db
package parser

// ❌ ЗАПРЕЩЕНО
package Config
package data_base
package my_parser
```

#### Переменные и функции
```go
// ✅ ПРАВИЛЬНО: camelCase для приватных
func parseTransactionLine(line string) {}
var transactionType int

// ✅ ПРАВИЛЬНО: PascalCase для экспортируемых
func LoadConfig() (*Config, error) {}
type TransactionRegistration struct {}
```

#### Константы
```go
// ✅ ПРАВИЛЬНО: PascalCase для экспортируемых констант
const DefaultBatchSize = 1000
const MaxRetries = 3

// ✅ ПРАВИЛЬНО: camelCase для приватных
const defaultTimeout = 30 * time.Second
```

#### Аббревиатуры
```go
// ✅ ПРАВИЛЬНО: аббревиатуры целиком в одном регистре
var userID string      // не userId
var ftpURL string      // не ftpUrl
type HTTPClient struct  // не HttpClient
var dbID string        // не dbId
```

### 2. Структуры

```go
// ✅ ПРАВИЛЬНО: группировка полей по смыслу с комментариями
type Config struct {
    // Database settings
    DBHost     string
    DBPort     int
    DBUser     string
    DBPassword string
    DBName     string
    DBSSLMode  string

    // FTP settings
    FTPHost        string
    FTPPort        int
    FTPUser        string
    FTPPassword    string
}
```

### 3. Комментарии

```go
// ✅ ПРАВИЛЬНО: комментарии для экспортируемых элементов
// LoadConfig loads configuration from .env file and environment variables
func LoadConfig() (*models.Config, error) {
    // ...
}

// Config represents application configuration
type Config struct {
    // ...
}

// ✅ ПРАВИЛЬНО: комментарии на русском для специфичных бизнес-терминов
type BillRegistration struct {
    BillCode         string  `json:"bill_code"`         // Поле №8: Код купюры
    GroupCode        string  `json:"group_code"`        // Поле №9: Код группы
    BillDenomination float64 `json:"bill_denomination"` // Поле №10: Достоинство купюры
}
```

### 4. Обработка ошибок

```go
// ✅ ПРАВИЛЬНО: оборачивание ошибок с контекстом
if err != nil {
    return nil, fmt.Errorf("failed to open file: %w", err)
}

// ✅ ПРАВИЛЬНО: ранний возврат при ошибках
file, err := os.Open(filePath)
if err != nil {
    return nil, fmt.Errorf("failed to open file: %w", err)
}
defer file.Close()

// ❌ ЗАПРЕЩЕНО: игнорирование ошибок
file, _ := os.Open(filePath)

// ✅ ПРАВИЛЬНО: если игнорирование осознанное — комментарий
if err := godotenv.Load(); err != nil {
    // .env file is optional, continue with environment variables
}
```

### 5. Импорты

```go
// ✅ ПРАВИЛЬНО: группировка импортов
import (
    // Стандартная библиотека
    "context"
    "fmt"
    "time"

    // Внешние зависимости
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    
    // Внутренние пакеты
    "github.com/user/go-frontol-loader/pkg/models"
)
```

### 6. JSON теги

```go
// ✅ ПРАВИЛЬНО: snake_case для JSON тегов
type TransactionRegistration struct {
    TransactionIDUnique int64  `json:"transaction_id_unique"`
    SourceFolder        string `json:"source_folder"`
    TransactionDate     string `json:"transaction_date"`
}
```

---

## 🧪 Тестирование

### Команды тестирования:

```bash
# Запуск всех тестов
make test-go

# Тесты с verbose output
make test-verbose

# Тесты с покрытием (генерирует coverage.html)
make test-coverage

# Тесты с race detector
make test-race

# Бенчмарки
make test-bench

# Интеграционные тесты (требуют запущенные сервисы)
make test-integration

# Полная проверка (fmt + lint + test)
make check

# CI pipeline (fmt + lint + test-race + coverage)
make ci
```

### Правила тестирования:

1. **Unit тесты** — в файлах `*_test.go` рядом с тестируемым кодом
2. **Интеграционные тесты** — в `tests/integration/` с build tag `integration`
3. **Тестовые данные** — в `tests/testdata/`
4. **Table-driven tests** — для множественных случаев
5. **Бенчмарки** — для критичных по производительности функций

```go
// ✅ ПРАВИЛЬНО: table-driven test
func TestParseDate(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    time.Time
        wantErr bool
    }{
        {"valid date", "02.01.2006", time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC), false},
        {"empty string", "", time.Time{}, true},
        {"invalid format", "2006-01-02", time.Time{}, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := parseDate(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("parseDate() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !got.Equal(tt.want) {
                t.Errorf("parseDate() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Бенчмарки:

```go
// ✅ ПРАВИЛЬНО: бенчмарк с ResetTimer
func BenchmarkParseTransactionLine(b *testing.B) {
    line := "12345;01.12.2024;10:30:00;1;001;100;1;ITEM001"
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = parseTransactionLine(line, "test")
    }
}
```

### Интеграционные тесты:

```go
//go:build integration

package integration

func TestDatabaseConnection(t *testing.T) {
    if os.Getenv("INTEGRATION_TEST") != "true" {
        t.Skip("Skipping integration test")
    }
    // тест с реальной БД
}
```

---

## 📊 Structured Logging (slog)

Используйте пакет `pkg/logger` для structured logging:

```go
// ✅ ПРАВИЛЬНО: создание логгера
import "github.com/user/go-frontol-loader/pkg/logger"

log := logger.New(logger.Config{
    Level:  "info",   // debug, info, warn, error
    Format: "json",   // json или text
})

// С контекстом
log := log.WithRequestID("req_123")
log := log.WithComponent("parser")
log := log.WithKassa("001", "folder1")

// Логирование
log.Info("Processing started", "files", 10)
log.Error("Failed to process", "error", err.Error())

// ETL хелперы
log.LogETLStart(ctx, "2024-12-01")
log.LogETLEnd(ctx, "2024-12-01", filesProcessed, transactionsLoaded, err)
log.LogFileProcessed(ctx, filePath, transactions, err)
log.LogDBOperation(ctx, "insert", "transactions", rowsAffected, err)
```

**Правила логирования:**
- Используйте structured logging вместо `fmt.Printf`
- Добавляйте контекст (request_id, component, kassa)
- Уровни: `debug` для отладки, `info` для операций, `error` для ошибок
- Никогда не логируйте пароли и токены

---

## 🔄 Graceful Shutdown

Используйте пакет `pkg/server` для HTTP сервера с graceful shutdown:

```go
import "github.com/user/go-frontol-loader/pkg/server"

// Создание сервера
srv := server.New(server.Config{
    Port:            8080,
    ReadTimeout:     15 * time.Second,
    WriteTimeout:    15 * time.Second,
    ShutdownTimeout: 30 * time.Second,
}, handler, log)

// Запуск с graceful shutdown
if err := srv.Run(ctx); err != nil {
    log.Error("Server error", "error", err)
}
```

### Middleware:

```go
// Request ID
handler = server.RequestIDMiddleware(handler)

// Logging
handler = server.LoggingMiddleware(log)(handler)

// Panic recovery
handler = server.RecoveryMiddleware(log)(handler)
```

---

## ✅ Валидация

Используйте пакет `pkg/validator` для валидации:

```go
import "github.com/user/go-frontol-loader/pkg/validator"

// Валидация конфигурации
v := validator.NewConfigValidator()
if errors := v.Validate(cfg); errors.HasErrors() {
    return fmt.Errorf("invalid config: %w", errors)
}

// Валидация даты
dateValidator := validator.DateValidator{}
if err := dateValidator.ValidateDate("2024-12-01"); err != nil {
    return err
}

// Валидация строки транзакции
txValidator := validator.TransactionValidator{}
if err := txValidator.ValidateTransactionLine(line); err != nil {
    return err
}
```

---

## 🗄️ Миграции базы данных (golang-migrate)

Проект использует [golang-migrate](https://github.com/golang-migrate/migrate) для управления миграциями.

### Команды миграций:

```bash
# Применить все миграции
make migrate-up

# Откатить все миграции
make migrate-down

# Применить N миграций (или откатить при отрицательном N)
make migrate-step N=1
make migrate-step N=-1

# Показать текущую версию
make migrate-version

# Принудительно установить версию (для исправления dirty state)
make migrate-force V=3

# Создать новую миграцию
make migrate-create NAME=add_users_table
```

### Структура миграций:

```
pkg/migrate/
├── migrate.go                          # Пакет для работы с миграциями
├── migrate_test.go                     # Тесты
└── migrations/                         # Embedded миграции
    ├── 000001_init_schema.up.sql       # Создание таблиц
    ├── 000001_init_schema.down.sql     # Откат
    ├── 000002_seed_data.up.sql         # Справочные данные
    ├── 000002_seed_data.down.sql
    ├── 000003_add_indexes.up.sql       # Индексы
    └── 000003_add_indexes.down.sql
```

### Правила написания миграций:

```sql
-- ✅ ПРАВИЛЬНО: использовать IF NOT EXISTS / IF EXISTS
CREATE TABLE IF NOT EXISTS users (...);
DROP TABLE IF EXISTS users;
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- ✅ ПРАВИЛЬНО: ON CONFLICT для seed данных
INSERT INTO users (id, name) VALUES (1, 'Alice')
ON CONFLICT (id) DO NOTHING;

-- ❌ ЗАПРЕЩЕНО: изменять существующие миграции после применения
-- Вместо этого создавайте новую миграцию
```

### Использование в коде:

```go
import "github.com/user/go-frontol-loader/pkg/migrate"

// Создание мигратора
migrator, err := migrate.NewMigrator(cfg)
if err != nil {
    return err
}
defer migrator.Close()

// Применение миграций
if err := migrator.Up(); err != nil {
    return err
}

// Проверка статуса
status := migrator.GetStatus()
fmt.Printf("Version: %d, Dirty: %v\n", status.Version, status.Dirty)
```

---

## 🔧 Инструменты разработки

### Обязательные команды перед коммитом:

```bash
# Форматирование кода
make fmt
# или
go fmt ./...

# Линтер
make lint
# или
golangci-lint run

# Тесты
make test-go
# или
go test ./...
```

### Конфигурация golangci-lint (рекомендуемая)

Создайте файл `.golangci.yml`:

```yaml
run:
  timeout: 5m

linters:
  enable:
    - errcheck      # проверка обработки ошибок
    - gosimple      # упрощение кода
    - govet         # проверка Go конструкций
    - ineffassign   # неиспользуемые присваивания
    - staticcheck   # статический анализ
    - unused        # неиспользуемый код
    - gofmt         # форматирование
    - goimports     # импорты
    - misspell      # орфография
    - unconvert     # лишние конверсии типов

linters-settings:
  errcheck:
    check-type-assertions: true
    check-blank: true
```

---

## 🐳 Docker

### Правила для Docker:

1. Использовать multi-stage builds для минимизации размера образа
2. Не копировать `.env` файлы в образ
3. Использовать переменные окружения для конфигурации
4. Указывать конкретные версии базовых образов

```dockerfile
# ✅ ПРАВИЛЬНО: multi-stage build
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/webhook-server

FROM alpine:3.19
COPY --from=builder /app/server /server
CMD ["/server"]
```

---

## 📋 Чеклист код-ревью

### Безопасность:
- [ ] Нет хардкода паролей/токенов
- [ ] Все SQL запросы параметризованы
- [ ] Входные данные валидируются
- [ ] Ресурсы (файлы, соединения) закрываются через `defer`
- [ ] HTTP таймауты установлены

### Производительность:
- [ ] Используется connection pooling для БД
- [ ] Батчевая обработка для массовых операций
- [ ] Context с таймаутами для долгих операций
- [ ] Нет утечек горутин

### Стиль:
- [ ] Код отформатирован (`go fmt`)
- [ ] Линтер проходит без ошибок
- [ ] Экспортируемые элементы задокументированы
- [ ] Ошибки обёрнуты с контекстом
- [ ] Импорты сгруппированы

---

## 📚 Полезные ссылки

- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
