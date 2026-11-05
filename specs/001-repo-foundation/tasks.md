---

description: "Task list for 001 repo-foundation"
---

# Tasks: Open-source foundation, CI and contributor experience

**Input**: Design documents from `specs/001-repo-foundation/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Included — the constitution (VIII) requires tests for merge-gating logic; `ci-gate` and DCO
scripts are tested first.

**Organization**: Tasks are grouped by user story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1–US5)

## Phase 1: Setup (Shared Infrastructure)

- [ ] T001 Initialize Go module `github.com/hermeto/tollgate` with `go 1.27.3` in go.mod
- [ ] T002 [P] Create pinned dev-tools module in tools/go.mod (golangci-lint v2.15.0, govulncheck v1.1.4, actionlint v1.7.12, editorconfig-checker v4.0.2)
- [ ] T003 [P] Add linter configs: .golangci.yml, .markdownlint-cli2.yaml (MD013 off, MD024 siblings_only), .yamllint
- [ ] T004 Add .pre-commit-config.yaml with hooks pinned by SHA (research R4) and local hooks for actionlint and editorconfig-checker

## Phase 2: Foundational (Blocking Prerequisites)

- [ ] T005 Create Makefile with help, dev-setup, fmt, lint, test, bench-micro per contracts/make-targets.md
- [ ] T006 [P] Create internal/version/version.go (Version, Commit, Date) and cmd/tollgate/main.go with `version` subcommand and `--version` flag printing `tollgate <version> (commit <sha>, built <date>)`
- [ ] T007 [P] Unit test for version output in cmd/tollgate/main_test.go

**Checkpoint**: `make lint test` runs locally.

## Phase 3: User Story 1 — Contributor can set up, verify and propose a change (P1)

**Goal**: fresh clone → `make dev-setup && make lint test` green; templates guide issues and PRs.

**Independent Test**: run quickstart.md "Validate locally" on a fresh clone.

- [ ] T008 [P] [US1] Write CODE_OF_CONDUCT.md (Contributor Covenant 2.1, contact via private reporting/maintainer email alias)
- [ ] T009 [P] [US1] Write CONTRIBUTING.md (dev setup for macOS and Linux, Spec Kit workflow, stacked branches and PRs, DCO sign-off, Conventional Commits, performance-budget rule, links to every governance and config file)
- [ ] T010 [P] [US1] Write SUPPORT.md, GOVERNANCE.md, MAINTAINERS.md and .github/CODEOWNERS (`@hermeto`)
- [ ] T011 [P] [US1] Create issue forms .github/ISSUE_TEMPLATE/{bug,feature,docs,policy-proposal}.yml with the "no prompts, keys or personal data" warning, and config.yml (blank issues off, Discussions and private reporting links)
- [ ] T012 [P] [US1] Create .github/PULL_REQUEST_TEMPLATE.md with playbook §5.3 sections
- [ ] T013 [US1] Run `make dev-setup && make lint test` on a fresh clone of the branch and record timing

## Phase 4: User Story 2 — Maintainer sees one required status (P1)

**Goal**: CI on any base; `ci-gate` aggregates and fails on failure, cancel or unexpected skip.

**Independent Test**: `go test ./internal/repotest/...` covers the gate rule; actionlint passes.

- [ ] T014 [P] [US2] Tests for scripts/ci-gate.sh per contracts/ci-gate.md in internal/repotest/cigate_test.go (all success, one failure, cancelled, skipped on PR, skipped PR-only on push, empty input)
- [ ] T015 [P] [US2] Tests for scripts/check-dco.sh in internal/repotest/dco_test.go using a temporary git repository (signed, unsigned, merge commit, trailer only in body)
- [ ] T016 [US2] Implement scripts/ci-gate.sh in Bash with jq (preinstalled on GitHub runners; checked by dev-setup)
- [ ] T017 [US2] Implement scripts/check-dco.sh
- [ ] T018 [US2] Create .github/workflows/ci.yml (lint job runs `make dev-setup` on Linux to prove the no-Homebrew path): lint, test, security (Trivy image by digest), pr-title, dco, dependency-review, ci-gate; pull_request on any base and push to main; read-only permissions, concurrency, timeouts
- [ ] T019 [P] [US2] Repository invariant tests in internal/repotest/workflows_test.go: every `uses:` pinned to a 40-hex SHA with a version comment, every workflow has top-level `permissions`, `concurrency`, and per-job `timeout-minutes`
- [ ] T020 [P] [US2] Add .github/rulesets/main.json (playbook §4.4) and .github/labels.yml (playbook §4.4 labels) with .github/workflows/labels.yml sync on push to main
- [ ] T021 [P] [US2] Add .github/dependabot.yml (gomod root and tools/, uv, github-actions, docker; weekly; grouped; Conventional prefixes)
- [ ] T022 [P] [US2] Add docs/adr/0000-record-architecture-decisions.md and docs/adr/template.md (MADR)

## Phase 5: User Story 3 — Security reporter (P1)

**Independent Test**: SECURITY.md contains channel, scope, out of scope, supported versions and response times.

- [ ] T023 [US3] Write SECURITY.md (private vulnerability reporting link, in-scope classes: gateway, policy bypass, PII leakage, audit tampering; out of scope; supported versions; response times from Clarifications)
- [ ] T024 [P] [US3] Add .github/workflows/codeql.yml (Go, Python, Actions; push to main, PRs, weekly) — `security-events: write` on the job only

## Phase 6: User Story 4 — Releases with a verifiable supply chain (P2)

**Independent Test**: `goreleaser check` passes; release workflow triggers only on tags.

- [ ] T025 [P] [US4] Add release-please-config.json and .release-please-manifest.json (release-type go, initial 0.0.0) and .github/workflows/release-please.yml (push to main)
- [ ] T026 [P] [US4] Add Dockerfile (distroless static nonroot by digest) and .goreleaser.yaml (darwin/linux × amd64/arm64, dockers_v2 to GHCR, syft SBOMs, cosign keyless signing)
- [ ] T027 [US4] Add .github/workflows/release.yml (tags `v*`; id-token, contents, packages, attestations write on the job; attest-build-provenance for binaries and image)

## Phase 7: User Story 5 — Repository health visible (P2)

- [ ] T028 [P] [US5] Add .github/workflows/scorecard.yml (weekly + push to main; publish results; SARIF upload)
- [ ] T029 [US5] Update README.md with CI, Scorecard, license and Go Report Card badges and links to governance files

## Phase 8: Polish & Cross-Cutting Concerns

- [ ] T030 Run actionlint, `make lint test`, `pre-commit run --all-files`; fix findings
- [ ] T031 Update docs/agent/PROGRESS.md

## Dependencies & Execution Order

- Phase 1 → Phase 2 → user stories. US2 depends on Makefile (T005) because CI calls `make lint` and
  `make test`. US1, US3, US4, US5 are independent of each other.
- Within US2: T014/T015 (tests) before T016/T017; T018 after T016/T017.

## Parallel Example: User Story 2

```bash
# Tests first, in parallel:
T014 internal/repotest/cigate_test.go
T015 internal/repotest/dco_test.go
# Config files in parallel:
T020 rulesets/labels   T021 dependabot   T022 ADRs
```

## Implementation Strategy

MVP = Phases 1–4 (local loop + `ci-gate`). Then US3, US4, US5, polish. One PR for the feature: the
non-generated change set is mostly documentation and configuration, under the 800-line rule once
generated/lock files are excluded.
