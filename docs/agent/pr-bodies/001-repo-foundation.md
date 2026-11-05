## Summary

Open-source foundation for Tollgate: governance documents, issue and PR templates, repository settings
as code, a pinned toolchain that works without Homebrew, and CI with one required status (`ci-gate`).
Spec: [`specs/001-repo-foundation/spec.md`](https://github.com/hermeto/tollgate/blob/001-repo-foundation/specs/001-repo-foundation/spec.md). Tracking: #3.

## Stack

First PR of the v1.0 stack; targets `main`. Every later feature PR is based on the previous one.

## Requirements covered

FR-001 to FR-015 of the feature spec; NFR-4 (supply chain), NFR-6 (pinning), NFR-7 (no paid calls),
NFR-11 (Apache-2.0) from `docs/product/spec.md`.

## Decisions for the reviewer

- `ci-gate` aggregates only `ci.yml` jobs; CodeQL and Scorecard report via code scanning. ([Clarifications](https://github.com/hermeto/tollgate/blob/001-repo-foundation/specs/001-repo-foundation/spec.md#clarifications))
- PR-only jobs (pr-title, dco, dependency-review) may be skipped on push to `main`; any other skip fails the gate.
- Go dev tools are pinned in `tools/go.mod` and run with `go tool`; other linters are SHA-pinned pre-commit hooks. Setup needs git, Go, uv and jq only.
- actionlint runs as its own pre-commit hook because it does not compile in the shared tools module.
- Trivy runs from an image pinned by digest instead of `aquasecurity/trivy-action`.
- release-please calls `release.yml` directly (tags made with `GITHUB_TOKEN` do not trigger workflows); images have no `latest` tag.
- Security response targets: acknowledge 3 business days, triage 7 days, fix HIGH/CRITICAL within 90 days; only the latest minor and `main` are supported.
- Code of Conduct contact: points to MAINTAINERS.md; the maintainer must publish a private address (#2).
- `.claude/settings.json` was not committed: the agent may not grant itself permissions; the content is in `docs/agent/claude-settings.json` (#1).
- The `main` ruleset, labels and milestones were applied with the API from the versioned files in this PR.
- ADR-0000 adopts MADR.

## How it was verified

- `make dev-setup && make lint test` on a fresh clone: exit 0, 7.4 s with warm caches.
- `actionlint` and `pre-commit run --all-files`: clean.
- `go test ./internal/repotest/...`: ci-gate rule (failure, cancel, skip on PR, PR-only skip on push, empty input), DCO (signed, unsigned, body-only trailer, merge commit), SHA pinning, permissions, concurrency and timeouts on every workflow.
- Trivy filesystem scan with the pinned image: 0 HIGH/CRITICAL findings.
- `goreleaser release --snapshot --clean --skip=sign`: four archives, SBOMs and a running distroless image.

## Not verified

- Scorecard, label sync and release-please run only on `main`; they are verified after this PR merges.
- Keyless signing (`cosign` OIDC) and the release workflow run only on a version tag.

## Performance impact

None: no hot-path code.

## Checklist

- [x] Tests added or updated
- [x] Docs updated in this PR (README, docs/, runbooks)
- [x] ADR added or updated for architectural decisions
- [x] No secrets, real keys or real personal data
- [x] New dependencies, images and actions pinned
- [x] No prompt/response content or PII in logs, traces, metrics or events
- [x] Commits signed off (DCO) and Conventional
