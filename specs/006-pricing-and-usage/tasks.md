# Tasks: 006 pricing-and-usage

- [x] T001 [P] Add price table examples and ADR-0006 in `pricing/examples` and `docs/adr/0006-cost-export-format-focus.md`.
- [x] T002 [P] Add feature design artifacts in `specs/006-pricing-and-usage/`.
- [x] T003 [P] Implement effective-dated price loading and cost lookup in `internal/pricing/pricing.go`.
- [x] T004 [P] Add price validation and reconciliation tests in `internal/pricing/pricing_test.go`.
- [x] T005 Implement content-free usage event model in `internal/usage/usage.go`.
- [x] T006 Implement bounded asynchronous JSONL spool writer and shutdown flush in `internal/usage/usage.go`.
- [x] T007 Add usage writer replay and durability tests in `internal/usage/usage_test.go`.
- [x] T008 Run race, coverage and lint verification and update `docs/agent/PROGRESS.md`.

## Dependencies

T001–T004 are independent of T005–T007. T008 follows all implementation tasks.
