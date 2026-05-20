# Локальная инфраструктура

Сеть Docker: **`traffic-its`**. Имя проекта Compose: **`traffic-infra`**.

Команды из каталога `infra/` — через **`make`** (см. `make help`) или `docker compose` с профилями.

## Профили

| Профиль | Содержимое |
|---------|------------|
| **`infra`** | Postgres, Kafka, ClickHouse, MinIO, MediaMTX, rtsp-generator — то же ядро, что в `deploy/k8s`. |
| **`observability`** | ELK, Prometheus, Grafana, экспортёры, cAdvisor, blackbox — в K8s не разворачивается. |
| **`apps`** | coordinator, analytics, pusher, data-service, ml-serving, router из `../services/`. |

```bash
cd infra
make infra-up              # только ядро
make observability-up      # ELK + Prometheus (таргеты compose; apps в compose)
make observability-k8s-up  # то же, но scrape подов из kubectl namespace traffic
make apps-up               # приложения + зависимости infra
make stack-up              # infra + apps (как K8s, без observability)
make stack-full-up         # infra + observability + apps
```

Остановка: `make infra-down`, `make observability-down`, `make stack-down`.

## Запуск вручную (compose)

```bash
cd infra
docker compose pull
docker compose --profile infra up -d
docker compose --profile observability up -d   # ELK + Prometheus + Grafana
```

```bash
docker compose --profile infra --profile observability --profile apps down
```

### MediaMTX и rtsp-generator (профиль `infra`)

RTSP и публикация тестовых потоков входят в **`infra`**:

```bash
docker compose --profile infra up -d mediamtx rtsp-generator
docker compose --profile infra stop mediamtx rtsp-generator
```

Данные **PostgreSQL**, **ClickHouse**, **Kafka**, **Elasticsearch**, **MinIO**, **Prometheus** (TSDB) и **Grafana** хранятся в именованных томах. Конфиг Prometheus — [`prometheus/prometheus.yml`](prometheus/prometheus.yml).

## Сервисы и подключение

- **Kafka** (KRaft) — с хоста: `localhost:9092`; в сети: `kafka:9092`.

- **Elasticsearch** (профиль `observability`) — http://localhost:9200. В сети: `elasticsearch:9200`.

- **Logstash** — Beats **5044**, API **9600**. Пайплайн: [`logstash/pipeline/logstash.conf`](logstash/pipeline/logstash.conf) → индексы **`traffic-docker-logs-YYYY.MM.DD`**.

- **Filebeat** — JSON-логи контейнеров Docker → Logstash ([`filebeat/filebeat.yml`](filebeat/filebeat.yml)). В Kibana — data view **`traffic-docker-logs-*`**.

- **Kibana** — http://localhost:5601.

- **ClickHouse** — HTTP `http://localhost:8123`, нативный `localhost:9000`. Пользователь **`default`**, пароль пустой (dev).

  **Имитационный справочник** (`its_infra_sim`): после `make infra-up` — `infra/clickhouse/bootstrap.sh` или one-shot `clickhouse-infra-sim-seed`.

- **MinIO** — API http://localhost:9050, консоль http://localhost:9051 (`minioadmin` / `minioadmin`).

- **Prometheus** — http://localhost:9090 ([`prometheus/prometheus.yml`](prometheus/prometheus.yml)).

- **Grafana** — http://localhost:3000 (`admin` / `admin`). Дашборды в [`grafana/provisioning/`](grafana/provisioning/).

- **MediaMTX** — `rtsp://localhost:8554`.

- **rtsp-generator** — http://localhost:8096 ([`rtsp-generator/README.md`](rtsp-generator/README.md)).

При старте `infra` выполняются one-shot: `clickhouse-init`, `minio-init`, `kafka-topics-init`, `clickhouse-infra-sim-seed`.

Тома: `postgres-data`, `kafka-data`, `es-data`, `clickhouse-data`, `minio-data`, `prometheus-data`, `grafana-data`.
