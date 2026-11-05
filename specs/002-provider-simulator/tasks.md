---

description: "Task list for 002 provider-simulator"
---

# Tasks: Deterministic OpenAI and Anthropic simulators with fault injection

**Input**: Design documents from `specs/002-provider-simulator/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Included — constitution VIII requires contract tests from recorded fixtures; tests precede
implementation inside each story.

**Organization**: Tasks are grouped by user story. Delivery is split into stacked PRs to respect the
800-line rule (see "PR slices" at the end).

## Format: `[ID] [P?] [Story] Description`

## Phase 1: Setup

- [ ] T001 Add dependencies (cobra v1.10.2, jsonschema/v6 v6.0.3) to go.mod
- [ ] T002 Move the CLI to cobra: root command, `version` subcommand and `--version` flag in cmd/tollgate/main.go and cmd/tollgate/root.go; keep the version output contract from 001 (cmd/tollgate/main_test.go)

## Phase 2: Foundational (blocking)

- [ ] T003 [P] Write schemas/sim.schema.json (draft 2020-12) per data-model.md: Distribution `kind` "fixed|normal|lognormal", rates "0–1", sum of `error_rates` and `reset_rate` "≤ 1" checked in code
- [ ] T004 [P] Add schemas/schemas.go embedding *.json and a `Validate(name, doc)` helper with tests in schemas/schemas_test.go
- [ ] T005 Implement sim/config.go: defaults per flavour (ports 9001/9002, seed 42, output normal(120, 40, 1..1024), dims 1536, log size 1000), YAML load, schema validation, flag overrides; tests in sim/config_test.go (valid file, unknown field, rate sum > 1)
- [ ] T006 [P] Implement sim/rng.go: content RNG = PCG(seed, fnv64(body)); latency RNG seeded with seed, body hash and occurrence count (FR-019); tests for determinism and independence
- [ ] T007 [P] Implement sim/sse.go (event/data framing, flush after each event) with tests
- [ ] T008 Implement sim/server.go: flavour routing, auth-header check (FR-016), `/_sim/healthz`, graceful shutdown with 5 s grace (FR-020), bind 127.0.0.1 by default (FR-018); tests in sim/server_test.go
- [ ] T009 Add `tollgate sim` subcommand in cmd/tollgate/sim.go (flags per contracts/cli.md)

## Phase 3: User Story 5 — Documented token counting (P1)

- [ ] T010 [P] [US5] Tests for internal/tokens (pieces, CJK, digits, code, empty string, message overhead) in internal/tokens/tokens_test.go
- [ ] T011 [US5] Implement internal/tokens/tokens.go: RE2 pre-tokenizer, `Count(text)`, `CountChat(messages)` (+3 per message, +3 priming)
- [ ] T012 [P] [US5] Implement sim/generate.go: deterministic word-piece generator (one piece = one token) and output length sampling clamped by request max tokens; tests

## Phase 4: User Story 1 — OpenAI flavour (P1)

- [ ] T013 [P] [US1] Record fixtures in testdata/wire/openai/ (chat completion, stream with usage, tool call, embeddings float and base64, models, errors) with testdata/wire/README.md citing sources (research R1, R2)
- [ ] T014 [P] [US1] Fixture structure tests in sim/openai_test.go (same keys, same chunk order; ids and timestamps ignored)
- [ ] T015 [US1] Implement chat completions non-streaming and streaming with `include_usage`, `max_tokens`/`max_completion_tokens` → `length`, `n > 1` → 400, forced tool calls (FR-017) in sim/openai.go
- [ ] T016 [US1] Implement `/v1/embeddings` (deterministic unit vectors, float and base64, `dimensions`) and `/v1/models` in sim/openai.go
- [ ] T017 [US1] Implement OpenAI error bodies (400, 401, 404 model_not_found, 429, 500, 503) in sim/openai.go

## Phase 5: User Story 2 — Anthropic flavour (P1)

- [ ] T018 [P] [US2] Copy fixtures from anthropic-sdk-python v2.5.0 (basic, tool use) and add non-streaming and error fixtures in testdata/wire/anthropic/ with sources
- [ ] T019 [P] [US2] Fixture structure and event-order tests in sim/anthropic_test.go
- [ ] T020 [US2] Implement `/v1/messages` non-streaming and streaming event sequence, `max_tokens` required, `max_tokens` stop reason, forced tool use, error bodies (incl. 529) in sim/anthropic.go

## Phase 6: User Story 3 — Latency model (P1)

- [ ] T021 [P] [US3] Tests for distributions (fixed, normal, lognormal, clamps) and for stream timing tolerances (US3 AS1) in sim/latency_test.go
- [ ] T022 [US3] Implement sim/latency.go and apply TTFT and inter-token delays in both flavours, cancelling timers on client disconnect

## Phase 7: User Story 4 — Fault injection (P1)

- [ ] T023 [P] [US4] Tests for each fault (status rates with retry-after, reset, pre-first-token timeout, mid-stream stall, outage window, live set/merge/reset) in sim/faults_test.go
- [ ] T024 [US4] Implement sim/faults.go (atomic state, outage resolution, decision per request) and admin endpoints `/_sim/faults` per contracts/admin-api.md

## Phase 8: User Story 6 — Request log (P2)

- [ ] T025 [P] [US6] Tests for ring buffer bounds, ordering and clearing in sim/reqlog_test.go
- [ ] T026 [US6] Implement sim/reqlog.go and `/_sim/requests` (GET, DELETE); admin calls are not recorded

## Phase 9: User Story 7 — Embeddings forward (P2)

- [ ] T027 [US7] Implement `embeddings.forward_url` proxy with 502 on upstream failure; tests with an httptest upstream in sim/openai_test.go

## Phase 10: SDK compatibility (FR-013, SC-001)

- [ ] T028 [P] Python clients (openai 3.46.0, anthropic 2.5.0) in test/sdk/python/ with pyproject.toml and uv.lock
- [ ] T029 [P] TypeScript clients (openai 7.51.0, @anthropic-ai/sdk 1.5.0) in test/sdk/node/ with package-lock.json
- [ ] T030 [P] Go clients (openai-go v3.95.0, anthropic-sdk-go v2.6.0) in test/sdk/go/ (own go.mod)
- [ ] T031 Implement scripts/sdk-compat.sh and `make sdk-compat`; add `sdk-compat` job to .github/workflows/ci.yml aggregated by ci-gate (update internal/repotest expectations)

## Phase 11: Polish & cross-cutting

- [ ] T032 Throughput benchmark (SC-003) in sim/bench_test.go and TTFT accuracy test under 1,000 concurrent streams (SC-007)
- [ ] T033 Token error measurement script test/sdk/python/token_error.py with corpus testdata/tokens/corpus.jsonl (SC-006)
- [ ] T034 [P] Write docs/simulator.md (usage, config, faults, token method, measured throughput, TTFT accuracy, token error with hardware) and link it from README
- [ ] T035 [P] Write docs/adr/0009-simulated-providers-default-upstream.md
- [ ] T036 Run `go test ./sim/... ./internal/tokens/... -race -cover`, `make sdk-compat`, the benchmark and `make lint`; update docs/agent/PROGRESS.md

## Dependencies & Execution Order

Phase 1 → Phase 2 → Phase 3 (tokens and generator feed usage) → Phases 4 and 5 (independent of each
other) → Phases 6, 7, 8 (need both flavours) → Phase 9 → Phase 10 (needs everything the SDKs touch) →
Phase 11.

## Parallel Example: User Story 1

```bash
T013 fixtures   T014 fixture tests   # in parallel, then T015–T017 sequentially in sim/openai.go
```

## PR slices (stacked, each < 800 changed lines excluding fixtures, lockfiles and Spec Kit artifacts)

| PR | Branch | Phases |
| --- | --- | --- |
| 1 | `002-provider-simulator` | Spec Kit artifacts, Phases 1–3 |
| 2 | `002-provider-simulator-iter1` | Phase 4 (US1) |
| 3 | `002-provider-simulator-iter2` | Phases 5 and 8 (US2, US6) |
| 4 | `002-provider-simulator-iter3` | Phases 6 and 7 (US3, US4) |
| 5 | `002-provider-simulator-iter4` | Phases 9–11 |

Feature 003 branches from the last slice.

## Implementation Strategy

MVP = Phases 1–4 (OpenAI flavour usable by 003). Each later slice is independently testable through
its own tests; the SDK suite in slice 5 is the end-to-end proof for SC-001.
