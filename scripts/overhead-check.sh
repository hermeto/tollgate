#!/usr/bin/env bash
set -euo pipefail

prometheus=${TOLLGATE_PROMETHEUS_URL:-http://127.0.0.1:9090}
query='histogram_quantile(0.99, sum(rate(tollgate_gateway_overhead_seconds_bucket[1m])) by (le))'
value=$(curl --fail --silent --get --data-urlencode "query=$query" "$prometheus/api/v1/query" | jq -r '.data.result[0].value[1] // empty')
if [[ -z "$value" ]]; then echo "no gateway overhead samples found" >&2; exit 1; fi
awk -v value="$value" 'BEGIN { if (value > 0.001) { printf "gateway overhead p99 %.6fs exceeds 1ms\n", value; exit 1 }; printf "gateway overhead p99 %.6fs\n", value }'
