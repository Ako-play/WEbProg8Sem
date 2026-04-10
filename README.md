# Cybersport — турнирный центр

Веб-приложение: регистрация и вход, дашборд с демо-дайджестом по **Dota 2** и **Counter-Strike 2**, инструменты «индекса матчапа» между командами (симуляция + история в базе).

## Возможности

- **Аккаунты**: email, логин (username), пароль; сессии на **JWT** (access + refresh), refresh хранится в PostgreSQL с возможностью отзыва.
- **Дайджест киберспорта**: турниры, сетки матчей, блоки патчей / новостей / режимов — отдаётся с бэкенда как готовая структура (демо-контент для проекта).
- **Команды и матчапы**: список команд-кодов, «живой» индекс пары, конвертер «вес → проекция» по индексу, история индекса по датам — значения **рассчитываются детерминированно** и при необходимости **сохраняются** в таблицу курсов в БД (история и актуальные запросы).

## Стек

### Backend

- Go 1.23
- PostgreSQL, `pgx`
- `golang-jwt/jwt`, `bcrypt`

### Frontend

- React 18, TypeScript, Vite

## Структура репозитория

| Путь | Описание |
|------|----------|
| [backend/](backend) | REST API, сервисы, репозитории, миграции |
| [frontend/](frontend) | SPA: авторизация и дашборд |
| [docker-compose.yml](docker-compose.yml) | Postgres, API, статика фронта |
| [.env.example](.env.example) | Пример переменных окружения |

## REST API

### Здоровье

- `GET /health`

### Авторизация (`Content-Type: application/json`)

- `POST /api/v1/auth/register` — тело: `email`, `password`, `username`
- `POST /api/v1/auth/login` — `email`, `password`
- `POST /api/v1/auth/refresh` — `refreshToken`
- `POST /api/v1/auth/logout` — опционально `refreshToken` (требуется Bearer access)
- `GET /api/v1/auth/me` — профиль (Bearer)

### Данные после входа (все с **Bearer** access)

- `GET /api/v1/teams` — список команд (код и название)
- `GET /api/v1/matchups/live?team=NAVI&opponents=G2,MOUZ` — «живые» индексы к выбранным оппонентам (алиасы query: `base` / `symbols` — см. код хендлеров)
- `GET /api/v1/matchups/index?team=NAVI&opponent=G2&weight=1` — индекс и проекция (алиасы: `base`, `target`, `amount`)
- `GET /api/v1/matchups/history?team=NAVI&opponent=G2&from=2026-04-01&to=2026-04-10` — история по дням (алиасы: `base`, `target`)
- `GET /api/v1/esports/digest` — JSON дайджест Dota 2 и CS2

CORS настраивается переменной `FRONTEND_ORIGIN` (можно несколько значений через запятую).

## Быстрый старт (Docker)

1. Скопируйте [`.env.example`](.env.example) в `.env` и при необходимости поправьте секреты и URL БД для compose (в `docker-compose` Postgres по умолчанию доступен сервису `db` на порту **5432 внутри сети**; с хоста часто маппится **5433 → 5432** — смотрите свой `docker-compose.yml`).
2. Запуск:

   ```bash
   docker compose up --build
   ```

3. В браузере:
   - фронтенд: <http://localhost:5173>
   - проверка API: <http://localhost:8080/health>

Фронтенд в Docker-режиме собирается с `VITE_API_URL` из окружения (см. compose).

## Локально без Docker

### Backend

```bash
cd backend
go mod tidy
go run ./cmd/api
```

Нужен запущенный PostgreSQL; строка подключения — в `.env` (`DATABASE_URL`), миграции применяются при старте приложения.

### Frontend

```bash
cd frontend
npm install
npm run dev
```

Укажите `VITE_API_URL` (например `http://localhost:8080/api/v1`) в `.env` фронта или через переменные окружения Vite.

## Где смотреть код

- конфиг: [backend/internal/config/config.go](backend/internal/config/config.go)
- сборка приложения и HTTP-сервер: [backend/internal/app/app.go](backend/internal/app/app.go)
- JWT и пользователи: [backend/internal/service/auth_service.go](backend/internal/service/auth_service.go)
- команды, индексы, история: [backend/internal/service/rates_service.go](backend/internal/service/rates_service.go)
- дайджест Dota 2 / CS2: [backend/internal/service/esports_feed.go](backend/internal/service/esports_feed.go)
- маршруты и JSON: [backend/internal/transport/http/handlers.go](backend/internal/transport/http/handlers.go)
- клиент API в браузере: [frontend/src/lib/api.ts](frontend/src/lib/api.ts)
- экран после входа: [frontend/src/components/Dashboard.tsx](frontend/src/components/Dashboard.tsx)

## Идеи для развития

- подключение реальных данных турниров и новостей (внешние API или админка)
- уведомления и избранные команды
- реальные котировки или внешний провайдер вместо симуляции индекса
- роли пользователей и админские сценарии
