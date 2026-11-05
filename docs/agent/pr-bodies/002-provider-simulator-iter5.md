## Summary

Slice 6 of 9 of the provider simulator: the Anthropic flavour, `POST /v1/messages`, non-streaming and
streaming with the official event sequence, `max_tokens` required and enforced (`stop_reason:
"max_tokens"`), forced tool use, and Anthropic error bodies. Adds the Anthropic fixtures (two copied
from the Anthropic Python SDK test suite, attributed in `NOTICE`).
Spec: [`specs/002-provider-simulator/spec.md`](https://github.com/hermeto/tollgate/blob/002-provider-simulator-iter5/specs/002-provider-simulator/spec.md) (US2). Tracking: #3.

## Stack

Based on #9; merge after it. Next slice: `002-provider-simulator-iter6` (latency model and faults).

## Requirements covered

Feature spec FR-003, FR-004 (Anthropic 400/401/404/413), FR-012 (Anthropic fixtures), FR-016, FR-017
(Anthropic), SC-002 and SC-004 for Anthropic.

## Decisions for the reviewer

- The SDK's two stream fixtures were recorded at different API versions (old usage object vs cache token fields, `service_tier` and tool `caller`). The simulator emits the current shape, and the contract test requires **every fixture field with the same type**, allowing newer fields. Tool `input` is treated as an opaque object.
- Correction to research R3: anthropic-sdk-python is Apache-2.0 (not MIT); `NOTICE` now attributes the copied fixtures.
- `tool_choice` `{"type":"any"}` picks the first tool; `{"type":"tool","name":X}` picks X; `auto`/`none` answer with text. The forced reply has no preamble text block (the real API does not add one when a tool is forced).
- Input tokens use the same documented approximation as OpenAI (system as a message, blocks flattened); the Anthropic tokenizer is not public, so the error is unknown and documented as such.
- `anthropic-version` is accepted but not enforced; both `x-api-key` and `Authorization: Bearer` authenticate.

## How it was verified

- `make lint test` clean. Tests: non-streaming shape (`message.json`), stream event order equals `message_stream.sse`, every event shape contains the fixture fields of both stream fixtures, tool use JSON and stream for `any` and `tool`, `auto` does not force; usage equals the counting method and the text; `max_tokens: 3` → `max_tokens` stop with 3 output tokens; 401/404 shapes; 400 for malformed JSON, missing or zero `max_tokens`, empty messages; block content counts the same as plain strings; replay determinism.

## Not verified

- Official SDK clients: slice 8.

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
