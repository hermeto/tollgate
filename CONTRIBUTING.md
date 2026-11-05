# Contributing to Tollgate

Thank you for helping. Tollgate is a governance product, so its own repository follows the rules it
asks its users to follow: every change is specified, reviewed, measured and signed off.

By participating you agree to the [Code of Conduct](CODE_OF_CONDUCT.md). Security issues go through
[private reporting](SECURITY.md), never public issues. Questions go to
[Discussions](https://github.com/hermeto/tollgate/discussions) (see [SUPPORT.md](SUPPORT.md)).

## Development setup

You need **git**, **Go** (any Go ≥ 1.21; the exact toolchain pinned in `go.mod` is downloaded
automatically), **uv** and **jq**. Docker is needed from feature 003 onwards (`make up`).

| Platform | Install |
| --- | --- |
| macOS | `brew install go uv jq` |
| Linux | Go tarball from <https://go.dev/dl/>, `curl -LsSf https://astral.sh/uv/install.sh \| sh`, `jq` from your package manager |

Then:

```bash
make dev-setup    # pinned Go tools (tools/go.mod) + pre-commit hooks; no Homebrew needed
make lint test    # every linter + race-enabled tests
make help         # all targets
```

No target calls a model provider or any paid API. You never need provider keys to contribute.

### Tooling and configuration

| File | Purpose |
| --- | --- |
| [`Makefile`](Makefile) | Entry points: `help`, `dev-setup`, `fmt`, `lint`, `test`, `bench-micro` |
| [`tools/go.mod`](tools/go.mod) | Pinned Go tools run with `go tool -modfile=tools/go.mod` |
| [`.pre-commit-config.yaml`](.pre-commit-config.yaml) | Pinned hooks: actionlint, shellcheck, yamllint, markdownlint, editorconfig |
| [`.golangci.yml`](.golangci.yml) | Go linters and formatters |
| [`.markdownlint-cli2.yaml`](.markdownlint-cli2.yaml), [`.yamllint`](.yamllint), [`.editorconfig`](.editorconfig) | Docs and YAML style |
| [`.github/ISSUE_TEMPLATE/`](.github/ISSUE_TEMPLATE), [`.github/PULL_REQUEST_TEMPLATE.md`](.github/PULL_REQUEST_TEMPLATE.md) | Issue forms (bug, feature, docs, policy proposal) and the PR template |
| [`.github/workflows/`](.github/workflows) | CI (`ci-gate`), CodeQL, Scorecard, release-please, release, label sync |
| [`.github/rulesets/main.json`](.github/rulesets/main.json) | Protection rules for `main` |
| [`.github/labels.yml`](.github/labels.yml) | Label set, synced on merge |
| [`.github/dependabot.yml`](.github/dependabot.yml) | Weekly grouped dependency updates |
| [`.goreleaser.yaml`](.goreleaser.yaml), [`Dockerfile`](Dockerfile) | Signed release artifacts and image |
| [`docs/adr/`](docs/adr) | Architecture decision records ([template](docs/adr/template.md)) |

## How work is organised: Spec Kit

Every feature follows [Spec Kit](https://github.com/github/spec-kit) and lives in `specs/NNN-short-name/`:

1. `spec.md` — what and why (user stories, requirements, success criteria), with a `## Clarifications` log.
2. `plan.md`, `research.md`, `data-model.md`, `contracts/`, `quickstart.md` — how, with sources.
3. `checklists/` — requirements quality gates.
4. `tasks.md` — ordered, traceable tasks.

The [constitution](.specify/memory/constitution.md) governs every plan. Architectural decisions become
ADRs in [`docs/adr/`](docs/adr) using the [MADR template](docs/adr/template.md).

## Branches and pull requests

- Feature branches are named `NNN-short-name` (for example `007-rate-limiting`); follow-ups use
  `NNN-short-name-iterK`.
- Large efforts are **stacked**: each feature branch starts from the previous one, and its PR targets
  that branch. Reviewers merge bottom-up; GitHub retargets the next PR when a base branch is deleted.
- PR titles are Conventional Commits, e.g. `feat(finops): budgets with soft/hard limits (#008)`.
- Fill in the [PR template](.github/PULL_REQUEST_TEMPLATE.md): requirements covered, decisions for the
  reviewer, verification, what was not verified, and performance impact.
- Keep PRs under ~800 changed lines, excluding generated files, lock files, fixtures and Spec Kit
  artifacts. Split by user story when larger.
- `main` only accepts squash merges through PRs with a green `ci-gate`.

## Commits: Conventional Commits and DCO

Commit messages follow [Conventional Commits](https://www.conventionalcommits.org/) (`feat:`, `fix:`,
`perf:`, `docs:`, `refactor:`, `test:`, `build:`, `ci:`, `chore:`), imperative mood, subject ≤ 72
characters, body explaining *why*. The changelog is generated from them; do not edit `CHANGELOG.md`.

Every commit must be signed off under the [Developer Certificate of Origin](https://developercertificate.org/):

```bash
git commit -s -m "fix(ratelimit): refund tokens when a stream ends early"
```

CI's `dco` job rejects commits without a `Signed-off-by` trailer. Fix a branch with
`git rebase --signoff <base>`.

## Performance budget rule

Tollgate sits on every request, so the hot path has a budget (constitution, Principle II):

- full pipeline p99 overhead ≤ 15 ms (excluding the semantic-cache embedding call);
- minimal pipeline (auth + rate limit + metrics) p99 ≤ 3 ms.

Every pipeline Stage ships with a micro-benchmark (`make bench-micro`) and a span. A PR that regresses a
Stage beyond its budget needs an ADR that justifies it. State the before/after numbers in the PR's
"Performance impact" section.

## Privacy rules for contributions

Never paste real prompts, responses, API keys or personal data into issues, PRs, fixtures or tests.
PII test data is synthetic and labeled as such. Example keys look like `tg-example-…`.

## Governance

See [GOVERNANCE.md](GOVERNANCE.md) and [MAINTAINERS.md](MAINTAINERS.md). Code owners are listed in
[`.github/CODEOWNERS`](.github/CODEOWNERS).
