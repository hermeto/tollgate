## Summary

Feature 016 adds a provider-free Go benchmark and shell runner for reproducible local gateway evidence.

## Base

Based on `015-kubernetes-helm`; merge after PR #32.

## Verification

- `./scripts/bench.sh`
- `go test ./...`
- `golangci-lint run ./...`
