# Implementation Plan: Open-source foundation, CI and contributor experience

**Branch**: `001-repo-foundation` | **Date**: 2026-11-04 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/001-repo-foundation/spec.md`

## Summary

Lay down the repository's governance files, templates and repository-as-code, a pinned local toolchain
that works without Homebrew, and GitHub Actions workflows: `ci.yml` (lint, test, security scan, PR
title, DCO, dependency review, aggregated `ci-gate`), `codeql.yml`, `scorecard.yml`,
`release-please.yml`, `release.yml` and `labels.yml`. A minimal Go module with a `tollgate` binary
that prints its version gives CI something real to lint, test and release.

## Technical Context

**Language/Version**: Go 1.27.3 (go.mod `go` directive pins the exact toolchain); Python only via uv
for pre-commit.

**Primary Dependencies**: none at runtime. Dev tools pinned in `tools/go.mod` and run with
`go tool -modfile=tools/go.mod`: golangci-lint v2.15.0, govulncheck v1.9.0, editorconfig-checker
v4.0.2. Pre-commit hooks pinned by commit SHA: pre-commit-hooks v6.0.0, shellcheck-py v0.11.0.1,
yamllint v1.38.0, markdownlint-cli2 v0.24.0, actionlint v1.7.12.

**Storage**: N/A.

**Testing**: `go test -race ./...`; workflow logic verified by actionlint and by a unit-testable
`ci-gate` script (`scripts/ci-gate.sh`) and DCO script (`scripts/check-dco.sh`) exercised by Go tests
in `internal/repotest`.

**Target Platform**: GitHub Actions `ubuntu-24.04` runners; contributor machines on macOS and Linux.

**Project Type**: single Go module (CLI binary), repository tooling.

**Performance Goals**: `make lint test` on a fresh clone < 5 min with warm caches (SC-001).

**Constraints**: no secrets required on PRs (fork-safe); top-level `permissions: {}` or read-only;
actions pinned by SHA; concurrency groups; timeouts on every job.

**Scale/Scope**: ~25 files of docs/config, 6 workflows, 2 small scripts, 1 Go main package.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
| --- | --- | --- |
| I. Evidence over opinion | Pass | Badges and CI results are the evidence; no claims without links. |
| II. Hot path is sacred | Pass (N/A) | No gateway code; `bench-micro` target exists for later features. |
| III. Privacy by default | Pass | No content handling. Issue forms warn against pasting prompts or keys. |
| IV. Fail modes explicit | Pass | `ci-gate` treats skipped/cancelled as failure (explicit, tested). |
| V. Governance as code | Pass | Ruleset, labels, CODEOWNERS, templates versioned; labels synced by workflow. |
| VI. Local-first, zero cost | Pass | `make dev-setup && make lint test` needs no secrets or paid APIs. |
| VII. Reproducible | Pass | Go tools in `tools/go.mod`, hooks by SHA, actions by SHA, Go version exact. |
| VIII. Test what matters | Pass | Scripts that gate merges (`ci-gate`, DCO) have Go tests. |
| IX. Secure supply chain | Pass | CodeQL, Scorecard, govulncheck, Trivy, dependency review; signed releases with SBOM and provenance; least-privilege permissions. |
| X. Simplicity | Pass | Plain GitHub Actions and Make; no custom actions. |
| XI. OSS citizenship | Pass | DCO check, Conventional PR titles, CoC, contribution guide. |

Post-design re-check: Pass (no deviations; Complexity Tracking empty).

## Project Structure

### Documentation (this feature)

```text
specs/001-repo-foundation/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── ci-gate.md
│   └── make-targets.md
├── checklists/
└── tasks.md
```

### Source Code (repository root)

```text
cmd/tollgate/main.go            # prints version (ldflags-injected)
internal/version/version.go     # Version, Commit, Date vars
internal/repotest/              # tests for scripts/ and repo invariants (SHA pinning, permissions)
scripts/ci-gate.sh              # aggregates needs.*.result
scripts/check-dco.sh            # Signed-off-by on non-merge commits in a range
tools/go.mod                    # pinned Go dev tools (go tool -modfile)
Makefile
Dockerfile                      # distroless runtime image used by GoReleaser
.goreleaser.yaml
.golangci.yml  .markdownlint-cli2.yaml  .yamllint  .pre-commit-config.yaml
release-please-config.json  .release-please-manifest.json
.github/
├── workflows/{ci,codeql,scorecard,release-please,release,labels}.yml
├── ISSUE_TEMPLATE/{bug,feature,docs,policy-proposal}.yml + config.yml
├── PULL_REQUEST_TEMPLATE.md
├── CODEOWNERS
├── dependabot.yml
├── labels.yml
└── rulesets/main.json
docs/adr/0000-record-architecture-decisions.md
docs/adr/template.md
CODE_OF_CONDUCT.md CONTRIBUTING.md SECURITY.md SUPPORT.md GOVERNANCE.md MAINTAINERS.md
```

**Structure Decision**: single Go module at the repository root following Appendix A §A.8; merge-gating
logic lives in shell scripts under `scripts/` so it is testable locally instead of hidden in YAML.

## Complexity Tracking

No constitution violations.
