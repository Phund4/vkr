# router (видео-поток → S3 → Kafka + два ML)

Сервис для **видео-контура** по схеме: RTSP → **S3** → параллельно **Kafka** (`its.video.ingest`, метаданные кадра) и **два HTTP-вызова ML** (`/v1/process/accident` и `/v1/process/congestion`); каждый ML шлёт частичный JSON в **analytics**, где результат склеивается и уходит в **Kafka persist** → **pusher** → ClickHouse/S3. Назначения камер — **coordinator** (`data_class=road_segment_video`).

**Камеры:** сегменты, `camera_id` и **`rtsp_url`** приходят из назначений. Если процесс **router** в той же Docker-сети, что и MediaMTX, в `rtsp_url` укажите хост **`mediamtx`** (например `rtsp://mediamtx:8554/cam-01`).

Потоки для примера в `config.cameras.yaml`: поднимите **MediaMTX** и **`rtsp-generator`** в **`infra`** (веб-UI **http://localhost:8096**); RTSP URL задаёте сами (Postgres/coordinator или YAML).

Ключи S3: `{prefix}/{YYYY-MM-DD}/{camera_id}/frame_{unixnano}.png`

## Требования

- Go **1.26+**
- **`ffmpeg`** в `PATH`, **`AWS_ACCESS_KEY_ID`** / **`AWS_SECRET_ACCESS_KEY`** (MinIO: `minioadmin` / `minioadmin`)
- Опционально **`.env`** (`ENV_FILE`).

## Конфигурация

| Файл | Назначение |
|------|------------|
| `config.cameras.yaml` | S3/ML/ingest/metrics и список камер (локально) |

| Переменная | Назначение |
|------------|------------|
| `CONFIG_PATH` | Путь к YAML (по умолчанию `config.cameras.yaml`) |
| `ML_BASE_URL` | корень ML-сервиса |
| `ML_PROCESS_PATH` | устаревший единый путь (для совместимости; конвейер по умолчанию использует accident/congestion) |
| `ML_ACCIDENT_PATH` | путь инцидентов (по умолчанию `/v1/process/accident`) |
| `ML_CONGESTION_PATH` | путь загруженности (по умолчанию `/v1/process/congestion`) |
| `KAFKA_BOOTSTRAP_SERVERS` | брокеры; если пусто — публикация в Kafka отключена |
| `KAFKA_TOPIC_VIDEO` | топик метаданных кадра (по умолчанию `its.video.ingest`) |
| `S3_ENDPOINT` | `s3.endpoint` |
| `METRICS_LISTEN_ADDR` | HTTP `/metrics` (по умолчанию `:9091`) |
| `COORDINATOR_BASE_URL` | URL coordinator (`http://127.0.0.1:8098`) |
| `COORDINATOR_ZONE_ID` | зона (например `zone-a`) |
| `COORDINATOR_CLUSTER_ID` | кластер (например `cluster-1`) |
| `COORDINATOR_INSTANCE_ID` | инстанс (например `ingest-a1`) |

## Метрики

- **`router_operation_errors_total{stage}`** — `ffmpeg_start`, `frame_read`, `jpeg_png`, `s3_put`, `ml_process`
- **`router_kafka_video_publish_errors_total{stage}`** — ошибки записи в `its.video.ingest`

Вывод **ffmpeg** в консоль отключён; пока RTSP недоступен, в лог не чаще чем раз в **~45 с на камеру** пишется краткое предупреждение.

Метрики на `http://127.0.0.1:9091/metrics` при локальном запуске. **Prometheus в compose** скрейпит сервис **`router:9091`** (см. [prometheus.yml](../../infra/prometheus/prometheus.yml)).

## Запуск

`go run ./cmd/router`

Нужны `COORDINATOR_BASE_URL` + идентификаторы зоны/кластера/инстанса. При старте без назначений сервис остаётся в standby.

Список камер для инстанса: `GET /v1/assignments?...&data_class=road_segment_video`.

Дальше: **analytics** читает `its.video.ingest` и принимает **два** частичных `POST /v1/ingest` от ML; после склейки — `its.persist.events` → **pusher**.

## Остановка

`Ctrl+C` — graceful shutdown воркеров и metrics-сервера.
