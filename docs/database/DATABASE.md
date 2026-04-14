# DATABASE

Краткая информация о БД и принципах (на основе файлов в `docs/database`).

## Назначение и модель
- БД хранит транзакционные выгрузки Frontol 6 в виде набора таблиц `tx_*`.
- Структура таблиц и типы колонок заданы в `DDL_SPEC.md`.
- Семантика полей и соответствие выгрузкам описаны в `TRANSACTION_TABLES_SPEC.md` и `TRANSACTION_TABLES_SCHEMA.md`.
- Используется PostgreSQL 17+ (кластер внешний по отношению к приложению).

## Общие правила по таблицам
- Во всех транзакционных таблицах есть:
  - `transaction_id_unique` BIGINT
  - `source_folder` TEXT
  - `transaction_date` DATE
  - `transaction_time` TIME
- Первичный ключ: `(transaction_id_unique, source_folder)`.
- Индексы:
  - `(<table>_date_idx)` по `transaction_date`
  - `(<table>_source_idx)` по `source_folder`
- Типы полей:
  - целые: BIGINT
  - дробные: NUMERIC(18,6)
  - дата: DATE
  - время: TIME
  - строка: TEXT
- Неописанные в документации поля именуются `reserved_<N>`.

## Служебные таблицы ETL
- Помимо `tx_*` таблиц, БД содержит служебные таблицы `etl_file_load_state` и `etl_operation_runs`.
- Назначение `etl_file_load_state`:
  - хранить durable-состояние успешно зафиксированной загрузки логического файла;
  - предотвращать повторную загрузку одного и того же `response.txt`, если локальный lifecycle-state не сохранился после DB commit;
  - хранить последний `content_hash` и `transaction_manifest` для корректного reconcile исправленных переотгрузок одной и той же даты.
- Ключ записи в `etl_file_load_state` - `logical_key`:
  - формат: `<remote_path>|<requested_date>`
  - пример: `/response/L32/L32_INTER/response.txt|2026-03-23`
- Основные поля таблицы:
  - `logical_key` TEXT PRIMARY KEY
  - `remote_path` TEXT
  - `requested_date` DATE
  - `source_folder` TEXT
  - `content_hash` TEXT
  - `transaction_manifest` JSONB
  - `updated_at` TIMESTAMPTZ
- Назначение `etl_operation_runs`:
  - хранить operation-level lifecycle для `load`, `download` и CLI запусков;
  - связывать все operational logs по `operation_id`;
  - помечать stale операции как `abandoned` после рестарта процесса.
- Основные поля таблицы:
  - `operation_id` TEXT PRIMARY KEY
  - `request_id` TEXT
  - `operation_type` TEXT
  - `status` TEXT
  - `date` TEXT
  - `source_folder` TEXT
  - `started_at` / `updated_at` / `finished_at` TIMESTAMPTZ
  - `error_message` TEXT
  - `failed_stage` TEXT
  - `timeout_report_sent` BOOLEAN
  - `crash_suspected` BOOLEAN

## Принципы хранения и обработки
- Данные группируются по типам транзакций (таблицы `tx_*`), набор колонок фиксирован.
- Номер телефона/карты хранится как TEXT (нечисловой формат считается валидным).
- Источником истины по структуре является `docs/database/DDL_SPEC.md`.

## ACID и транзакционность
- Изменения данных выполняются транзакционно и должны соответствовать ACID:
  - Atomicity: запись набора связанных строк выполняется целиком.
  - Consistency: все записи соответствуют схеме и правилам типов.
  - Isolation: параллельные загрузки не должны нарушать корректность чтения.
  - Durability: подтвержденные записи сохраняются при сбоях.
- Для загрузки одного логического файла durable-метаданные в `etl_file_load_state` записываются в ту же транзакцию, что и `tx_*` строки этого файла.
- Lifecycle записи в `etl_operation_runs` пишутся best-effort и не должны блокировать сам ETL pipeline, если registry временно недоступен.

## Миграции (по коду)
- Используется `golang-migrate` с драйвером `pgx` и источником миграций из embedded FS.
- Миграции лежат в `pkg/migrate/migrations/*.sql` и встраиваются в бинарь.
- При запуске мигратора база создается автоматически, если ее нет (`ensureDatabaseExists`).
- Поддерживаются операции: `Up`, `Down`, `Steps(n)`, `Version`, `Force`, `Drop`.
- Есть режим запуска миграций из пути на файловой системе через `NewMigratorFromPath`.

## UPDATE
- Любые UPDATE должны выполняться в транзакции.
- Для адресации строк используйте первичный ключ `(transaction_id_unique, source_folder)`.
- UPDATE применяется только для корректировок уже загруженных записей; структура таблиц не меняется через UPDATE.
