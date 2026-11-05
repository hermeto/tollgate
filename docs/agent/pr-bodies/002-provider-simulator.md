## Summary

Slice 1 of 5 of the provider simulator (`tollgate sim`): Spec Kit artifacts for the whole feature, the
cobra CLI, the simulator configuration with JSON Schema validation, the HTTP server skeleton (auth
check, health, graceful shutdown, loopback default), deterministic RNG, SSE writer, the text generator
and the shared token-counting approximation (`internal/tokens`).
Spec: [`specs/002-provider-simulator/spec.md`](https://github.com/hermeto/tollgate/blob/002-provider-simulator/specs/002-provider-simulator/spec.md). Tracking: #3.

## Stack

Based on #4; merge after it. Next slice: `002-provider-simulator-iter1` (OpenAI flavour).

## Requirements covered

Feature spec FR-001 (CLI), FR-008 (token method), FR-011 (config schema), FR-014, FR-016, FR-018,
FR-019, FR-020. Product spec NFR-7, NFR-8 (groundwork).

## Decisions for the reviewer

_Filled in when the slice is ready._

## How it was verified

_Filled in when the slice is ready._

## Not verified

_Filled in when the slice is ready._

## Performance impact

None on the gateway hot path (no gateway code yet).

## Checklist

- [ ] Tests added or updated
- [ ] Docs updated in this PR (README, docs/, runbooks)
- [ ] ADR added or updated for architectural decisions
- [ ] No secrets, real keys or real personal data
- [ ] New dependencies, images and actions pinned
- [ ] No prompt/response content or PII in logs, traces, metrics or events
- [ ] Commits signed off (DCO) and Conventional
