#!/usr/bin/env bash
set -euo pipefail

root_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
compose=(docker compose --project-name tollgate --file "$root_dir/deploy/compose/compose.yaml")
command=${1:-up}

case "$command" in
  up)
    profile=${PROFILE:-local}
    if [[ "$profile" != local ]]; then echo "unknown profile: $profile" >&2; exit 2; fi
    "${compose[@]}" up --build --detach
    deadline=$((SECONDS + 300))
    until curl --fail --silent http://127.0.0.1:8081/readyz >/dev/null; do
      if (( SECONDS >= deadline )); then "${compose[@]}" ps; exit 1; fi
      sleep 2
    done
    echo "Tollgate local stack is ready"
    echo "API:      http://127.0.0.1:8080"
    echo "Metrics:  http://127.0.0.1:8081/metrics"
    echo "Grafana:  http://127.0.0.1:3000"
    echo "Example client key (not validated until feature 005): tg-example-local"
    echo "Example upstream key (fake): tg-example-upstream"
    ;;
  down)
    "${compose[@]}" down --volumes --remove-orphans
    if [[ -n $(docker ps -aq --filter label=com.docker.compose.project=tollgate) ]]; then echo "tollgate containers remain" >&2; exit 1; fi
    if [[ -n $(docker network ls -q --filter label=com.docker.compose.project=tollgate) ]]; then echo "tollgate networks remain" >&2; exit 1; fi
    if [[ -n $(docker volume ls -q --filter label=com.docker.compose.project=tollgate) ]]; then echo "tollgate volumes remain" >&2; exit 1; fi
    ;;
  *) echo "usage: $0 {up|down}" >&2; exit 2 ;;
esac
