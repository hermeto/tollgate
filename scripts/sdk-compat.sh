#!/usr/bin/env bash
# Runs the official OpenAI and Anthropic SDKs (Python, TypeScript, Go) against Tollgate, changing
# only the base URL (spec 002 FR-013, SC-001). TARGET=sim (default) starts both simulators.
# TARGET=gateway is reserved for feature 004. Requires go, uv and npm; never calls a real provider.
set -euo pipefail

TARGET="${TARGET:-sim}"
OPENAI_PORT="${SIM_OPENAI_PORT:-19101}"
ANTHROPIC_PORT="${SIM_ANTHROPIC_PORT:-19102}"
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if [[ "${TARGET}" != "sim" && "${TARGET}" != "gateway" ]]; then
  echo "sdk-compat: unknown TARGET=${TARGET} (expected sim or gateway)" >&2
  exit 2
fi

workdir="$(mktemp -d)"
pids=()
cleanup() {
  for pid in "${pids[@]}"; do kill "${pid}" 2>/dev/null || true; done
  wait 2>/dev/null || true
  rm -rf "${workdir}"
}
trap cleanup EXIT

echo "==> building tollgate"
(cd "${ROOT}" && go build -o "${workdir}/tollgate" ./cmd/tollgate)

start_sim() {
  local flavor="$1" port="$2"
  "${workdir}/tollgate" sim --flavor "${flavor}" --port "${port}" 2>"${workdir}/sim-${flavor}.log" &
  pids+=("$!")
  for _ in $(seq 1 50); do
    if curl -fsS "http://127.0.0.1:${port}/_sim/healthz" >/dev/null 2>&1; then
      return 0
    fi
    sleep 0.1
  done
  echo "sdk-compat: ${flavor} simulator did not become healthy" >&2
  cat "${workdir}/sim-${flavor}.log" >&2
  return 1
}

start_sim openai "${OPENAI_PORT}"
start_sim anthropic "${ANTHROPIC_PORT}"
if [[ "${TARGET}" == gateway ]]; then
  export TOLLGATE_SDK_UPSTREAM_KEY=tg-example-sdk-compat
  "${workdir}/tollgate" serve --config "${ROOT}/config/gateway-sdk.yaml" 2>"${workdir}/gateway.log" &
  pids+=("$!")
  for _ in $(seq 1 50); do
    if curl -fsS "http://127.0.0.1:18081/readyz" >/dev/null 2>&1; then break; fi
    sleep 0.1
  done
  export OPENAI_BASE_URL="http://127.0.0.1:18080/v1"
  export ANTHROPIC_BASE_URL="http://127.0.0.1:18080"
else
  export OPENAI_BASE_URL="http://127.0.0.1:${OPENAI_PORT}/v1"
  export ANTHROPIC_BASE_URL="http://127.0.0.1:${ANTHROPIC_PORT}"
fi
# Make sure no real provider key can be picked up by an SDK default.
unset OPENAI_API_KEY ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN OPENAI_ORG_ID OPENAI_PROJECT_ID || true

failed=0
echo "==> python (uv)"
(cd "${ROOT}/test/sdk/python" && uv run --locked --quiet python compat.py) || failed=1
echo "==> typescript (npm)"
(cd "${ROOT}/test/sdk/node" && npm ci --ignore-scripts --no-audit --no-fund --silent && node compat.mjs) || failed=1
echo "==> go"
(cd "${ROOT}/test/sdk/go" && go run .) || failed=1

if ((failed)); then
  echo "sdk-compat: FAILED" >&2
  exit 1
fi
echo "sdk-compat: all 6 clients passed"
