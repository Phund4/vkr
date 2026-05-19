#!/usr/bin/env bash
# Импорт traffic/*:local из Docker в containerd namespace k8s.io (Rancher Desktop с containerd).
set -euo pipefail

IMAGES=(coordinator analytics pusher data-service router ml-serving rtsp-generator)

if ! command -v docker >/dev/null; then
  echo "docker not found" >&2
  exit 1
fi

if ! command -v nerdctl >/dev/null; then
  echo "nerdctl not found."
  echo "Rancher Desktop → Settings → Container Engine → dockerd (moby),"
  echo "затем перезапустите RD и снова: make build-images apply"
  exit 1
fi

NS="${CONTAINERD_NAMESPACE:-k8s.io}"
TMPDIR="${TMPDIR:-/tmp}"
trap 'rm -f "$TMPDIR"/traffic-k8s-load-*.tar' EXIT

for img in "${IMAGES[@]}"; do
  ref="traffic/${img}:local"
  if ! docker image inspect "$ref" >/dev/null 2>&1; then
    echo "missing image $ref — run: make build-images" >&2
    exit 1
  fi
  tar="${TMPDIR}/traffic-k8s-load-${img}.tar"
  echo "==> docker save $ref → nerdctl -n $NS load -i"
  docker save "$ref" -o "$tar"
  nerdctl --namespace "$NS" load -i "$tar"
  rm -f "$tar"
done

echo "Loaded into containerd namespace $NS:"
nerdctl --namespace "$NS" images | grep traffic || true
