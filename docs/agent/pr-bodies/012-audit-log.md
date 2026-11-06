## Summary

Feature 012 adds content-free tamper-evident audit events with SHA-256 predecessor chaining and
verification.

## Base

Based on `011-pii-redaction`; merge after PR #28.

## Verification

- `go test ./internal/audit/... -race -cover`
- `go test ./...`
- `golangci-lint run ./...`
