## Summary

Slice 1 of 7 of the provider simulator (`tollgate sim`): the Spec Kit artifacts for the whole feature,
the move of the `tollgate` CLI to cobra, and the simulator configuration (`sim.Config`) loaded from YAML
and validated by `schemas/sim.schema.json`, with flavor defaults, flag overrides and the semantic checks
the schema cannot express (fault rates summing above 1, min above max, 529 outside Anthropic).
Spec: [`specs/002-provider-simulator/spec.md`](https://github.com/hermeto/tollgate/blob/002-provider-simulator/specs/002-provider-simulator/spec.md). Tracking: #3.

## Stack

Based on #4; merge after it. Next slice: `002-provider-simulator-iter1` (server skeleton).

## Requirements covered

Feature spec FR-011 (config validated by a JSON Schema) and the configuration part of FR-005, FR-007
and FR-009. The CLI, endpoints and behaviour arrive in later slices.

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
