# Frame Control System (Go, SQLite)

Монолитный REST API на Go с SQLite.

## Запуск

1. Требуется Go 1.22+.
2. `make tidy` (один раз), затем `make run`.
3. Конфиг через переменные окружения: `APP_PORT=8080 DB_PATH=app.db JWT_SECRET=dev-secret-change-me`.

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

Документация: `docs/openapi.yaml`.

## Тесты

`go test ./...`


