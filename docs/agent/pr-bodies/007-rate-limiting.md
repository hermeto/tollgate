## Summary

Feature 007 adds a deterministic, concurrency-safe local token bucket for request/token admission,
including remaining capacity and reset metadata.

## Base

Based on `006-pricing-and-usage`; merge after PR #23.

## Requirements covered

FR-RL-1 through FR-RL-4; tasks T001–T004.

## Verification

- `go test ./internal/ratelimit/... -race -cover`
- `go test ./...`
- `golangci-lint run ./...`
