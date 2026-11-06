#!/usr/bin/env bash
set -euo pipefail

api=${TOLLGATE_API_URL:-http://127.0.0.1:8080}
ops=${TOLLGATE_OPS_URL:-http://127.0.0.1:8081}
request_id="smoke-$(date +%s)"
body='{"model":"simulated","messages":[{"role":"user","content":"hello"}],"stream":false}'
response=$(curl --fail --silent --show-error -H "x-request-id: $request_id" -H 'content-type: application/json' -d "$body" "$api/v1/chat/completions")
grep -q '"choices"' <<<"$response"
stream_body='{"model":"simulated","messages":[{"role":"user","content":"hello"}],"stream":true}'
stream=$(curl --fail --silent --show-error -H 'content-type: application/json' -d "$stream_body" "$api/v1/chat/completions")
grep -q 'data: \[DONE\]' <<<"$stream"
curl --fail --silent "$ops/metrics" | grep -q 'tollgate_requests_total'
tempo=${TOLLGATE_TEMPO_URL:-http://127.0.0.1:3200}
deadline=$((SECONDS + 30))
until curl --fail --silent --get --data-urlencode "tags=tollgate.request_id=$request_id" "$tempo/api/search" | grep -q "$request_id"; do
  if (( SECONDS >= deadline )); then echo "trace not found in Tempo" >&2; exit 1; fi
  sleep 1
done
echo "smoke passed (request_id=$request_id)"
