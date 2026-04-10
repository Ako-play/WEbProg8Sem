# Crypto Parser

Курсовой проект по теме конвертации криптовалют и просмотра истории котировок.

## Что реализовано

- REST API на Go
- авторизация через JWT access + refresh
- ручной IoC/DI через конструкторы и слой `app`
- PostgreSQL и SQL-миграции
- Docker Compose и `.env`-конфигурация
- React frontend для входа, криптоконвертации и просмотра истории пары

## Стек

### Backend

- Go
- PostgreSQL
- `pgx`
- `golang-jwt/jwt`
- `bcrypt`

### Frontend

- React
- TypeScript
- Vite

## Структура проекта

- [backend](backend)
- [frontend](frontend)
- [docs](docs)
- [docker-compose.yml](docker-compose.yml)

## REST endpoints

### Auth

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`
- `GET /api/v1/auth/me`

### Rates

- `GET /api/v1/currencies`
- `GET /api/v1/rates/latest?base=BTC&symbols=ETH,SOL,BNB`
- `GET /api/v1/rates/convert?base=BTC&target=ETH&amount=1.5`
- `GET /api/v1/rates/history?base=BTC&target=ETH&from=2026-03-01&to=2026-03-10`

## Быстрый старт

1. Скопируйте [`.env.example`](.env.example) в `.env`.
2. Запустите проект:
   - `docker compose up --build`
3. Откройте:
   - frontend: <http://localhost:5173>
   - backend: <http://localhost:8080/health>

По умолчанию backend получает данные из публичного API CoinGecko.

## Локальный запуск без Docker

### Backend

1. Перейдите в папку [backend](backend)
2. Выполните:
   - `go mod tidy`
   - `go run ./cmd/api`

### Frontend

1. Перейдите в папку [frontend](frontend)
2. Выполните:
   - `npm install`
   - `npm run dev`

## Где смотреть ключевые части

- конфигурация: [backend/internal/config/config.go](backend/internal/config/config.go)
- DI и инициализация: [backend/internal/app/app.go](backend/internal/app/app.go)
- авторизация: [backend/internal/service/auth_service.go](backend/internal/service/auth_service.go)
- работа с курсами: [backend/internal/service/rates_service.go](backend/internal/service/rates_service.go)
- HTTP-обработчики: [backend/internal/transport/http/handlers.go](backend/internal/transport/http/handlers.go)
- миграции: [backend/migrations/001_init.sql](backend/migrations/001_init.sql)
- frontend: [frontend/src/App.tsx](frontend/src/App.tsx)

## Идеи для развития

- роли пользователей и аудит действий
- планировщик фоновой синхронизации криптокотировок
- графики с расширенной аналитикой по парам
- кеш Redis для часто запрашиваемых криптопар
