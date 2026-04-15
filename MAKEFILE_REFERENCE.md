# Makefile Reference

Краткий справочник по актуальным таргетам из `Makefile`.

## Docker и compose

| Команда | Описание |
|---------|----------|
| `make build` | Собрать Docker-образы |
| `make up` | Запустить compose-стек |
| `make down` | Остановить compose-стек |
| `make dev` | Запуск с `docker-compose.override.yml` |
| `make prod` | Запуск с `docker-compose.prod.yml` |
| `make logs` | Все логи compose |
| `make logs-webhook` | Логи `webhook-server` |
| `make logs-ftp` | Логи `ftp-server` |
| `make logs-db` | Сообщение о том, что PostgreSQL внешний |
| `make status` | `docker-compose ps` |
| `make restart` | Перезапуск сервисов |
| `make clean` | Остановить и удалить контейнеры, volumes и образы |
| `make shell` | Shell внутри `webhook-server` |
| `make stats` | `docker stats` |
| `make health` | `curl` в локальный `/api/health` |

## ETL и ручные операции

| Команда | Описание |
|---------|----------|
| `make etl` | Полный ETL для текущей даты через CLI pipeline |
| `make etl-date DATE=YYYY-MM-DD` | Полный ETL для указанной даты |
| `make etl-webhook` | Триггер ETL через `/api/load` для текущей даты |
| `make etl-webhook-date DATE=YYYY-MM-DD` | Триггер ETL через `/api/load` для даты |
| `make loader` | Ручной запуск `loader` |
| `make loader-date DATE=YYYY-MM-DD` | Ручной запуск `loader` для даты |
| `make send-request` | Отправить `request.txt` |
| `make clear-requests` | Очистить FTP request/response |
| `make clear-db` | Очистить данные через `clear-db` CLI |
| `make clear-db-sql` | Очистить данные через SQL-скрипт |

## Локальная разработка

| Команда | Описание |
|---------|----------|
| `make build-local` | Собрать локальные бинарники |
| `make clean-local` | Удалить локальные бинарники |
| `make run-local` | Запустить `cmd/webhook-server` |
| `make run-loader-local` | Запустить `cmd/loader` |

## Тесты и качество

| Команда | Описание |
|---------|----------|
| `make fmt` | `go fmt ./...` |
| `make lint` | `golangci-lint run` |
| `make test-go` | `go test ./...` |
| `make test-reliability` | Фокусный regression suite |
| `make test-verbose` | `go test -v ./...` |
| `make test-coverage` | Покрытие и `coverage.html` |
| `make test-ftp-structure` | Интеграционный тест FTP структуры |
| `make test-race` | Race detector на всем проекте |
| `make test-race-critical` | Race detector на критичных пакетах |
| `make test-bench` | Бенчмарки |
| `make test-integration` | Интеграционные тесты по тегу `integration` |
| `make test-all` | `test-go` + `test-race` + `test-bench` |
| `make check` | `fmt` + `lint` + `test-go` |
| `make ci` | `fmt` + `lint` + `test-reliability` + `test-race-critical` |

## Миграции и внешняя БД

| Команда | Описание |
|---------|----------|
| `make init-db` | Прогнать compose-сервис `migrate` |
| `make migrate-up` | `go run ./cmd/migrate up` |
| `make migrate-down` | `go run ./cmd/migrate down` |
| `make migrate-step N=1` | `go run ./cmd/migrate step $(N)` |
| `make migrate-version` | Показать версию миграций |
| `make migrate-force V=3` | Принудительно выставить версию |
| `make migrate-drop` | Удалить все таблицы |
| `make migrate-create NAME=...` | Создать новую пару файлов миграции |
| `make backup-db` | Backup внешней БД через локальный `pg_dump` |
| `make restore-db FILE=backup.sql` | Restore внешней БД через локальный `psql` |

Для `backup-db` и `restore-db` должны быть доступны `pg_dump`/`psql` и выставлены `DB_HOST`, `DB_USER`, `DB_PASSWORD`.

## Быстрые сценарии

```bash
# первый запуск
make setup-dev
make build
make up
make health

# ручной ETL для даты
make etl-date DATE=2024-12-18

# миграции
make migrate-version
make migrate-up
```

---

## 🎯 Быстрые комбо

```bash
# Полная проверка кода
make check

# CI пайплайн
make ci

# Запуск сервисов и ETL
make dev && make migrate-up && make etl

# Запуск ETL с просмотром логов
make etl & make logs-webhook

# Перезапуск всего
make down && make clean && make build && make up && make migrate-up

# Бэкап и миграция
make backup-db && make migrate-up

# Отладка webhook
make logs-webhook -f &
make etl-webhook-date DATE=2024-12-18
```

---

## 💡 Советы

1. **Перед коммитом:** `make check`
2. **Перед pull request:** `make ci`
3. **Проблемы с Docker:** `make down && make clean && make build && make up`
4. **Dirty миграции:** `make migrate-force V=0 && make migrate-up`
5. **Сборка бинарников:** `make build-local`

---

## 📚 См. также

- [QUICKSTART.md](QUICKSTART.md) - Быстрый старт
- [TESTING.md](TESTING.md) - Подробное руководство по тестированию
- [CODING_RULES.md](CODING_RULES.md) - Правила написания кода
