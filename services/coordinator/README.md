# coordinator

Сервис-координатор для **router**: источники и пул инстансов читаются **только из PostgreSQL** (`sources`, `ingestion_instances`, `worker_heartbeats`).

- `GET /v1/sources`, `GET /v1/assignments`, `GET /v1/ingestion_instances` — из БД
- `POST /v1/assignments/reload` — увеличить `revision` и **POST /v1/reload** на живые router из `ingestion_instances`
- `POST /v1/workers/heartbeat` — в БД

## Конфигурация

- `LISTEN_ADDR` — HTTP (по умолчанию `:8098`)
- `HEARTBEAT_TIMEOUT_SEC` — таймаут «живого» heartbeat (по умолчанию `30`)
- `DATABASE_URL` — **обязателен**, строка подключения к PostgreSQL

`.env` подхватывается автоматически (или `ENV_FILE`).

## Схема и seed

`infra/postgres/init/001_coordinator_schema.sql`, `002_coordinator_seed.sql` — миграции вне кода coordinator.

## API

- `GET /health`, `GET /metrics`
- `GET /v1/sources?zone_id=...`
- `GET /v1/assignments?...` — в ответе поле `revision`
- `POST /v1/assignments/reload?zone_id=zone-a` — после изменения `sources` в БД (push на router)
- `POST /v1/workers/heartbeat`
- `GET /v1/workers`
- `GET /v1/ingestion_instances?zone_id=...`

## Запуск

```bash
cd services/coordinator
export DATABASE_URL=postgres://...
go run ./cmd/coordinator
```
