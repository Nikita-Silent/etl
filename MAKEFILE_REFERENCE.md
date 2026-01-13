# Makefile Reference

Справочник всех доступных команд Makefile для проекта Frontol ETL.

---

## 📦 Docker команды

| Команда | Описание |
|---------|----------|
| `make build` | Собрать Docker образы |
| `make up` | Запустить все сервисы |
| `make down` | Остановить все сервисы |
| `make dev` | Запуск в development режиме |
| `make prod` | Запуск в production режиме |
| `make logs` | Просмотр логов всех сервисов |
| `make logs-webhook` | Логи webhook сервера |
| `make logs-db` | Логи PostgreSQL |
| `make logs-ftp` | Логи FTP сервера |
| `make clean` | Удалить контейнеры и volumes |
| `make status` | Показать статус сервисов |
| `make restart` | Перезапустить сервисы |
| `make shell` | Открыть shell в webhook контейнере |
| `make stats` | Показать использование ресурсов |

---

## 🔧 Локальная разработка

| Команда | Описание |
|---------|----------|
| `make build-local` | Собрать все бинарники локально |
| `make clean-local` | Удалить локальные бинарники |
| `make run-local` | Запустить webhook сервер локально |
| `make run-loader-local` | Запустить loader локально |

---

## 🧪 Тестирование

| Команда | Описание |
|---------|----------|
| `make test-go` | Запустить все unit тесты |
| `make test-verbose` | Тесты с подробным выводом |
| `make test-coverage` | Тесты с покрытием кода |
| `make test-race` | Тесты с race detector |
| `make test-bench` | Запустить бенчмарки |
| `make test-integration` | Интеграционные тесты (требует сервисы) |
| `make test-all` | test-go + test-race + test-bench |

---

## 🎨 Качество кода

| Команда | Описание |
|---------|----------|
| `make fmt` | Форматирование кода (go fmt) |
| `make lint` | Запустить golangci-lint |
| `make check` | fmt + lint + test-go |
| `make ci` | fmt + lint + test-race + test-coverage |

---

## 🗄️ Миграции базы данных

| Команда | Описание | Пример |
|---------|----------|--------|
| `make migrate-up` | Применить все миграции | `make migrate-up` |
| `make migrate-down` | Откатить все миграции | `make migrate-down` |
| `make migrate-step` | Применить N миграций | `make migrate-step N=1` |
| `make migrate-version` | Текущая версия | `make migrate-version` |
| `make migrate-force` | Принудительно установить версию | `make migrate-force V=3` |
| `make migrate-drop` | Удалить все таблицы (ОПАСНО!) | `make migrate-drop` |
| `make migrate-create` | Создать новую миграцию | `make migrate-create NAME=add_users` |

---

## 🚀 ETL операции

### Полный ETL Pipeline (рекомендуется)

| Команда | Описание | Пример |
|---------|----------|--------|
| `make etl` | Запустить полный ETL для сегодня | `make etl` |
| `make etl-date` | Запустить полный ETL для даты | `make etl-date DATE=2024-12-18` |
| `make etl-webhook` | Триггер ETL через webhook (сегодня) | `make etl-webhook` |
| `make etl-webhook-date` | Триггер ETL через webhook (дата) | `make etl-webhook-date DATE=2024-12-18` |

### Ручные операции (для отладки)

| Команда | Описание |
|---------|----------|
| `make loader` | Запустить только loader |
| `make loader-date` | Loader для конкретной даты (DATE=YYYY-MM-DD) |
| `make send-request` | Отправить request.txt к кассам |
| `make clear-requests` | Очистить request/response папки |

---

## 💾 База данных

| Команда | Описание | Пример |
|---------|----------|--------|
| `make init-db` | Инициализировать БД | `make init-db` |
| `make backup-db` | Создать backup базы | `make backup-db` |
| `make restore-db` | Восстановить из backup | `make restore-db FILE=backup.sql` |

---

## 🔍 Утилиты

| Команда | Описание |
|---------|----------|
| `make health` | Health check всех сервисов |
| `make update` | Обновить и перезапустить |
| `make push` | Build и push в registry (требует REGISTRY) |
| `make setup-dev` | Настройка окружения для разработки |
| `make quick-start` | setup-dev + build + up |

---

## 📋 Примеры использования

### Первый запуск

```bash
# 1. Настройка
make setup-dev

# 2. Запуск
make build
make up

# 3. Инициализация БД
make migrate-up

# 4. Проверка
make health
```

### Ежедневная работа

```bash
# Проверка кода перед коммитом
make check

# Запуск полного ETL для сегодня (самый простой способ)
make etl

# Или через webhook (асинхронно)
make etl-webhook

# Просмотр логов
make logs
```

### Запуск ETL для конкретной даты

```bash
# Через CLI
make etl-date DATE=2024-12-18

# Через webhook
make etl-webhook-date DATE=2024-12-18

# Просмотр логов
make logs-webhook
```

### Отладка

```bash
# Смотрим логи конкретного сервиса
make logs-webhook

# Открываем shell в контейнере
make shell

# Проверяем БД
docker-compose exec postgres psql -U frontol_user -d kassa_db

# Смотрим статистику
make stats
```

### Тестирование

```bash
# Полная проверка
make ci

# Только unit тесты
make test-go

# С покрытием
make test-coverage
open coverage.html

# Бенчмарки
make test-bench
```

### Миграции

```bash
# Применить все
make migrate-up

# Откатить последнюю
make migrate-step N=-1

# Проверить версию
make migrate-version

# Создать новую
make migrate-create NAME=add_customers_table

# Исправить dirty state
make migrate-force V=3
make migrate-up
```

### Проблемы

```bash
# Перезапуск с нуля
make down
make clean
make build
make up
make migrate-up

# Очистка тестов
go clean -testcache
make test-go

# Rebuild бинарников
make clean-local
make build-local
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
