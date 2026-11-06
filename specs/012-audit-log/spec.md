# Feature Specification: Audit Log

## User Scenarios & Testing

### User Story 1 - Tamper-evident governance events

Operators can record key, policy and authorization actions without storing request content.

### User Story 2 - Integrity verification

An operator verifies the event chain and detects modification or deletion.

## Functional Requirements

- FR-AUD-1: Events MUST contain timestamp, action, actor/resource metadata and chained hashes.
- FR-AUD-2: Verification MUST detect altered content and broken predecessor links.
- FR-AUD-3: Prompt, response and bearer secrets MUST never be recorded.

## Success Criteria

- SC-001: A valid chain verifies.
- SC-002: A tampered persisted event fails verification.
