# 📦 Бизнес-логика и ETL Pipeline

## 📋 Содержание

1. [Обзор ETL процесса](#обзор-etl-процесса)
2. [Подробное описание этапов](#подробное-описание-этапов)
3. [Парсинг файлов Frontol](#парсинг-файлов-frontol)
4. [Загрузка в базу данных](#загрузка-в-базу-данных)
5. [Обработка ошибок](#обработка-ошибок)
6. [Идемпотентность](#идемпотентность)

---

## 🎯 Обзор ETL процесса

### Что такое ETL?

**ETL** (Extract, Transform, Load) - процесс извлечения, преобразования и загрузки данных.

### Цель

Автоматизированная загрузка данных из кассовой системы Frontol 6 в PostgreSQL для последующего анализа и построения дашбордов.

### Архитектура ETL

```
┌──────────────────────────────────────────────────────┐
│              Frontol 6 (Кассовая система)            │
│  - Регистрация продаж                                │
│  - Скидки, бонусы                                    │
│  - Отчеты ККТ                                        │
└────────────────────┬─────────────────────────────────┘
                     │
                     │ FTP Export (по запросу)
                     ▼
┌──────────────────────────────────────────────────────┐
│               FTP Server                             │
│  /request  - Запросы к кассам                        │
│  /response - Ответы от касс (файлы транзакций)       │
└────────────────────┬─────────────────────────────────┘
                     │
                     │ ETL Pipeline
                     ▼
┌──────────────────────────────────────────────────────┐
│          PostgreSQL Database                         │
└────────────────────┬─────────────────────────────────┘
                     │
                     │ SQL Queries
                     ▼
┌──────────────────────────────────────────────────────┐
│          Analytics & Dashboards                      │
│  - Grafana, Power BI, Tableau, etc.                  │
└──────────────────────────────────────────────────────┘
```

---

## 🔄 Подробное описание этапов

### Полный цикл ETL

```
┌─────────────────────────────────────────────────────┐
│ Этап 1: Clear (Очистка)                             │
│ - Удаление старых request файлов                    │
│ - Удаление старых response файлов                   │
│ Цель: Подготовка к новому запросу                   │
└────────────────────┬────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────┐
│ Этап 2: Request (Запрос данных)                     │
│ - Генерация request.txt с датой                     │
│ - Отправка на FTP для каждой кассы                  │
│ Формат: DATE_FROM=DD.MM.YYYY;DATE_TO=DD.MM.YYYY     │
└────────────────────┬────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────┐
│ Этап 3: Wait (Ожидание)                             │
│ - Ожидание обработки запроса кассами                │
│ - Настраиваемая задержка (по умолчанию 1 минута)    │
│ Цель: Дать время Frontol сформировать файлы         │
└────────────────────┬────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────┐
│ Этап 4: Download (Загрузка)                         │
│ - Чтение списка файлов из response папок            │
│ - Загрузка файлов локально                          │
│ - Валидация формата файлов                          │
└────────────────────┬────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────┐
│ Этап 5: Parse (Парсинг)                             │
│ - Парсинг заголовка (3 строки)                      │
│ - Парсинг строк транзакций                          │
│ - Маппинг на Go структуры (models)                  │
│ - Группировка по типам транзакций                   │
└────────────────────┬────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────┐
│ Этап 6: Load (Загрузка в БД)                        │
│ - Batch insert с ON CONFLICT DO UPDATE              │
│ - Транзакционная обработка                          │
│ - Валидация данных перед вставкой                   │
│ - Логирование статистики                            │
└────────────────────┬────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────┐
│ Этап 7: Report (Отчет)                              │
│ - Формирование итогового отчета                     │
│ - Отправка на WEBHOOK_REPORT_URL (если настроен)    │
│ Цель: Мониторинг и уведомления                      │
└─────────────────────────────────────────────────────┘
```

---

## 📄 Парсинг файлов Frontol

### Формат файла

Каждый файл Frontol имеет следующую структуру:

```
#                           <-- Строка 1: Флаг обработки (# = не обработан)
12345                       <-- Строка 2: Database ID (DBID)
67890                       <-- Строка 3: Report Number
1;01.12.2024;10:30:00;1;... <-- Строка 4+: Транзакции (semicolon-separated)
2;01.12.2024;10:31:00;2;...
...
```

### Парсинг заголовка

```go
func ParseHeader(file *os.File) (*Header, error) {
    scanner := bufio.NewScanner(file)

    // Строка 1: Флаг обработки
    scanner.Scan()
    processingFlag := scanner.Text()

    // Строка 2: DBID
    scanner.Scan()
    dbid := scanner.Text()

    // Строка 3: Report Number
    scanner.Scan()
    reportNum := scanner.Text()

    return &Header{
        ProcessingFlag: processingFlag,
        DBID:          dbid,
        ReportNum:     reportNum,
    }, nil
}
```

### Парсинг транзакций

Каждая строка транзакции:
- Поля разделены точкой с запятой (`;`)
- 4-е поле - тип транзакции (определяет таблицу назначения)

**Пример:**
```
1;01.12.2024;10:30:00;1;001;12345;100;1;ITEM001;DRINK;500.00;5;500.00;0;1;1;10.00;50.00;...
│ │           │        │
│ │           │        └─ Поле №4: Тип транзакции (1 = Регистрация товара)
│ │           └────────── Поле №3: Время транзакции
│ └────────────────────── Поле №2: Дата транзакции
└──────────────────────── Поле №1: Уникальный ID
```

### Пример парсинга

```go
func parseTransactionLine(line string, sourceFolder string) (*models.Transaction, error) {
    fields := strings.Split(line, ";")

    // Базовые поля (есть во всех транзакциях)
    transactionIDUnique, _ := strconv.ParseInt(fields[0], 10, 64)
    transactionDate := parseDate(fields[1])
    transactionTime := parseTime(fields[2])
    transactionType, _ := strconv.Atoi(fields[3])

    // Специфичные поля зависят от типа транзакции
    switch transactionType {
    case 1, 11: // Регистрация товара
        return &models.TransactionRegistration{
            TransactionIDUnique: transactionIDUnique,
            SourceFolder:        sourceFolder,
            TransactionDate:     transactionDate,
            TransactionTime:     transactionTime,
            TransactionType:     transactionType,
            CashRegisterCode:    parseInt(fields[4]),
            DocumentNumber:      parseInt64(fields[5]),
            ItemCode:            fields[6],
            Quantity:            parseFloat(fields[7]),
            ItemSum:             parseFloat(fields[8]),
            // ... остальные поля
        }, nil

    case 3: // Специальная цена
        return &models.SpecialPrice{
            // ... поля для спеццены
        }, nil

    // ... остальные типы
    }
}
```

### Обработка специальных случаев

#### 1. Пустые поля
```go
func parseOptionalFloat(value string) *float64 {
    if value == "" {
        return nil
    }
    result, _ := strconv.ParseFloat(value, 64)
    return &result
}
```

#### 2. Даты в разных форматах
```go
func parseDate(dateStr string) time.Time {
    // Формат Frontol: DD.MM.YYYY
    layout := "02.01.2006"
    date, err := time.Parse(layout, dateStr)
    if err != nil {
        return time.Time{}
    }
    return date
}
```

#### 3. Валидация данных
```go
func validateTransaction(tx *models.Transaction) error {
    if tx.TransactionIDUnique == 0 {
        return errors.New("transaction_id_unique is required")
    }
    if tx.SourceFolder == "" {
        return errors.New("source_folder is required")
    }
    // ... остальные проверки
    return nil
}
```

---

## ❌ Обработка ошибок

### Уровни обработки ошибок

#### 1. Валидация входных данных

```go
func validateDate(dateStr string) error {
    _, err := time.Parse("2006-01-02", dateStr)
    if err != nil {
        return fmt.Errorf("invalid date format: %w", err)
    }
    return nil
}
```

#### 2. Сетевые ошибки (retry logic)

```go
func downloadWithRetry(ctx context.Context, ftpClient *ftp.Client, file string) error {
    maxRetries := 3
    retryDelay := 5 * time.Second

    for i := 0; i < maxRetries; i++ {
        err := ftpClient.Download(file)
        if err == nil {
            return nil
        }

        log.Warn("Download failed, retrying...", "attempt", i+1, "error", err)
        time.Sleep(retryDelay)
    }

    return fmt.Errorf("download failed after %d retries", maxRetries)
}
```

#### 3. Парсинг ошибок

```go
func parseFile(filePath string) (*Transactions, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to open file: %w", err)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    lineNum := 0

    for scanner.Scan() {
        lineNum++
        line := scanner.Text()

        tx, err := parseTransactionLine(line, sourceFolder)
        if err != nil {
            // Логируем ошибку, но продолжаем парсинг
            log.Warn("Failed to parse line", "file", filePath, "line", lineNum, "error", err)
            continue
        }

        // Добавляем транзакцию
        transactions.Add(tx)
    }

    if err := scanner.Err(); err != nil {
        return nil, fmt.Errorf("scanner error: %w", err)
    }

    return transactions, nil
}
```

#### 4. Database ошибки

```go
func (r *Repository) LoadData(ctx context.Context, data []models.Transaction) error {
    tx, err := r.pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback(ctx)

    for _, row := range data {
        _, err := tx.Exec(ctx, query, row...)
        if err != nil {
            return fmt.Errorf("failed to insert row: %w", err)
        }
    }

    if err := tx.Commit(ctx); err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }

    return nil
}
```

### Логирование ошибок

Используется structured logging (slog):

```go
log.Error("ETL failed",
    "date", date,
    "error", err.Error(),
    "files_processed", filesProcessed,
    "transactions_loaded", transactionsLoaded,
)
```

### Отчеты об ошибках

Если настроен `WEBHOOK_REPORT_URL`, ошибки отправляются в мониторинг:

```json
{
  "request_id": "req_123",
  "date": "2024-12-18",
  "status": "error",
  "error": "failed to parse file: invalid transaction type",
  "files_processed": 5,
  "transactions_loaded": 0,
  "duration_seconds": 15
}
```

---

## 🔁 Идемпотентность

### Что такое идемпотентность?

**Идемпотентность** - свойство операции, позволяющее выполнять её многократно без изменения результата после первого выполнения.

### Зачем это нужно?

- ✅ Безопасный повторный запуск при ошибке
- ✅ Обработка дубликатов данных
- ✅ Восстановление после сбоев

### Как это реализовано?

#### 1. Уникальные ключи

Все таблицы используют составной первичный ключ:
```sql
PRIMARY KEY (transaction_id_unique, source_folder)
```

- `transaction_id_unique` - уникальный ID из Frontol
- `source_folder` - идентификатор кассы/папки

#### 2. ON CONFLICT DO UPDATE

При попытке вставить дубликат данные обновляются:

```sql
INSERT INTO tx_item_registration_1_11 (...)
VALUES (...)
ON CONFLICT (transaction_id_unique, source_folder)
DO UPDATE SET
    transaction_date = EXCLUDED.transaction_date,
    amount_total = EXCLUDED.amount_total,
    ...
```

#### 3. Пример работы

**Первая загрузка:**
```sql
INSERT INTO transactions (id, folder, amount)
VALUES (1, 'P13', 100.00)
-- Результат: INSERT 0 1 (новая запись создана)
```

**Повторная загрузка (те же данные):**
```sql
INSERT INTO transactions (id, folder, amount)
VALUES (1, 'P13', 150.00)
ON CONFLICT (id, folder) DO UPDATE SET amount = EXCLUDED.amount
-- Результат: UPDATE 1 (запись обновлена)
```

**Результат:**
- Нет дубликатов
- Данные актуализированы
- Можно повторять сколько угодно раз

---

## 📊 Статистика и мониторинг

### Логирование статистики

```go
log.Info("ETL completed",
    "date", date,
    "duration_seconds", duration.Seconds(),
    "files_processed", filesProcessed,
    "files_skipped", filesSkipped,
    "transactions_loaded", transactionsLoaded,
    "errors", errorCount,
)
```

### Метрики по типам транзакций

```go
log.Debug("Transaction breakdown",
    "tx_item_registration_1_11", len(transactions.Registrations),
    "tx_bonus_accrual_9", len(transactions.Bonuses),
    "discount_transactions", len(transactions.Discounts),
    "special_prices", len(transactions.SpecialPrices),
    // ... остальные типы
)
```

---

## 📚 См. также

- [ARCHITECTURE.md](ARCHITECTURE.md) - Архитектура системы
- [DATABASE.md](DATABASE.md) - Структура базы данных
- [API.md](API.md) - API документация
- [frontol_6_integration.md](../frontol_6_integration.md) - Формат файлов Frontol

---

**Последнее обновление:** 2026-01-03
