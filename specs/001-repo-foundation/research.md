# Research: 001 repo-foundation

All versions resolved on 2026-11-04 from each project's latest GitHub release (`gh release view`) and
pinned by full commit SHA (actions, pre-commit hooks) or exact version (Go tools, images by digest).

## R1. Pinned GitHub Actions

| Action | Version | Commit SHA |
| --- | --- | --- |
| actions/checkout | v7.0.1 | 3d3c42e5aac5ba805825da76410c181273ba90b1 |
| actions/setup-go | v7.0.0 | b7ad1dad31e06c5925ef5d2fc7ad053ef454303e |
| astral-sh/setup-uv | v10.3.3 | 68559e99cc807c1eb7e9238fd43a15bba460eec2 |
| actions/cache | v6.1.0 | 55cc8345863c7cc4c66a329aec7e433d2d1c52a9 |
| amannn/action-semantic-pull-request | v6.1.1 | 48f256284bd46cdaab1048c3721360e808335d50 |
| actions/dependency-review-action | v5.0.0 | a1d282b36b6f3519aa1f3fc636f609c47dddb294 |
| github/codeql-action | v4.39.3 | 33d7b85da1a04171e211befe2c739303b9a4c066 |
| ossf/scorecard-action | v2.5.0 | fa7117db83c52d787ae5a89f350999701505d78e |
| googleapis/release-please-action | v5.0.0 | 45996ed1f6d02564a971a2fa1b5860e934307cf7 |
| goreleaser/goreleaser-action | v7.3.0 | a4f02ab5355a589de60d917a659ce617ed19b3de |
| sigstore/cosign-installer | v4.1.3 | 756fe2d284c7292c4e9b62c8a28f1e5629c59532 |
| anchore/sbom-action (download-syft) | v0.24.3 | 66cbf4bc1f1c0d2edc94016e65bc221b6bb0ad6c |
| docker/setup-qemu-action | v4.4.0 | 99012661954931238ded8c8b007157a8430204e1 |
| docker/setup-buildx-action | v4.4.1 | f87e5991a6d7451dcb8d9637bfbc97413f497069 |
| docker/login-action | v4.7.0 | 81c6f93d4bc991e510b342a00b419b3bb4efc5c6 |
| actions/attest-build-provenance | v4.2.2 | 4d101475d8b20a2381f78447822ac1eab6504dd8 |
| actions/upload-artifact | v7.0.2 | cf430e030ddbb5b0abf93d22962f4752f3646cd9 |
| crazy-max/ghaction-github-labeler | v6.0.0 | 548a7c3603594ec17c819e1239f281a3b801ab4d |

- **Decision**: pin every `uses:` to a full SHA with a `# vX.Y.Z` comment; Dependabot's
  `github-actions` ecosystem keeps both in sync.
- **Rationale**: tags are mutable; SHA pinning is a Scorecard "Pinned-Dependencies" requirement.
- **Alternatives considered**: major-version tags (rejected: mutable).

## R2. Trivy: container image instead of `aquasecurity/trivy-action`

- **Decision**: run Trivy 0.76.1 from `aquasec/trivy@sha256:790fd66d1697395d8693d1acaa1ea702406a20f109b6fc7bac8db623eca6c06c`
  with `docker run` in the `security` job (filesystem scan for vulnerabilities, secrets and
  misconfigurations; fails on HIGH/CRITICAL).
- **Rationale**: an image pinned by digest is immutable and needs no third-party action code with
  repository token access; the trivy-action repository had tags rewritten in a 2026 supply-chain
  incident, which SHA pinning mitigates but a digest-pinned image avoids entirely.
- **Alternatives considered**: trivy-action pinned by SHA (acceptable, more moving parts).

## R3. Go developer tools via `go tool -modfile=tools/go.mod`

- **Decision**: a separate `tools/go.mod` with `tool` directives for
  `github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0`, `golang.org/x/vuln/cmd/govulncheck@v1.1.4`,
  `github.com/rhysd/actionlint/cmd/actionlint@v1.7.12` and
  `github.com/editorconfig-checker/editorconfig-checker/v4/cmd/editorconfig-checker@v4.0.2`.
- **Rationale**: the `tool` directive (Go ≥ 1.24) pins tools in go.sum without polluting the main
  module's dependency graph; the same command works on macOS, Linux and CI without Homebrew (FR-011).
- **Alternatives considered**: `tools.go` blank imports in the main module (pollutes go.mod), Homebrew
  (macOS-only), golangci-lint-action in CI (would diverge from local; we use the same `make lint`).

## R4. Non-Go linters via pre-commit

- **Decision**: pre-commit 4.x (installed with `uv tool install pre-commit`), hooks pinned by SHA:
  pre-commit-hooks v6.0.0 (3e8a870…), shellcheck-py v0.11.0.1 (745efac…), yamllint v1.38.0 (cba56bc…),
  markdownlint-cli2 v0.24.0 (341e2a3…). Actionlint and editorconfig-checker run from `go tool` as
  local hooks so there is one pinned version of each.
- **Rationale**: pre-commit manages hook runtimes (Python, Node via nodeenv) itself, so Linux users
  without Node or Homebrew still get pinned linters.
- **Alternatives considered**: installing each linter globally (unpinned, platform-specific).

## R5. `ci-gate` aggregation

- **Decision**: `ci-gate` job with `needs: [lint, test, security, pr-title, dco, dependency-review]` and
  `if: always()`, running `scripts/ci-gate.sh` with `toJSON(needs)` and the event name. The script
  fails if any result is not `success`, except `skipped` for PR-only jobs on non-PR events.
- **Rationale**: GitHub treats a skipped required check as passing, so a plain `needs:` job would turn
  green when a dependency is skipped; `if: always()` plus an explicit result check closes that hole.
  Putting the logic in a script makes it unit-testable.
- **Source**: GitHub Docs, "Using jobs in a workflow — defining prerequisite jobs" and "Contexts —
  needs context" (https://docs.github.com/en/actions/reference/workflows-and-actions/contexts#needs-context).

## R6. DCO check without a third-party app

- **Decision**: `scripts/check-dco.sh <base> <head>` iterates `git rev-list --no-merges base..head` and
  requires a `Signed-off-by:` trailer (via `git interpret-trailers --parse`) on each commit; checkout
  uses `fetch-depth: 0`.
- **Rationale**: playbook requires no external app; trailer parsing avoids false positives in bodies.

## R7. Release pipeline

- **Decision**: release-please (`release-type: go`, manifest mode) on pushes to `main` opens/updates
  the release PR; `release.yml` triggers on `v*` tags and runs GoReleaser v2.20.0 with
  `dockers_v2` (multi-arch distroless image on GHCR), syft SBOMs, keyless cosign signing of checksums
  and images, and `actions/attest-build-provenance` for binaries and image.
- **Base image**: `gcr.io/distroless/static-debian13:nonroot@sha256:25ac8f0c3b2334a9eb05452266bd4697864a8c8ef0ef71e5e6480d7f92868329`
  (linux/amd64 and linux/arm64/v8).
- **Rationale**: Conventional Commits already drive versioning; keyless signing avoids stored keys.
- **Sources**: https://github.com/googleapis/release-please-action, https://goreleaser.com/customization/docker_v2/,
  https://docs.sigstore.dev/cosign/signing/overview/, https://github.com/actions/attest-build-provenance.

## R8. Ruleset, labels and settings as code

- **Decision**: `.github/rulesets/main.json` exactly as in playbook §4.4; `.github/labels.yml` in the
  ghaction-github-labeler format, synced on pushes to `main` that touch the file (needs
  `issues: write`; never runs on PRs).
- **Rationale**: keeps settings reviewable (Principle V) and fork-safe.

## R9. Badges

- **Decision**: README badges for CI (`actions/workflows/ci.yml/badge.svg`), OpenSSF Scorecard
  (`api.scorecard.dev/projects/github.com/hermeto/tollgate/badge`), license and Go Report Card
  (`goreportcard.com/badge/github.com/hermeto/tollgate`). They resolve once the repository is public.
