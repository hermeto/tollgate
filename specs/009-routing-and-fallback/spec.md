# Feature Specification: Routing and Fallback

## User Scenarios & Testing

### User Story 1 - Weighted healthy routing

A logical model selects among healthy targets using configured weights.

### User Story 2 - Bounded retries

Transient upstream failures use exponential jittered backoff within a request retry budget and never retry after a stream begins.

## Functional Requirements

- FR-RT-1: Unhealthy targets MUST not be selected.
- FR-RT-2: Selection MUST honor positive weights over repeated requests.
- FR-RT-3: Backoff MUST be bounded and deterministic in its configured limits.

## Success Criteria

- SC-001: A fully unhealthy target set fails fast.
- SC-002: Retry tests prove no post-first-byte retry.
