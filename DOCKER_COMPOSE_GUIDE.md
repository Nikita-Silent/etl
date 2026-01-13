# 🐳 Docker Compose Guide - Запуск ETL

Полное руководство по запуску Frontol ETL через Docker Compose без использования Makefile.

---

## 📦 Структура сервисов

В `docker-compose.yml` описаны следующие сервисы:

### Постоянные сервисы (всегда работают):
- **`ftp-server`** - FTP сервер для тестирования
- **`webhook-server`** - HTTP сервер с webhook API

### CLI сервисы (запускаются по требованию):
- **`loader`** - Загрузчик данных с FTP в PostgreSQL
- **`parser-test`** - Тестер парсера файлов
- **`send-request`** - Отправка request.txt к кассам
- **`clear-requests`** - Очистка request/response папок

---

## 🚀 Быстрый старт

### 1. Запуск постоянных сервисов

```bash
# Запустить FTP и Webhook сервер
docker-compose up -d

# Проверить статус
docker-compose ps

# Проверить логи
docker-compose logs -f webhook-server
```

### 2. Применить миграции БД

```bash
# Применить миграции (требует отдельного migrate сервиса или ручного запуска)
docker-compose exec webhook-server ./migrate up
```

### 3. Проверить работоспособность

```bash
# Health check
curl http://localhost:$SERVER_PORT/api/health

# Должен вернуть:
# {"status":"healthy","timestamp":"...","service":"frontol-etl-webhook"}
```

---

## ⚡ Запуск ETL

### Вариант 1: Через Webhook API (рекомендуется)

Это **самый простой способ** - webhook-server автоматически выполнит весь pipeline.

```bash
# ETL для сегодняшней даты
curl -X POST http://localhost:$SERVER_PORT/api/load \
  -H 'Content-Type: application/json' \
  -d '{}'

# ETL для конкретной даты
curl -X POST http://localhost:$SERVER_PORT/api/load \
  -H 'Content-Type: application/json' \
  -d '{"date": "2024-12-18"}'

# Просмотр логов выполнения
docker-compose logs -f webhook-server
```

### Вариант 2: Ручной запуск через CLI

Полный контроль над каждым шагом ETL.

```bash
# Шаг 1: Очистить FTP папки
docker-compose run --rm clear-requests

# Шаг 2: (Опционально) Отправить request.txt
docker-compose run --rm send-request

# Шаг 3: Подождать ответа от Frontol (60 секунд)
sleep 60

# Шаг 4: Загрузить данные для сегодня
docker-compose run --rm loader

# Или для конкретной даты
docker-compose run --rm loader ./frontol-loader 2024-12-18
```

### Вариант 3: Автоматический pipeline (bash скрипт)

Создайте файл `run_etl.sh`:

```bash
#!/bin/bash
DATE=${1:-$(date +%Y-%m-%d)}

echo "🚀 Running ETL for date: $DATE"

echo "Step 1/3: Clearing FTP folders..."
docker-compose run --rm clear-requests

echo "Step 2/3: Waiting 60 seconds..."
sleep 60

echo "Step 3/3: Loading data..."
docker-compose run --rm loader ./frontol-loader $DATE

echo "✅ ETL completed!"
```

Использование:

```bash
chmod +x run_etl.sh

# Для сегодня
./run_etl.sh

# Для конкретной даты
./run_etl.sh 2024-12-18
```

---

## 📋 Основные команды

### Управление сервисами

```bash
# Запустить все постоянные сервисы
docker-compose up -d

# Остановить все сервисы
docker-compose down

# Перезапустить сервисы
docker-compose restart

# Пересобрать образы
docker-compose build

# Пересобрать и запустить
docker-compose up -d --build

# Остановить и удалить всё (включая volumes)
docker-compose down -v
```

### Просмотр состояния

```bash
# Список запущенных контейнеров
docker-compose ps

# Логи всех сервисов
docker-compose logs

# Логи конкретного сервиса
docker-compose logs webhook-server
docker-compose logs ftp-server

# Следить за логами в реальном времени
docker-compose logs -f webhook-server

# Последние 100 строк логов
docker-compose logs --tail=100 webhook-server
```

### Запуск CLI команд

```bash
# Загрузить данные
docker-compose run --rm loader

# Загрузить данные за дату
docker-compose run --rm loader ./frontol-loader 2024-12-18

# Очистить FTP папки
docker-compose run --rm clear-requests

# Отправить запросы
docker-compose run --rm send-request

# Протестировать парсер
docker-compose run --rm parser-test ./parser-test /app/test-data/sample.txt
```

### Отладка

```bash
# Открыть shell в webhook контейнере
docker-compose exec webhook-server sh

# Открыть shell в FTP контейнере
docker-compose exec ftp-server sh

# Посмотреть файлы на FTP
docker-compose exec ftp-server ls -la /home/frontol/

# Посмотреть структуру папок
docker-compose exec ftp-server find /home/frontol/ -type d

# Выполнить команду в контейнере
docker-compose exec webhook-server ./migrate version
```

---

## 🎯 Примеры использования

### Пример 1: Ежедневная загрузка

```bash
# 1. Убедиться что сервисы запущены
docker-compose ps

# 2. Запустить ETL через webhook
curl -X POST http://localhost:$SERVER_PORT/api/load \
  -H 'Content-Type: application/json' \
  -d '{}'

# 3. Проверить логи
docker-compose logs -f webhook-server
```

### Пример 2: Загрузка исторических данных

```bash
# Скрипт для загрузки за несколько дней
for date in 2024-12-15 2024-12-16 2024-12-17 2024-12-18; do
  echo "Loading data for $date..."
  curl -X POST http://localhost:$SERVER_PORT/api/load \
    -H 'Content-Type: application/json' \
    -d "{\"date\": \"$date\"}"
  echo "Waiting 2 minutes..."
  sleep 120
done
```

### Пример 3: Отладка с ручным контролем

```bash
# 1. Очистить папки
docker-compose run --rm clear-requests

# 2. Проверить что очистилось
docker-compose exec ftp-server ls -la /home/frontol/request/
docker-compose exec ftp-server ls -la /home/frontol/response/

# 3. Отправить запросы
docker-compose run --rm send-request

# 4. Проверить что отправилось
docker-compose exec ftp-server cat /home/frontol/request/P13/request.txt

# 5. Подождать
sleep 60

# 6. Загрузить данные с подробными логами
docker-compose run --rm loader

# 7. Проверить результат в БД
docker-compose exec webhook-server psql -h $DB_HOST -U $DB_USER -d $DB_NAME \
  -c "SELECT COUNT(*) FROM transactions_registration;"
```

---

## 🔧 Конфигурация

### Переменные окружения

Создайте файл `.env` в корне проекта:

```bash
# Database (внешний PostgreSQL)
DB_HOST=your-postgres-host
DB_PORT=5432
DB_USER=frontol_user
DB_PASSWORD=secure_password
DB_NAME=kassa_db
DB_SSLMODE=disable

# FTP Configuration
FTP_HOST=ftp-server
FTP_PORT=21
FTP_USER=frontol
FTP_PASSWORD=frontol123
FTP_REQUEST_DIR=/request
FTP_RESPONSE_DIR=/response

# Kassa Structure
KASSA_STRUCTURE=P13:P13;N22:N22_Inter,N22_FURN;SH54:SH54;S6:S6;L98:L98;L32:L32;S39:S39;O49:O49;L28:L28

# Application Configuration
LOCAL_DIR=/tmp/frontol
BATCH_SIZE=1000
MAX_RETRIES=3
RETRY_DELAY_SECONDS=5
WAIT_DELAY_MINUTES=1
LOG_LEVEL=info

# Webhook Configuration
SERVER_PORT=8080
WEBHOOK_REPORT_URL=
```

### Режимы запуска

```bash
# Development режим (с debug логами)
docker-compose -f docker-compose.yml -f docker-compose.override.yml up -d

# Production режим
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# По умолчанию (без указания файла) использует docker-compose.yml + docker-compose.override.yml
docker-compose up -d
```

---

## 📊 Мониторинг

### Health Check

```bash
# Webhook server health
curl http://localhost:$SERVER_PORT/api/health

# Ожидаемый ответ:
# {
#   "status": "healthy",
#   "timestamp": "2024-12-18T10:30:00Z",
#   "service": "frontol-etl-webhook"
# }
```

### Логи

```bash
# Все логи
docker-compose logs

# Только ошибки
docker-compose logs | grep -i error

# Логи за последний час
docker-compose logs --since 1h

# Логи с временными метками
docker-compose logs -t webhook-server
```

### Статистика ресурсов

```bash
# Использование CPU и памяти
docker stats

# Размер образов
docker-compose images

# Размер volumes
docker volume ls
docker volume inspect parcer_ftp_data
```

---

## 🐛 Troubleshooting

### Проблема: Webhook не запускается

```bash
# Проверить статус
docker-compose ps

# Проверить логи
docker-compose logs webhook-server

# Пересобрать и перезапустить
docker-compose up -d --build webhook-server
```

### Проблема: Ошибки подключения к БД

```bash
# Проверить переменные окружения
docker-compose config | grep DB_

# Проверить подключение из контейнера
docker-compose exec webhook-server sh -c \
  'psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "SELECT 1;"'
```

### Проблема: FTP не работает

```bash
# Проверить статус FTP
docker-compose ps ftp-server

# Проверить логи FTP
docker-compose logs ftp-server

# Перезапустить FTP
docker-compose restart ftp-server

# Проверить папки
docker-compose exec ftp-server ls -la /home/frontol/
```

### Проблема: Loader не находит файлы

```bash
# Проверить что есть на FTP
docker-compose exec ftp-server find /home/frontol/ -name "*.txt"

# Запустить loader с debug логами
docker-compose run --rm -e LOG_LEVEL=debug loader
```

### Очистка и перезапуск

```bash
# Остановить всё
docker-compose down

# Удалить volumes (ОСТОРОЖНО: удалит данные FTP)
docker-compose down -v

# Очистить образы
docker-compose down --rmi all

# Пересобрать с нуля
docker-compose build --no-cache

# Запустить заново
docker-compose up -d
```

---

## 🔄 Автоматизация

### Cron для ежедневного запуска

Добавьте в crontab:

```bash
crontab -e

# Запуск каждый день в 2:00
0 2 * * * cd /home/user/parcer && curl -X POST http://localhost:$SERVER_PORT/api/load -H 'Content-Type: application/json' -d '{}' >> /var/log/frontol-etl.log 2>&1
```

### Systemd Service

Создайте `/etc/systemd/system/frontol-etl.service`:

```ini
[Unit]
Description=Frontol ETL Docker Compose
After=docker.service network-online.target
Requires=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=/home/user/parcer
ExecStart=/usr/bin/docker-compose up -d
ExecStop=/usr/bin/docker-compose down
TimeoutStartSec=0

[Install]
WantedBy=multi-user.target
```

Активируйте:

```bash
sudo systemctl daemon-reload
sudo systemctl enable frontol-etl
sudo systemctl start frontol-etl
sudo systemctl status frontol-etl
```

### Systemd Timer для ETL

Создайте `/etc/systemd/system/frontol-etl-daily.timer`:

```ini
[Unit]
Description=Frontol ETL Daily Timer

[Timer]
OnCalendar=daily
OnCalendar=02:00
Persistent=true

[Install]
WantedBy=timers.target
```

Создайте `/etc/systemd/system/frontol-etl-daily.service`:

```ini
[Unit]
Description=Frontol ETL Daily Run
After=network-online.target

[Service]
Type=oneshot
ExecStart=/usr/bin/curl -X POST http://localhost:$SERVER_PORT/api/load -H 'Content-Type: application/json' -d '{}'
StandardOutput=journal
StandardError=journal
```

Активируйте:

```bash
sudo systemctl daemon-reload
sudo systemctl enable frontol-etl-daily.timer
sudo systemctl start frontol-etl-daily.timer
sudo systemctl list-timers
```

---

## 📚 См. также

- [docker-compose.yml](docker-compose.yml) - Основной файл конфигурации
- [docker-compose.override.yml](docker-compose.override.yml) - Development настройки
- [docker-compose.prod.yml](docker-compose.prod.yml) - Production настройки
- [WEBHOOK_GUIDE.md](WEBHOOK_GUIDE.md) - Подробнее о webhook API
- [EXTERNAL_POSTGRES.md](EXTERNAL_POSTGRES.md) - Настройка внешней БД

---

## ✅ Checklist

- [ ] Файл `.env` создан и настроен
- [ ] Сервисы запущены: `docker-compose up -d`
- [ ] Health check проходит: `curl http://localhost:$SERVER_PORT/api/health`
- [ ] FTP доступен: `docker-compose ps ftp-server`
- [ ] Миграции применены
- [ ] ETL запущен через webhook или CLI
- [ ] Логи показывают успешное выполнение

---

## 🎯 Основные команды - Шпаргалка

```bash
# Запуск сервисов
docker-compose up -d

# Остановка
docker-compose down

# Статус
docker-compose ps

# Логи
docker-compose logs -f webhook-server

# ETL через webhook
curl -X POST http://localhost:$SERVER_PORT/api/load -H 'Content-Type: application/json' -d '{"date":"2024-12-18"}'

# ETL через CLI
docker-compose run --rm clear-requests
sleep 60
docker-compose run --rm loader ./frontol-loader 2024-12-18

# Отладка
docker-compose exec webhook-server sh
```

Готово! Теперь у вас есть полное руководство по работе с ETL через Docker Compose. 🐳

