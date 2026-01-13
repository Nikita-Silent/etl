# 🧪 Комплексный План Тестирования

**Проект:** Frontol 6 ETL Loader
**Версия:** 1.0
**Дата:** 2026-01-07
**Статус:** Draft

---

## 📋 Содержание

1. [Обзор стратегии тестирования](#обзор-стратегии-тестирования)
2. [Unit Tests (Модульные тесты)](#unit-tests-модульные-тесты)
3. [Integration Tests (Интеграционные тесты)](#integration-tests-интеграционные-тесты)
4. [E2E Tests (Сквозные тесты)](#e2e-tests-сквозные-тесты)
5. [Метрики качества](#метрики-качества)
6. [План реализации](#план-реализации)

---

## Обзор стратегии тестирования

### Пирамида тестирования

```
       E2E (5-10 тестов)
      ▲
     ███
    █████  Integration (20-30 тестов)
   ███████
  █████████ Unit (100+ тестов)
 ███████████
```

### Приоритеты

**P0 (Критически важно):**
- Безопасность (auth, SQL injection)
- Целостность данных (upserts, constraints)
- ETL pipeline (основная бизнес-логика)

**P1 (Важно):**
- HTTP API endpoints
- Parser для всех 44 типов транзакций
- Error handling и retry logic

**P2 (Желательно):**
- Performance benchmarks
- Edge cases
- CLI утилиты

---

## Unit Tests (Модульные тесты)

### 1. pkg/server (КРИТИЧЕСКИЙ ПРОБЕЛ)

**Файл:** `pkg/server/server_test.go`

#### 1.1 Server Lifecycle
```go
TestNew()                           // Создание сервера с конфигурацией
TestNew_DefaultLogger()             // Создание с nil logger (должен использовать Default)
TestNew_CustomConfig()              // Кастомные timeout'ы
TestStart()                         // Успешный запуск сервера
TestStart_InvalidPort()             // Запуск на занятом порту
TestGracefulShutdown()              // Graceful shutdown без активных запросов
TestGracefulShutdown_WithActiveRequests()  // Shutdown с ожиданием запросов
TestGracefulShutdown_Timeout()      // Таймаут при shutdown
TestSignalHandling()                // Обработка SIGINT/SIGTERM
```

#### 1.2 Middleware
```go
TestRequestIDMiddleware()           // Добавление request_id в контекст
TestRequestIDMiddleware_Existing()  // Сохранение существующего request_id
TestLoggingMiddleware()             // Логирование HTTP запросов
TestRecoveryMiddleware()            // Recovery от panic
TestRecoveryMiddleware_WithPanic()  // Логирование и recovery от panic
TestCORSMiddleware()                // CORS headers (если есть)
```

#### 1.3 Handlers
```go
TestHealthCheckHandler()            // GET /api/health
TestHealthCheckHandler_DBDown()     // Health check когда БД недоступна
```

**Приоритет:** P0
**Оценка:** ~15-20 тестов
**Зависимости:** httptest, mock logger

---

### 2. pkg/parser (КРИТИЧЕСКИЙ ПРОБЕЛ)

#### 2.1 Dispatcher (dispatcher_test.go)

**Текущее состояние:** Нет тестов для 434 строк кода

```go
TestGetTransactionType_AllTypes()   // Все 44 типа транзакций
TestGetTransactionType_Invalid()    // Неизвестный тип
TestParseTransaction_Type1_11()     // Регистрация товара
TestParseTransaction_Type2_12()     // Сторно товара
TestParseTransaction_Type4_14()     // Налог на товар
TestParseTransaction_Type6_16()     // ККТ регистрация
TestParseTransaction_Type9()        // Начисление бонуса
TestParseTransaction_Type10()       // Списание бонуса
TestParseTransaction_Type15()       // Скидка
TestParseTransaction_Type17()       // Надбавка
TestParseTransaction_Type18()       // Оплата счета
TestParseTransaction_Type19()       // Возврат оплаты
TestParseTransaction_Type20_29()    // Операции сотрудников
// ... тесты для всех 44 типов
```

**Стратегия тестирования:**
- Создать fixture файлы с примерами каждого типа
- Table-driven tests для массовой проверки
- Негативные сценарии (неверное количество полей, неверный формат)

#### 2.2 Generic Parsing (tx_parsing_test.go)

```go
TestParseTxModel_ValidData()        // Парсинг валидных данных
TestParseTxModel_MissingFields()    // Недостаточно полей
TestParseTxModel_ExtraFields()      // Лишние поля (должны игнорироваться)
TestParseTxModel_DateParsing()      // Парсинг дат (DD.MM.YYYY)
TestParseTxModel_TimeParsing()      // Парсинг времени (HH:MM:SS)
TestParseTxModel_FloatWithComma()   // Float с запятой (1234,56 → 1234.56)
TestParseTxModel_FloatWithDot()     // Float с точкой
TestParseTxModel_IntegerFields()    // Парсинг integer полей
TestParseTxModel_EmptyStrings()     // Пустые строки → NULL
TestParseTxModel_Reflection()       // Reflection-based field mapping
TestParseTxModel_UnknownFieldTag()  // Неизвестный тег в структуре
```

#### 2.3 Field Mapping (mappers_test.go)

```go
TestColumnToFieldName()             // column_name → ColumnName
TestGetFieldByTag()                 // Поиск поля по тегу `field:"N"`
TestConvertWindowsEncoding()        // Windows-1251 → UTF-8
```

**Приоритет:** P0 (критическая бизнес-логика)
**Оценка:** ~50-60 тестов
**Зависимости:** Fixture файлы с примерами транзакций

---

### 3. pkg/repository (КРИТИЧЕСКИЙ ПРОБЕЛ)

#### 3.1 Loader Operations (loader_test.go - расширение)

**Текущие тесты:** Только utility функции (4 теста)
**Нужны тесты для:**

```go
// Batch operations
TestPrepareBatchInsert()            // Генерация SQL для batch insert
TestPrepareBatchUpsert()            // Генерация SQL с ON CONFLICT
TestBuildPlaceholders()             // Построение placeholders ($1, $2, ...)
TestBuildConflictClause()           // ON CONFLICT (pk) DO UPDATE SET ...

// Transaction loaders (с mock DB)
TestLoadTxItemRegistration()        // Загрузка tx_item_registration_1_11
TestLoadTxItemStorno()              // Загрузка tx_item_storno_2_12
TestLoadTxBonusAccrual()            // Загрузка tx_bonus_accrual_9
TestLoadTxPayment()                 // Загрузка tx_payment_18
// ... для всех 21 таблицы

// Error handling
TestLoadData_Retry()                // Retry на deadlock/serialization
TestLoadData_MaxRetries()           // Превышение max retries
TestLoadData_NonRetryableError()    // Не retry-able ошибка (например, constraint violation)
TestLoadData_EmptyData()            // Загрузка пустого slice
TestLoadData_LargeData()            // Загрузка большого batch (>1000 записей)

// Upsert behavior (требует real DB в integration tests)
// Но можем проверить SQL generation в unit тестах
TestGenerateUpsertSQL()             // Правильный SQL для upsert
TestGenerateConflictColumns()       // Composite PK: (transaction_id_unique, source_folder)
```

**Приоритет:** P0
**Оценка:** ~25-30 тестов
**Зависимости:** Mock DB (уже есть в pkg/db/mocks.go)

---

### 4. pkg/pipeline (КРИТИЧЕСКИЙ ПРОБЕЛ)

#### 4.1 Pipeline Orchestration (pipeline_test.go - расширение)

**Текущие тесты:** Только структуры данных
**Нужны тесты для:**

```go
// Pipeline execution with mocks
TestRun_Success()                   // Успешное выполнение всех 4 шагов
TestRun_ClearRequestsFails()        // Ошибка на шаге 1 (clear)
TestRun_SendRequestsFails()         // Ошибка на шаге 2 (send)
TestRun_WaitStep()                  // Проверка wait delay
TestRun_ProcessFilesFails()         // Ошибка на шаге 4 (process)
TestRun_PartialSuccess()            // Часть файлов обработана, часть - ошибки
TestRun_NoFiles()                   // Нет файлов для обработки
TestRun_ContextCancellation()       // Отмена через context

// File processing
TestProcessFile_Success()           // Успешная обработка файла
TestProcessFile_AlreadyProcessed()  // Файл уже помечен .processed
TestProcessFile_ParseError()        // Ошибка парсинга
TestProcessFile_DBError()           // Ошибка загрузки в БД
TestProcessFile_MarkAsProcessed()   // Маркировка файла после успеха

// Parallel processing
TestParallelProcessing()            // Параллельная обработка нескольких файлов
TestParallelProcessing_Mutex()      // Mutex защита FTP клиента
TestParallelProcessing_WaitGroup()  // WaitGroup координация

// Statistics
TestAggregateStats()                // Агрегация статистики
TestCalculateDuration()             // Расчет времени выполнения
TestTransactionDetails()            // Детализация по типам транзакций
```

**Приоритет:** P0
**Оценка:** ~18-20 тестов
**Зависимости:** Mock FTP, Mock DB, Mock Parser

---

### 5. cmd/ binaries (НОВОЕ)

#### 5.1 webhook-server (cmd/webhook-server/main_test.go)

```go
TestMain_ConfigLoading()            // Загрузка конфигурации из env
TestMain_InvalidConfig()            // Обработка невалидной конфигурации
TestMain_DBConnection()             // Проверка подключения к БД
TestMain_FTPConnection()            // Проверка подключения к FTP (опционально)
TestMain_ServerStart()              // Запуск HTTP сервера
TestMain_GracefulShutdown()         // Graceful shutdown при SIGTERM
```

#### 5.2 loader (cmd/loader/main_test.go)

```go
TestMain_CLIArgs()                  // Парсинг аргументов командной строки
TestMain_DateValidation()           // Валидация даты
TestMain_PipelineExecution()        // Запуск pipeline.Run()
TestMain_ErrorHandling()            // Обработка ошибок pipeline
TestMain_ExitCodes()                // Правильные exit codes (0/1)
```

#### 5.3 migrate (cmd/migrate/main_test.go)

```go
TestMigrate_Up()                    // Применение миграций
TestMigrate_Down()                  // Откат миграций
TestMigrate_Version()               // Проверка версии
TestMigrate_Force()                 // Принудительная установка версии
TestMigrate_InvalidCommand()        // Неизвестная команда
```

**Приоритет:** P1
**Оценка:** ~15-20 тестов
**Зависимости:** Mocks для всех зависимостей

---

### 6. Дополнительные Unit тесты

#### 6.1 pkg/auth (расширение auth_test.go)

```go
TestBearerAuthMiddleware_Disabled() // Auth disabled (empty token)
TestBearerAuthMiddleware_Valid()    // Валидный токен
TestBearerAuthMiddleware_Invalid()  // Невалидный токен
TestBearerAuthMiddleware_Missing()  // Отсутствует header
TestBearerAuthMiddleware_WrongScheme() // "Basic" вместо "Bearer"

// ✅ УЖЕ ЕСТЬ (но можно расширить)
```

#### 6.2 pkg/workers (расширение pool_test.go)

```go
TestWorkerPool_Cancel()             // Отмена через context
TestWorkerPool_Panic()              // Обработка panic в worker
TestWorkerPool_FullQueue()          // Переполнение очереди

// ✅ ЧАСТИЧНО ЕСТЬ
```

---

### Итого Unit Tests

**Текущее состояние:** ~21 файл, ~100 тестов
**После реализации плана:** ~35 файлов, ~250-300 тестов
**Целевое покрытие:** 75-80%

---

## Integration Tests (Интеграционные тесты)

### Текущее состояние

**Есть:**
- ✅ `tests/integration/framework/` - инфраструктура
- ✅ `tests/integration/db_test.go` - простые DB тесты
- ✅ `tests/integration/loader_test.go` - базовые loader тесты
- ✅ `tests/integration/framework_test.go` - тесты самого фреймворка
- ✅ `tests/integration/schema_test.go` - проверка constraints

**Нужно добавить:**

---

### 1. Database Integration (tests/integration/db_integration_test.go)

#### 1.1 Upsert Behavior

```go
TestUpsert_InsertNew()              // Первая вставка записи
TestUpsert_UpdateExisting()         // Update существующей записи
TestUpsert_CompositePK()            // PK: (transaction_id_unique, source_folder)
TestUpsert_MultipleSourceFolders()  // Разные source_folder, одинаковый ID
TestUpsert_Idempotent()             // Повторная загрузка того же файла
```

#### 1.2 Constraints & Validation

```go
TestConstraint_PrimaryKey()         // Нарушение PK (должно обновить)
TestConstraint_NotNull()            // NOT NULL поля
TestConstraint_ForeignKey()         // FK constraints
TestConstraint_CheckDate()          // CHECK constraints для дат
```

#### 1.3 Encoding

```go
TestEncoding_Windows1251ToUTF8()    // Автоматическая конвертация
TestEncoding_SpecialCharacters()    // Спецсимволы (кириллица, №, и т.д.)
TestEncoding_EmptyStrings()         // Пустые строки
```

#### 1.4 Connection Pooling

```go
TestConnectionPool_Concurrent()     // Параллельные запросы
TestConnectionPool_Exhaustion()     // Исчерпание пула
TestConnectionPool_Recovery()       // Восстановление после потери соединения
```

**Приоритет:** P0
**Оценка:** ~12-15 тестов

---

### 2. FTP Integration (tests/integration/ftp_integration_test.go)

#### 2.1 Real FTP Operations

```go
TestFTP_ListFiles()                 // Листинг файлов
TestFTP_DownloadFile()              // Скачивание файла
TestFTP_UploadRequest()             // Загрузка request.txt
TestFTP_MarkAsProcessed()           // Переименование в .processed
TestFTP_DeleteFile()                // Удаление файла
TestFTP_CreateDirectory()           // Создание директории
```

#### 2.2 FTP Pool

```go
TestFTPPool_Concurrent()            // Параллельные операции
TestFTPPool_Reconnect()             // Переподключение при потере связи
TestFTPPool_Timeout()               // Таймаут операций
```

#### 2.3 Error Scenarios

```go
TestFTP_FileNotFound()              // Файл не найден
TestFTP_PermissionDenied()          // Нет прав доступа
TestFTP_NetworkFailure()            // Потеря сети
TestFTP_Retry()                     // Retry логика
```

**Приоритет:** P1
**Оценка:** ~10-12 тестов

---

### 3. Parser Integration (tests/integration/parser_integration_test.go)

#### 3.1 Real File Parsing

**Стратегия:** Создать fixture файлы для каждого типа транзакций

```go
TestParseRealFile_Type1()           // Реальный файл с type 1
TestParseRealFile_Type2()           // Реальный файл с type 2
// ... для всех типов
TestParseRealFile_Mixed()           // Файл со смешанными типами
TestParseRealFile_LargeFile()       // Большой файл (10000+ строк)
TestParseRealFile_Malformed()       // Файл с ошибками
```

#### 3.2 End-to-End Parsing + Loading

```go
TestParseAndLoad_SingleFile()       // Парсинг + загрузка в БД
TestParseAndLoad_MultipleFiles()    // Несколько файлов
TestParseAndLoad_VerifyData()       // Проверка данных в БД после загрузки
```

**Приоритет:** P0
**Оценка:** ~15-20 тестов
**Зависимости:** Fixture файлы в `tests/fixtures/`

---

### 4. Pipeline Integration (tests/integration/pipeline_integration_test.go)

#### 4.1 Full Pipeline with Real Services

```go
TestPipeline_FullRun()              // Полный запуск с реальными сервисами
TestPipeline_WithRealFiles()        // С реальными файлами на FTP
TestPipeline_MultipleKassas()       // Обработка нескольких касс
TestPipeline_LargeDataset()         // Большой объем данных
TestPipeline_ParallelFiles()        // Параллельная обработка
```

#### 4.2 Error Recovery

```go
TestPipeline_DBReconnect()          // Восстановление после потери DB
TestPipeline_FTPReconnect()         // Восстановление после потери FTP
TestPipeline_PartialFailure()       // Часть файлов обработана успешно
```

**Приоритет:** P0
**Оценка:** ~8-10 тестов

---

### 5. API Integration (tests/integration/api_integration_test.go)

#### 5.1 HTTP Endpoints

```go
TestAPI_Health()                    // GET /api/health
TestAPI_LoadWebhook()               // POST /api/load
TestAPI_LoadWebhook_Auth()          // С Bearer token
TestAPI_LoadWebhook_NoAuth()        // Без токена (должен отклонить)
TestAPI_QueueStatus()               // GET /api/queue/status
TestAPI_Files()                     // GET /api/files
TestAPI_Docs()                      // GET /api/docs
```

#### 5.2 Request Queue

```go
TestQueue_Sequential()              // Последовательная обработка
TestQueue_Capacity()                // Переполнение очереди
TestQueue_Concurrent()              // Параллельные запросы
```

**Приоритет:** P1
**Оценка:** ~8-10 тестов

---

### Итого Integration Tests

**Текущее состояние:** 4 файла, ~10 тестов
**После реализации плана:** ~10 файлов, ~50-70 тестов
**Требования:** PostgreSQL, FTP server (через docker-compose.test.yml)

---

## E2E Tests (Сквозные тесты)

### Стратегия

E2E тесты должны симулировать реальное использование системы:
1. Запуск всех сервисов через Docker Compose
2. Загрузка тестовых файлов на FTP
3. Вызов API / CLI
4. Проверка данных в БД

**Инструменты:**
- Docker Compose для сервисов
- testcontainers-go (опционально)
- HTTP client для API тестов
- SQL queries для проверки БД

---

### 1. Webhook Flow (tests/e2e/webhook_test.go)

```go
TestE2E_WebhookTrigger()
  ├── Запустить Docker Compose
  ├── Загрузить тестовые файлы на FTP
  ├── POST /api/load {"date": "2024-12-01"}
  ├── Дождаться завершения (polling /api/queue/status)
  ├── Проверить данные в БД
  └── Teardown

TestE2E_WebhookAuth()
  ├── POST без токена → 401
  ├── POST с неверным токеном → 401
  └── POST с верным токеном → 202

TestE2E_WebhookMultipleRequests()
  ├── Отправить 5 запросов подряд
  ├── Проверить, что обрабатываются последовательно
  └── Проверить, что все завершились успешно
```

**Приоритет:** P0
**Оценка:** 3-5 тестов

---

### 2. CLI Flow (tests/e2e/cli_test.go)

```go
TestE2E_LoaderCLI()
  ├── Запустить Docker Compose
  ├── Загрузить тестовые файлы на FTP
  ├── ./frontol-loader 2024-12-01
  ├── Проверить exit code 0
  ├── Проверить данные в БД
  └── Teardown

TestE2E_MigrateCLI()
  ├── ./migrate up
  ├── Проверить версию 3
  ├── ./migrate down -n 1
  ├── Проверить версию 2
  └── ./migrate up
```

**Приоритет:** P1
**Оценка:** 2-3 теста

---

### 3. Full ETL Cycle (tests/e2e/etl_full_test.go)

```go
TestE2E_FullETLCycle()
  ├── Setup: запуск сервисов, миграции, очистка БД
  │
  ├── Шаг 1: Загрузка fixture файлов на FTP
  │   ├── P13/P13/export_001.txt (type 1, 9, 15)
  │   ├── N22/N22_Inter/export_002.txt (type 2, 10, 18)
  │   └── N22/N22_FURN/export_003.txt (mixed types)
  │
  ├── Шаг 2: Триггер ETL через API
  │   └── POST /api/load {"date": "2024-12-01"}
  │
  ├── Шаг 3: Ожидание завершения (с таймаутом 60s)
  │   └── Polling /api/queue/status каждые 2s
  │
  ├── Шаг 4: Проверка результатов в БД
  │   ├── Проверить tx_item_registration_1_11
  │   ├── Проверить tx_bonus_accrual_9
  │   ├── Проверить tx_discount_15
  │   ├── Проверить composite PK (multiple source_folder)
  │   └── Проверить, что файлы помечены .processed
  │
  ├── Шаг 5: Повторный запуск (идемпотентность)
  │   ├── POST /api/load {"date": "2024-12-01"}
  │   └── Проверить, что данные не задублировались
  │
  └── Teardown: остановка сервисов
```

**Приоритет:** P0
**Оценка:** 1 большой тест

---

### 4. Error Scenarios (tests/e2e/errors_test.go)

```go
TestE2E_DatabaseDown()
  ├── Остановить PostgreSQL
  ├── POST /api/load
  ├── Проверить graceful error handling
  └── Запустить PostgreSQL обратно

TestE2E_FTPDown()
  ├── Остановить FTP server
  ├── POST /api/load
  ├── Проверить retry и timeout
  └── Запустить FTP обратно

TestE2E_MalformedFile()
  ├── Загрузить файл с ошибками на FTP
  ├── POST /api/load
  ├── Проверить, что другие файлы обработаны
  └── Проверить error в логах
```

**Приоритет:** P1
**Оценка:** 3-4 теста

---

### Итого E2E Tests

**Текущее состояние:** 0 тестов
**После реализации плана:** ~8-12 тестов
**Время выполнения:** ~2-5 минут
**Требования:** Docker Compose с PostgreSQL + FTP

---

## Метрики качества

### Целевые показатели

| Метрика | Текущее | Цель |
|---------|---------|------|
| **Code Coverage** | ~50% | **75-80%** |
| **Unit Tests** | ~100 | **250-300** |
| **Integration Tests** | ~10 | **50-70** |
| **E2E Tests** | 0 | **8-12** |
| **Critical Path Coverage** | 60% | **95%+** |
| **CI Test Time** | ~3 min | **< 5 min** |

### Critical Path Coverage (приоритет тестирования)

**Critical Path = код, который ДОЛЖЕН работать для работы системы**

1. ✅ **Auth** (уже 100%)
2. ❌ **Parser dispatcher** (0% → 90%)
3. ❌ **Repository loader** (20% → 90%)
4. ❌ **Pipeline execution** (10% → 85%)
5. ❌ **HTTP API** (0% → 80%)
6. ✅ **Config** (уже ~90%)
7. ✅ **FTP operations** (уже ~85%)

---

## План реализации

### Фаза 1: Критические Unit тесты (1-2 недели)

**Цель:** Покрыть критическую бизнес-логику

- [ ] **pkg/server** - все тесты (P0)
- [ ] **pkg/parser/dispatcher** - все 44 типа (P0)
- [ ] **pkg/parser/tx_parsing** - reflection parsing (P0)
- [ ] **pkg/repository/loader** - batch operations, upserts (P0)
- [ ] **pkg/pipeline** - orchestration (P0)

**Deliverables:**
- +130 unit тестов
- Coverage: 50% → 70%

---

### Фаза 2: Integration тесты (1 неделя)

**Цель:** Проверить взаимодействие с реальными сервисами

- [ ] **Database integration** - upserts, constraints (P0)
- [ ] **FTP integration** - real operations (P1)
- [ ] **Parser integration** - real files (P0)
- [ ] **Pipeline integration** - full run (P0)
- [ ] **API integration** - endpoints (P1)

**Deliverables:**
- +50 integration тестов
- Fixture файлы для всех типов транзакций

---

### Фаза 3: E2E тесты (3-5 дней)

**Цель:** Проверить полный flow

- [ ] **Webhook flow** - API → ETL → DB (P0)
- [ ] **CLI flow** - loader, migrate (P1)
- [ ] **Full ETL cycle** - с идемпотентностью (P0)
- [ ] **Error scenarios** - resilience (P1)

**Deliverables:**
- +10 E2E тестов
- Docker Compose тестовое окружение

---

### Фаза 4: CLI binaries (опционально, 2-3 дня)

**Цель:** Покрыть cmd/ тестами

- [ ] **webhook-server** - main (P1)
- [ ] **loader** - main (P1)
- [ ] **migrate** - main (P1)

**Deliverables:**
- +15 тестов для CLI

---

### Фаза 5: Финальная полировка (1-2 дня)

- [ ] Документация тестов
- [ ] CI/CD оптимизация
- [ ] Code coverage отчеты
- [ ] Benchmark тесты для критических путей

**Deliverables:**
- Coverage report
- Performance baseline

---

## Структура директорий

```
tests/
├── fixtures/                       # Тестовые данные
│   ├── transactions/
│   │   ├── type_01_sample.txt     # Пример type 1
│   │   ├── type_02_sample.txt     # Пример type 2
│   │   ├── ...
│   │   └── mixed_types.txt        # Смешанные типы
│   └── ftp/
│       ├── P13/
│       │   └── P13/
│       │       └── export_001.txt
│       └── N22/
│           ├── N22_Inter/
│           └── N22_FURN/
│
├── integration/
│   ├── framework/                 # ✅ Уже есть
│   ├── db_integration_test.go     # ⬜ Новый
│   ├── ftp_integration_test.go    # ⬜ Новый
│   ├── parser_integration_test.go # ⬜ Новый
│   ├── pipeline_integration_test.go # ⬜ Новый
│   └── api_integration_test.go    # ⬜ Новый
│
├── e2e/
│   ├── webhook_test.go            # ⬜ Новый
│   ├── cli_test.go                # ⬜ Новый
│   ├── etl_full_test.go           # ⬜ Новый
│   ├── errors_test.go             # ⬜ Новый
│   └── helpers/
│       ├── docker.go              # Docker Compose helpers
│       ├── api.go                 # API client helpers
│       └── db.go                  # DB assertion helpers
│
pkg/
├── server/
│   └── server_test.go             # ⬜ Новый
├── parser/
│   ├── dispatcher_test.go         # ⬜ Новый
│   ├── tx_parsing_test.go         # ⬜ Новый
│   └── mappers_test.go            # ⬜ Новый
├── repository/
│   └── loader_test.go             # ⬜ Расширить
└── pipeline/
    └── pipeline_test.go           # ⬜ Расширить
```

---

## Рекомендации по написанию тестов

### 1. Именование

```go
// ✅ Хорошо
func TestLoadTxItemRegistration_ValidData(t *testing.T)
func TestLoadTxItemRegistration_EmptySlice(t *testing.T)
func TestLoadTxItemRegistration_DBError(t *testing.T)

// ❌ Плохо
func TestLoader1(t *testing.T)
func TestValid(t *testing.T)
```

### 2. Table-Driven Tests

```go
func TestParseTransactionType(t *testing.T) {
    tests := []struct {
        name        string
        typeCode    int
        wantErr     bool
        wantDesc    string
    }{
        {"Type 1 - Item Registration", 1, false, "Регистрация товара"},
        {"Type 2 - Item Storno", 2, false, "Сторно товара"},
        {"Invalid Type", 999, true, ""},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test logic
        })
    }
}
```

### 3. Subtests

```go
func TestServer(t *testing.T) {
    t.Run("Start", func(t *testing.T) { /* ... */ })
    t.Run("Shutdown", func(t *testing.T) { /* ... */ })
    t.Run("Graceful", func(t *testing.T) { /* ... */ })
}
```

### 4. Fixtures

```go
// tests/fixtures/loader.go
func LoadFixture(t *testing.T, name string) []byte {
    t.Helper()
    data, err := os.ReadFile(filepath.Join("fixtures", name))
    if err != nil {
        t.Fatalf("failed to load fixture %s: %v", name, err)
    }
    return data
}
```

### 5. Cleanup

```go
func TestWithDatabase(t *testing.T) {
    db := setupTestDB(t)
    t.Cleanup(func() {
        db.Close()
    })

    // test logic
}
```

---

## CI/CD Integration

### GitHub Actions Workflow

```yaml
# .github/workflows/test.yml
name: Tests

on: [push, pull_request]

jobs:
  unit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24'

      - name: Unit Tests
        run: make test-go

      - name: Coverage
        run: |
          make test-coverage
          bash <(curl -s https://codecov.io/bash)

  integration:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s

      ftp:
        image: stilliard/pure-ftpd
        env:
          PUBLICHOST: localhost

    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4

      - name: Integration Tests
        run: make test-integration
        env:
          INTEGRATION_TEST: true

  e2e:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: E2E Tests
        run: |
          docker-compose -f docker-compose.test.yml up -d
          sleep 10
          make test-e2e
          docker-compose -f docker-compose.test.yml down
```

---

## Заключение

### Текущее состояние: 🟡 Частичное покрытие

- ✅ Хорошая основа (21 test файл)
- ✅ Интеграционная инфраструктура
- ❌ Критические пробелы в server, parser, repository, pipeline
- ❌ Нет E2E тестов

### После реализации плана: 🟢 Комплексное покрытие

- ✅ 250+ unit тестов
- ✅ 50+ integration тестов
- ✅ 10+ E2E тестов
- ✅ 75-80% code coverage
- ✅ 95%+ critical path coverage

### Приоритет реализации

**Начать с:**
1. **pkg/parser/dispatcher** - критическая бизнес-логика
2. **pkg/repository/loader** - целостность данных
3. **pkg/pipeline** - основной flow
4. **E2E full cycle** - smoke test для всей системы

**Время реализации:** 3-4 недели при работе одного разработчика

---

**Последнее обновление:** 2026-01-07
**Автор:** Claude Code
**Статус:** ✅ Ready for Review
