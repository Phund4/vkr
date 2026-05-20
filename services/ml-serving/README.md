# ml-serving

Две модели (accident / congestion) работают через **Kafka**, без HTTP-вызовов от router:

| Топик | Направление |
|-------|-------------|
| `its.ml.accident.in` | router → ml-serving |
| `its.ml.accident.out` | ml-serving → analytics |
| `its.ml.congestion.in` | router → ml-serving |
| `its.ml.congestion.out` | ml-serving → analytics |

Сообщение in: JSON с `jpeg_base64`, `segment_id`, `camera_id`, `observed_at`, `s3_key`, `pipeline_started_at`.

Сообщение out: те же meta + `ml` с одной веткой (`incident` или `congestion`).

HTTP остаётся только для `/health` и `/metrics`.

## Переменные

| Переменная | По умолчанию |
|------------|--------------|
| `KAFKA_BOOTSTRAP_SERVERS` | — (обязателен) |
| `KAFKA_TOPIC_ML_ACCIDENT_IN` | `its.ml.accident.in` |
| `KAFKA_TOPIC_ML_ACCIDENT_OUT` | `its.ml.accident.out` |
| `KAFKA_TOPIC_ML_CONGESTION_IN` | `its.ml.congestion.in` |
| `KAFKA_TOPIC_ML_CONGESTION_OUT` | `its.ml.congestion.out` |
| `CONGESTION_INTERVAL_SEC` | `2` (кеш инференса congestion) |
