# Tollgate Constitution

Tollgate is an LLM gateway with platform policies. This constitution sets the non-negotiable rules
that every specification, plan, task list and pull request in this repository must follow.

## Core Principles

### I. Evidence over opinion

- Every policy MUST ship with its measured cost (latency, CPU) and its measured effect (dollars,
  availability, risk avoided).
- Claims in documentation MUST link to the data, benchmark or test that supports them.
- A number that cannot be reproduced from the repository MUST NOT appear in the report.

Rationale: Tollgate exists to show what governance costs and what it saves; unmeasured claims
defeat that purpose.

### II. The hot path is sacred

- Each pipeline Stage MUST have a declared latency budget, a micro-benchmark (`go test -bench`)
  and its own trace span.
- The full pipeline MUST stay within p99 ≤ 15 ms added overhead (excluding the semantic-cache
  embedding call) and the minimal pipeline within p99 ≤ 3 ms.
- A change that regresses the hot path beyond its budget MUST NOT be merged without an ADR that
  justifies it.

Rationale: a gateway that slows every request becomes the thing teams route around.

### III. Privacy by default

- Prompt and response content MUST NOT appear in logs, traces, metrics, cache keys, usage events
  or audit records unless a policy explicitly enables redacted capture.
- PII datasets MUST be synthetic and labeled as such.
- Tests MUST prove that no sink leaks PII.

Rationale: the gateway sees everything teams send to providers; it must not become the leak.

### IV. Fail modes are explicit

- Every dependency failure (Redis/Valkey, PostgreSQL, vector store, NER service, telemetry
  backend) MUST have a declared fail-open or fail-closed behaviour per policy.
- Every fail-mode activation MUST be audited.

Rationale: implicit failure behaviour turns outages into silent policy bypasses.

### V. Governance as code

- Policies, prices, dashboards, alerts and repository settings MUST live in git.
- Each of them MUST be validated by a schema and change only through reviewed pull requests.

Rationale: a governance product has to be governed the same way it asks its users to govern.

### VI. Local-first and zero cost

- Every feature MUST work with the simulated providers on a laptop and in CI.
- No automated process (test, benchmark, demo, CI job) MAY call a paid API. The default upstream
  in every profile is the simulator.

Rationale: anyone can reproduce the results without spending money or holding provider keys.

### VII. Reproducible by default

- Every dependency MUST be pinned: Go modules (go.sum), Python (uv.lock), container images by
  digest, GitHub Actions by full commit SHA. `latest` tags are forbidden.
- Benchmark results MUST carry complete metadata (commit, versions, hardware, config and dataset
  hashes, seeds) or be rejected.
- Generators (datasets, reports, reference docs) MUST be deterministic.

Rationale: results that drift with the environment are not evidence.

### VIII. Test what matters

- Policy packages MUST keep unit-test coverage at or above 80%.
- Protocol handling MUST have contract tests built from recorded wire fixtures.
- Parsers and redaction MUST be fuzzed.
- User-facing flows MUST have end-to-end tests on the local profile.
- Generated artifacts MUST have golden tests.

Rationale: tests concentrate where a regression would break governance guarantees.

### IX. Secure supply chain

- Releases MUST be signed and ship with an SBOM and build provenance.
- CI MUST run CodeQL, OpenSSF Scorecard, govulncheck, Trivy and dependency review.
- Workflows MUST use least-privilege permissions.

Rationale: a security control with a weak supply chain is itself an attack surface.

### X. Simplicity and small changes

- Prefer the simplest design that can be measured. No speculative abstractions.
- Pull requests MUST be small: one feature or one iteration each. v2 ideas go to the roadmap.

Rationale: small, measurable changes are reviewable; large clever ones are not.

### XI. Open-source citizenship

- Commits MUST follow Conventional Commits and carry a DCO sign-off.
- Documentation MUST be updated in the same pull request as the behaviour it describes.
- Language MUST be inclusive, and contribution paths MUST be documented and welcoming.

Rationale: the repository is a public reference; how it is built is part of what it teaches.

## Technical Constraints

- Gateway, CLI, simulator and load generator: Go (latest stable), shipped as one `tollgate` binary
  with subcommands. Python (managed by uv) is used only for the report generator.
- HTTP uses `net/http` with at most a minimal router; logging uses `log/slog` JSON; configuration
  is YAML validated by JSON Schemas in `schemas/`.
- State stores: Redis/Valkey for counters, PostgreSQL for usage events, keys and audit.
- Every policy is a pipeline Stage implementing one interface, with its own span, metrics and
  micro-benchmark tracked in CI.
- Real providers are configured only by explicit operator configuration; code MUST NOT fall back
  to real endpoints when keys are absent.

## Development Workflow and Quality Gates

- Every feature follows the Spec Kit flow: specify → clarify → plan → checklist → tasks →
  analyze → implement → converge, with artifacts under `specs/NNN-short-name/`.
- Architectural decisions are recorded as ADRs in `docs/adr/` (MADR format).
- Every pull request MUST pass the aggregated `ci-gate` status check. Disabling a check, linter
  or test to get green is forbidden; fix the cause or open a `needs-human` issue.
- Pull request descriptions MUST state requirements covered, decisions taken for the reviewer,
  how the change was verified and its hot-path performance impact.

## Governance

- This constitution supersedes other practices in this repository.
- Amendments are made through a pull request labeled `constitution`, with a semantic version bump
  and a sync of dependent templates:
  - MAJOR: a principle is removed or redefined.
  - MINOR: a principle or section is added or materially expanded.
  - PATCH: wording and clarifications.
- Every plan's Constitution Check MUST pass, or justify each deviation in its Complexity Tracking
  table.
- Reviewers verify compliance on every pull request; runtime guidance for agents lives in
  `CLAUDE.md` and `docs/agent/PLAYBOOK.md`.

**Version**: 1.0.0 | **Ratified**: 2026-11-04 | **Last Amended**: 2026-11-04
