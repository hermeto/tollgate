# Feature Specification: Benchmark Harness

## User Scenarios & Testing

### User Story 1 - Reproducible performance evidence

Maintainers run a pinned local benchmark that reports gateway operation latency and allocations without contacting providers.

## Functional Requirements

- FR-BENCH-1: Benchmarks MUST use local handlers or simulators only.
- FR-BENCH-2: Results MUST report Go benchmark timing and allocations.
- FR-BENCH-3: The command MUST be runnable from a clean checkout.

## Success Criteria

- SC-001: Benchmark command exits successfully without provider credentials.
