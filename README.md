# Currency Converter API

REST API для конвертации валют с поддержкой кэширования (Redis или in-memory).

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
| `REDIS_USE` | Использовать Redis | `false` |
| `REDIS_ADDR` | Адрес Redis | `localhost:6379` |
| `RPS` | Запросов в секунду (rate limiter) | `10` |
| `BURST` | Burst лимит | `20` |
| `LOG_LEVEL` | Уровень логирования | `info` |

Полный список в `.env.example`.

## 🧪 Тесты

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
API_KEY=your-api-key
ADMIN_API_KEY=admin-secret-key-123

# Redis
REDIS_USE=true

# Rate limiter
RPS=10
BURST=20

# Logger
LOG_LEVEL=info
```

## 🏗️ Архитектура

```
cmd/
└── api/              # Точка входа
internal/
├── bootstrap/        # Инициализация приложения
├── client/           # HTTP клиент для внешнего API
├── config/           # Конфигурация
├── handler/          # HTTP хендлеры
│   └── docs/         # Swagger документация
├── logger/           # Логгер (slog)
├── middleware/       # Middleware (rate limiter, auth)
├── repository/       # Репозитории (Redis, in-memory)
├── routes/           # Роутинг
└── service/          # Бизнес-логика
```
