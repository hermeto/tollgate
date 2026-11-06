# Implementation Plan: Pricing and Usage

## Technical Context

Go 1.27, YAML price files, JSON-lines usage events and a bounded asynchronous writer. The gateway
remains simulator-only in tests. Pricing is isolated from transport so later budgets and reports can
reuse the exact event model.

## Constitution Check

- No real provider calls: pass.
- No secrets or content in telemetry: pass.
- Bounded hot-path work and graceful shutdown: pass.

## Design

`internal/pricing` loads `pricing/<provider>.yaml` and selects effective prices. `internal/usage`
defines immutable events, a sink interface, a buffered writer and JSONL spool sink. The initial server
integration records a completed request's provider usage when the response exposes it; protocol
parsers remain responsible for future stream aggregation.

## Risks and Mitigations

Provider-specific usage shapes vary; unknown/missing usage is explicit via `estimated`. Price changes
are versioned by effective date, so historical events retain their computed cost.
