## Summary

Feature 011 adds deterministic high-confidence synthetic PII detection and redaction with content-free
match metadata.

## Base

Based on `010-observability`; merge after PR #27.

## Verification

- `go test ./internal/pii/... -race -cover`
- `go test ./...`
- `golangci-lint run ./...`
