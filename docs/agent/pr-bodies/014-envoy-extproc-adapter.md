## Summary

Feature 014 adds a protocol-neutral Envoy ext_proc decision contract across request and response phases.

## Base

Based on `013-response-cache`; merge after PR #30.

## Verification

- `go test ./internal/extproc/... -race -cover`
- `go test ./...`
- `golangci-lint run ./...`
