#!/usr/bin/env bash
# Подставляет путь к репо (hostPath) и применяет манифесты из customization.yaml.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
K8S_DIR="$(cd "$(dirname "$0")/../k8s" && pwd)"

if ! command -v kubectl >/dev/null; then
  echo "kubectl not found" >&2
  exit 1
fi

CTX="$(kubectl config current-context 2>/dev/null || true)"
echo "kubectl context: ${CTX:-unknown}"

if kubectl get namespace traffic >/dev/null 2>&1; then
  kubectl -n traffic delete job kafka-topics-init clickhouse-init minio-init \
    --ignore-not-found --wait=true 2>/dev/null || true
  kubectl -n traffic delete ingress observability --ignore-not-found 2>/dev/null || true
fi

bash "$(dirname "$0")/apply-customization.sh" "$ROOT" "$K8S_DIR"

echo "Applied namespace traffic (customization.yaml)."
