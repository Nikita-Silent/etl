# 📚 Документация Frontol ETL

Центральная навигация по документации проекта. Обновляйте этот файл при
добавлении новых документов или изменении источников истины.

---

## 🗂️ Навигация

### 🏗️ Архитектура и дизайн
- [infrastructure/ARCHITECTURE.md](infrastructure/ARCHITECTURE.md) - Архитектура проекта
- [infrastructure/TECH_STACK.md](infrastructure/TECH_STACK.md) - Технический стек и обоснования

### 💾 База данных
- [database/DATABASE.md](database/DATABASE.md) - База данных (актуально)
- [database/DDL_SPEC.md](database/DDL_SPEC.md) - Источник истины по структуре таблиц
- [database/TRANSACTION_TABLES_SCHEMA.md](database/TRANSACTION_TABLES_SCHEMA.md) - Схема транзакционных таблиц
- [database/TRANSACTION_TABLES_SPEC.md](database/TRANSACTION_TABLES_SPEC.md) - Семантика полей

### 🔌 API и интерфейсы
- [infrastructure/API.md](infrastructure/API.md) - API и интерфейсы
- [CONFIGURATION.md](CONFIGURATION.md) - Переменные окружения
- `api/openapi.yaml` - источник истины по HTTP API

### 📦 Бизнес-логика и формат выгрузок
- [infrastructure/BUSINESS_LOGIC.md](infrastructure/BUSINESS_LOGIC.md) - ETL pipeline
- [frontol_6_integration.md](frontol_6_integration.md) - Навигация по интеграции Frontol 6
- [Frontol_6_Integration.pdf](Frontol_6_Integration.pdf) - источник истины по формату выгрузок
- [../PARSERS_TO_TABLES_MAPPING.md](../PARSERS_TO_TABLES_MAPPING.md) - Маппинг парсеров

### 🧪 Разработка и тестирование
- [coding/CODING_RULES.md](coding/CODING_RULES.md) - Правила написания кода
- [test/TESTING.md](test/TESTING.md) - Руководство по тестированию
- [test/TEST_PLAN.md](test/TEST_PLAN.md) - План тестирования

### 🐳 Развертывание и Docker
- [infrastructure/DEPLOYMENT.md](infrastructure/DEPLOYMENT.md) - Production deployment
- [../DOCKER_COMPOSE_GUIDE.md](../DOCKER_COMPOSE_GUIDE.md) - Руководство по Docker Compose
- [docker/DOCKER_BEST_PRACTICES.md](docker/DOCKER_BEST_PRACTICES.md) - Best practices для Docker

### 🔧 Решение проблем и поддержка
- [coding/TROUBLESHOOTING.md](coding/TROUBLESHOOTING.md) - Troubleshooting
- [../QUICKSTART.md](../QUICKSTART.md) - Быстрый старт

### 🗺️ Улучшения и планы
- [IMPROVEMENTS.md](IMPROVEMENTS.md) - Идеи и техдолг

---

## 📊 Структура docs/

```
docs/
├── README.md
├── CONFIGURATION.md
├── IMPROVEMENTS.md
├── frontol_6_integration.md
├── Frontol_6_Integration.pdf
├── docker/DOCKER_BEST_PRACTICES.md
├── infrastructure/ARCHITECTURE.md
├── infrastructure/API.md
├── infrastructure/BUSINESS_LOGIC.md
├── infrastructure/DEPLOYMENT.md
├── infrastructure/TECH_STACK.md
├── database/DATABASE.md
├── database/DDL_SPEC.md
├── database/TRANSACTION_TABLES_SCHEMA.md
├── database/TRANSACTION_TABLES_SPEC.md
├── coding/CODING_RULES.md
├── coding/TROUBLESHOOTING.md
└── test/TESTING.md
```
