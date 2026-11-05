## Summary

Slice 1 of 6 of feature 003: the Spec Kit artifacts for the gateway skeleton and local profile, ADR-0001
(standalone Go gateway with an optional Envoy ext_proc adapter) and ADR-0002 (Valkey and PostgreSQL;
goose for migrations), and `internal/pipeline`: the Stage contract every policy implements, the runner,
and the request-ID and timing Stages, with a micro-benchmark.
Spec: [`specs/003-local-stack/spec.md`](https://github.com/hermeto/tollgate/blob/003-local-stack/specs/003-local-stack/spec.md). Tracking: #3.

## Stack

Based on #14; merge after it. Next slice: `003-local-stack-iter1` (gateway config, SSE).

## Requirements covered

_Filled in when the slice is ready._

## Decisions for the reviewer

_Filled in when the slice is ready._

## How it was verified

_Filled in when the slice is ready._

## Not verified

_Filled in when the slice is ready._

## Performance impact

_Filled in when the slice is ready._

## Checklist

- [ ] Tests added or updated
- [ ] Docs updated in this PR (README, docs/, runbooks)
- [ ] ADR added or updated for architectural decisions
- [ ] No secrets, real keys or real personal data
- [ ] New dependencies, images and actions pinned
- [ ] No prompt/response content or PII in logs, traces, metrics or events
- [ ] Commits signed off (DCO) and Conventional
