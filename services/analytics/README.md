# analytics

Сервис принимает `POST /v1/ingest` (от **ml-serving** с блоком ML) и сообщения из **Kafka** (`its.video.ingest`), обновляет **метрики Prometheus** и пишет в **ClickHouse** только по **ML**-событиям: загруженность (`road_congestion`), инциденты (`road_incidents`).

Локального списка событий нет — только gauge/counter в Prometheus и OLAP в ClickHouse.

## Запуск

Требуется доступный **ClickHouse** на `CLICKHOUSE_ADDR` (по умолчанию `127.0.0.1:9000`). Удобно поднять через [infra/docker-compose.yml](../../infra/docker-compose.yml); сам **analytics** к Docker не привязан.

```bash
cd services/analytics
go run ./cmd/analytics
```

Настройки в **`.env`** (в git не коммитится). Переопределение файла: **`ENV_FILE`**.

| Переменная | Назначение |
|------------|------------|
| `LISTEN_ADDR` | HTTP, по умолчанию `:8093` |
| `CLICKHOUSE_ADDR` | `host:9000` native, по умолчанию `127.0.0.1:9000` |
| `CLICKHOUSE_DATABASE` | по умолчанию `default` |
| `CLICKHOUSE_USER` / `CLICKHOUSE_PASSWORD` | учётка CH |
| `CLICKHOUSE_INCIDENTS_TABLE` | по умолчанию `road_incidents` |
| `CLICKHOUSE_CONGESTION_TABLE` | по умолчанию `road_congestion` |
| `CRASH_ALERT_THRESHOLD` | порог для `crash_probability` (плюс label `crash`) |
| `CONGESTION_PERSIST_INTERVAL_SEC` | минимальный интервал (сек.) между строками в таблице загруженности **на одну камеру** (по умолчанию `2`; согласуйте с `CONGESTION_INTERVAL_SEC` в ML) |
| `KAFKA_BOOTSTRAP_SERVERS` | брокеры через запятую; пусто — консьюмер Kafka не стартует |
| `KAFKA_CONSUMER_GROUP` | группа чтения |
| `KAFKA_TOPIC_VIDEO` | топик видео-событий (по умолчанию `its.video.ingest`) |

## API

- `POST /v1/ingest` — JSON road-events (`segment_id`, `camera_id`, `observed_at`, `s3_key`, `ml`).
- `GET /metrics` — Prometheus.
- `GET /health` — проверка процесса (не проверяет CH).

## Запись в ClickHouse

- **`road_congestion`**: не чаще чем раз в `CONGESTION_PERSIST_INTERVAL_SEC` **на пару** `(segment_id, camera_id)` — те же поля; метрика `analytics_road_congestion_score` по-прежнему обновляется на каждый ingest.
- **`road_incidents`** (или `CLICKHOUSE_INCIDENTS_TABLE`): только если **`incident.label` == `crash`** (без учёта регистра) **или** `incident.crash_probability >= CRASH_ALERT_THRESHOLD` — поля `crash_probability`, `incident_label`, `raw_ml` и т.д.

При обновлении с одной объединённой таблицы пересоздайте схему или выполните миграцию вручную: `CREATE TABLE` выполняется только для несуществующих таблиц.
