# Feature Specification: Envoy External Processing Adapter

## User Scenarios & Testing

### User Story 1 - Protocol-neutral processing

Envoy can ask Tollgate to inspect request/response headers and bodies and receive continue, mutate or deny decisions.

### User Story 2 - Safe failure

Malformed or denied traffic receives an explicit status and never leaks mutable buffers across protocol boundaries.

## Functional Requirements

- FR-EXT-1: Decisions MUST identify the processing phase and action.
- FR-EXT-2: Deny bodies MUST be copied before returning to the transport adapter.
- FR-EXT-3: The adapter MUST remain independent of a real provider.

## Success Criteria

- SC-001: Contract tests cover all four processing phases.
- SC-002: Mutation tests prove body ownership.
