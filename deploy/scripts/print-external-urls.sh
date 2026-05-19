#!/usr/bin/env bash
# Печать URL сервисов с type=LoadBalancer (EXTERNAL-IP на Rancher Desktop — IP VM, напр. 192.168.64.2).
set -euo pipefail

NS="${1:-traffic}"

echo "LoadBalancer (namespace ${NS}):"
kubectl -n "$NS" get svc -o json \
  | python3 -c "
import json, sys
data = json.load(sys.stdin)
rows = []
for s in data.get('items', []):
    if s.get('spec', {}).get('type') != 'LoadBalancer':
        continue
    name = s['metadata']['name']
    ing = (s.get('status', {}).get('loadBalancer') or {}).get('ingress') or []
    ip = (ing[0].get('ip') if ing else None) or (ing[0].get('hostname') if ing else None) or '<pending>'
    for p in s.get('spec', {}).get('ports') or []:
        port = p.get('port')
        pname = p.get('name') or str(port)
        if name == 'mediamtx' and port == 8554:
            print(f'  {name} ({pname}): rtsp://{ip}:{port}')
        elif name == 'postgres' and port == 5432:
            print(f'  {name}: jdbc:postgresql://{ip}:{port}/coordinator?user=coordinator&password=coordinator&sslmode=disable')
        elif name == 'clickhouse' and port == 8123:
            print(f'  {name} (HTTP/JDBC): jdbc:clickhouse://{ip}:{port}/default')
        elif name == 'clickhouse' and port == 9000:
            print(f'  {name} (native): {ip}:{port}')
        elif name == 'minio' and port == 9000:
            print(f'  {name} (S3 API): http://{ip}:{port}  (minioadmin / minioadmin)')
        elif name == 'minio' and port == 9001:
            print(f'  {name} (console): http://{ip}:{port}')
        else:
            scheme = 'https' if port == 443 else 'http'
            print(f'  {name} ({pname}): {scheme}://{ip}:{port}')
"
