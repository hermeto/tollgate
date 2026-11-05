## Summary

Slice 4 of 8 of the provider simulator: OpenAI-compatible `POST /v1/chat/completions`, non-streaming and
streaming, with `stream_options.include_usage`, max-token truncation (`finish_reason: "length"`) and
forced tool calls. Adds the OpenAI wire fixtures (`testdata/wire/openai/`, sources in its README) and
structural contract tests against them.
Spec: [`specs/002-provider-simulator/spec.md`](https://github.com/hermeto/tollgate/blob/002-provider-simulator-iter3/specs/002-provider-simulator/spec.md) (US1). Tracking: #3.

## Stack

Based on #7; merge after it. Next slice: `002-provider-simulator-iter4` (embeddings, models, request log).

## Requirements covered

Feature spec FR-002 (chat part), FR-004 (OpenAI 400/401/404/413 bodies), FR-006, FR-008 (usage), FR-012
(OpenAI fixtures), FR-017 (OpenAI), SC-002 (OpenAI chat), SC-004 (OpenAI chat).

## Decisions for the reviewer

- Contract tests compare **shape** (keys, value types, first array element, event order with repeated content chunks collapsed), not values: ids, timestamps and text legitimately differ.
- Fixtures are transcribed from the API reference examples and the SDK type definitions at openai-python v3.46.0 / openai-node v7.51.0; nothing was captured from a paid API.
- `usage` is omitted from chunks unless `include_usage` is set; then it is `null` on every chunk and the final chunk has `choices: []`.
- `max_completion_tokens` takes precedence over `max_tokens`; `n > 1` returns 400 (documented as unsupported).
- Prompt tokens = `tokens.CountChat` over message text (string or text parts) plus the tool definitions' JSON; tool-call completion tokens = tokens of the name plus the arguments.
- SSE fixtures are excluded from `end-of-file-fixer`: the trailing blank line dispatches the last event.
- The model is echoed as requested (the real API returns a dated snapshot name).

## How it was verified

- `make lint test` clean. Tests: shape against `chat_completion.json`, both stream fixtures, tool-call JSON and stream (for `required` and a named function), `auto` does not force a call; usage equals the counting method and the returned text; `length` cut for both limit fields; streamed usage equals streamed text; 401/404/400/413 bodies; replay with the same seed is byte-identical apart from `id`/`created`, a different seed differs; a stream stops when the client is gone.

## Not verified

- Official SDK clients against this endpoint: the compatibility suite lands in slice 8.

## Performance impact

None on the gateway hot path.

## Checklist

- [x] Tests added or updated
- [x] Docs updated in this PR (README, docs/, runbooks)
- [x] ADR added or updated for architectural decisions — ADR-0009 ships with the last slice
- [x] No secrets, real keys or real personal data
- [x] New dependencies, images and actions pinned
- [x] No prompt/response content or PII in logs, traces, metrics or events
- [x] Commits signed off (DCO) and Conventional
