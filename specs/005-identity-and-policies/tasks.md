---
description: "Tasks for identity and declarative policies"
---

# Tasks: 005 identity-and-policies

## Phase 1: Setup

- [x] T001 [P] Add policy schema and synthetic policies under policies/examples.
- [x] T002 [P] Add identity and policy decision ADRs/documentation.

## Phase 2: User Story 1 — Virtual keys

- [x] T003 [P] Implement internal/identity key generation, SHA-256 storage, constant-time authentication and expiry.
- [x] T004 Implement atomic JSON persistence and create/list/rotate/revoke operations.
- [x] T005 Add `tollgate keys create|list|rotate|revoke` with once-only secret output.
- [x] T006 Integrate optional gateway authentication and identity context with 401 responses.

## Phase 3: User Story 2 — Policies

- [x] T007 [P] Implement internal/policy schema validation and deterministic org/team/project merge with provenance.
- [x] T008 Implement atomic policy reload with invalid-update retention and transition events.
- [x] T009 Add `tollgate policy validate|explain` commands.

## Phase 4: User Story 3 — Model governance

- [x] T010 Add model allowlist stage with enforce and dry-run decisions and 403 protocol errors.
- [x] T011 Add auth/policy e2e, inheritance property tests and synthetic admin-action events.

## Phase 5: Verification

- [x] T012 Run race/coverage tests, CLI examples and update docs/agent/PROGRESS.md.
