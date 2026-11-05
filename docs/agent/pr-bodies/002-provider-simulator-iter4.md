## Summary

Slice 5 of 9 of the provider simulator: OpenAI `POST /v1/embeddings` (deterministic unit vectors, float
or base64, `dimensions`), `GET /v1/models` and `GET /v1/models/{model}`, and the request log
(`GET`/`DELETE /_sim/requests`) that end-to-end tests use to assert what the gateway sent upstream.
Spec: [`specs/002-provider-simulator/spec.md`](https://github.com/hermeto/tollgate/blob/002-provider-simulator-iter4/specs/002-provider-simulator/spec.md) (US1, US6). Tracking: #3.

## Stack

Based on #8; merge after it. Next slice: `002-provider-simulator-iter5` (Anthropic flavour).

## Requirements covered

Feature spec FR-002 (embeddings, models), FR-009 (deterministic embeddings), FR-010, SC-002 and SC-004
for embeddings.

## Decisions for the reviewer

- Base64 embeddings are supported from day one because openai-python and openai-node request `encoding_format: base64` by default (research R2).
- Embedding vectors: PCG seeded by (seed, hash of model and input), normal samples normalised to unit length; usage = `tokens.Count` of each input.
- Token-array inputs are accepted and hashed as text (enough for compatibility; no tokenizer needed).
- CodeQL `go/uncontrolled-allocation-size` (first run): the 1–8192 bound checked the request field, not the value reaching `make`; it now applies to the final value and again inside `embed`, with a test for `dimensions: 1e9`.
- Request log: ring buffer of `request_log.size` records (default 1000), bodies capped at 64 KiB with `body_truncated`, headers verbatim (the simulator only ever sees fake keys); admin endpoints are not recorded; size 0 disables it.

## How it was verified

- `make lint test` clean. Embeddings: shape against both fixtures, unit norm, base64 decodes to the same floats as the float response, `dimensions: 256`, 400 for empty input, empty array, `dimensions: 0` and unknown encoding. Models: list shape, retrieve known/unknown (404 shape). Replay determinism now includes embeddings.
- Request log: ring order and eviction, clear, disabled at size 0, endpoint records method/path/headers/body/status, admin calls skipped, 401 recorded, large bodies truncated while the handler still reads the full body.

## Not verified

- Official SDK clients: slice 8.

## Performance impact

None on the gateway hot path. Recording copies at most 64 KiB per request in the simulator.

## Checklist

- [x] Tests added or updated
- [x] Docs updated in this PR (README, docs/, runbooks)
- [x] ADR added or updated for architectural decisions — ADR-0009 ships with the last slice
- [x] No secrets, real keys or real personal data
- [x] New dependencies, images and actions pinned
- [x] No prompt/response content or PII in logs, traces, metrics or events
- [x] Commits signed off (DCO) and Conventional
