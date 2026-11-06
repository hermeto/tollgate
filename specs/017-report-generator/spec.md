# Feature Specification: Report Generator

## User Scenarios & Testing

### User Story 1 - Reproducible governance report

FinOps and platform teams generate a deterministic Markdown summary of requests and cost by team from content-free usage events.

## Functional Requirements

- FR-REP-1: Reports MUST aggregate event cost without request content.
- FR-REP-2: Output MUST be deterministic and sorted by owner.
- FR-REP-3: Report totals MUST reuse event costs rather than recompute from text.

## Success Criteria

- SC-001: Golden fixture output is stable across runs.
