# analytics

Читает **Kafka**: `its.video.ingest` (метаданные кадра), `its.ml.accident.out` и `its.ml.congestion.out` (результаты **ml-serving**). Ветки accident и congestion обрабатываются **независимо**; публикует нормализованные события в `its.persist.events` → **pusher** пишет в ClickHouse.

Опционально: `POST /v1/ingest` для отладки (тот же JSON, что в ML out-топиках).

## Запуск

```bash
cd services/analytics
go run ./cmd/analytics
```

| Переменная | Назначение |
|------------|------------|
| `LISTEN_ADDR` | HTTP (`/metrics`, `/health`, опционально `/v1/ingest`), по умолчанию `:8093` |
| `CRASH_ALERT_THRESHOLD` | порог для `crash_probability` и метрики alert (по умолчанию `0.8`) |
| `CONGESTION_PERSIST_INTERVAL_SEC` | минимальный интервал между persist congestion на камеру (согласуйте с `CONGESTION_INTERVAL_SEC` в ML) |
| `KAFKA_BOOTSTRAP_SERVERS` | брокеры; пусто — консьюмер не стартует |
| `KAFKA_CONSUMER_GROUP` | группа чтения |
| `KAFKA_TOPIC_VIDEO` | по умолчанию `its.video.ingest` |
| `KAFKA_TOPIC_ML_ACCIDENT_OUT` | по умолчанию `its.ml.accident.out` |
| `KAFKA_TOPIC_ML_CONGESTION_OUT` | по умолчанию `its.ml.congestion.out` |
| `KAFKA_TOPIC_PERSIST` | по умолчанию `its.persist.events` |

## API

- `POST /v1/ingest` — отладка (JSON как в `its.ml.*.out`).
- `GET /metrics` — Prometheus.
- `GET /health` — liveness.
