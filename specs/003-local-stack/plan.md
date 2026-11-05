# Implementation Plan: `tollgate serve` skeleton, pipeline interface and the local profile

**Branch**: `003-local-stack` | **Date**: 2026-11-05 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/003-local-stack/spec.md`

## Summary

Add the gateway skeleton: `internal/pipeline` (Stage interface, decisions, runner, request-ID and timing
Stages), `internal/config` (YAML validated by `schemas/gateway.schema.json`), `internal/sse` (event
framing shared with 004 and 016), `internal/upstream` (tuned HTTP client, streaming relay),
`internal/server` (API and operations servers, drain on shutdown), `internal/telemetry` (OpenTelemetry
traces and logs over OTLP, metrics on `/metrics`), and `internal/health` (dependency checks with fail
modes). `tollgate serve` and `tollgate probe` wire them up. `deploy/compose/` holds the local profile:
gateway, two simulators, Valkey, PostgreSQL, OpenTelemetry Collector, Prometheus, Tempo, Loki and
Grafana, all pinned by digest. `make up|down|smoke` and the `e2e-local` CI job drive it.

## Technical Context

**Language/Version**: Go 1.27.3.

**Primary Dependencies**: OpenTelemetry Go v1.47.0 (`otel`, `sdk`, `sdk/log`, `otlptracehttp`,
`otlploghttp`), `exporters/prometheus` v0.69.0 with `prometheus/client_golang` v1.25.0,
`contrib/bridges/otelslog` v0.21.0. Standard library `net/http` for both servers and the upstream client.
Existing: cobra, yaml v4, jsonschema v6.

**Storage**: none on the request path yet; Valkey 9.1.3 and PostgreSQL 18.6 run in the stack and are
health-checked. Migrations tool chosen now for 006: goose v3.29.0 (ADR-0002).

**Testing**: unit tests per package; pipeline property tests (hook order, decisions); upstream relay
tests with the in-process simulator (`sim` package); `go test -bench` for the empty pipeline and the
passthrough; end-to-end tests in `test/e2e` against the Compose stack (`TOLLGATE_E2E=1`).

**Target Platform**: Linux containers (amd64, arm64) and macOS for development.

**Project Type**: single Go module; Compose profile under `deploy/compose/`.

**Performance Goals**: empty pipeline ≪ 100 µs; passthrough gateway overhead ≤ 1 ms p99 (SC-003),
measured inside the gateway (`tollgate_gateway_overhead_seconds`).

**Constraints**: no buffering of streams; telemetry never blocks requests (batching exporters with
bounded queues, drop on overflow); client credentials never forwarded; every image pinned by digest
with arm64 and amd64 variants.

**Scale/Scope**: ~3,500 changed lines across 6 stacked slices (tasks.md "PR slices").

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
| --- | --- | --- |
| I. Evidence | Pass | Overhead histogram and a committed micro-benchmark baseline. |
| II. Hot path | Pass | Stage interface with per-Stage span, metric and benchmark from day one. |
| III. Privacy | Pass | Access logs and spans carry ids, sizes and timings, never bodies. |
| IV. Fail modes | Pass | Dependencies declare `fail_mode`; readiness transitions are logged. Telemetry is fail-open by design. |
| V. Governance as code | Pass | Gateway config validated by `schemas/gateway.schema.json`; stack configs in git. |
| VI. Local-first | Pass | Default upstream is the simulator; the stack runs on a laptop. |
| VII. Reproducible | Pass | Images by digest, Go modules in go.sum. |
| VIII. Tests | Pass | Unit, relay, benchmark and e2e tests; `e2e-local` required in CI. |
| IX. Supply chain | Pass | New deps scanned; gateway image distroless non-root. |
| X. Simplicity | Pass | stdlib HTTP; TCP reachability checks until real clients arrive (006/007). |
| XI. OSS citizenship | Pass | ADR-0001 and ADR-0002, docs in the same PRs. |

Post-design re-check: Pass.

## Project Structure

### Documentation (this feature)

```text
specs/003-local-stack/
├── plan.md  research.md  data-model.md  quickstart.md
├── contracts/{pipeline.md,config.md,http.md,make-targets.md}
├── checklists/
└── tasks.md
```

### Source Code (repository root)

```text
cmd/tollgate/{serve.go,probe.go}
internal/pipeline/        # Stage, Decision, Request, Response, Chunk, Pipeline; requestid.go, timing.go
internal/config/          # gateway config load + validation
internal/sse/             # SSE event reader/writer
internal/upstream/        # client, Forward, streaming relay
internal/server/          # API server, ops server, drain
internal/telemetry/       # providers, names, metrics instruments
internal/health/          # dependency checks, readiness state
schemas/gateway.schema.json
config/local.yaml          # host run against local simulators
deploy/compose/            # compose.yaml, gateway.Dockerfile, gateway.yaml, otel-collector.yaml,
                           # prometheus.yml, tempo.yaml, grafana/provisioning/
scripts/{stack.sh,smoke.sh}
test/e2e/                  # TestPassthrough, TestStreamingRelay, TestBodyLimit, …
docs/adr/0001-standalone-gateway-with-optional-extproc.md
docs/adr/0002-state-stores.md
docs/local-stack.md
```

**Structure Decision**: Appendix A §A.8 layout; core request path packages are separate so later
policies depend only on `internal/pipeline`.

## Complexity Tracking

No constitution violations.
