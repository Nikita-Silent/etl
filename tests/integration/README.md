# Интеграционные тесты

Интеграционные тесты используют PostgreSQL в Docker контейнере.

## Быстрый старт

### Локальный запуск

```bash
# Запустить тесты с автоматическим поднятием БД (из корня репозитория)
./scripts/run-tests-with-db.sh
```

Или вручную:

```bash
# 1. Поднять PostgreSQL контейнер (из корня репозитория)
docker compose -f docker-compose.test.yml up -d postgres-test

# 2. Дождаться готовности БД
timeout 30 bash -c 'until docker exec frontol-postgres-test pg_isready -U frontol_user -d kassa_db_test; do sleep 1; done'

# 3. Применить миграции
export TEST_DB_HOST=localhost
export TEST_DB_PORT=5433
export TEST_DB_USER=frontol_user
export TEST_DB_PASSWORD=test_password
export TEST_DB_NAME=kassa_db_test
go run ./cmd/migrate/main.go up

# 4. Запустить тесты
export SKIP_INTEGRATION_TESTS=false
go test -v -tags=integration ./tests/integration/...

# 5. Остановить контейнер
docker compose -f docker-compose.test.yml down -v
```

## Переменные окружения

Тесты используют следующие переменные (значения по умолчанию для Docker контейнера):

- `TEST_DB_HOST` - хост БД (по умолчанию: `localhost`)
- `TEST_DB_PORT` - порт БД (по умолчанию: `5433`)
- `TEST_DB_USER` - пользователь БД (по умолчанию: `frontol_user`)
- `TEST_DB_PASSWORD` - пароль БД (по умолчанию: `test_password`)
- `TEST_DB_NAME` - имя БД (по умолчанию: `kassa_db_test`)
- `SKIP_INTEGRATION_TESTS` - пропустить тесты (по умолчанию: `true`)

Если `TEST_DB_*` переменные не установлены, тесты попытаются использовать стандартные `DB_*` переменные.

## CI/CD

В GitHub Actions тесты запускаются с Docker контейнером PostgreSQL. Контейнер поднимается перед тестами и останавливается после их завершения.

## Структура

- `db_test.go` - тесты подключения к БД и транзакций
- `loader_test.go` - тесты Loader с реальной БД
- `test_helpers.go` - вспомогательные функции для тестов

## Test Framework

Более подробный обзор окружения и testcontainers см. в `tests/integration/framework/README.md`.
