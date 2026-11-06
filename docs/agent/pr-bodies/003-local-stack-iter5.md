## Summary

Final slice of feature 003. Adds stack-gated end-to-end tests, the `e2e-local` CI job, Prometheus
overhead enforcement, local-stack documentation and the complete feature Definition of Done.

## Base

Based on `003-local-stack-iter4`; merge after PR #19.

## Requirements covered

FR-013–FR-015, SC-001, SC-003–SC-005; tasks T026–T030.

## Decisions for the reviewer

- E2E tests skip unless `TOLLGATE_E2E=1`, keeping ordinary unit-test runs Docker-free.
- CI owns the Compose lifecycle and always tears down volumes/orphans; logs are collected on failure.
- The overhead gate queries p99 from Prometheus and fails above 1 ms; missing samples fail closed.

## Verification

- `make lint test`
- `go test ./...` and `go test -race ./...`
- `actionlint`, `shellcheck`, `yamllint`
- `docker compose config`
- Full Docker execution remains pending because the local Docker daemon is unavailable; CI runs it.
