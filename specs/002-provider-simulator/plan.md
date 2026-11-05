# Implementation Plan: Deterministic OpenAI and Anthropic simulators with fault injection

**Branch**: `002-provider-simulator` | **Date**: 2026-11-04 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/002-provider-simulator/spec.md`

## Summary

Add package `sim/` and the `tollgate sim` subcommand: an HTTP server with an OpenAI flavour
(`/v1/chat/completions`, `/v1/embeddings`, `/v1/models`) and an Anthropic flavour (`/v1/messages`) that
emit the official JSON and SSE shapes, a seeded latency and output-length model, live fault injection
(`/_sim/faults`), a request log (`/_sim/requests`) and deterministic embeddings. Token counting lives in
`internal/tokens` so later features (006, 007) reuse the same approximation. A compatibility suite runs
the official SDKs (Python, TypeScript, Go) against both flavours.

## Technical Context

**Language/Version**: Go 1.27.3.

**Primary Dependencies**: `github.com/spf13/cobra` v1.10.2 (CLI; generates the CLI reference in 018),
`go.yaml.in/yaml/v4` v4.0.0 (config), `github.com/santhosh-tekuri/jsonschema/v6` v6.0.3 (schema
validation). Standard library for HTTP, SSE, randomness (`math/rand/v2` PCG) and hashing.
SDK compatibility only (separate modules/lockfiles): `openai` 3.46.0 and `anthropic` 2.5.0 (PyPI, uv),
`openai` 7.51.0 and `@anthropic-ai/sdk` 1.5.0 (npm), `github.com/openai/openai-go/v3` v3.95.0 and
`github.com/anthropics/anthropic-sdk-go/v2` v2.6.0 (Go), `tiktoken` 0.14.0 (reference tokenizer).

**Storage**: in-memory only (request log ring buffer, fault state).

**Testing**: `go test -race` unit tests per handler against `testdata/wire` fixtures (structure
comparison), fault and latency tests with short durations, a throughput benchmark, and the SDK suite
(`make sdk-compat`).

**Target Platform**: Linux and macOS, amd64 and arm64; runs locally, in Compose (003) and kind (014/015).

**Project Type**: single Go module; `sim` is a library package plus a cobra subcommand.

**Performance Goals**: ≥ 10,000 non-streaming req/s per instance with zero latency (SC-003).

**Constraints**: streams never buffered; every timer cancelled on client disconnect; no outbound calls
except the optional embeddings forward URL; bind 127.0.0.1 by default.

**Scale/Scope**: ~1,500 lines of Go plus tests and fixtures; SDK scripts ~150 lines per language.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
| --- | --- | --- |
| I. Evidence | Pass | Throughput and token error are measured and documented with hardware. |
| II. Hot path | Pass | Simulator is not the gateway, but it must not distort measurements: benchmark committed. |
| III. Privacy | Pass | Request log holds synthetic test traffic only; documented as a test double. |
| IV. Fail modes | Pass | This feature *creates* the dependency failures other features declare fail modes for. |
| V. Governance as code | Pass | Config validated by `schemas/sim.schema.json`. |
| VI. Local-first, zero cost | Pass | No outbound calls; SDKs point at localhost. |
| VII. Reproducible | Pass | Seeded RNG; SDK versions pinned in lockfiles; fixtures cited. |
| VIII. Test what matters | Pass | Contract tests from recorded fixtures; SDK compatibility in CI. |
| IX. Supply chain | Pass | New Go deps scanned by govulncheck/Trivy; npm and uv lockfiles committed. |
| X. Simplicity | Pass | One server type, flavour switch, no plugin system. |
| XI. OSS citizenship | Pass | Docs and ADR-0009 in the same PR. |

Post-design re-check: Pass.

## Project Structure

### Documentation (this feature)

```text
specs/002-provider-simulator/
├── plan.md  research.md  data-model.md  quickstart.md
├── contracts/{openai.md,anthropic.md,admin-api.md,cli.md}
├── checklists/
└── tasks.md
```

### Source Code (repository root)

```text
cmd/tollgate/            # cobra root: version, sim
sim/
├── config.go            # Config, defaults, YAML load + schema validation
├── server.go            # Server, routing, auth check, request log middleware
├── openai.go            # chat (JSON + SSE), embeddings, models, errors
├── anthropic.go         # messages (JSON + SSE), errors
├── generate.go          # deterministic text/tool-call generation
├── latency.go           # Distribution sampling, per-request RNG
├── faults.go            # FaultState, admin API, injection points
├── reqlog.go            # bounded request log
├── sse.go               # SSE writer with flush
└── *_test.go, bench_test.go
internal/tokens/         # documented token approximation (shared with 006/007)
schemas/                 # sim.schema.json + embed package
testdata/wire/{openai,anthropic}/   # fixtures + README with sources
test/sdk/{python,node,go}/          # official SDK compatibility clients (pinned)
scripts/sdk-compat.sh               # starts both sims, runs the 6 clients
docs/adr/0009-simulated-providers-default-upstream.md
docs/simulator.md                   # usage, latency/fault config, token method, measured numbers
```

**Structure Decision**: `sim/` at the repository root per Appendix A §A.8; token counting in
`internal/tokens` because the gateway's estimator (006) must use the same method.

## Complexity Tracking

No constitution violations.
