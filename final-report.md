### Финальная интеграция ML-сервиса
**Тема**: инференс, связка компонентов, мониторинг, контейнеризация

**Автор**: Попов Александр Иванович  
**Группа**: БВТ2203

Распределённая ИТС: приём видео, ML-инференс, аналитика в ClickHouse, мониторинг.

---

### Шаг 1. Разработка инференс-модуля

Реализован отдельный микросервис **`services/ml-serving`** (FastAPI, Uvicorn):

- **`POST /v1/process`** — приём кадра (multipart), инференс моделей **ДТП (классификация)** и **загруженности (регрессия)**; при настроенном **`ANALYTICS_INGEST_URL`** — передача результата в **analytics** (`POST /v1/ingest`), иначе — только JSON-ответ клиенту.
- **`GET /health`** — статус загрузки моделей, пути к чекпойнтам, флаг использования **`winners.json`**.
- Логика инференса вынесена в **`inference_core.py`**.

Выбор весов в рантайме:

- явные переменные **`ACCIDENT_CKPT`**, **`CONGESTION_CKPT`**, либо автоматическое чтение из **`models/winners.json`** внутри **`services/ml-serving`** (каталоги **`models/`**, **`artifacts/`** в образе).

Упаковка для рантайма: **`services/ml-serving/Dockerfile`**, переменная **`SERVING_ROOT`** для путей к артефактам внутри контейнера.

---

### Шаг 2. Интеграция компонентов

**Цепочка данных:**

1. Источники: `video-source-sim` (RTSP через MediaMTX, профиль `ingest`).
2. **router** — RTSP, S3 (MinIO), вызовы **ml-serving**.
3. **coordinator** — назначение источников инстансам, состояние в **PostgreSQL**.
4. **ml-serving** — инференс; при необходимости прямой **POST** в **analytics** и/или отдельный продюсер в Kafka (топик видео), если он включён в контуре.
5. **analytics** — HTTP ingest и/или Kafka consumer, запись в **ClickHouse**, бизнес-метрики Prometheus.

**Документация по архитектуре:** краткий запуск и мониторинг — **`infra/README.md`** и README сервисов в **`services/`**. Высокоуровневое описание — **`high-level-design.md`**. При расхождениях итоговой системы с этим планом правки отражаются в этих файлах и в **`infra/`**.

---

### Шаг 3. Внедрение базового мониторинга

| Механизм | Реализация |
|----------|------------|
| Логирование запросов и событий | Логирование в приложениях (например, **`ml-serving`**, предупреждения при отсутствии **`ANALYTICS_INGEST_URL`**); у сервисов на Go — логи в stdout контейнеров. |
| Метрики доступности | **`GET /health`** у сервисов; **blackbox-exporter** опрашивает HTTP health **coordinator** и **ml-serving** (**`infra/prometheus/prometheus.yml`**). |
| Нагрузка на систему | **cAdvisor** — CPU/RAM контейнеров; в Grafana дашборд **«Сервисы»** — ряды по имени контейнера (`name`), т.к. лейблы Compose в метриках cAdvisor на Docker Desktop часто недоступны. |
| Экспорт метрик | **`/metrics`** (формат Prometheus): **router**, **analytics**, **coordinator**, **ml-serving**; сбор — **`infra/prometheus/prometheus.yml`**, TSDB в томе, визуализация — **Grafana** (`infra/grafana/provisioning/`). |

Дашборды по Kafka, трафику и аналитике — JSON в **`infra/grafana/provisioning/dashboards/json/`**.

---

### Шаг 4. Контейнеризация и оркестрация

**Dockerfile по компонентам:**

| Компонент | Путь |
|-----------|------|
| coordinator | `services/coordinator/Dockerfile` |
| router | `services/router/Dockerfile` |
| ml-serving | `services/ml-serving/Dockerfile` |
| analytics | `services/analytics/Dockerfile` |
| video-source-sim | `infra/video-source-sim/Dockerfile` |

**Оркестрация:** единый **`infra/docker-compose.yml`**: сеть **`traffic-its`**, именованные тома (PostgreSQL, Kafka KRaft, ClickHouse, MinIO, Prometheus, Grafana и т.д.), переменные окружения, healthcheck'и, init топиков Kafka, профиль **`ingest`** по необходимости.

Запуск стека из каталога `infra`:

```bash
docker compose up --build -d
```

Минимальные требования: **Docker** с **Compose V2** (`docker compose`). Подробности — **`infra/README.md`**; быстрый старт и таблицы — **`README.md`**, детальная схема — **`readme.md`**.

*Примечание:* **Elasticsearch/Kibana** в compose закомментированы; стек собирается без них.

---

### Шаг 5. Демонстрация и финальная документация

**Тестовые данные:**

- видео для пайплайна: **`.data/videos/*.mp4`** (см. **`infra/README.md`**, профиль **`ingest`**);
- веса моделей для демонстрации — положить **`best.pt`** в пути из **`services/ml-serving/models/winners.json`** (см. **`services/ml-serving/models/README.md`**).

**Демонстрация:** запуск `cd infra && docker compose up --build -d`; проверка ML — `curl http://localhost:8000/health` (порт **ml-serving** при пробросе из compose); UI — **Grafana** http://localhost:3000; остальные порты — в **`infra/README.md`**. Опционально: скриншоты дашбордов, схемы из **`readme.md`** / **`README.md`**, сопроводительное видео.

---

### Ссылки на документацию

| Содержание | Файл |
|------------|------|
| Установка и запуск (Docker), стек наблюдаемости | `infra/README.md` |
| Архитектура (высокий уровень) | `high-level-design.md` |
| Веса и winners.json | `services/ml-serving/models/README.md` |
| Инференс-сервис | `services/ml-serving/README.md` |
| Видео-контур (router) | `services/router/README.md` |
