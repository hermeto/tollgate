# Feature Specification: Budgets and Chargeback

## User Scenarios & Testing

### User Story 1 - Budget enforcement (P1)

Teams and projects reserve estimated spend before a request. Soft thresholds warn and hard thresholds block without allowing concurrent overspend.

### User Story 2 - Chargeback export (P1)

FinOps exports content-free usage events with cost, provider, model and owner fields to CSV with stable FOCUS-aligned columns.

## Functional Requirements

- FR-FIN-3: Reservations MUST include in-flight spend and reject requests exceeding hard limits.
- FR-FIN-4: Soft threshold decisions MUST be observable without blocking.
- FR-FIN-5: Chargeback output MUST reconcile to usage event costs and include USD currency.

## Success Criteria

- SC-001: Concurrent reservation tests never exceed the hard limit.
- SC-002: Export totals equal input event totals exactly to cent precision.

## Out of Scope

ERP integration and currency conversion.
