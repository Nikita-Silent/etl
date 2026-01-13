#!/bin/bash
# Скрипт очистки базы данных
# Удаляет все данные из таблиц транзакций, сохраняя справочники

set -euo pipefail

# Цвета для вывода
RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

confirm="${1:-}"
echo -e "${YELLOW}WARNING: This will delete all transaction data from the database!${NC}"
echo -e "${YELLOW}Reference tables will be preserved.${NC}"
echo ""
if [ "$confirm" != "--yes" ]; then
    read -r -p "Are you sure? Type 'yes' to confirm: " confirm
    if [ "$confirm" != "yes" ]; then
        echo "Aborted."
        exit 1
    fi
fi

# Загружаем переменные из .env если есть (без source, чтобы не падать на не-shell формате)
if [ -f .env ]; then
    while IFS='=' read -r key value; do
        case "$key" in
            ''|\#*) continue ;;
        esac
        value="${value%\"}"
        value="${value#\"}"
        value="${value%\'}"
        value="${value#\'}"
        case "$value" in
            *" #"*) value="${value%% #*}" ;;
        esac
        value="${value%"${value##*[![:space:]]}"}"
        export "$key=$value"
    done < .env
fi

# Проверяем наличие переменных окружения
if [ -z "${DB_HOST:-}" ] || [ -z "${DB_USER:-}" ] || [ -z "${DB_NAME:-}" ]; then
    echo -e "${RED}Error: DB_HOST, DB_USER, and DB_NAME environment variables must be set${NC}"
    exit 1
fi

echo ""
echo "Starting database cleanup..."

# Используем SQL скрипт если доступен psql
if command -v psql &> /dev/null; then
    echo "Using psql..."
    sql_file="$(dirname "$0")/clear-database.sql"
    PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -f "$sql_file"
elif [ -n "${DOCKER_COMPOSE:-}" ] || docker compose version &> /dev/null; then
    echo "Using docker compose..."
    docker compose run --rm --no-deps webhook-server ./clear-db --confirm
else
    echo -e "${RED}Error: Neither psql nor docker compose found${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}Database cleanup completed!${NC}"
