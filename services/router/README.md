# router (видео-поток → S3 → ML)

Сервис для **видео-контура**: RTSP → кадры в **S3** → **ML**. Какие камеры обрабатывать, задаёт **coordinator** через назначения источников (`data_class=road_segment_video`).

**Камеры:** сегменты, `camera_id` и **`rtsp_url`** приходят из назначений. Если процесс **router** в той же Docker-сети, что и MediaMTX, в `rtsp_url` укажите хост **`mediamtx`** (например `rtsp://mediamtx:8554/cam-01`).

Потоки для примера в `config.cameras.yaml`: профиль **`ingest`** в **`infra`** (`video-source-sim`).

Ключи S3: `{prefix}/{YYYY-MM-DD}/{camera_id}/frame_{unixnano}.png`

## Требования

- Go **1.22+**
- **`ffmpeg`** в `PATH`, **`AWS_ACCESS_KEY_ID`** / **`AWS_SECRET_ACCESS_KEY`** (MinIO: `minioadmin` / `minioadmin`)
- Опционально **`.env`** (`ENV_FILE`).

## Конфигурация

| Файл | Назначение |
|------|------------|
| `config.cameras.yaml` | S3/ML/ingest/metrics и список камер (локально) |

| Переменная | Назначение |
|------------|------------|
| `CONFIG_PATH` | Путь к YAML (по умолчанию `config.cameras.yaml`) |
| `ML_BASE_URL` | переопределение `ml.base_url` |
| `ML_PROCESS_PATH` | `ml.process_path` |
| `S3_ENDPOINT` | `s3.endpoint` |
| `METRICS_LISTEN_ADDR` | HTTP `/metrics` (по умолчанию `:9091`) |
| `COORDINATOR_BASE_URL` | URL coordinator (`http://127.0.0.1:8098`) |
| `COORDINATOR_ZONE_ID` | зона (например `zone-a`) |
| `COORDINATOR_CLUSTER_ID` | кластер (например `cluster-1`) |
| `COORDINATOR_INSTANCE_ID` | инстанс (например `ingest-a1`) |

## Метрики

- **`router_operation_errors_total{stage}`** — `ffmpeg_start`, `frame_read`, `s3_put`, `ml_process`

Вывод **ffmpeg** в консоль отключён; пока RTSP недоступен, в лог не чаще чем раз в **~45 с на камеру** пишется краткое предупреждение.

Метрики на `http://127.0.0.1:9091/metrics` при локальном запуске. **Prometheus в compose** скрейпит сервис **`router:9091`** (см. [prometheus.yml](../../infra/prometheus/prometheus.yml)).

## Запуск

`go run ./cmd/router`

Нужны `COORDINATOR_BASE_URL` + идентификаторы зоны/кластера/инстанса. При старте без назначений сервис остаётся в standby.

Список камер для инстанса: `GET /v1/assignments?...&data_class=road_segment_video`.

Дальше по цепочке: **ml-serving** → **analytics** (`POST /v1/ingest` и при необходимости Kafka `its.video.ingest`).

## Остановка

`Ctrl+C` — graceful shutdown воркеров и metrics-сервера.
