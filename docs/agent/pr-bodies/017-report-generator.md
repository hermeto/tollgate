## Summary

Feature 017 adds deterministic content-free usage aggregation and Markdown report rendering by team.

## Base

Based on `016-benchmark-harness`; merge after PR #33.

## Verification

- `go test ./internal/report/... -race -cover`
- `go test ./...`
- `golangci-lint run ./...`
