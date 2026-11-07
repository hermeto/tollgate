# Release Checklist & Publishing Guide

## 1. Pre-Release Verification Gates

- [ ] Review stacked pull requests bottom-up and merge in order to `main`.
- [ ] Run `make lint test` locally from a clean working tree.
- [ ] Run `go test -race ./...` across all packages.
- [ ] Run simulator-only SDK compatibility suite (`make sdk-compat`).
- [ ] Verify that no real provider credentials or personal data are included.
- [ ] Ensure all required status checks in `ci-gate` are green.

## 2. Release & Artifact Publishing Scope

Tollgate supports both automated and manual release workflows:

### Automated Release (Standard Path)

- Conventional Commits on `main` trigger `.github/workflows/release-please.yml`.
- Merging the Release PR automatically creates the Git release tag (e.g. `v0.1.0`).
- The tag triggers `.github/workflows/release.yml` via `workflow_call`.

### Manual Dispatch (Maintainer Override)

- Maintainers can manually trigger `.github/workflows/release.yml` using `workflow_dispatch` with the target `tag` parameter (e.g. `v0.1.0`).

### Automated Artifact Generation

Both release triggers execute the following automated steps:

1. **Verification Gate**: Executes `go test ./...` and `go test -race ./...` before any build artifact is produced.
2. **Binary Compilation**: GoReleaser builds multi-platform binaries (`darwin/amd64`, `darwin/arm64`, `linux/amd64`, `linux/arm64`).
3. **Container Image**: Builds and pushes distroless multi-arch container images to GitHub Container Registry (`ghcr.io/hermeto/tollgate:<version>`).
4. **Supply Chain Security**:
   - Generates Syft Software Bill of Materials (SBOM) for all binaries and container images.
   - Signs release artifacts keylessly via Cosign with Sigstore.
   - Attests build provenance via GitHub Actions artifact attestations.
