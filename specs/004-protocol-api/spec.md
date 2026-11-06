# Feature Specification: Multi-protocol API and logical models

**Feature Branch**: `004-protocol-api`
**Status**: Draft

## Clarifications

- Q: Which provider features are translated in v1? → A: Text chat, system prompts, roles, tools, stop reasons, usage and streaming events; unsupported multimodal or provider-specific fields fail explicitly. (decided by agent: this is the smallest interoperable surface required by the playbook.)
- Q: How are logical models exposed? → A: Configuration maps a logical name to an ordered target list, and `/v1/models` exposes only logical names. (decided by agent: callers must not depend on provider model identifiers.)
- Q: What happens to unknown JSON fields? → A: Same-protocol passthrough preserves them; translation rejects fields it cannot represent instead of silently dropping them. (decided by agent: preservation protects SDK compatibility and explicit rejection protects correctness.)

## User Stories & Acceptance Scenarios

### US1 — OpenAI-compatible gateway (P1)

1. Given an OpenAI client, when it calls chat completions, embeddings or models, then the gateway returns the simulator response with request ID and provider status.
2. Given a streaming chat request, when chunks arrive, then the client receives protocol-valid SSE events and `[DONE]`.
3. Given tool calls or `response_format`, when the configured upstream supports them, then fields are preserved.

### US2 — Anthropic-compatible gateway (P1)

1. Given an Anthropic client, when it calls `/v1/messages`, then passthrough responses and streams remain valid Anthropic wire format.

### US3 — Logical models (P1)

1. Given `chat-default` mapped to a real target, when a caller sends that logical model, then the upstream receives the real model and the caller sees only the logical model.
2. Given an unknown model, then the caller receives a protocol-correct 404/400 error without an upstream request.

### US4 — Translation (P1)

1. Given an OpenAI chat request routed to Anthropic, then system prompts, roles, tools, stop reasons, usage and stream events translate both ways.
2. Given an unsupported feature, then the caller receives a clear protocol-correct error and no silent degradation occurs.

## Functional Requirements

- **FR-API-1**: Serve OpenAI chat completions, embeddings and models.
- **FR-API-2**: Serve Anthropic messages and streaming events.
- **FR-API-3**: Preserve unknown fields in same-protocol passthrough.
- **FR-API-4**: Map logical models to real provider/model targets and list only authorized logical models.
- **FR-API-5**: Translate supported OpenAI and Anthropic chat fields and reject unsupported fields explicitly.
- **FR-API-6**: Return caller-protocol errors and propagate `x-request-id` on every outcome.
- **FR-API-7**: Keep raw request bytes available to policy stages while exposing one canonical internal request.

## Success Criteria

- **SC-001**: Official OpenAI and Anthropic SDK compatibility tests pass through the gateway for passthrough routes.
- **SC-002**: Table-driven wire-fixture tests cover translation of text, tools, stop reasons, usage and streams.
- **SC-003**: Native JSON and SSE fuzz tests run for at least 60 seconds in the release validation workflow.
- **SC-004**: Unsupported translation features produce a protocol-correct error in under 10 ms without an upstream call.

## Edge Cases

- Partial JSON and multiple events in one SSE frame.
- Very large tool schemas bounded by the configured request limit.
- Unknown fields, empty content blocks, omitted usage and provider-specific stop reasons.

## Key Entities

- Canonical request, message, content part, tool, response and usage.
- Logical model and provider target.

## Assumptions

Responses API, Batch API and new multimodal translation are v2; same-protocol passthrough may carry them.
