# 🚀 Быстрый старт для проверки

Минимальный набор команд для проверки работоспособности проекта.

---

## ⚡ За 5 минут

```bash
# 1. Настройка окружения
cp env.example .env
# Отредактируйте .env (DB_HOST, DB_USER, DB_PASSWORD)

# 2. Запуск сервисов через docker-compose
docker compose up -d
# ✓ Миграции БД (автоматически)
# ✓ FTP сервер
# ✓ Webhook server

# 3. Проверка webhook
curl http://localhost:$SERVER_PORT/api/health
# ✓ {"status":"healthy",...}

# 4. Запуск ETL через webhook
curl -X POST http://localhost:$SERVER_PORT/api/load \
  -H 'Content-Type: application/json' \
  -d '{"date": "2024-12-18"}'
# ✓ {"status":"queued",...}

# 5. Просмотр логов
docker compose logs -f webhook-server
# ✓ ETL pipeline выполнен
```

**Готово!** Проект работает и ETL запущен ✅

📖 **Подробные руководства:** 
- [DOCKER_COMPOSE_GUIDE.md](DOCKER_COMPOSE_GUIDE.md) - Docker Compose
- [MIGRATIONS.md](MIGRATIONS.md) - Миграции БД (автоматические!)

---

## 📊 Полная проверка (10 минут)

### 1️⃣ Сборка образов

```bash
# Собрать Docker образы
docker-compose build

# Проверить образы
docker-compose images
```

### 2️⃣ Запуск сервисов

```bash
# Запустить все сервисы
docker-compose up -d

# Проверить статус
docker-compose ps

# Должны быть запущены:
# - migrate (exited 0 - это нормально, миграции выполнены!)
# - ftp-server
# - webhook-server
```

### 3️⃣ Проверка API

```bash
# Health check
curl http://localhost:$SERVER_PORT/api/health

# Должен вернуть:
# {
#   "status": "healthy",
#   "service": "frontol-etl-webhook",
#   "timestamp": "..."
# }
```

### 4️⃣ Тест парсера (опционально)

```bash
# Протестировать парсер на файле
docker-compose run --rm parser-test ./parser-test /app/test-data/sample.txt
```

### 5️⃣ Запуск ETL

```bash
# Вариант 1: Через webhook (рекомендуется)
curl -X POST http://localhost:$SERVER_PORT/api/load \
  -H 'Content-Type: application/json' \
  -d '{"date": "2024-12-18"}'

# Вариант 2: Через CLI (ручной контроль)
docker-compose run --rm clear-requests
sleep 60
docker-compose run --rm loader ./frontol-loader 2024-12-18

# Проверить логи
docker-compose logs -f webhook-server
```

---

## 🎯 Результаты проверки

| Компонент | Команда | Ожидаемый результат |
|-----------|---------|---------------------|
| **Сборка** | `docker-compose build` | Успешно |
| **Миграции** | `docker-compose up migrate` | Exited 0 (автоматически) |
| **Запуск** | `docker-compose up -d` | 2 сервиса Up |
| **Статус** | `docker-compose ps` | webhook-server Up |
| **Webhook** | `curl localhost:$SERVER_PORT/api/health` | healthy |
| **ETL** | `curl -X POST localhost:$SERVER_PORT/api/load...` | queued |
| **Логи** | `docker-compose logs webhook-server` | Успешно ✅ |

---

## 🔥 Горячие команды

```bash
# Запуск сервисов
docker-compose up -d

# ETL через webhook для сегодня
curl -X POST http://localhost:$SERVER_PORT/api/load -H 'Content-Type: application/json' -d '{}'

# ETL через webhook для даты
curl -X POST http://localhost:$SERVER_PORT/api/load \
  -H 'Content-Type: application/json' \
  -d '{"date":"2024-12-18"}'

# ETL через CLI (ручной режим)
docker-compose run --rm clear-requests
sleep 60
docker-compose run --rm loader ./frontol-loader 2024-12-18

# Просмотр логов в реальном времени
docker-compose logs -f webhook-server

# Остановка всех сервисов
docker-compose down

# Перезапуск
docker-compose restart
```

---

## ❓ Проблемы?

**Сервисы не запускаются:**
```bash
docker-compose down
docker-compose build
docker-compose up -d
```

**Webhook не отвечает:**
```bash
docker-compose logs webhook-server
docker-compose restart webhook-server
```

**ETL не работает:**
```bash
# Проверить логи
docker-compose logs -f webhook-server

# Проверить FTP
docker-compose logs ftp-server

# Проверить БД подключение
docker-compose exec webhook-server env | grep DB_
```

---

## 📚 Подробнее

- 🐳 **Docker Compose:** [DOCKER_COMPOSE_GUIDE.md](DOCKER_COMPOSE_GUIDE.md) - **Главное руководство!**
- 🔌 **Webhook API:** [WEBHOOK_GUIDE.md](WEBHOOK_GUIDE.md)
- 📝 **Правила кода:** [CODING_RULES.md](.cursor/rules/CODING_RULES.mdc)
- 📖 **Тестирование:** [TESTING.md](.cursor/rules/TESTING.mdc)
- 🛠️ **Makefile:** [MAKEFILE_REFERENCE.md](MAKEFILE_REFERENCE.md) - опционально

---

## ✅ Успех!

Если все команды выше выполнились успешно, проект полностью рабочий! 🎉
