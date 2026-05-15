# router и coordinator: новый инстанс

Идентичность router (`COORDINATOR_ZONE_ID`, `COORDINATOR_CLUSTER_ID`, `COORDINATOR_INSTANCE_ID`) должна совпадать с записью в PostgreSQL, таблица **`ingestion_instances`** (зона, кластер, инстанс), см. seed и миграции в **`infra/postgres/init/`**.

Правило: для каждого процесса router есть строка в **`ingestion_instances`** с тем же `zone_id`, `cluster_id`, `instance_id`. Поле **`url`** — опционально (метрики и т.п.).

Список отдаёт coordinator: `GET /v1/ingestion_instances?zone_id=zone-a`.

## Обязательные переменные

| Переменная | Пример |
|------------|--------|
| `COORDINATOR_BASE_URL` | `http://127.0.0.1:8098` |
| `COORDINATOR_ZONE_ID` | `zone-a` |
| `COORDINATOR_CLUSTER_ID` | `cluster-1` |
| `COORDINATOR_INSTANCE_ID` | `ingest-a1` |

## Чеклист

1. В PostgreSQL добавлена строка в **`ingestion_instances`** для зоны/кластера/инстанса.
2. Перезапущен coordinator (данные уже в БД).
3. У процесса router заданы те же `COORDINATOR_*`.

Два процесса на одном хосте — разные `COORDINATOR_INSTANCE_ID` и порты **`METRICS_LISTEN_ADDR`**.

## См. также

- `services/coordinator/README.md`
