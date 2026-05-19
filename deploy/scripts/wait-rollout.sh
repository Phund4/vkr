#!/usr/bin/env bash
set -euo pipefail

NS=traffic

wait_deploy() {
  local name=$1
  local timeout=${2:-300s}
  echo "  → deployment/$name"
  kubectl -n "$NS" rollout status "deployment/$name" --timeout="$timeout"
}

wait_job() {
  local name=$1
  local timeout=${2:-240s}
  if ! kubectl -n "$NS" get job "$name" >/dev/null 2>&1; then
    echo "  ⚠ job/$name не найден — сначала выполните: make apply"
    return 0
  fi
  echo "  → job/$name"
  kubectl -n "$NS" wait --for=condition=complete "job/$name" --timeout="$timeout"
}

echo "Waiting for infra..."
wait_deploy postgres
echo "  → deployment/kafka (может занять 2–4 мин)"
if ! wait_deploy kafka 420s; then
  echo ""
  echo "Kafka не поднялся. Диагностика:"
  echo "  kubectl -n $NS logs deploy/kafka --tail=100"
  echo "Сброс тома:"
  echo "  kubectl -n $NS scale deployment kafka --replicas=0"
  echo "  kubectl -n $NS delete pvc kafka-data"
  echo "  kubectl -n $NS scale deployment kafka --replicas=1"
  exit 1
fi
wait_deploy clickhouse
wait_deploy minio
wait_deploy mediamtx

echo "Waiting for init jobs (создаются через make apply / kustomize)..."
wait_job kafka-topics-init 240s
wait_job clickhouse-init 120s
wait_job minio-init 120s

echo "Waiting for apps..."
for d in coordinator analytics ml-serving pusher data-service rtsp-generator router; do
  wait_deploy "$d" 300s || true
done

echo "Waiting for observability..."
wait_deploy prometheus 180s || true
wait_deploy grafana 180s || true

kubectl -n "$NS" get pods
echo ""
bash "$(dirname "$0")/print-external-urls.sh" "$NS" || true
echo "  Grafana: admin / admin"
echo "  RTSP Studio: потоки cam-01 … cam-04 → rtsp://mediamtx:8554/<id> (внутри кластера)"
