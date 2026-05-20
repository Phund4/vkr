# router

Сервис **видео-контура**: RTSP → **S3** → параллельно **Kafka**:

- `its.video.ingest` — метаданные кадра (без ML);
- `its.ml.accident.in` и `its.ml.congestion.in` — JPEG кадра для двух моделей в **ml-serving**.

**ml-serving** пишет в `its.ml.accident.out` / `its.ml.congestion.out`; **analytics** обрабатывает ветки **независимо** (аварии в CH без ожидания загруженности) → `its.persist.events` → **pusher**.

Назначения камер — **coordinator** (`data_class=road_segment_video`). Router **не опрашивает** coordinator: ждёт push `POST /v1/reload` (шлёт coordinator из `POST /v1/assignments/reload`).

На том же порту, что метрики (`METRICS_LISTEN_ADDR`, по умолчанию `:9091`): `GET /metrics`, `POST /v1/reload`.

## Переменные окружения

| Переменная | Описание |
|------------|----------|
| `KAFKA_BOOTSTRAP_SERVERS` | брокер (обязателен) |
| `KAFKA_TOPIC_VIDEO` | мета кадра (по умолчанию `its.video.ingest`) |
| `KAFKA_TOPIC_ML_ACCIDENT_IN` | вход accident (по умолчанию `its.ml.accident.in`) |
| `KAFKA_TOPIC_ML_CONGESTION_IN` | вход congestion (по умолчанию `its.ml.congestion.in`) |
| `COORDINATOR_*` | идентичность инстанса и URL coordinator |

HTTP к ml-serving **не используется**.
