# Модели для ml-serving

- **`winners.json`** — пути к чекпойнтам относительно каталога сервиса (`SERVING_ROOT`, в Docker: `/app`).
- В **docker compose** каталог **`../.data/artifacts`** монтируется в **`/app/artifacts`**: фактические `best.pt` подкладываются из корня репозитория командой **`make -f infra/Makefile sync-ml-artifacts`** (копирует `artifacts/accident/baseline-cnn/best.pt` и `artifacts/congestion/tiny-cnn/best.pt` в `.data/artifacts/...`). Каталог `.data/` в `.gitignore`.
- Локально без compose: положите веса в **`services/ml-serving/artifacts/...`** как в `winners.json`, либо задайте **`ACCIDENT_CKPT`** и **`CONGESTION_CKPT`** (абсолютный путь или относительно `SERVING_ROOT`).

Без файлов `*.pt` сервис стартует, но Kafka-воркеры не смогут выполнять инференс до появления весов.
