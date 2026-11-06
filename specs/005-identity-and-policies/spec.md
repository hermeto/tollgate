# Feature Specification: Identity and declarative policies

**Feature Branch**: `005-identity-and-policies`

## Clarifications

- Q: Where are virtual keys stored in this first implementation? → A: A JSON file store with atomic replacement is the local profile default; the identity interface permits Redis/PostgreSQL replication later. (decided by agent: local-first operation is required and the state-store ADR already reserves shared stores.)
- Q: What credential is accepted by the gateway? → A: `Authorization: Bearer tg_<prefix><secret>`; only SHA-256 is persisted and the full secret is shown once. (decided by agent: high-entropy random keys make a fast hash safe and avoid recoverable secrets.)
- Q: How are policy inheritance conflicts resolved? → A: org defaults, then team, then project; later explicit fields replace earlier fields, and provenance is retained. (decided by agent: deterministic reviewable merge semantics.)

## User Stories & Acceptance Scenarios

### US1 — Virtual keys (P1)

1. Admin creates a key for team/project with optional expiry; plaintext is shown once and never stored.
2. Admin lists visible prefixes, rotates and revokes keys; a revoked key fails within one second locally.
3. Gateway accepts valid keys and returns 401 for missing, expired or revoked keys.

### US2 — Policies (P1)

1. Admin validates org/team/project YAML and references.
2. Admin explains the effective policy and provenance for a team/project.
3. Hot reload rejects invalid files and retains the previous valid snapshot.

### US3 — Model governance (P1)

1. A team model allowlist returns 403 for a disallowed logical model.
2. Dry-run records the decision without denying the request.

## Functional Requirements

- **FR-ID-1**: Create, list, rotate and revoke virtual keys with expiry and visible prefixes.
- **FR-ID-2**: Store only SHA-256 key hashes and authenticate in constant time.
- **FR-ID-3**: Enforce 401 for invalid identity and propagate team/project identity to the request.
- **FR-POL-1**: Load org→team→project YAML with explicit deterministic inheritance and provenance.
- **FR-POL-2**: Validate policy schema, references and model allowlists.
- **FR-POL-3**: Hot reload atomically and preserve the last valid snapshot on invalid updates.
- **FR-POL-4**: Enforce or dry-run model allowlists with protocol-correct 403 errors.
- **FR-AUD-5**: Emit structured admin-action events for key and policy changes.

## Success Criteria

- **SC-001**: Auth and policy e2e tests cover 401, revoked/expired key, allowlist 403, dry-run and reload.
- **SC-002**: Property tests verify inheritance is associative and provenance identifies the winning file.
- **SC-003**: Local key revocation is effective within one second and authentication adds less than 1 ms p99.

## Edge Cases

Duplicate team files, concurrent reload, clock skew, malformed YAML, empty allowlists and key rotation during traffic.

## Assumptions

OIDC/JWT and distributed revocation notifications are interfaces reserved for a later deployment increment;
the local JSON store is authoritative when configured.
