## Summary

Slice 4 of feature 003 adds observability and operational health to the local gateway. It includes
OpenTelemetry traces/logs/metrics with bounded fail-open exporters, dependency readiness checks, the
operations endpoints, and `tollgate probe`.

## Base

Based on `003-local-stack-iter2`; merge after PR #17.

## Requirements covered

FR-010 through FR-012, FR-019, SC-005, and tasks T002/T016–T020.

## Decisions for the reviewer

- Telemetry is fail-open: bounded batch queues and five-second exporter timeouts ensure a stopped
  collector cannot block requests.
- JSON stdout logs are fanned out with the `otelslog` bridge; bodies, prompts, completions and
  credentials are never emitted.
- Dependencies use TCP checks every five seconds until protocol clients arrive in later features;
  fail-closed dependencies gate readiness and fail-open dependencies are reported as degraded.
- Codex Spec Kit support is installed alongside Claude as `.agents/skills`; Claude remains default.

## Verification

- `go test ./...`
- `go test -race ./...`
- `golangci-lint run ./...`
- pre-commit hooks, including markdownlint, pass.
