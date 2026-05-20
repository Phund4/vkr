#!/usr/bin/env bash
# Применяет манифесты из deploy/k8s/customization.yaml (без Kustomize).
# При изменении списка resources/configMaps — обновите и customization.yaml, и массивы ниже.
set -euo pipefail

ROOT="${1:?repo root}"
K8S_DIR="${2:?k8s dir}"
NS=traffic

apply_configmap() {
  local name=$1
  shift
  kubectl create configmap "$name" --namespace="$NS" --dry-run=client -o yaml "$@" | kubectl apply -f -
}

apply_manifest() {
  local rel=$1
  sed "s|__PIIS_REPO_ROOT__|${ROOT}|g" "${K8S_DIR}/${rel}" | kubectl apply -n "$NS" -f -
}

apply_configmap postgres-init \
  --from-file="${ROOT}/infra/postgres/init/001_coordinator_schema.sql" \
  --from-file="${ROOT}/infra/postgres/init/002_coordinator_seed.sql"

apply_configmap router-config \
  --from-file=config.cameras.docker.yaml="${ROOT}/services/router/config.cameras.docker.yaml"

apply_configmap grafana-dashboards-json \
  --from-file=traffic-services.json="${ROOT}/infra/grafana/provisioning/dashboards/json/traffic-services.json" \
  --from-file=traffic-results.json="${ROOT}/infra/grafana/provisioning/dashboards/json/traffic-results.json"

RESOURCES=(
  namespace.yaml
  secret.yaml
  config/clickhouse-init-configmap.yaml
  infra/postgres.yaml
  infra/kafka.yaml
  infra/clickhouse.yaml
  infra/minio.yaml
  infra/mediamtx.yaml
  jobs/kafka-topics-init.yaml
  jobs/clickhouse-init.yaml
  jobs/minio-init.yaml
  apps/coordinator.yaml
  apps/analytics.yaml
  apps/pusher.yaml
  apps/data-service.yaml
  apps/ml-serving.yaml
  apps/router.yaml
  apps/rtsp-generator.yaml
  observability/prometheus-rbac.yaml
  observability/prometheus-configmap.yaml
  observability/prometheus-deployment.yaml
  observability/grafana-datasources-configmap.yaml
  observability/grafana-dashboards-configmap.yaml
  observability/grafana-pvc.yaml
  observability/grafana.yaml
)

for rel in "${RESOURCES[@]}"; do
  apply_manifest "$rel"
done
