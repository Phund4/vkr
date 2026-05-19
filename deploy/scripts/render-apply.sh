#!/usr/bin/env bash
# Подставляет путь к репо (hostPath) и применяет манифесты.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
K8S_DIR="$(cd "$(dirname "$0")/../k8s" && pwd)"

if ! command -v kubectl >/dev/null; then
  echo "kubectl not found" >&2
  exit 1
fi

CTX="$(kubectl config current-context 2>/dev/null || true)"
echo "kubectl context: ${CTX:-unknown}"

# Job.spec.template неизменяем — перед apply удаляем init jobs (пересоздадутся с новым spec).
if kubectl get namespace traffic >/dev/null 2>&1; then
  kubectl -n traffic delete job kafka-topics-init clickhouse-init minio-init \
    --ignore-not-found --wait=true 2>/dev/null || true
  kubectl -n traffic delete ingress observability --ignore-not-found 2>/dev/null || true
fi

if command -v kustomize >/dev/null; then
  kustomize build --load-restrictor LoadRestrictionsNone "$K8S_DIR" \
    | sed "s|__PIIS_REPO_ROOT__|${ROOT}|g" \
    | kubectl apply -f -
else
  kubectl kustomize --load-restrictor LoadRestrictionsNone "$K8S_DIR" \
    | sed "s|__PIIS_REPO_ROOT__|${ROOT}|g" \
    | kubectl apply -f -
fi

echo "Applied namespace traffic."
