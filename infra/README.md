# Локальная инфраструктура

Сеть Docker: **`traffic-its`**. Имя проекта Compose: **`traffic-infra`**.

## Запуск и остановка

```bash
cd infra
docker compose pull
docker compose up -d
```

```bash
docker compose down
```

### Профиль `ingest` (MediaMTX + видео-симулятор)

Поднять только RTSP и публикацию тестовых потоков:

```bash
docker compose --profile ingest up -d mediamtx rtsp-generator
```

**Остановить** эти контейнеры, не трогая остальной стек (Kafka, ClickHouse и т.д.):

```bash
docker compose --profile ingest stop mediamtx rtsp-generator
```

Снова запустить:

```bash
docker compose --profile ingest start mediamtx rtsp-generator
```

Удалить контейнеры профиля `ingest` (данные в томах основного стека не затрагиваются):

```bash
docker compose --profile ingest rm -sf mediamtx rtsp-generator
```

Данные **PostgreSQL**, **ClickHouse**, **Kafka**, **Elasticsearch**, **MinIO**, **Prometheus** (TSDB) и **Grafana** (в том числе дашборды и настройки, созданные в UI) хранятся в именованных томах и переживают перезапуск контейнеров. Конфиг Prometheus — [`prometheus/prometheus.yml`](prometheus/prometheus.yml); после правок: `docker compose restart prometheus` или lifecycle reload.

## Сервисы и подключение

- **Kafka** (KRaft, без ZooKeeper) — с хоста: `localhost:9092`; внутри сети: `kafka:9092`. Пример: `export KAFKA_BOOTSTRAP_SERVERS=localhost:9092`.

- **Elasticsearch** — http://localhost:9200 (без логина, dev). Внутри сети: `elasticsearch:9200`.

- **Logstash** — приём от Filebeat (Beats) на **5044**, HTTP API на **9600**. Пайплайн: [`logstash/pipeline/logstash.conf`](logstash/pipeline/logstash.conf) → индексы **`traffic-docker-logs-YYYY.MM.DD`** (префикс не `logstash-*`: в ES 8 шаблоны Elastic для `logstash-*` ожидают **data stream** и дают 400 на обычный `index` из Logstash).

- **Filebeat** — читает JSON-логи контейнеров Docker и отправляет в Logstash (см. [`filebeat/filebeat.yml`](filebeat/filebeat.yml)): **filestream** по путям `/var/lib/docker/containers/*/*-json.log`, `prospector.scanner.fingerprint.enabled: false`, `compression_level: 0` к Logstash. В `docker-compose.yml` задан явный `command` (`filebeat -e --strict.perms=false -c …`). В Kibana — **Data view** **`traffic-docker-logs-*`**, время **`@timestamp`**. Запасной обход Logstash: [`filebeat/filebeat.direct-es.yml`](filebeat/filebeat.direct-es.yml).

- **Kibana** — http://localhost:5601 (логи: Discover → data view **`traffic-docker-logs-*`**).

- **ClickHouse** — HTTP с хоста: `http://localhost:8123`; нативный протокол: `localhost:9000`. Внутри сети: `clickhouse:8123`, `clickhouse:9000`. Пользователь **`default`**, пароль пустой (только dev).

  **JDBC** (драйвер `com.clickhouse.jdbc.ClickHouseDriver`, интерфейс HTTP на порту 8123):
  - с хоста: `jdbc:clickhouse://localhost:8123/default`
  - из контейнера в сети `traffic-its`: `jdbc:clickhouse://clickhouse:8123/default`

  Пример `clickhouse-client` с хоста: `clickhouse-client --host localhost --port 9000`.

  **Имитационный справочник** (отдельная БД `its_infra_sim`, не `default`): таблицы `municipalities`, `bus_stops`, `bus_stop_routes`. После `docker compose up -d clickhouse` выполните `infra/clickhouse/bootstrap.sh` (нужен `clickhouse-client` на хосте или Docker compose). Скрипты: [`clickhouse/001_schema.sql`](clickhouse/001_schema.sql), [`clickhouse/002_seed.sql`](clickhouse/002_seed.sql). Проверка: `clickhouse-client -q "SELECT count() FROM its_infra_sim.municipalities"` и `SELECT count() FROM its_infra_sim.bus_stops`.

- **MinIO (S3)** — API с хоста: `http://localhost:9050`; консоль: http://localhost:9051. Учётные данные: **`minioadmin` / `minioadmin`**. Внутри сети: endpoint `http://minio:9000`.

- **Prometheus** — http://localhost:9090. Конфиг [`prometheus/prometheus.yml`](prometheus/prometheus.yml): scrape **`/metrics`** у прикладных сервисов в сети `traffic-its` (`router`, `analytics`, `ml-serving`, `coordinator`), экспортёры **Elasticsearch**, **PostgreSQL**, **Kafka**, **ClickHouse**, **`cadvisor`**, **`blackbox-exporter`** (HTTP health `coordinator` / `ml-serving`). Часть таргетов может быть DOWN, если сервис не запущен.

- **Grafana** — http://localhost:3000, логин по умолчанию **`admin` / `admin`**. Провижининг из [`grafana/provisioning/`](grafana/provisioning/) (папка дашбордов **Traffic**); том **`grafana-data`**. Дашборды: **`Services`** (`traffic-services`) — конвейер, Kafka, **единая панель ошибок** (метка `source`), data-service RPS, cAdvisor по `container_label_com_docker_compose_service`; **`Results (данные конвейера)`** (`traffic-results`) — загруженность, инциденты, запись в CH/Kafka.

- **MediaMTX** (профиль `ingest`) — `rtsp://localhost:8554`. Запуск: `docker compose --profile ingest up -d --build mediamtx rtsp-generator`.

- **rtsp-generator** (RTSP Studio) — веб-UI **http://localhost:8096**, API для запуска/остановки RTSP-потоков из `../.data/videos/*.mp4` или синтетики; не связан с coordinator (см. [`rtsp-generator/README.md`](rtsp-generator/README.md)).

Профиль **`ingest`** (MediaMTX + видео-симулятор). Пример:

```bash
docker compose --profile ingest up -d mediamtx rtsp-generator
```

## Приложения вне compose

**analytics**, **router**, **ml-serving** при необходимости запускаются вручную (см. `.env` в каталогах сервисов): [`services/analytics`](../services/analytics/README.md), [`services/router`](../services/router/README.md), [`services/ml-serving`](../services/ml-serving/README.md). В compose также поднимаются **coordinator** и **router** для полного контура. Для видео: **router** → **ml-serving** → **analytics** → **ClickHouse** (и метрики); при использовании Kafka события могут дублироваться через топик **`its.video.ingest`**.

При старте compose автоматически выполняются one-shot инициализаторы:
- `clickhouse-init` — создаёт таблицы `default.road_incidents` и `default.road_congestion`;
- `minio-init` — создаёт бакет `its-frames`.

Тома данных Compose (список имён): `postgres-data`, `zookeeper-data`, `kafka-data`, `es-data`, `clickhouse-data`, `minio-data`, `prometheus-data`, `grafana-data`.
