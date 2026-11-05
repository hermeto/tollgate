---

description: "Task list for 003 local-stack"
---

# Tasks: `tollgate serve` skeleton, pipeline interface and the local profile

**Input**: Design documents from `specs/003-local-stack/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Included (constitution VIII; the spec requires micro-benchmarks and e2e tests).

## Format: `[ID] [P?] [Story] Description`

## Phase 1: Setup

- [ ] T001 [P] Write docs/adr/0001-standalone-gateway-with-optional-extproc.md (maintainer decision, documented) and docs/adr/0002-state-stores.md (Valkey + PostgreSQL, goose for migrations, pgx driver)
- [ ] T002 Add OpenTelemetry and Prometheus dependencies at the versions in research R5

## Phase 2: User Story 2 — Policy pipeline contract (P1, foundational for every other story)

- [ ] T003 [P] [US2] Tests for hook order, Deny, Respond, Mutate and `ran` semantics in internal/pipeline/pipeline_test.go
- [ ] T004 [US2] Implement Stage, hooks, Decision, Request, Response, Chunk and Pipeline in internal/pipeline/pipeline.go per contracts/pipeline.md
- [ ] T005 [P] [US2] Request-ID Stage in internal/pipeline/requestid.go: propagate a valid `x-request-id` ("1–128 printable ASCII characters") or generate `req_` + 24 base62 characters; add it upstream and to the response; tests
- [ ] T006 [P] [US2] Timing Stage in internal/pipeline/timing.go: stamps arrival and exposes per-Stage durations for the overhead metric; tests
- [ ] T007 [US2] Micro-benchmark of the empty pipeline (request-ID + timing) in internal/pipeline/bench_test.go; record baseline numbers in docs/local-stack.md (SC-002)

## Phase 3: User Story 1 — Faithful streaming passthrough (P1)

- [ ] T008 [P] [US1] Write schemas/gateway.schema.json per data-model.md ("listen.ops must differ from listen.api" checked in code)
- [ ] T009 [US1] Implement internal/config (defaults, YAML load, schema validation, duration parsing) with tests in internal/config/config_test.go
- [ ] T010 [P] [US1] Implement internal/sse (event reader with 1 MiB bound, writer) with tests and a fuzz test for the reader
- [ ] T011 [US1] Implement internal/upstream: tuned transport, hop-by-hop and credential stripping, credential from `credential_env`, Forward returning status/headers/body stream; tests against the in-process simulator
- [ ] T012 [US1] Implement internal/server API handler: body limit (413), pipeline request/response/chunk hooks, non-streaming copy, SSE relay with flush per event, `upstream_stream_error` on early close, 502/504 mapping, client cancel → upstream cancel; tests in internal/server/api_test.go
- [ ] T013 [US1] Implement drain on shutdown (stop accepting, not ready, drain timeout) in internal/server/server.go; test that an in-flight stream completes after shutdown starts
- [ ] T014 [US1] Add `tollgate serve` in cmd/tollgate/serve.go and config/local.yaml (upstream = local simulator)
- [ ] T015 [US1] Passthrough benchmark against the in-process simulator in internal/server/bench_test.go

## Phase 4: User Story 5 — Health and metrics (P2)

- [ ] T016 [P] [US5] Implement internal/telemetry: tracer provider (OTLP/HTTP, batch, bounded queue), logger provider (otelslog bridge + JSON stdout), meter provider with Prometheus exporter; instrument names as constants; tests that a stopped endpoint never blocks (SC-005)
- [ ] T017 [US5] Instrument the request path: server span, per-Stage spans, upstream client span with trace context propagation; metrics `tollgate_requests_total`, `tollgate_request_duration_seconds`, `tollgate_gateway_overhead_seconds{phase}`, `tollgate_stage_duration_seconds{stage}`; access log without content (FR-019)
- [ ] T018 [P] [US5] Implement internal/health: TCP reachability checks every 5 s, fail modes, readiness report, logged transitions; tests
- [ ] T019 [US5] Operations server on `listen.ops`: `/healthz`, `/readyz`, `/metrics`; tests
- [ ] T020 [P] [US5] Add `tollgate probe <url>` in cmd/tollgate/probe.go for container healthchecks

## Phase 5: User Story 3 — One-command local environment (P1)

- [ ] T021 [P] [US3] deploy/compose/gateway.Dockerfile (golang build stage and distroless runtime, both by digest)
- [ ] T022 [US3] deploy/compose/compose.yaml: project `tollgate`, all services by digest, healthchecks, published ports on 127.0.0.1 (FR-021), postgres volume
- [ ] T023 [P] [US3] Service configs: deploy/compose/gateway.yaml, otel-collector.yaml, prometheus.yml, tempo.yaml, grafana provisioning (Prometheus, Tempo, Loki data sources)
- [ ] T024 [US3] scripts/stack.sh (`up` with readiness wait and URL/key banner, `down` with volumes) and Makefile targets `up`, `down` with `PROFILE` (only `local` for now)
- [ ] T025 [US3] scripts/smoke.sh and `make smoke`: non-streaming and streaming requests, trace in Tempo by `tollgate.request_id`, metric in Prometheus

## Phase 6: User Story 4 — The local stack is tested on every PR (P1)

- [ ] T026 [P] [US4] End-to-end tests in test/e2e (TestPassthrough, TestStreamingRelay, TestBodyLimit, TestClientCancelCancelsUpstream, TestTelemetryDownDoesNotBlock) gated by `TOLLGATE_E2E=1`; `make e2e`
- [ ] T027 [US4] CI job `e2e-local` (up → smoke → e2e → overhead check → down, logs on failure) aggregated by ci-gate; update internal/repotest expectations
- [ ] T028 [US4] Overhead check script: query p99 of `tollgate_gateway_overhead_seconds` after a load step and fail above 1 ms (SC-003)

## Phase 7: Polish

- [ ] T029 [P] Write docs/local-stack.md (architecture of the stack, URLs, commands, troubleshooting, measured numbers) and link it from README and CONTRIBUTING
- [ ] T030 Run `make lint test`, the Definition of done and update docs/agent/PROGRESS.md

## Dependencies & Execution Order

Phase 1 → Phase 2 (pipeline) → Phase 3 (needs the pipeline) → Phase 4 (instruments Phase 3) →
Phase 5 (needs `serve` and `probe`) → Phase 6 (needs the stack) → Phase 7.

## PR slices (stacked, each < 800 changed lines excluding fixtures, lockfiles and Spec Kit artifacts)

| PR | Branch | Tasks |
| --- | --- | --- |
| 1 | `003-local-stack` | Spec Kit artifacts, T001–T007 (ADRs, pipeline) |
| 2 | `003-local-stack-iter1` | T008–T010 (config schema, config, SSE) |
| 3 | `003-local-stack-iter2` | T011–T015 (upstream, API server, drain, `serve`, benchmark) |
| 4 | `003-local-stack-iter3` | T016–T020 (telemetry, health, ops server, `probe`) |
| 5 | `003-local-stack-iter4` | T021–T025 (Compose stack, `make up/down/smoke`) |
| 6 | `003-local-stack-iter5` | T026–T030 (e2e tests, `e2e-local`, docs) |

Feature 004 branches from the last slice.

## Implementation Strategy

MVP = slices 1–3 (a working passthrough gateway). Slices 4–6 add observability, the stack and CI.
