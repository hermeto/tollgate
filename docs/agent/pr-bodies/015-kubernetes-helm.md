## Summary

Feature 015 adds a renderable Helm chart with explicit image version, service and health probes.

## Base

Based on `014-envoy-extproc-adapter`; merge after PR #31.

## Verification

- `helm lint charts/tollgate`
- `helm template tollgate charts/tollgate`
- `go test ./...`
