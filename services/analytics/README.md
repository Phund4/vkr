# Analytics

Читает **Kafka**: `its.video.ingest` (метаданные кадра), `its.ml.accident.out` и `its.ml.congestion.out` (результаты **ml-serving**). Публикует `its.persist.events` → **pusher** (ClickHouse). Кадры в S3 — только через новый топик `its.frames.ingest` (router → pusher).

| Переменная | По умолчанию |
|------------|--------------|
| `KAFKA_TOPIC_VIDEO` | `its.video.ingest` |
| `KAFKA_TOPIC_ML_ACCIDENT_OUT` | `its.ml.accident.out` |
| `KAFKA_TOPIC_ML_CONGESTION_OUT` | `its.ml.congestion.out` |
| `KAFKA_TOPIC_PERSIST` | `its.persist.events` |
