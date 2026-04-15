# 🚀 Production Deployment

## 📋 Содержание

1. [Обзор](#обзор)
2. [Docker Compose (Рекомендуется)](#docker-compose-рекомендуется)
3. [Systemd Service](#systemd-service)
4. [Автоматизация (Cron/Systemd Timer)](#автоматизация)
5. [Monitoring и Health Checks](#monitoring-и-health-checks)
6. [Backup и Recovery](#backup-и-recovery)
7. [Security Best Practices](#security-best-practices)

---

## 🎯 Обзор

Frontol ETL может быть развернут в production несколькими способами:

1. **Docker Compose** ✅ (Рекомендуется) - простое развертывание с автоматическими миграциями
2. **Systemd Service** - нативный запуск в Linux
3. **Kubernetes** - теоретически возможен, но в этом репозитории не документирован и не является текущим целевым сценарием развертывания.

---

## 🤖 CI/CD (Автосборка Docker)

В репозитории настроена GitHub Actions сборка Docker-образа при каждом пуше.
Публикация образа в GHCR выполняется только для `main`, `master` и тегов `v*`.

**Где смотреть:**
- Workflow: `.github/workflows/docker.yml`
- Registry: `ghcr.io/<owner>/<repo>`

---

## 🐳 Docker Compose (Рекомендуется)

### Быстрый старт

```bash
# 1. Клонировать репозиторий
git clone <repo-url>
cd etl

# 2. Настроить окружение
cp env.example .env
nano .env  # Установить пароли и хосты

# 3. Запустить сервисы
docker-compose up -d

# 4. Проверить здоровье
curl http://localhost:$SERVER_PORT/api/health
```

### Структура сервисов

Текущий `docker-compose.yml` поднимает:

- `migrate` - one-shot миграции для внешнего PostgreSQL
- `ftp-structure-init` - one-shot инициализация структуры FTP
- `ftp-server` - локальный FTP сервер
- `webhook-server` - HTTP API
- `loader`, `parser-test`, `send-request`, `clear-requests`, `ftp-check` - CLI/diagnostic сервисы

PostgreSQL внешний и не входит в compose-стек.

### Production конфигурация

**Используйте существующий `docker-compose.prod.yml`:**

Он уже лежит в корне репозитория и переопределяет уровни логирования, лимиты памяти и публикацию портов для production.

**Запуск:**

```bash
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

### Обновление версии

```bash
# 1. Загрузить новую версию
docker-compose pull

# 2. Пересоздать контейнеры
docker-compose up -d

# 3. Проверить логи
docker-compose logs -f webhook-server
```

---

## ⚙️ Systemd Service

### Установка webhook-server как системный сервис

#### 1. Сборка бинарника

```bash
# Сборка для production
CGO_ENABLED=0 GOOS=linux go build -o webhook-server ./cmd/webhook-server

# Копирование в /opt
sudo mkdir -p /opt/frontol-loader
sudo cp webhook-server /opt/frontol-loader/
sudo cp migrate /opt/frontol-loader/
sudo chmod +x /opt/frontol-loader/*
```

#### 2. Создание .env файла

```bash
sudo nano /opt/frontol-loader/.env
```

**Содержимое:**

```bash
# Database
DB_HOST=postgres.example.com
DB_PORT=5432
DB_USER=frontol_user
DB_PASSWORD=your_secure_password
DB_NAME=kassa_db
DB_SSLMODE=require

# FTP
FTP_HOST=ftp.example.com
FTP_PORT=21
FTP_USER=frontol
FTP_PASSWORD=your_ftp_password

# Application
SERVER_PORT=8080
LOG_LEVEL=info
BATCH_SIZE=5000
WEBHOOK_BEARER_TOKEN=your_secret_token
```

#### 3. Создание systemd service

**Создайте `/etc/systemd/system/frontol-webhook.service`:**

```ini
[Unit]
Description=Frontol ETL Webhook Server
After=network.target postgresql.service
Wants=postgresql.service

[Service]
Type=simple
User=frontol
Group=frontol
WorkingDirectory=/opt/frontol-loader
EnvironmentFile=/opt/frontol-loader/.env
ExecStart=/opt/frontol-loader/webhook-server
Restart=always
RestartSec=10s

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/log/frontol

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=frontol-webhook

[Install]
WantedBy=multi-user.target
```

#### 4. Создание пользователя

```bash
sudo useradd -r -s /bin/false frontol
sudo chown -R frontol:frontol /opt/frontol-loader
sudo mkdir -p /var/log/frontol
sudo chown frontol:frontol /var/log/frontol
```

#### 5. Запуск сервиса

```bash
# Перезагрузить systemd
sudo systemctl daemon-reload

# Включить автозапуск
sudo systemctl enable frontol-webhook

# Запустить сервис
sudo systemctl start frontol-webhook

# Проверить статус
sudo systemctl status frontol-webhook

# Просмотр логов
sudo journalctl -u frontol-webhook -f
```

---

## 🔄 Автоматизация

### Вариант 1: Cron (для webhook-based запуска)

Автоматический запуск ETL каждый день в 2:00 AM.

**Создайте `/opt/frontol-loader/daily-etl.sh`:**

```bash
#!/bin/bash

# Дата для загрузки (вчера)
DATE=$(date -d "yesterday" +%Y-%m-%d)

# Webhook URL
WEBHOOK_URL="http://localhost:$SERVER_PORT/api/load"
BEARER_TOKEN="your_secret_token"

# Логи
LOG_FILE="/var/log/frontol/etl-trigger.log"

echo "[$(date)] Starting ETL for date: $DATE" >> $LOG_FILE

# Запуск ETL через webhook
RESPONSE=$(curl -s -X POST $WEBHOOK_URL \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $BEARER_TOKEN" \
  -d "{\"date\": \"$DATE\"}")

echo "[$(date)] Response: $RESPONSE" >> $LOG_FILE

if echo "$RESPONSE" | grep -q "queued"; then
    echo "[$(date)] ETL started successfully" >> $LOG_FILE
    exit 0
else
    echo "[$(date)] ETL failed to start" >> $LOG_FILE
    exit 1
fi
```

**Сделайте исполняемым:**

```bash
chmod +x /opt/frontol-loader/daily-etl.sh
chown frontol:frontol /opt/frontol-loader/daily-etl.sh
```

**Добавьте в crontab:**

```bash
sudo crontab -e -u frontol
```

**Добавьте строку:**

```cron
# Запуск ETL каждый день в 2:00 AM
0 2 * * * /opt/frontol-loader/daily-etl.sh
```

**Проверка cron логов:**

```bash
tail -f /var/log/frontol/etl-trigger.log
```

---

### Вариант 2: Systemd Timer

Более современная альтернатива cron.

#### 1. Создайте service unit

**`/etc/systemd/system/frontol-etl-daily.service`:**

```ini
[Unit]
Description=Frontol ETL Daily Job
After=network.target frontol-webhook.service
Requires=frontol-webhook.service

[Service]
Type=oneshot
User=frontol
Group=frontol
ExecStart=/opt/frontol-loader/daily-etl.sh
StandardOutput=journal
StandardError=journal
SyslogIdentifier=frontol-etl-daily
```

#### 2. Создайте timer unit

**`/etc/systemd/system/frontol-etl-daily.timer`:**

```ini
[Unit]
Description=Frontol ETL Daily Timer
Requires=frontol-etl-daily.service

[Timer]
# Запуск каждый день в 2:00 AM
OnCalendar=*-*-* 02:00:00
Persistent=true

# Если пропустили время (сервер был выключен), запустить при следующем старте
OnBootSec=5min

[Install]
WantedBy=timers.target
```

#### 3. Активация timer

```bash
# Перезагрузить systemd
sudo systemctl daemon-reload

# Включить timer
sudo systemctl enable frontol-etl-daily.timer

# Запустить timer
sudo systemctl start frontol-etl-daily.timer

# Проверить статус
sudo systemctl status frontol-etl-daily.timer

# Список всех timers
sudo systemctl list-timers

# Просмотр логов
sudo journalctl -u frontol-etl-daily.service -f
```

**Тестовый запуск:**

```bash
sudo systemctl start frontol-etl-daily.service
```

---

## 📊 Monitoring и Health Checks

### Health Check Endpoint

```bash
curl http://localhost:$SERVER_PORT/api/health
```

**Ожидаемый ответ:**

```json
{
  "status": "healthy",
  "timestamp": "2026-01-03T12:00:00Z",
  "service": "frontol-etl-webhook",
  "checks": {
    "database": {"status": "healthy", "latency_ms": 12},
    "ftp": {"status": "healthy", "latency_ms": 8},
    "queues": {
      "load_queue_size": 0,
      "download_queue_size": 0,
      "total_queue_size": 0,
      "active_operations": 0,
      "is_shutting_down": false
    }
  },
  "response_time_ms": 20
}
```

### Prometheus Metrics (Future)

Prometheus-метрики в текущем репозитории не реализованы и отдельной roadmap-документации для них нет.

### Логирование

**Structured logs (JSON формат):**

```bash
# Просмотр логов (systemd)
sudo journalctl -u frontol-webhook -f

# Просмотр логов (Docker)
docker-compose logs -f webhook-server

# Фильтрация по уровню
sudo journalctl -u frontol-webhook -p err

# Поиск по request_id
sudo journalctl -u frontol-webhook | grep "request_id=req_123"
```

### Мониторинг скриптом

**Создайте `/opt/frontol-loader/monitor.sh`:**

```bash
#!/bin/bash

WEBHOOK_URL="http://localhost:$SERVER_PORT/api/health"
ALERT_EMAIL="admin@example.com"

RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" $WEBHOOK_URL)

if [ "$RESPONSE" != "200" ]; then
    echo "Frontol ETL webhook is DOWN! HTTP code: $RESPONSE" | \
        mail -s "ALERT: Frontol ETL Down" $ALERT_EMAIL
    exit 1
fi

exit 0
```

**Добавьте в cron (каждые 5 минут):**

```cron
*/5 * * * * /opt/frontol-loader/monitor.sh
```

---

## 💾 Backup и Recovery

### Database Backup

```bash
# Ежедневный backup PostgreSQL
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/var/backups/frontol"
mkdir -p $BACKUP_DIR

pg_dump -h postgres.example.com \
        -U frontol_user \
        -d kassa_db \
        -F c \
        -f $BACKUP_DIR/kassa_db_$DATE.dump

# Удаление старых backup (>7 дней)
find $BACKUP_DIR -name "*.dump" -mtime +7 -delete
```

**Добавьте в cron:**

```cron
0 3 * * * /opt/frontol-loader/backup-db.sh
```

### Восстановление

```bash
pg_restore -h postgres.example.com \
           -U frontol_user \
           -d kassa_db \
           -c \
           /var/backups/frontol/kassa_db_20260103_030000.dump
```

---

## 🔐 Security Best Practices

### 1. Firewall

```bash
# Разрешить только необходимые порты
sudo ufw allow $SERVER_PORT/tcp  # Webhook server (только из доверенных сетей)
sudo ufw allow from 192.168.1.0/24 to any port $SERVER_PORT  # Только локальная сеть

# PostgreSQL доступ только с localhost
sudo ufw deny $DB_PORT/tcp
```

### 2. SSL/TLS

Используйте reverse proxy (nginx/caddy) для HTTPS:

**Пример nginx:**

```nginx
server {
    listen 443 ssl http2;
    server_name etl.example.com;

    ssl_certificate /etc/ssl/certs/example.com.crt;
    ssl_certificate_key /etc/ssl/private/example.com.key;

    location / {
        proxy_pass http://localhost:$SERVER_PORT;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### 3. Bearer Token

Обязательно установите `WEBHOOK_BEARER_TOKEN` в production:

```bash
# В .env
WEBHOOK_BEARER_TOKEN=$(openssl rand -hex 32)
```

### 4. Database SSL

```bash
# В .env
DB_SSLMODE=require
```

### 5. Регулярные обновления

```bash
# Обновление Docker образов
docker-compose pull
docker-compose up -d

# Обновление системных пакетов
sudo apt update && sudo apt upgrade -y
```

---

## 📚 См. также

- [QUICKSTART.md](../QUICKSTART.md) - Быстрый старт
- [DOCKER_COMPOSE_GUIDE.md](../DOCKER_COMPOSE_GUIDE.md) - Детальное руководство по Docker Compose
- [API.md](API.md) - API документация
- [../CONFIGURATION.md](../CONFIGURATION.md) - Конфигурация
- [../coding/TROUBLESHOOTING.md](../coding/TROUBLESHOOTING.md) - Решение проблем

---

**Последнее обновление:** 2026-01-03
