# Деплой piis в локальный Kubernetes (Rancher Desktop)

Rancher Desktop уже даёт кластер и `kubectl`. Деплой через `kubectl` / этот каталог.

## Что входит

| Слой | Компоненты |
|------|------------|
| **infra** | postgres, kafka, clickhouse, minio, mediamtx |
| **jobs** | топики Kafka, схема ClickHouse, bucket MinIO |
| **apps** | coordinator, analytics, pusher, data-service, ml-serving, router, rtsp-generator |
| **observability** | Prometheus, Grafana |

## Предусловия

1. **Rancher Desktop** → Kubernetes **Enabled**.
2. Образы `traffic/*:local`: `make build-images`, при containerd — `make load-images-k8s`.
3. `kubectl config use-context rancher-desktop`

## Быстрый старт

```bash
make -f deploy/Makefile deploy
make -f deploy/Makefile urls
```

## Доступ снаружи (LoadBalancer)

UI и API опубликованы как **LoadBalancer**. На Mac **EXTERNAL-IP** — IP виртуалки Lima (часто `192.168.64.2`), не `127.0.0.1`.

```bash
make -f deploy/Makefile urls
kubectl -n traffic get svc
```

| Сервис | Порт | Назначение |
|--------|------|------------|
| **postgres** | 5432 | JDBC: `jdbc:postgresql://<IP>:5432/coordinator?user=coordinator&password=coordinator&sslmode=disable` |
| **clickhouse** | 8123 / 9000 | HTTP/JDBC / нативный протокол |
| **minio** | 9000 / 9001 | S3 API / веб-консоль (`minioadmin` / `minioadmin`) |
| **data-service** | 8610 / 8611 | REST API / metrics |
| **rtsp-generator** | 8096 | RTSP Studio |
| **coordinator** | 8098 | API назначений камер |
| **ml-serving** | 8000 | ML API |
| **grafana** | 3000 | Дашборды (`admin` / `admin`) |
| **prometheus** | 9090 | `/targets` |
| **mediamtx** | 8554 | RTSP `rtsp://<IP>:8554/cam-01` |

Внутри кластера router по-прежнему использует `rtsp://mediamtx:8554/...`.

**Rancher Desktop:** Kubernetes → Services → namespace `traffic` → колонка **EXTERNAL-IP**.

Видео для генератора: `<repo>/.data/videos/*.mp4`.

## Запуск потоков данных

1. `http://<EXTERNAL-IP>:8096` — RTSP Studio, старт потоков `cam-01` … `cam-04`.
2. Router читает назначения с coordinator (`http://<EXTERNAL-IP>:8098/v1/assignments?...`).
3. Метрики — Grafana `http://<EXTERNAL-IP>:3000`.

Логи: `kubectl -n traffic logs -f deploy/router`

## Kafka не стартует

```bash
kubectl -n traffic scale deployment kafka --replicas=0
kubectl -n traffic delete pvc kafka-data
make -f deploy/Makefile apply
kubectl -n traffic scale deployment kafka --replicas=1
```

## Удаление

```bash
make -f deploy/Makefile delete
```
