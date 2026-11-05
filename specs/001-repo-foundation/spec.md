# Feature Specification: Open-source foundation, CI and contributor experience

**Feature Branch**: `001-repo-foundation`

**Created**: 2026-11-04

**Status**: Draft

**Input**: User description: "Open-source foundation for the Tollgate repository: governance documents,
issue and PR templates, repository settings as code, automated quality gates aggregated into one
required check, security scanning, automated signed releases and visible repository health."

## Clarifications

### Session 2026-11-04

- Q: Which jobs does `ci-gate` aggregate, given that CodeQL and Scorecard run as separate workflows? → A: Only the jobs of the main CI workflow: lint, test, security scan (Trivy filesystem), PR title, sign-off and dependency review. CodeQL and Scorecard report through code scanning and are not part of `ci-gate`. (decided by agent: `needs:` only works inside one workflow, and one required status keeps the ruleset simple.)
- Q: How do PR-only jobs behave on pushes to the default branch? → A: PR title, sign-off and dependency review are skipped by condition on push events, and `ci-gate` accepts `skipped` only for those three jobs and only on non-PR events. (decided by agent: avoids a false failure on `main` while keeping skip-as-failure on PRs.)
- Q: How does a Linux contributor without Homebrew get the tools? → A: Setup requires only git, Go and uv. Go tools are pinned in a separate tools module and run with `go tool`; the other linters are pinned pre-commit hooks that pre-commit installs itself. Homebrew is an optional shortcut on macOS. (decided by agent: one pinned path for macOS, Linux and CI; no global installs.)
- Q: What response times does the security policy promise? → A: Acknowledge within 3 business days, triage within 7 calendar days, fix or mitigation for HIGH/CRITICAL within 90 days, with coordinated disclosure at 90 days. (decided by agent: common open-source defaults that a single maintainer can keep.)
- Q: Which versions receive security fixes? → A: Only the latest released minor version (pre-1.0: the latest 0.x release) and `main`. (decided by agent: a single-maintainer project cannot backport.)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Contributor can set up, verify and propose a change (Priority: P1)

A first-time contributor clones the repository, reads the README, the contribution guide and the code
of conduct, runs one setup command and one verification command on macOS or Linux, and opens a pull
request whose description is pre-filled from a template.

**Why this priority**: Without a working local loop and clear rules, nobody (including the maintainer's
agent) can contribute safely; every later feature depends on it.

**Independent Test**: On a fresh clone with no secrets and no provider credentials, run the setup
command followed by the lint-and-test command; both exit 0 without contacting any paid API.

**Acceptance Scenarios**:

1. **Given** a fresh clone on macOS with Homebrew, **When** the contributor runs the setup command and
   then the lint-and-test command, **Then** both succeed and no network call is made to a model provider.
2. **Given** a fresh clone on Linux without Homebrew, **When** the contributor follows the documented
   alternative installation path, **Then** the lint-and-test command succeeds.
3. **Given** a contributor opens a pull request, **When** the PR form loads, **Then** the description
   is pre-filled with the sections required by the project (summary, stack note, requirements,
   decisions for the reviewer, verification, performance impact, checklist).
4. **Given** a contributor wants to report a bug, propose a feature, a docs change or a policy, **When**
   they open a new issue, **Then** they can choose a structured form for each and are pointed to
   Discussions for questions.

---

### User Story 2 - Maintainer sees one required status per pull request (Priority: P1)

Every pull request, whatever its base branch, is checked automatically, and one aggregated status
(`ci-gate`) tells the maintainer whether all required checks passed.

**Why this priority**: The branch ruleset requires exactly one status; stacked PRs target feature
branches, so checks must run on any base.

**Independent Test**: Open a PR with a deliberately failing lint; `ci-gate` fails. Fix it; `ci-gate`
passes. A skipped or cancelled required job also makes `ci-gate` fail.

**Acceptance Scenarios**:

1. **Given** a PR against any branch, **When** CI runs, **Then** linters (Go, workflows, shell, YAML,
   Markdown, editorconfig), vulnerability scan, race-enabled tests, PR title check, sign-off check and
   dependency review run, and `ci-gate` reports their aggregate result.
2. **Given** any required job fails, is cancelled or is skipped unexpectedly, **When** `ci-gate`
   evaluates, **Then** `ci-gate` fails and names the failing job.
3. **Given** a PR from a fork without repository secrets, **When** CI runs, **Then** all required jobs
   can pass (no job depends on a secret).
4. **Given** a commit without a `Signed-off-by` trailer, **When** the sign-off check runs, **Then** it
   fails and lists the offending commit.
5. **Given** a PR title that is not a Conventional Commit, **When** the title check runs, **Then** it fails.

---

### User Story 3 - Security reporter knows how and what to report (Priority: P1)

A security researcher finds the security policy, reports privately through the platform's private
vulnerability reporting, and knows the scope and the response times.

**Why this priority**: Tollgate is a security control; an unclear reporting path is itself a risk.

**Independent Test**: Read the security policy and confirm it states the private channel, the in-scope
classes (gateway, policy bypass, PII leakage, audit tampering), out-of-scope items and response times.

**Acceptance Scenarios**:

1. **Given** a reporter opens the security policy, **When** they read it, **Then** they find the
   private reporting link, in-scope and out-of-scope classes, supported versions and response targets
   (acknowledgement within 3 business days, triage within 7 days, fix or mitigation for HIGH/CRITICAL
  within 90 days) and that only the latest minor release and `main` receive fixes.

---

### User Story 4 - Maintainer releases from Conventional Commits with a verifiable supply chain (Priority: P2)

Merges to the default branch keep a release proposal (changelog and next version) up to date. When a
version tag is created, binaries for macOS and Linux (amd64 and arm64) and a minimal multi-arch
container image are published with an SBOM, keyless signatures and build provenance.

**Why this priority**: Required for v1.0 trust (NFR-4) but not needed to build features.

**Independent Test**: Run the release build in snapshot mode locally; it produces the four binaries,
checksums and an SBOM. Release workflows only trigger on the default branch (proposal) and on tags.

**Acceptance Scenarios**:

1. **Given** a Conventional Commit merged to the default branch, **When** the release-proposal job runs,
   **Then** a release PR with a generated changelog and the next semantic version is opened or updated.
2. **Given** a version tag, **When** the release workflow runs, **Then** it publishes binaries, an
   image, SBOMs, keyless signatures and provenance attestations; nothing runs for untagged commits.

---

### User Story 5 - Visitor sees repository health at a glance (Priority: P2)

A visitor sees badges for CI, OpenSSF Scorecard, license and Go Report Card in the README; Scorecard
runs weekly and on pushes to the default branch.

**Why this priority**: Health signals matter for adoption but do not block development.

**Independent Test**: The README shows the four badges with working links; the Scorecard workflow
definition has a weekly schedule and a default-branch push trigger.

**Acceptance Scenarios**:

1. **Given** a visitor opens the README, **When** the page renders, **Then** the four badges are
   visible and link to their sources.

---

### Edge Cases

- A fork PR has no access to secrets: every required job must pass without them; jobs that need
  write tokens (labels sync, Scorecard publishing, releases) never run on PRs.
- A contributor on Linux has no Homebrew: the setup command installs every tool through Go and uv and
  prints exact install instructions for Go or uv if either is missing.
- A required job is skipped because an upstream job failed: `ci-gate` treats `skipped` and
  `cancelled` as failure.
- Merge commits created by the platform (e.g. update-branch) lack sign-off: the sign-off check only
  inspects commits authored in the PR, excluding merge commits.
- Dependency review is unavailable for the repository (e.g. feature disabled): the job fails loudly
  rather than silently passing.
- Markdown with long lines (tables, links) is allowed; duplicate headings are allowed only between
  sibling sections.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The repository MUST contain a code of conduct (Contributor Covenant 2.1), a contribution
  guide (development setup, Spec Kit workflow, branches and PRs, DCO sign-off, Conventional Commits,
  performance-budget rule), a security policy, a support guide, a governance document, a maintainers
  list and a code-owners file.
- **FR-002**: The repository MUST provide structured issue forms for bug reports, feature requests,
  documentation changes and policy proposals, plus an issue-template configuration that routes
  questions to Discussions and disables blank issues.
- **FR-003**: The repository MUST provide a pull-request template containing the sections defined in
  playbook §5.3.
- **FR-004**: Repository settings MUST be versioned as code: the default-branch ruleset, the label set
  (with an automated sync on the default branch), an ADR-0000 "Record architecture decisions" and an
  ADR template.
- **FR-005**: CI MUST run on pull requests against any base branch and on pushes to the default branch,
  executing: Go linting, Go vulnerability scanning, workflow linting, shell linting, YAML linting,
  Markdown linting, editorconfig checking, Go tests with the race detector, PR title validation,
  sign-off validation and dependency review.
- **FR-006**: CI MUST publish one aggregated status named `ci-gate` that succeeds only if every required
  job succeeded, and fails when any required job failed, was cancelled or was skipped. The only
  accepted skips are the PR-only jobs (PR title, sign-off, dependency review) on non-PR events.
- **FR-007**: The sign-off check MUST verify a `Signed-off-by` trailer on every non-merge commit in the
  PR without relying on a third-party app.
- **FR-008**: Security scanning MUST include CodeQL (Go, Python, Actions), OpenSSF Scorecard (weekly and
  on pushes to the default branch) and a Trivy filesystem scan.
- **FR-009**: Automated dependency updates MUST cover Go modules, Python (uv), GitHub Actions and Docker,
  grouped weekly with Conventional Commit prefixes.
- **FR-010**: Releases MUST be driven by Conventional Commits (changelog and tag) and MUST produce
  darwin/linux × amd64/arm64 binaries and a distroless multi-arch image on GHCR, with SBOMs, keyless
  signatures and artifact attestations; release jobs MUST run only on tags (the release proposal runs
  on the default branch).
- **FR-011**: Setup MUST require only git, Go and uv (plus Docker for later features); every lint and
  test tool MUST be pinned and installed by the repository's own tooling, without Homebrew.
- **FR-012**: Developer experience MUST include a Makefile with `help`, `dev-setup`, `fmt`, `lint`,
  `test` and `bench-micro` targets, a pre-commit configuration, a Go linter configuration, a Markdown
  linter configuration (line length rule off, duplicate headings allowed only between siblings) and a
  YAML linter configuration.
- **FR-013**: Every workflow MUST declare read-only top-level permissions, per-job least-privilege
  permissions, actions pinned by full commit SHA with a version comment, a concurrency group and job
  timeouts.
- **FR-014**: The Go module `github.com/hermeto/tollgate` MUST exist with a `tollgate` binary that
  prints its version (placeholder replaced by later features).
- **FR-015**: Every governance, template and configuration file introduced by this feature MUST be
  linked from the README or the contribution guide.

### Key Entities

- **Required check (`ci-gate`)**: the single status the ruleset requires; aggregates required jobs.
- **Ruleset**: versioned description of default-branch protections (no deletion, no force-push, linear
  history, PR required, review threads resolved, squash only, `ci-gate` required).
- **Label set**: versioned list of labels (name, color, description) synced to the repository.
- **ADR**: a numbered decision record following the MADR template.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: On a fresh clone, the setup command followed by the lint-and-test command completes with
  exit code 0 in under 5 minutes on a laptop with a warm package cache, with zero calls to paid APIs.
- **SC-002**: 100% of workflows pass the workflow linter, and 100% of third-party actions are pinned by
  full commit SHA.
- **SC-003**: Introducing a failure in any single required job makes `ci-gate` fail in 100% of cases
  (verified for failure and skip).
- **SC-004**: The OpenSSF Scorecard workflow runs on schedule and on default-branch pushes; its first
  run completes successfully after the repository is published.
- **SC-005**: Every file listed in FR-001 to FR-004 and FR-012 is reachable from the README or the
  contribution guide in at most one click.

## Assumptions

- The GitHub repository is `hermeto/tollgate`, public, default branch `main`; until the maintainer
  creates it (PROGRESS blocker B1), remote checks (`gh pr checks`, Scorecard, dependency review) are
  verified only by workflow linting and local runs.
- The maintainer is the only code owner and maintainer at launch (`@hermeto`).
- Dependency review requires the dependency graph, which is enabled by default on public repositories.
- Release signing uses keyless OIDC; no long-lived signing keys are stored.
- Go Report Card and Scorecard badges resolve only after the repository is public.
