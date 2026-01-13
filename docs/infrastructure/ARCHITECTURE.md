# 🏗️ Архитектура проекта Frontol ETL

## 📋 Содержание

1. [Обзор](#обзор)
2. [Компоненты системы](#компоненты-системы)
3. [Структура проекта](#структура-проекта)
4. [Архитектурные паттерны](#архитектурные-паттерны)
5. [Потоки данных](#потоки-данных)
6. [Диаграммы](#диаграммы)

---

## 🎯 Обзор

**Frontol ETL** - это высокопроизводительная система для автоматизированной загрузки и обработки данных из кассовой системы Frontol 6 в PostgreSQL базу данных.

### Ключевые характеристики:

- **Язык:** Go 1.24+
- **Архитектура:** Модульная с разделением на пакеты
- **Развертывание:** Docker Compose
- **База данных:** PostgreSQL 12+
- **Протокол обмена:** FTP
- **API:** RESTful HTTP (webhook server)

---

## 🧩 Компоненты системы

### 1. Основные приложения

#### 1.1 Webhook Server (`cmd/webhook-server/`)
**Назначение:** HTTP сервер для асинхронного запуска ETL pipeline через API

**Функции:**
- Прием HTTP POST запросов для запуска ETL
- Health check endpoint
- Graceful shutdown
- Structured logging (slog)
- Интерактивная API документация (Scalar)

**Endpoints:**
- `POST /api/load` - Запуск ETL для указанной даты
- `GET /api/files` - Выгрузка данных из БД в файл
- `GET /api/queue/status` - Статус очереди обработки
- `GET /api/kassas` - Список доступных касс
- `GET /api/health` - Health check
- `GET /api/docs` - Интерактивная документация API
- `GET /api/openapi.yaml` - OpenAPI спецификация
Источник истины по схемам и параметрам: `api/openapi.yaml`

#### 1.2 Loader (`cmd/loader/`)
**Назначение:** CLI приложение для запуска ETL pipeline

**Функции:**
- Загрузка данных с FTP
- Парсинг файлов Frontol
- Загрузка в PostgreSQL
- Поддержка указания конкретной даты

#### 1.3 Migrate (`cmd/migrate/`)
**Назначение:** Управление миграциями базы данных

**Функции:**
- Применение миграций (`up`)
- Откат миграций (`down`)
- Проверка версии схемы
- Принудительная установка версии (`force`)

### 2. Вспомогательные утилиты

#### 2.1 Parser Test (`cmd/parser-test/`)
**Назначение:** Тестирование парсера на файлах

#### 2.2 Send Request (`cmd/send-request/`)
**Назначение:** Отправка request.txt файлов к кассам

#### 2.3 Clear Requests (`cmd/clear-requests/`)
**Назначение:** Очистка папок request/response на FTP

---

## 📁 Структура проекта

```
/
├── cmd/                           # Точки входа приложений
│   ├── loader/                    # Основной ETL загрузчик
│   ├── webhook-server/            # HTTP сервер с API
│   ├── migrate/                   # Миграции БД
│   ├── loader-local/              # Локальный загрузчик
│   ├── parser-test/               # Тестер парсера
│   ├── send-request/              # Отправка запросов
│   └── clear-requests/            # Очистка папок
│
├── pkg/                           # Переиспользуемые пакеты
│   ├── pipeline/                  # Оркестрация ETL pipeline
│   ├── config/                    # Управление конфигурацией
│   ├── db/                        # Работа с БД (pgx connection pool)
│   ├── ftp/                       # FTP клиент
│   ├── logger/                    # Structured logging (slog)
│   ├── migrate/                   # Database migrations
│   ├── models/                    # Структуры данных
│   ├── parser/                    # Парсинг файлов Frontol
│   ├── repository/                # Data access layer
│   ├── server/                    # HTTP сервер + middleware
│   └── validator/                 # Валидация данных
│
├── tests/
│   ├── integration/               # Интеграционные тесты
│   └── testdata/                  # Тестовые данные
│
├── docs/                          # Документация
│   ├── ARCHITECTURE.md            # Архитектура (этот файл)
│   ├── DATABASE.md                # База данных
│   ├── API.md                     # API и интерфейсы
│   ├── BUSINESS_LOGIC.md          # Бизнес-логика
│   ├── TECH_STACK.md              # Технический стек
│   ├── ROADMAP.md                 # Roadmap
│   ├── CODING_RULES.md            # Правила кода
│   └── TESTING.md                 # Тестирование
│
├── Dockerfile                     # Docker образ приложений
├── docker-compose.yml             # Оркестрация сервисов
├── Makefile                       # Команды автоматизации
└── .golangci.yml                  # Конфигурация линтера
```

### Разделение ответственности

#### cmd/ - Минимальная логика
- Только `main.go` с инициализацией
- Парсинг аргументов командной строки
- Запуск приложения

#### pkg/ - Вся бизнес-логика
- Каждый пакет имеет одну четкую ответственность
- Избегаем циклических зависимостей
- Переиспользуемые компоненты

---

## 🎨 Архитектурные паттерны

### 1. Layered Architecture (Слоеная архитектура)

```
┌─────────────────────────────────────────┐
│   Presentation Layer (cmd/)             │
│   - HTTP API (webhook-server)           │
│   - CLI (loader, migrate)               │
└─────────────────────────────────────────┘
              ▼
┌─────────────────────────────────────────┐
│   Business Logic Layer (pkg/)           │
│   - pipeline (ETL orchestration)        │
│   - parser (file parsing)               │
│   - validator (data validation)         │
└─────────────────────────────────────────┘
              ▼
┌─────────────────────────────────────────┐
│   Data Access Layer                     │
│   - repository (data loading)           │
│   - db (connection pool)                │
│   - ftp (file operations)               │
└─────────────────────────────────────────┘
              ▼
┌─────────────────────────────────────────┐
│   Infrastructure                        │
│   - PostgreSQL                          │
│   - FTP Server                          │
└─────────────────────────────────────────┘
```

### 2. Repository Pattern

Изоляция логики доступа к данным от бизнес-логики.

```go
// pkg/repository/
type Repository interface {
    LoadTransactionRegistrations(ctx context.Context, data []models.TransactionRegistration) error
    LoadBonusTransactions(ctx context.Context, data []models.BonusTransaction) error
    // ...
}
```

### 3. Dependency Injection

Компоненты получают зависимости через конструкторы.

```go
func NewPipeline(
    cfg *models.Config,
    ftpClient *ftp.Client,
    dbPool *db.Pool,
    logger *logger.Logger,
) *Pipeline
```

### 4. Factory Pattern

Создание экземпляров через фабричные функции.

```go
// pkg/logger/logger.go
func New(cfg Config) *Logger

// pkg/db/pool.go
func NewPool(cfg *models.Config) (*Pool, error)
```

### 5. Pipeline Pattern

ETL процесс представлен как последовательность шагов.

```go
func (p *Pipeline) Run(ctx context.Context, date string) error {
    // 1. Clear
    if err := p.clearFTPFolders(ctx); err != nil { return err }

    // 2. Request
    if err := p.sendRequests(ctx, date); err != nil { return err }

    // 3. Wait
    time.Sleep(cfg.WaitDelayMinutes)

    // 4. Download
    files, err := p.downloadFiles(ctx)
    if err != nil { return err }

    // 5. Parse
    transactions, err := p.parseFiles(ctx, files)
    if err != nil { return err }

    // 6. Load
    return p.loadToDatabase(ctx, transactions)
}
```

---

## 🌊 Потоки данных

### ETL Pipeline (основной поток)

```
┌────────────────┐
│ HTTP Request   │
│ POST /api/load │
└────────┬───────┘
         │
         ▼
┌────────────────────────────────┐
│ Webhook Server                 │
│ - Парсинг JSON                 │
│ - Валидация даты               │
│ - Запуск горутины              │
└────────┬───────────────────────┘
         │ 202 Accepted (немедленный ответ)
         │
         ▼
┌────────────────────────────────┐
│ ETL Pipeline (async)           │
├────────────────────────────────┤
│ 1. Clear FTP folders           │
│    - Удаление старых request   │
│    - Удаление старых response  │
└────────┬───────────────────────┘
         │
         ▼
┌────────────────────────────────┐
│ 2. Send requests               │
│    - Генерация request.txt     │
│    - Отправка на FTP           │
│    - Формат: DATE_FROM=..      │
└────────┬───────────────────────┘
         │
         ▼ (wait 1-2 min)
┌────────────────────────────────┐
│ 3. Download responses          │
│    - Чтение response папок     │
│    - Загрузка файлов           │
│    - Валидация формата         │
└────────┬───────────────────────┘
         │
         ▼
┌────────────────────────────────┐
│ 4. Parse files                 │
│    - Парсинг заголовка         │
│    - Парсинг транзакций        │
│    - Маппинг на Go структуры   │
└────────┬───────────────────────┘
         │
         ▼
┌────────────────────────────────┐
│ 5. Load to database            │
│    - Batch insert/upsert       │
│    - Transaction handling      │
│    - Идемпотентность           │
└────────┬───────────────────────┘
         │
         ▼
┌────────────────────────────────┐
│ 6. Send report (optional)      │
│    - POST к WEBHOOK_REPORT_URL │
│    - Статистика выполнения     │
└────────────────────────────────┘
```

### Поток миграций

```
┌────────────────┐
│ Init Container │ (при docker-compose up)
└────────┬───────┘
         │
         ▼
┌────────────────────────────────┐
│ Apply migrations               │
│ - 000001_init_schema.up.sql    │   │
└────────┬───────────────────────┘
         │
         ▼
┌────────────────────────────────┐
│ Database ready                 │
│ - Все таблицы созданы          │
│ - Справочники заполнены        │
│ - Индексы построены            │
└────────────────────────────────┘
```

---

## 📊 Диаграммы

### Диаграмма компонентов

```
┌──────────────────────────────────────────────────────────────┐
│                        Docker Compose                         │
├──────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌─────────────────┐         ┌──────────────────┐           │
│  │  Init Container │────────▶│  PostgreSQL DB   │           │
│  │   (Migrations)  │         │   (External)     │           │
│  └─────────────────┘         └──────────────────┘           │
│                                        ▲                      │
│                                        │                      │
│  ┌─────────────────┐                  │                      │
│  │  Webhook Server │──────────────────┘                      │
│  │  - HTTP API     │                                         │
│  │  - Graceful ⚡   │                                         │
│  └────────┬────────┘                                         │
│           │                                                   │
│           │ calls pkg/pipeline                               │
│           ▼                                                   │
│  ┌─────────────────┐         ┌──────────────────┐           │
│  │  ETL Pipeline   │────────▶│   FTP Server     │           │
│  │  - Clear        │◀────────│  (Test/Prod)     │           │
│  │  - Request      │         └──────────────────┘           │
│  │  - Download     │                                         │
│  │  - Parse        │                                         │
│  │  - Load         │                                         │
│  └─────────────────┘                                         │
│                                                               │
│  ┌─────────────────────────────────────────────┐            │
│  │  CLI Utilities (on-demand)                  │            │
│  │  - loader, parser-test, send-request, etc.  │            │
│  └─────────────────────────────────────────────┘            │
│                                                               │
└──────────────────────────────────────────────────────────────┘
```

### Диаграмма зависимостей пакетов

```
cmd/webhook-server
    │
    ├── pkg/config
    ├── pkg/logger
    ├── pkg/server ──┐
    ├── pkg/pipeline │
    └── pkg/validator
            │
            ▼
        pkg/models
            ▲
            │
    ┌───────┴──────────┐
    │                  │
pkg/db            pkg/ftp
    │                  │
    │                  │
    └───────┬──────────┘
            │
            ▼
      pkg/repository
            │
            ▼
      pkg/parser
```

### Диаграмма состояний ETL

```
┌──────────┐
│   IDLE   │ (Webhook waiting)
└────┬─────┘
     │ POST /api/load
     ▼
┌──────────┐
│STARTING  │ (Validation, init)
└────┬─────┘
     │
     ▼
┌──────────┐
│CLEARING  │ (Clear FTP folders)
└────┬─────┘
     │
     ▼
┌──────────┐
│REQUESTING│ (Send request.txt)
└────┬─────┘
     │
     ▼
┌──────────┐
│ WAITING  │ (Sleep N minutes)
└────┬─────┘
     │
     ▼
┌───────────┐
│DOWNLOADING│ (Fetch files from FTP)
└────┬──────┘
     │
     ▼
┌──────────┐
│ PARSING  │ (Parse transactions)
└────┬─────┘
     │
     ▼
┌──────────┐
│ LOADING  │ (Insert to database)
└────┬─────┘
     │
     ▼
┌──────────┐
│COMPLETED │ (Success/Error)
└──────────┘
     │
     └─────▶ IDLE
```

---

## 🔐 Безопасность

### 1. Конфиденциальные данные
- Все пароли и токены только через переменные окружения
- `.env` файл не коммитится в git
- Валидация обязательных параметров при старте

### 2. SQL Injection Protection
- Использование параметризованных запросов (pgx)
- Никакой конкатенации SQL строк

### 3. Input Validation
- Валидация всех входных данных (даты, файлы, JSON)
- Whitelist подход для разрешенных значений

### 4. Connection Security
- Connection pooling для оптимизации ресурсов
- Таймауты для всех сетевых операций
- Graceful shutdown для корректного завершения

---

## ⚡ Производительность

### 1. Database
- **Connection pooling** с pgx
- **Batch insert** с ON CONFLICT для идемпотентности
- **Индексы** на часто используемых полях

### 2. File Processing
- **Буферизованное чтение** файлов
- **Потоковая обработка** без загрузки всего файла в память

### 3. Concurrency
- **Горутины** для асинхронной обработки
- **Context** для управления жизненным циклом
- **Worker pool** для ограничения ресурсов

---

## 🔄 Масштабируемость

### Горизонтальное масштабирование
- Запуск нескольких экземпляров loader для разных касс
- Идемпотентность загрузок (уникальные ключи)
- Без состояния между запусками

### Вертикальное масштабирование
- Настройка размера connection pool
- Регулировка batch size для загрузки
- Оптимизация буферов и таймаутов

---

**Последнее обновление:** 2026-01-03
