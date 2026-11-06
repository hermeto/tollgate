# Implementation Plan: Multi-protocol API and logical models

## Summary

Introduce a canonical request/response model and protocol adapters for OpenAI and Anthropic while retaining
raw bytes for passthrough. Add logical model mapping and a table-driven translation package.

## Technical Context

Go 1.27.3; standard library JSON/SSE handling; existing `internal/server`, `internal/sse`, config schema and
simulators. Translation is deterministic and never calls real providers. Unknown same-protocol fields remain in
raw bytes; translated requests use explicit support validation.

## Architecture

- `internal/canonical`: protocol-neutral messages, tools, usage and model targets.
- `internal/api/openai`, `internal/api/anthropic`: bounded decoding and protocol error encoding.
- `internal/translate`: OpenAI↔Anthropic conversion with support matrix and stream events.
- `internal/config`: logical model configuration validated by `schemas/models.schema.json`.
- `internal/server`: route dispatch and target selection.

## Phases

1. Canonical types and logical-model schema.
2. OpenAI and Anthropic adapters with fixture tests.
3. Translation and unsupported-feature errors.
4. Server wiring, SDK compatibility and fuzz tests.

## Constitution Check

Pass: provider calls remain simulator-only, request bodies stay out of logs/telemetry, parsers are bounded and
fuzzed, and dependencies remain pinned in go.mod/go.sum.
