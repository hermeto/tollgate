## Summary

Slice 1 of 6 of feature 003: the Spec Kit artifacts for the gateway skeleton and local profile, ADR-0001
(standalone Go gateway with an optional Envoy ext_proc adapter) and ADR-0002 (Valkey and PostgreSQL;
goose for migrations), and `internal/pipeline`: the Stage contract every policy implements, the runner,
and the request-ID and timing Stages, with a micro-benchmark.
Spec: [`specs/003-local-stack/spec.md`](https://github.com/hermeto/tollgate/blob/003-local-stack/specs/003-local-stack/spec.md). Tracking: #3.

## Stack

Based on #14; merge after it. Next slice: `003-local-stack-iter1` (gateway config, SSE).

## Requirements covered

Feature spec FR-006, FR-007, FR-008, FR-009 (Stage timing), SC-002. ADR-0001 and ADR-0002 (product
spec §A.10).

## Decisions for the reviewer

- Clarifications for the whole feature (Valkey instead of Redis for licensing, ports 8080/8081, upstream credential from `credential_env`, mid-stream error event, labelled fake example keys): [spec.md](https://github.com/hermeto/tollgate/blob/003-local-stack/specs/003-local-stack/spec.md#clarifications).
- Hook model: `Stage` + three optional hook interfaces resolved once at construction; decisions are a value type (no allocation). Research R2.
- A Stage that denies or short-circuits does not run its own response hook; the outer Stages do (FR-007 wording updated). Headers that must be on every outcome go in `Request.ResponseHeader`.
- Respond in the response phase is ignored (treated as Continue); Respond without a response becomes a 500 Deny.
- Chunk hooks are not recorded per event (only the chunk phase total matters for overhead), to keep streams allocation-free.
- Request ids: a client id is kept when it is 1–128 printable ASCII characters; otherwise `req_` + `crypto/rand.Text()`.
- ADR-0002: Valkey + PostgreSQL 18, pgx driver, goose migrations; the Redis client is chosen in 007 by benchmark.
- 003 is delivered in 6 stacked slices (tasks.md "PR slices").

## How it was verified

- `make lint test` clean; `internal/pipeline` coverage 95.6%. Tests: hook order (request A→B→C, response and chunk C→B→A, mutations applied), Deny at B (C and upstream skipped, only A sees the response), Respond short-circuit, Respond without response, response- and chunk-phase Deny, Mutate recorded, construction validation, request-id propagation/generation/uniqueness, timing stamp.

## Not verified

- Nothing runs the pipeline on real traffic yet: the server arrives in slice 3.

## Performance impact

New hot-path baseline: `BenchmarkEmptyPipeline` 540 ns/op, 12 allocs (request-id generation dominates); `BenchmarkChunkNoHooks` 4 ns/op, 0 allocs (Apple M4).

## Checklist

- [x] Tests added or updated
- [x] Docs updated in this PR (README, docs/, runbooks)
- [x] ADR added or updated for architectural decisions
- [x] No secrets, real keys or real personal data
- [x] New dependencies, images and actions pinned
- [x] No prompt/response content or PII in logs, traces, metrics or events
- [x] Commits signed off (DCO) and Conventional
