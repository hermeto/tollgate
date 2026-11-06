## Summary

Feature 013 adds bounded exact-response caching with SHA-256 request keys, TTL expiry and copied bodies.

## Base

Based on `012-audit-log`; merge after PR #29.

## Verification

- `go test ./internal/cache/... -race -cover`
- `go test ./...`
- `golangci-lint run ./...`
