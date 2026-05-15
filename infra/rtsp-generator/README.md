# RTSP Studio (`rtsp-generator`)

Веб-UI и HTTP API для **динамической** публикации RTSP-потоков в **MediaMTX**: выбор `.mp4` из каталога, зацикливание, синтетические источники (lavfi). **Не связан с coordinator** — RTSP URL вы прописываете в своём Postgres / конфиге router самостоятельно.

## Запуск локально

```bash
export RTSP_PUBLISH_BASE=rtsp://127.0.0.1:8554
export SIM_VIDEO_DIR=/path/to/mp4
go run ./cmd/rtsp-generator
```

Откройте http://127.0.0.1:8096

## Docker Compose

Сервис `rtsp-generator` в `infra/docker-compose.yml`: порт **8096**, том `../.data/videos` → `/videos`.

## API

| Метод | Путь | Описание |
|--------|------|----------|
| GET | `/health` | Liveness |
| GET | `/api/config` | Базовый RTSP URL, каталог видео |
| GET | `/api/videos` | Список `*.mp4` |
| GET | `/api/streams` | Активные потоки |
| POST | `/api/streams` | Старт (JSON см. ниже) |
| DELETE | `/api/streams/:id` | Остановка |

### Тело POST `/api/streams`

```json
{
  "stream_id": "cam-east-01",
  "file": "clip.mp4",
  "loop": true,
  "realtime": true,
  "synthetic": false
}
```

Синтетика:

```json
{
  "stream_id": "syn-1",
  "synthetic": true,
  "synthetic_kind": "testsrc2",
  "realtime": true
}
```

Итоговый URL для router: `{RTSP_PUBLISH_BASE}/{stream_id}`.

## Переменные окружения

- `LISTEN_ADDR` — `:8096`
- `RTSP_PUBLISH_BASE` — `rtsp://mediamtx:8554`
- `SIM_VIDEO_DIR` — `/videos`
- `SIM_FPS`, `SIM_SIZE` — для синтетики
