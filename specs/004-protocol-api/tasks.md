---
description: "Tasks for multi-protocol API and logical models"
---

# Tasks: 004 protocol-api

## Phase 1: Setup

- [x] T001 [P] Add schemas/models.schema.json and config model types for logical targets.
- [x] T002 [P] Add synthetic OpenAI and Anthropic wire fixtures under testdata/wire.

## Phase 2: Foundational canonical model

- [x] T003 [US1] Add canonical request, message, tool, response and usage types in internal/canonical.
- [x] T004 [US1] Add bounded JSON helpers and protocol error types in internal/api.

## Phase 3: User Story 1 — OpenAI API

- [x] T005 [P] [US1] Implement OpenAI chat, embeddings and models adapters in internal/api/openai with fixture tests.
- [x] T006 [US1] Wire OpenAI routes and logical-model target selection into internal/server.

## Phase 4: User Story 2 — Anthropic API

- [x] T007 [P] [US2] Implement Anthropic messages and SSE event adapters in internal/api/anthropic with fixture tests.
- [x] T008 [US2] Wire Anthropic passthrough routes into internal/server with protocol errors and request IDs.

## Phase 5: User Story 4 — Translation

- [x] T009 [P] [US4] Implement OpenAI↔Anthropic canonical conversion and support matrix in internal/translate.
- [x] T010 [US4] Add table-driven translation tests for roles, tools, stop reasons, usage and streaming events.
- [x] T011 [US4] Reject unsupported translation fields with protocol-correct errors and tests.

## Phase 6: User Story 3 — Logical models and validation

- [x] T012 [US3] Implement logical model loading, `/v1/models` filtering and unknown-model errors.
- [x] T013 [US3] Preserve raw same-protocol unknown fields and retain canonical request for later policies.

## Phase 7: Verification

- [x] T014 [P] Add native fuzz tests for JSON adapters and SSE events.
- [ ] T015 Run SDK compatibility, race tests and 60-second fuzz validation; update docs/agent/PROGRESS.md.
