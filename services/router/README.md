# Router

Сервис **видео-контура**: RTSP → **Kafka** (без S3; загрузка кадров — **pusher**):

| Топик | Назначение |
|-------|------------|
| `its.video.ingest` | метаданные кадра → **analytics** (как раньше) |
| `its.frames.ingest` | **новый**: мета + PNG base64 → **pusher** → S3 |
| `its.ml.accident.in` / `its.ml.congestion.in` | JPEG для **ml-serving** |

## Переменные окружения

| Переменная | По умолчанию |
|------------|--------------|
| `KAFKA_TOPIC_VIDEO` | `its.video.ingest` |
| `KAFKA_TOPIC_FRAMES` | `its.frames.ingest` |
| `KAFKA_TOPIC_ML_ACCIDENT_IN` | `its.ml.accident.in` |
| `KAFKA_TOPIC_ML_CONGESTION_IN` | `its.ml.congestion.in` |
