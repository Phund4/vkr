#!/usr/bin/env bash
# Сборка локальных образов traffic/*:local для Rancher Desktop (dockerd).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"

build() {
  local name=$1 ctx=$2 df=${3:-Dockerfile}
  echo "==> traffic/${name}:local"
  docker build -t "traffic/${name}:local" -f "${ctx}/${df}" "${ctx}"
}

build coordinator services/coordinator
build analytics   services/analytics
build pusher      services/pusher
build data-service services/data-service
build router      services/router
build ml-serving  services/ml-serving
build rtsp-generator infra/rtsp-generator

echo "Done. Images:"
docker images 'traffic/*' --format '  {{.Repository}}:{{.Tag}}'

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
if [ "${LOAD_INTO_K8S:-1}" = "1" ]; then
  bash "$SCRIPT_DIR/load-images-k8s.sh" || true
fi
