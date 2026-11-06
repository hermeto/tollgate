# Feature Specification: Rate Limiting

## Clarifications

- Q: What is the default implementation for local development? → A: An in-memory token bucket with the same interface as distributed storage (decided by agent: preserve local simulator workflows while leaving Redis integration replaceable).

## User Scenarios & Testing

### User Story 1 - Request and token quotas (P1)

An operator defines per-key, project and team request/token limits. The most restrictive applicable limit controls admission and a completed request reconciles its token charge.

### User Story 2 - Caller feedback (P1)

An over-limit caller receives protocol-correct 429 with retry and remaining-window headers.

## Functional Requirements

- FR-RL-1: Limits MUST support capacity, refill period and independent request/token dimensions.
- FR-RL-2: The most restrictive applicable limit MUST win.
- FR-RL-3: Denials MUST include retry-after, limit, remaining and reset metadata.
- FR-RL-4: The local strategy MUST be concurrency-safe and deterministic under an injected clock.

## Success Criteria

- SC-001: No test window admits more than capacity plus configured burst.
- SC-002: Race tests pass under concurrent callers.

## Out of Scope

Redis leases, traffic priorities and multi-region coordination are subsequent iterations.
