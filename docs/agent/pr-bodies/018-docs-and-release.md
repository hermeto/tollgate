## Summary

Feature 018 closes the v1.0 backlog with documentation, changelog, comprehensive release guide, and an automated and manual release publishing workflow with GoReleaser, Syft SBOM, Cosign signatures, and provenance attestations.

## Base

Based on `017-report-generator`; merge after PR #34.

## Decisions for the reviewer

- **Release publishing workflow**: Restored `.github/workflows/release.yml` with both `workflow_call` (for automated releases triggered by `release-please`) and `workflow_dispatch` (for manual maintainer invocation).
- **Verification gates**: Pre-release verification job runs `go test ./...` and `go test -race ./...` before GoReleaser publishes multi-platform binaries, distroless multi-arch GHCR images, SBOMs, and Cosign signatures.
- **Security hardening**: Added non-root `USER 65532:65532` to `deploy/compose/gateway.Dockerfile` and restricted `securityContext` to Helm `deployment.yaml` to pass Trivy misconfiguration scanners cleanly.

## Verification

- `make lint test`
- `go test -race ./...`
- `actionlint .github/workflows/*.yml`
- `trivy fs --scanners misconfig --severity HIGH,CRITICAL .`
- `helm lint charts/tollgate`
