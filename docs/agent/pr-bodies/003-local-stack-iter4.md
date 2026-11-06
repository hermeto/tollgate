## Summary

Slice 5 of feature 003 adds the reproducible local Compose profile: a digest-pinned gateway image,
two simulators, Valkey, PostgreSQL, OpenTelemetry Collector, Prometheus, Tempo, Loki and Grafana.
It also adds `make up`, `make down` and `make smoke` with readiness polling, cleanup assertions and
trace/metric smoke checks.

## Base

Based on `003-local-stack-iter3`; merge after PR #18.

## Requirements covered

FR-013, FR-014, FR-021, SC-001 and SC-004; tasks T021–T025.

## Decisions for the reviewer

- Every published port binds to loopback; internal service ports remain private to the Compose network.
- The gateway and simulator containers use the same locally built distroless image; all external images
  are pinned by the digests recorded in research R1.
- `make smoke` waits for the asynchronous Tempo export and verifies the request ID trace plus metrics.
- Docker execution was not possible in this environment because the local daemon is unavailable;
  `docker compose config` and all static checks pass, and CI/E2E will exercise the full stack.

## Verification

- `go test ./...`
- `go test -race ./...`
- `golangci-lint run ./...`
- `docker compose config`
- `shellcheck` and `yamllint`
