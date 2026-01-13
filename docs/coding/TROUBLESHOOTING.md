# 🔧 Troubleshooting - Решение проблем

## 📋 Содержание

1. [Частые проблемы](#частые-проблемы)
2. [Docker и Docker Compose](#docker-и-docker-compose)
3. [База данных](#база-данных)
4. [FTP](#ftp)
5. [Webhook Server](#webhook-server)
6. [ETL Pipeline](#etl-pipeline)
7. [Миграции](#миграции)
8. [Диагностика](#диагностика)

---

## 🐛 Частые проблемы

### 1. FTP Connection Failed

**Проблема:**
```
ERROR: failed to connect to FTP server: dial tcp: i/o timeout
```

**Возможные причины:**
- ❌ FTP сервер недоступен
- ❌ Неверный хост или порт
- ❌ Firewall блокирует подключение
- ❌ Passive mode не настроен

**Решение:**

```bash
# Проверить доступность FTP
nc -zv ftp.example.com $FTP_PORT
telnet ftp.example.com $FTP_PORT

# Проверить переменные окружения
echo $FTP_HOST
echo $FTP_PORT

# Проверить логи FTP сервера
docker-compose logs ftp-server

# Для Docker: проверить passive mode
# См. раздел ниже "Настройка passive mode"
```

**Настройка passive mode:**

```yaml
# docker-compose.yml
ftp-server:
  environment:
    - PUBLICHOST=localhost  # или внешний IP
    - PASV_MIN_PORT=21100
    - PASV_MAX_PORT=21110
  ports:
    - "${FTP_PORT}:${FTP_PORT}"
    - "${PASV_MIN_PORT}-${PASV_MAX_PORT}:${PASV_MIN_PORT}-${PASV_MAX_PORT}"
```

---

### 2. Database Connection Failed

**Проблема:**
```
ERROR: failed to connect to database: FATAL: password authentication failed
```

**Решение:**

```bash
# Проверить подключение к БД напрямую
psql -h postgres.example.com -U frontol_user -d kassa_db

# Проверить переменные окружения
docker-compose config | grep DB_

# Проверить, что БД запущена
docker-compose ps postgres
# или
systemctl status postgresql

# Проверить пароль в .env
cat .env | grep DB_PASSWORD
```

**Если БД в Docker:**

```bash
# Перезапустить PostgreSQL
docker-compose restart postgres

# Проверить логи
docker-compose logs postgres
```

---

### 3. Webhook не отвечает

**Проблема:**
```bash
curl http://localhost:$SERVER_PORT/api/health
# curl: (7) Failed to connect
```

**Решение:**

```bash
# Проверить, запущен ли webhook-server
docker-compose ps webhook-server
# или
systemctl status frontol-webhook

# Проверить порт
netstat -tulpn | grep $SERVER_PORT
# или
ss -tulpn | grep $SERVER_PORT

# Проверить логи
docker-compose logs webhook-server
# или
journalctl -u frontol-webhook -f

# Перезапустить сервис
docker-compose restart webhook-server
# или
systemctl restart frontol-webhook
```

---

### 4. Миграции не применяются

**Проблема:**
```
ERROR: Dirty database version X. Fix and force version.
```

**Решение:**

```bash
# Проверить текущую версию
make migrate-version

# Принудительно установить версию (ОСТОРОЖНО!)
make migrate-force V=3

# Применить миграции заново
make migrate-up

# Если используется Docker Compose:
docker-compose run --rm migrate -path=/migrations -database="$DB_DSN" version
docker-compose run --rm migrate -path=/migrations -database="$DB_DSN" force 3
docker-compose run --rm migrate -path=/migrations -database="$DB_DSN" up
```

---

### 5. ETL Pipeline падает с ошибкой

**Проблема:**
```
ERROR: failed to parse file: invalid transaction type
```

**Решение:**

```bash
# Проверить формат файла
cat /path/to/file.txt | head -10

# Запустить parser-test для отладки
docker-compose run --rm parser-test ./parser-test /path/to/file.txt

# Проверить логи ETL
docker-compose logs -f webhook-server | grep ERROR

# Увеличить уровень логирования
# В .env:
LOG_LEVEL=debug
docker-compose restart webhook-server
```

---

### 6. Файлы парсятся, но не загружаются в БД

**Проблема:**
```
WARN: failed to load tx_bonus_accrual_9: table does not exist
```

**Решение:**

```bash
# Проверить, применены ли миграции
make migrate-version
# Должна быть версия 3 или выше

# Проверить существование таблиц
psql -h postgres.example.com -U frontol_user -d kassa_db -c "\dt"

# Применить миграции
make migrate-up
```

---

### 7. Переменная окружения не подставляется

**Проблема:**
```
Warning: The "PUBLICHOST" variable is not set. Defaulting to a blank string.
```

**Решение:**

```bash
# Проверить .env файл
cat .env | grep PUBLICHOST

# Если переменной нет, добавить:
echo "PUBLICHOST=localhost" >> .env

# Перезапустить Docker Compose
docker-compose down
docker-compose up -d

# Проверить, что переменная подставилась
docker-compose config | grep PUBLICHOST
```

---

### 8. Дублирующиеся данные в БД

**Проблема:**
```
ERROR: duplicate key value violates unique constraint
```

**Решение:**

Это не должно происходить благодаря `ON CONFLICT DO UPDATE`. Проверьте:

```sql
-- Проверить первичный ключ
SELECT tablename, indexname, indexdef
FROM pg_indexes
WHERE schemaname = 'public'
AND tablename = 'tx_item_registration_1_11';

-- Должен быть:
-- PRIMARY KEY (transaction_id_unique, source_folder)
```

**Если ключ отсутствует:**

```bash
# Применить миграции заново
make migrate-down
make migrate-up
```

---

## 🐳 Docker и Docker Compose

### Сервисы не запускаются

```bash
# Проверить логи всех сервисов
docker-compose logs

# Очистить всё и пересоздать
docker-compose down -v  # ВНИМАНИЕ: удаляет volumes!
docker-compose build --no-cache
docker-compose up -d

# Проверить использование ресурсов
docker stats
```

### Образы не собираются

```bash
# Очистить Docker кэш
docker system prune -a

# Пересобрать без кэша
docker-compose build --no-cache

# Проверить Dockerfile
docker-compose config
```

### Недостаточно памяти

```bash
# Проверить использование памяти
docker stats

# Увеличить лимиты в docker-compose.yml
deploy:
  resources:
    limits:
      memory: 2G
```

---

## 🗄️ База данных

### Таблицы не созданы

```bash
# Проверить, применены ли миграции
make migrate-version

# Список таблиц
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "\dt"

# Применить миграции вручную
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -f kassa_ddl.sql
```

### Медленные запросы

```sql
-- Включить логирование медленных запросов
SET log_min_duration_statement = 1000; -- 1 секунда

-- Анализ запроса
EXPLAIN ANALYZE SELECT * FROM tx_item_registration_1_11 WHERE transaction_date >= '2024-01-01';

-- Проверить индексы
SELECT * FROM pg_indexes WHERE tablename = 'tx_item_registration_1_11';

-- Пересоздать индексы (если нужно)
REINDEX TABLE tx_item_registration_1_11;
```

### Connection Pool исчерпан

**Проблема:**
```
ERROR: sorry, too many clients already
```

**Решение:**

```bash
# Увеличить max_connections в PostgreSQL
# postgresql.conf:
max_connections = 200

# Или уменьшить pool size в приложении
# .env:
DB_MAX_CONNS=10
```

---

## 📁 FTP

### Папки не создаются

```bash
# Проверить FTP init container
docker-compose logs ftp-init

# Проверить права доступа
docker-compose exec ftp-server ls -la /home/vsftpd/

# Пересоздать структуру вручную
docker-compose run --rm ftp-init
```

### Файлы не скачиваются

```bash
# Проверить список файлов на FTP
docker-compose exec ftp-server ls -la /home/vsftpd/frontol/response/

# Проверить FTP логи
docker-compose logs ftp-server | grep RETR

# Тест подключения
ftp localhost
# user: frontol
# pass: <из .env>
# ls response/
```

---

## 🔌 Webhook Server

### 401 Unauthorized

**Проблема:**
```bash
curl -X POST http://localhost:$SERVER_PORT/api/load
# {"error": "unauthorized"}
```

**Решение:**

```bash
# Если установлен WEBHOOK_BEARER_TOKEN, используйте его:
curl -X POST http://localhost:$SERVER_PORT/api/load \
  -H "Authorization: Bearer your_token"

# Или уберите токен из .env (для dev)
# Закомментируйте:
# WEBHOOK_BEARER_TOKEN=...
```

### Таймауты

**Проблема:**
```
ERROR: context deadline exceeded
```

**Решение:**

```bash
# Увеличить таймауты в .env
WAIT_DELAY_MINUTES=5  # Вместо 1
REQUEST_TIMEOUT_SECONDS=300  # 5 минут

# Перезапустить сервис
docker-compose restart webhook-server
```

---

## 🔄 ETL Pipeline

### Файлы не обрабатываются

```bash
# Проверить, что файлы есть на FTP
docker-compose exec ftp-server ls -la /home/vsftpd/frontol/response/

# Запустить ETL вручную с debug
LOG_LEVEL=debug docker-compose run --rm loader ./frontol-loader 2024-12-18

# Проверить логи
docker-compose logs -f webhook-server | grep "file processed"
```

### Partial data loaded

**Проблема:**
```
INFO: Loaded 500 transactions, but file has 1000
```

**Решение:**

```bash
# Проверить логи парсера
grep "WARNING" logs/*.log

# Увеличить batch size
# .env:
BATCH_SIZE=5000

# Проверить memory limits
docker stats webhook-server
```

---

## 🔍 Диагностика

### Общая проверка системы

```bash
#!/bin/bash
echo "=== System Check ==="

# 1. Docker
echo "Docker:"
docker --version
docker-compose --version

# 2. Services
echo "Services:"
docker-compose ps

# 3. Health checks
echo "Webhook Health:"
curl -s http://localhost:$SERVER_PORT/api/health | jq .

# 4. Database
echo "Database:"
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "SELECT version();"

# 5. FTP
echo "FTP:"
nc -zv localhost $FTP_PORT

# 6. Disk space
echo "Disk:"
df -h

# 7. Memory
echo "Memory:"
free -h

# 8. Logs (last 10 errors)
echo "Recent errors:"
docker-compose logs --tail=100 | grep ERROR
```

### Сбор диагностической информации

```bash
#!/bin/bash
DIAG_DIR="./diagnostics_$(date +%Y%m%d_%H%M%S)"
mkdir -p $DIAG_DIR

# Конфигурация
docker-compose config > $DIAG_DIR/docker-compose-config.yml
cat .env > $DIAG_DIR/env.txt

# Логи
docker-compose logs > $DIAG_DIR/all-logs.txt
docker-compose logs webhook-server > $DIAG_DIR/webhook-logs.txt
docker-compose logs postgres > $DIAG_DIR/postgres-logs.txt

# Статус
docker-compose ps > $DIAG_DIR/services-status.txt
docker stats --no-stream > $DIAG_DIR/docker-stats.txt

# База данных
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "\dt" > $DIAG_DIR/db-tables.txt
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "SELECT COUNT(*) FROM tx_item_registration_1_11" > $DIAG_DIR/db-count.txt

echo "Diagnostics collected in: $DIAG_DIR"
```

---

## 📚 Дополнительные ресурсы

- [DEPLOYMENT.md](DEPLOYMENT.md) - Production deployment
- [CONFIGURATION.md](CONFIGURATION.md) - Конфигурация
- [API.md](API.md) - API документация
- [DATABASE.md](DATABASE.md) - База данных

---

## 🆘 Получить помощь

Если проблема не решена:

1. Соберите диагностическую информацию (скрипт выше)
2. Создайте Issue в GitHub с:
   - Описанием проблемы
   - Шагами для воспроизведения
   - Логами и диагностикой
   - Версией проекта
3. Проверьте [существующие Issues](https://github.com/user/go-frontol-loader/issues)

---

**Последнее обновление:** 2026-01-03
