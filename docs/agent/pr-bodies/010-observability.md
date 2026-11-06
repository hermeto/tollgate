## Summary

Feature 010 centralizes content-free GenAI telemetry attributes and adds versioned dashboard and alert
artifacts for governance overhead.

## Base

Based on `009-routing-and-fallback`; merge after PR #26.

## Verification

- `go test ./internal/observability/... -race -cover`
- `go test ./...`
- `golangci-lint run ./...`
