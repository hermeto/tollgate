## Summary

Feature 009 adds concurrency-safe weighted healthy-target selection and bounded jittered retry backoff.

## Base

Based on `008-budgets-and-chargeback`; merge after PR #25.

## Verification

- `go test ./internal/routing/... -race -cover`
- `go test ./...`
- `golangci-lint run ./...`
