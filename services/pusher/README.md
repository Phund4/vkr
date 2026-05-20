# pusher

Сервис читает **Kafka** топик **`its.persist.events`** (сообщения после склейки ML в **analytics**), пишет строки в **ClickHouse** (`road_incidents`, `road_congestion`) и при необходимости работает с **S3** (MinIO). HTTP используется для **health/readiness** и отдельный порт — для **Prometheus**.

## Запуск

Нужны **Kafka**, **ClickHouse**, учётка **S3** (или MinIO). Удобно поднять стек из [infra/docker-compose.yml](../../infra/docker-compose.yml) (сервис **`pusher`** в профиле `apps`).

```bash
cd services/pusher
go run ./cmd
```

Обязательная переменная **`ENV`** (например `dev`). Остальное — через префиксы ниже или см. `internal/config`.

## Переменные окружения

| Префикс / переменная | Назначение |
|----------------------|------------|
| `ENV` | окружение процесса (обязательно) |
| `SERVER_PORT` | Echo, по умолчанию `:8094` |
| `KAFKA_BOOTSTRAP_SERVERS` | брокеры |
| `KAFKA_CONSUMER_GROUP` | группа (по умолчанию `pusher`) |
| `KAFKA_TOPIC_PERSIST` | топик (по умолчанию `its.persist.events`) |
| `CLICKHOUSE_ADDR` | `host:9000` native |
| `CLICKHOUSE_DATABASE` / `USER` / `PASSWORD` | учётка CH |
| `CLICKHOUSE_INCIDENTS_TABLE` / `CONGESTION_TABLE` | таблицы (defaults `road_incidents` / `road_congestion`) |
| `S3_ENDPOINT` / `S3_BUCKET` | MinIO/S3 |
| `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` | ключи (как у router) |
| `METRICS_ENABLED` | включить сервер метрик (по умолчанию `true`) |
| `PROMETHEUS_PORT` | порт `/metrics` (по умолчанию `8082`) |
| `PROMETHEUS_HTTP_PATH` | путь метрик (по умолчанию `/metrics`) |

## HTTP

- `GET /health`, `GET /probe/health`, `GET /probe/ready` — на `SERVER_PORT`
- `GET /metrics` — на отдельном адресе `:{PROMETHEUS_PORT}{PROMETHEUS_HTTP_PATH}`

## Линтер

```bash
golangci-lint run
```

Конфиг: [`.golangci.yml`](./.golangci.yml).
