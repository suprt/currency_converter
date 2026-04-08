![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)
![License](https://img.shields.io/badge/License-MIT-green)
![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/suprt/currency_converter/test.yml?label=CI&logo=github)
[![Go Report Card](https://goreportcard.com/badge/github.com/suprt/currency_converter)](https://goreportcard.com/report/github.com/suprt/currency_converter)

# Currency Converter API

REST API для конвертации валют с поддержкой кэширования (Redis или in-memory) и механизмами отказоустойчивости.

## Возможности

- Конвертация валют и получение курсов в реальном времени
- Гибкое кэширование: Redis или in-memory с автоматической очисткой
- Retry с exponential backoff при сбоях внешнего API
- Circuit Breaker для защиты от каскадных отказов
- Rate Limiting по алгоритму Token Bucket с лимитированием по IP
- Валидация параметров запросов (коды валют, суммы)
- Admin API для управления кэшем
- Swagger документация
- Docker контейнеризация
- Unit и интеграционные тесты (Testcontainers)
- CI/CD через GitHub Actions

## 🚀 Быстрый старт

### Требования
- Go 1.25+
- Redis (опционально)
- Docker и Docker Compose (опционально)

### Локальная установка

```bash
go mod download
```

### Запуск локально

```bash
# Скопируйте .env.example в .env и настройте переменные
cp .env.example .env

# Запустите сервер
go run cmd/api/main.go
```

Или используйте скрипты:

**Windows (PowerShell):**
```powershell
.\scripts\build.ps1 run
```

**Linux/macOS:**
```bash
make run
```

Сервер запустится на `http://localhost:8080`

### Запуск с Docker

```bash
# Скопируйте .env.example в .env
cp .env.example .env

# Запустите все сервисы (API + Redis)
docker-compose up -d

# Просмотр логов
docker-compose logs -f api

# Остановка
docker-compose down
```

API будет доступно на `http://localhost:8080`

## 📚 API Документация

### Swagger UI

После запуска сервера откройте в браузере:
```
http://localhost:8080/swagger/index.html
```

### Генерация документации

При изменении хендлеров перегенерируйте Swagger документацию:

```bash
swag init -g cmd/api/main.go -o internal/handler/docs
```

Или используйте Makefile:
```bash
make docs
```

## 📡 Endpoints

### Public API

| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/health` | Проверка здоровья |
| GET | `/rates` | Получить курс валют |
| GET | `/convert` | Конвертировать сумму |
| GET | `/currencies` | Список всех валют |

### Admin API (требует API ключ)

| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/admin/cache/size` | Размер кэша |
| GET | `/admin/cache/get` | Получить значение из кэша |
| GET | `/admin/cache/set` | Установить значение в кэш |
| GET | `/admin/cache/delete` | Удалить из кэша |
| POST | `/admin/cache/clear` | Очистить и обновить кэш |
| GET | `/admin/cache/check` | Проверить наличие в кэше |
| GET | `/admin/cache/ttl` | TTL ключа |

## 🔐 Аутентификация

Admin endpoints требуют заголовок `X-API-KEY`:

```bash
curl -H "X-API-KEY: your-secret-key" http://localhost:8080/admin/cache/size
```

## 📝 Примеры использования

### Конвертация валют

```bash
# Конвертировать 100 USD в EUR
curl "http://localhost:8080/convert?from=USD&to=EUR&amount=100"
```

### Получить курс

```bash
# Получить курс EUR к GBP
curl "http://localhost:8080/rates?from=EUR&to=GBP"
```

### Проверка здоровья

```bash
curl http://localhost:8080/health
```

## ⚙️ Конфигурация

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `API_KEY` | API ключ для внешнего API | - |
| `API_URL` | URL внешнего API | `https://currencyapi.net/api/v2/` |
| `SERVER_HOST` | Хост сервера | `localhost` |
| `SERVER_PORT` | Порт сервера | `8080` |
| `SERVER_TIMEOUT` | Таймаут сервера | `45s` |
| `CONVERTER_TIMEOUT` | Таймаут HTTP клиента | `10s` |
| `REDIS_USE` | Использовать Redis | `false` |
| `REDIS_ADDR` | Адрес Redis | `localhost:6379` |
| `RPS` | Запросов в секунду (rate limiter) | `10` |
| `BURST` | Burst лимит | `20` |
| `CIRCUIT_BREAKER_THRESHOLD` | Порог срабатывания Circuit Breaker | `5` |
| `CIRCUIT_BREAKER_TIMEOUT` | Время восстановления Circuit Breaker | `30s` |
| `LOG_LEVEL` | Уровень логирования | `info` |

Полный список в `.env.example`.

## 🛡️ Надёжность

### Retry логика

При сбоях внешнего API клиент автоматически повторяет запросы с **exponential backoff**:
- Максимум **3 попытки**
- Начальная задержка: **1 секунда**
- Множитель: **2x** (1s → 2s → 4s)
- Повторяются только при **сетевых ошибках** и **5xx ответах**
- **4xx ошибки** не повторяются (клиентская ошибка)

### Circuit Breaker

Защищает систему от **каскадных отказов** при длительных сбоях внешнего API:
- **CLOSED** — нормальная работа
- **OPEN** — блокировка запросов после N ошибок (даёт API время на восстановление)
- **HALF-OPEN** — пробный запрос для проверки восстановления

### Graceful Shutdown

При получении сигнала SIGTERM/SIGINT:
1. Сервер перестаёт принимать новые запросы
2. Текущие запросы обрабатываются (до 20s)
3. Останавливается фоновый обновитель курсов
4. Останавливается rate limiter
5. Закрывается соединение с Redis

## 🧪 Тестирование

### Unit-тесты

```bash
# Запустить все тесты
go test ./...

# Запустить с покрытием
go test ./... -cover

# Запустить конкретный пакет
go test ./internal/handler/... -v
```

**Windows (PowerShell):**
```powershell
.\scripts\build.ps1 test
.\scripts\build.ps1 test-cover
```

**Linux/macOS:**
```bash
make test
make test-cover
```

### Интеграционные тесты

Интеграционные тесты используют **Testcontainers** для поднятия реального Redis контейнера:

```bash
# Запустить интеграционные тесты (требуется Docker)
go test -tags=integration -v ./internal/repository/
```

**Что тестируется:**
- CRUD операции с Redis
- TTL и автоматическое удаление просроченных ключей
- Очистка кэша
- Корректная обработка ошибок

## 🐳 Docker

### Сборка образа

```bash
docker build -t currency_converter .
```

### Запуск с Docker Compose

```bash
# Запуск
docker-compose up -d

# Остановка
docker-compose down

# Пересборка
docker-compose up -d --build
```

### Переменные окружения для Docker

Создайте `.env` файл на основе `.env.example`:

```bash
# API
API_KEY=your-api-key-here
ADMIN_API_KEY=admin-secret-key-123

# Redis
REDIS_USE=true

# Rate limiter
RPS=10
BURST=20

# Logger
LOG_LEVEL=info
```

## 🔧 CI/CD

Проект использует **GitHub Actions** для автоматической проверки качества кода:

| Job | Что делает |
|-----|------------|
| **Test** | Установка Go → Download deps → Генерация Swagger → Сборка → Vet → Тесты (-race) → Интеграционные тесты |
| **Lint** | Установка Go → Генерация Swagger → golangci-lint |

Конфигурация: [`.github/workflows/test.yml`](.github/workflows/test.yml)

## 🏗️ Архитектура

```
cmd/
└── api/              # Точка входа
internal/
├── bootstrap/        # Инициализация приложения
├── client/           # HTTP клиент для внешнего API (retry, circuit breaker)
├── config/           # Конфигурация через env
├── handler/          # HTTP хендлеры + валидация ввода
│   └── docs/         # Swagger документация
├── logger/           # Логгер (slog)
├── middleware/       # Middleware (rate limiter, auth)
├── repository/       # Репозитории (Redis, in-memory)
├── routes/           # Роутинг (chi)
└── service/          # Бизнес-логика
```

## 📦 Стек технологий

- **Go 1.25** — основной язык
- **chi/v5** — HTTP роутер
- **go-redis/v9** — Redis клиент
- **testcontainers-go** — Интеграционное тестирование
- **swaggo/swag** — Swagger документация
- **godotenv** — Загрузка переменных окружения
- **slog** — Структурированное логирование (стандартная библиотека)

## 📜 Лицензия

Этот проект распространяется под лицензией **MIT**.
См. файл [LICENSE](LICENSE) для подробностей.
