## Summary

Slice 8 of 9 of the provider simulator: the official-SDK compatibility suite and the embeddings forward
mode. `make sdk-compat` (and the new required CI job `sdk-compat`) starts both simulators and runs the
official OpenAI and Anthropic SDKs for Python, TypeScript and Go against them, changing only the base
URL. `embeddings.forward_url` proxies `/v1/embeddings` to a local OpenAI-compatible embedding server.
Spec: [`specs/002-provider-simulator/spec.md`](https://github.com/hermeto/tollgate/blob/002-provider-simulator-iter7/specs/002-provider-simulator/spec.md) (US7, FR-013). Tracking: #3.

## Stack

Based on #11; merge after it. Next slice: `002-provider-simulator-iter8` (performance, ADR-0009).

## Requirements covered

Feature spec FR-009 (forward mode), FR-013, FR-015, SC-001. Product spec NFR-10 (official SDKs work by
changing only `base_url`), NFR-7.

## Decisions for the reviewer

- SDK pins: `openai` 3.46.0 and `anthropic` 2.5.0 (uv), `openai` 7.51.0 and `@anthropic-ai/sdk` 1.5.0 (npm), `openai-go/v3` v3.95.0 and `anthropic-sdk-go/v2` v2.6.0 (own Go module, so their large dependency graphs stay out of the main module). Node 26.13.0 (current LTS) in CI.
- The clients use plain `create(..., stream=True)` iteration rather than SDK stream helpers, which differ across these SDK versions.
- The harness unsets `OPENAI_API_KEY`, `ANTHROPIC_API_KEY` and related variables so no real key can be picked up by an SDK default.
- Default model names follow the pinned SDKs' constants (`claude-*-5-5`, `gpt-5.5`, `gpt-5.4-*`): the Anthropic TypeScript SDK warned that `claude-sonnet-4-5` reaches end of life on 2026-11-30. `gpt-4o-mini` stays because the OpenAI fixtures use it.
- `sdk-compat` is a required job (aggregated by `ci-gate`); Dependabot now covers `test/sdk/{python,node,go}`.
- Forward mode returns 502 on target failure and never falls back to synthetic vectors; it is the only outbound call and targets operator-configured local URLs only.

## How it was verified

- `make sdk-compat`: all 6 clients pass (chat JSON and streaming with usage, forced tool call in Python/TypeScript, embeddings, model list; messages JSON and streaming).
- `make lint test` clean; forward mode tested with an httptest upstream (body passed unchanged, 502 when the target is down); Trivy filesystem scan clean for `uv.lock`, `package-lock.json` and `test/sdk/go/go.mod`.

## Not verified

- `TARGET=gateway`: arrives with feature 004.

## Performance impact

None on the gateway hot path. CI gains one job (about 1 minute locally).

## Checklist

- [x] Tests added or updated
- [x] Docs updated in this PR (README, docs/, runbooks)
- [x] ADR added or updated for architectural decisions — ADR-0009 ships with the next slice
- [x] No secrets, real keys or real personal data
- [x] New dependencies, images and actions pinned
- [x] No prompt/response content or PII in logs, traces, metrics or events
- [x] Commits signed off (DCO) and Conventional
