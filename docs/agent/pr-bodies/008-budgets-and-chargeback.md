## Summary

Feature 008 adds concurrent spend reservations with soft/hard decisions and a content-free,
FOCUS-aligned CSV chargeback writer.

## Base

Based on `007-rate-limiting`; merge after PR #24.

## Requirements covered

FR-FIN-3 through FR-FIN-5; tasks T001–T005.

## Verification

- `go test ./internal/budget/... ./internal/finops/... -race -cover`
- `go test ./...`
- `golangci-lint run ./...`
