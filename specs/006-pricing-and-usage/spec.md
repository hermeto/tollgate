# Feature Specification: Pricing and Usage

## Clarifications

- Q: What is the billing source when a provider omits usage? → A: Use the documented token estimate and mark the event estimated (decided by agent: preserve accounting continuity without inventing provider usage).
- Q: Which persistence guarantees apply locally? → A: The asynchronous writer must flush on shutdown and retain bounded spool files when the database is unavailable (decided by agent: keep local development deterministic and recoverable).

## User Scenarios & Testing

### User Story 1 - Price lookup (P1)

An operator defines provider/model prices with currency, effective date and source. The gateway loads the table and computes a request cost using provider usage.

### User Story 2 - Usage event (P1)

After every completed request, including a stream, the gateway emits a content-free event attributed to identity, logical model, real target, usage, latency and cost.

### User Story 3 - Durable asynchronous delivery (P1)

The event writer batches events, applies bounded back-pressure, spools unavailable database writes locally, and flushes pending events during graceful shutdown.

## Functional Requirements

- FR-PRC-1: Price tables MUST validate USD currency, positive prices, model, provider and effective date.
- FR-PRC-2: Lookup MUST choose the latest price effective at request time and flag unknown models.
- FR-USG-1: Usage events MUST contain request ID, timestamp, identity context, logical/real model, provider, token counts, estimated flag, cost, latency, status and policy decisions, and MUST NOT contain prompt or response content.
- FR-USG-2: Reported provider usage MUST be preferred; estimates MUST be used only when usage is absent and MUST be flagged.
- FR-USG-3: The writer MUST be asynchronous, bounded and flush on shutdown; unavailable sinks MUST use a bounded on-disk spool.
- FR-USG-4: Cost computation MUST reconcile with raw reported usage within 1%.

## Success Criteria

- SC-001: Fixed simulator fixtures recompute event cost within 1%.
- SC-002: Usage write work adds less than 0.2 ms p99 to the request path in the benchmark.
- SC-003: Shutdown preserves every event accepted by the writer or reports bounded spool exhaustion explicitly.

## Assumptions

- Prices are USD per one million tokens; cached input uses its own rate when present.
- PostgreSQL delivery is represented by a sink interface so tests use a deterministic fixture sink.

## Out of Scope

Budgets, chargeback reports and provider routing are later features.

## Dependencies

Feature 005 identity context and feature 004 canonical usage fields.
