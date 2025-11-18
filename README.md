# Frame Control System — микросервисное приложение

Микросервисный приложение на Go, запускаемый через Docker Compose.

## Сервисы (docker-compose)

- `app` — HTTP API (Go), порт 8080
- `postgres` — СУБД, порт 5432

Файл оркестрации: `docker-compose.yml`

## Быстрый старт (Docker)

1. Требуется Docker/Docker Compose.
2. Запустите стэк:
   - Только приложение:  
     `docker compose up -d app`
3. Проверка:  
   `GET http://localhost:8080/api/v1/healthz` → ожидается `{ "success": true }`

## Ручной запуск (без Docker)

1. Требуется Go 1.22+.
2. Установите переменные окружения (см. ниже) или создайте `.env`.
3. Локально:
   - `make tidy`
   - `make run`

## Переменные окружения

- `APP_ENV` — профиль (`dev`/`test`/`prod`), по умолчанию `dev`
- `APP_PORT` — порт HTTP (по умолчанию `8080`)
- `DB_DSN` — строка подключения к БД (пример: `postgres://appuser:apppass@postgres:5432/appdb?sslmode=disable`)
- `JWT_SECRET` — секрет для подписи JWT (обязателен в prod)
- `CORS_ORIGINS` — `*` или список источников через запятую
- `LOG_LEVEL` — уровень логов (`info`, `debug`, …)
- `RATE_LIMIT_RPS` — глобальный RPS лимит (float)
- `RATE_LIMIT_BURST` — burst для rate limit

См. пример: `.env.example`.

## Эндпоинты

- `GET /api/v1/healthz`
- `POST /api/v1/users/register`
- `POST /api/v1/users/login`
- `GET /api/v1/users/me` (JWT)
- `PATCH /api/v1/users/me` (JWT)
- `GET /api/v1/users` (admin)
- `POST /api/v1/orders` (JWT)
- `GET /api/v1/orders` (JWT; admin видит всех)
- `GET /api/v1/orders/{id}` (JWT; владелец или admin)
- `PATCH /api/v1/orders/{id}/status` (JWT; валидные переходы)
- `DELETE /api/v1/orders/{id}` (JWT)
- `GET /api/v1/events/outbox` (admin)
- Dev (не в prod): `POST /api/v1/dev/seed-admin` — создать/назначить admin и вернуть JWT

Документация: `docs/openapi.yaml`

## Администратор (dev)

Быстрый способ выдать права администратора:

`POST /api/v1/dev/seed-admin`

Body (необязательно):
```json
{ "email": "admin@example.com", "password": "admin123", "name": "Admin" }
```
Если пользователь существует — ему добавят роль `admin` и при необходимости обновят пароль, в ответе вернётся `token` для admin.

## Postman коллекция

- Импортируйте `docs/postman_collection.json` в Postman.
- Переменные окружения:
  - `baseUrl` — `http://localhost:8080/api/v1`
  - `token` — JWT (Login сохраняет автоматически)
  - `adminToken` — JWT админа (Dev: seed‑admin)
  - `orderId*` — идентификаторы заказов (ставятся тестами коллекции)

## Поведение и соглашения

- Формат ответа: `{ success, data?, error? }`, ошибка `{ code, message }`
- Версионирование путей: префикс `/api/v1`
- Авторизация: `Authorization: Bearer <JWT>`
- Логи: структурированные, включают `request_id`, статус, длительность
- Rate limit: глобальный, настраивается через env
- Доменные события: `order.created`, `order.status_updated` — доступны через outbox (admin)
