# Сервис ml-serving

Сервис инференса ML в рантайме (эндпоинт `/v1/process`).

- Чекпойнты и **`winners.json`** лежат **в этом сервисе**: каталоги **`models/`** и **`artifacts/`** (см. **`models/README.md`**). В Docker образ копируется содержимое этих каталогов; для работы нужны реальные файлы **`.pt`** по путям из `winners.json`.
- Если **`ACCIDENT_CKPT` / `CONGESTION_CKPT` не заданы**, пути к весам берутся из **`WINNERS_JSON`** (по умолчанию **`models/winners.json`** относительно **`SERVING_ROOT`** / каталога сервиса). Явные переменные в `.env` имеют приоритет.
- При заданном **`ANALYTICS_INGEST_URL`** (полный URL, например `http://analytics:8093/v1/ingest`) результат инференса дополнительно отправляется в **analytics** (запись в ClickHouse и метрики).

## Запуск

```bash
cd services/ml-serving
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
uvicorn api.main:app --host 0.0.0.0 --port 8000 --no-access-log
```

Настройки — в **`services/ml-serving/.env`** (или переменные окружения). Запускайте из каталога сервиса или задайте **`SERVING_ROOT`** и абсолютные пути к чекпойнтам и к `WINNERS_JSON`.

## Проверка

```bash
curl -s http://127.0.0.1:8000/health
```

В ответе: `accident_checkpoint`, `congestion_checkpoint`, `winners_json`, флаги `*_from_winners_json`.
