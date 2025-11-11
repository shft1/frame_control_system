# Frame Control System (Go, SQLite)

Монолитный REST API на Go с SQLite.

## Запуск

1. Требуется Go 1.22+ (рекомендуется актуальный toolchain).
2. Установите переменные окружения (см. ниже) или создайте файл `.env` по образцу `.env.example`.
3. Установите зависимости и запустите:
   - `make tidy` (однократно)
   - `make run`
4. Health-check: `GET http://localhost:8080/api/v1/healthz` → `{ "success": true }`

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

## Переменные окружения

- `APP_ENV` — профиль (`dev`/`test`/`prod`), по умолчанию `dev`
- `APP_PORT` — порт HTTP (по умолчанию `8080`)
- `DB_PATH` — путь к файлу SQLite (по умолчанию `app.db`)
- `JWT_SECRET` — секрет для подписи JWT (обязателен в prod)
- `CORS_ORIGINS` — `*` или список источников через запятую
- `LOG_LEVEL` — уровень логов (`info`, `debug`, …)
- `RATE_LIMIT_RPS` — глобальный RPS лимит (float)
- `RATE_LIMIT_BURST` — burst для rate limit

См. пример: `.env.example`.

## Тесты

`go test ./...`

## Postman коллекция

- Импортируйте `docs/postman_collection.json` в Postman.
- Настройте переменные окружения:
  - `baseUrl` (по умолчанию `http://localhost:8080/api/v1`)
  - `token` — клиентский JWT (получить из запроса Login)
  - `adminToken` — JWT админа (если используете роль admin)
  - `orderId` — идентификатор созданного заказа (сеттер установлен в тестах коллекции)

## Поведение и соглашения

- Формат ответа: `{ success, data?, error? }`, ошибка `{ code, message }`.
- Версионирование путей: префикс `/api/v1`.
- Авторизация: `Authorization: Bearer <JWT>`.
- Логи: структурированные, включают `request_id`, статус, длительность.
- Rate limit: глобальный, настраивается через env.
- Доменные события: `order.created`, `order.status_updated` — сохраняются в таблицу `outbox_events` (эндпоинт просмотра только для admin).

## Примечания по SQLite

- Включены `WAL` и `foreign_keys=ON`.
- Миграции выполняются автоматически при старте (встроены через `embed`).


