# Agent progress

Resume instructions: read this file and [`PLAYBOOK.md`](PLAYBOOK.md), then continue from **Next step**.

## Status

| Field | Value |
| --- | --- |
| Current feature | 014 envoy-extproc-adapter — complete |
| Spec Kit phase | 014 complete; T001–T004 complete; next feature 015 kubernetes-helm |
| Last commit | `feat(api): add gateway SDK compatibility` on `004-protocol-api` (see `git log -1`) |
| Next step | feature 015 (`015-kubernetes-helm`): Kubernetes and Helm deployment |

## Features (merge order)

| # | Branch | Local status | PR | PR body draft |
| --- | --- | --- | --- | --- |
| 001 | `001-repo-foundation` | done; CI green | [#4](https://github.com/hermeto/tollgate/pull/4) ready | [pr-bodies/001-repo-foundation.md](pr-bodies/001-repo-foundation.md) |
| 002 (1/9) | `002-provider-simulator` | CLI, schema, config | [#5](https://github.com/hermeto/tollgate/pull/5) ready | [pr-bodies/002-provider-simulator.md](pr-bodies/002-provider-simulator.md) |
| 002 (2/9) | `002-provider-simulator-iter1` | server skeleton | [#6](https://github.com/hermeto/tollgate/pull/6) ready | [iter1](pr-bodies/002-provider-simulator-iter1.md) |
| 002 (3/9) | `002-provider-simulator-iter2` | tokens, generator | [#7](https://github.com/hermeto/tollgate/pull/7) ready | [iter2](pr-bodies/002-provider-simulator-iter2.md) |
| 002 (4/9) | `002-provider-simulator-iter3` | OpenAI chat | [#8](https://github.com/hermeto/tollgate/pull/8) ready | [iter3](pr-bodies/002-provider-simulator-iter3.md) |
| 002 (5/9) | `002-provider-simulator-iter4` | embeddings, models, request log | [#9](https://github.com/hermeto/tollgate/pull/9) ready | [iter4](pr-bodies/002-provider-simulator-iter4.md) |
| 002 (6/9) | `002-provider-simulator-iter5` | Anthropic | [#10](https://github.com/hermeto/tollgate/pull/10) ready | [iter5](pr-bodies/002-provider-simulator-iter5.md) |
| 002 (7/9) | `002-provider-simulator-iter6` | latency, faults | [#11](https://github.com/hermeto/tollgate/pull/11) ready | [iter6](pr-bodies/002-provider-simulator-iter6.md) |
| 002 (8/9) | `002-provider-simulator-iter7` | SDK compat, embeddings forward | [#13](https://github.com/hermeto/tollgate/pull/13) | [iter7](pr-bodies/002-provider-simulator-iter7.md) |
| 002 (9/9) | `002-provider-simulator-iter8` | pacing deadlines, perf, ADR-0009 | pending | [iter8](pr-bodies/002-provider-simulator-iter8.md) |
| 003 (1/6) | `003-local-stack` | pipeline, request-ID, timing, ADRs | [#15](https://github.com/hermeto/tollgate/pull/15) draft | — |
| 003 (2/6) | `003-local-stack-iter1` | gateway schema, config, SSE | [#16](https://github.com/hermeto/tollgate/pull/16) draft | — |
| 003 (3/6) | `003-local-stack-iter2` | upstream, passthrough server, drain, serve | [#17](https://github.com/hermeto/tollgate/pull/17) draft | — |
| 003 (4/6) | `003-local-stack-iter3` | OTel telemetry, health/readiness, ops metrics, probe | [#18](https://github.com/hermeto/tollgate/pull/18) ready | [iter3](pr-bodies/003-local-stack-iter3.md) |
| 003 (5/6) | `003-local-stack-iter4` | pinned Compose stack, up/down/smoke | pending | [iter4](pr-bodies/003-local-stack-iter4.md) |
| 003 (6/6) | `003-local-stack-iter5` | E2E tests, CI gate, overhead check, docs | pending | [iter5](pr-bodies/003-local-stack-iter5.md) |
| 004 | `004-protocol-api` | canonical model, OpenAI/Anthropic adapters, logical models, translation, SDK compatibility | [#21](https://github.com/hermeto/tollgate/pull/21) ready | [PR body](pr-bodies/004-protocol-api.md) |
| 005 | `005-identity-and-policies` | virtual keys, policy schemas/inheritance, reload, allowlists, CLI and gateway enforcement | pending | [PR body](pr-bodies/005-identity-and-policies.md) |
| 006 | `006-pricing-and-usage` | effective-dated pricing, cost calculation, content-free usage events, bounded async spool | pending | [PR body](pr-bodies/006-pricing-and-usage.md) |
| 007 | `007-rate-limiting` | local token bucket admission and deterministic reset metadata | pending | [PR body](pr-bodies/007-rate-limiting.md) |
| 008 | `008-budgets-and-chargeback` | spend reservations, soft/hard budget decisions, FOCUS-aligned CSV export | pending | [PR body](pr-bodies/008-budgets-and-chargeback.md) |
| 009 | `009-routing-and-fallback` | weighted healthy-target selection and bounded retry backoff | pending | [PR body](pr-bodies/009-routing-and-fallback.md) |
| 010 | `010-observability` | normalized GenAI attributes, dashboards and overhead alert | pending | [PR body](pr-bodies/010-observability.md) |
| 011 | `011-pii-redaction` | deterministic synthetic PII detection and privacy-safe redaction metadata | pending | [PR body](pr-bodies/011-pii-redaction.md) |
| 012 | `012-audit-log` | content-free hash-chained governance events and verification | pending | [PR body](pr-bodies/012-audit-log.md) |
| 013 | `013-response-cache` | bounded exact response cache with TTL and content-addressed keys | pending | [PR body](pr-bodies/013-response-cache.md) |
| 014 | `014-envoy-extproc-adapter` | protocol-neutral Envoy external-processing phase decisions | pending | [PR body](pr-bodies/014-envoy-extproc-adapter.md) |

PR metadata: title `<type>(<area>): <summary> (#NNN)`, base = previous feature
branch (`main` for 001), labels `kind/*`, `area/*`, `spec`, `stacked` (not 001), milestone per playbook
§6, assignee `hermeto`.

## Blockers

- **B1 — resolved 2026-11-05.** The classifier refused `gh repo create --public`; the maintainer
  created `hermeto/tollgate` and pushed `main`, `001-repo-foundation` and `002-provider-simulator`.
- **B2 — `.claude/settings.json` not written** ([#1](https://github.com/hermeto/tollgate/issues/1)).
  The classifier refuses self-granted permissions; the intended content is in
  [`claude-settings.json`](claude-settings.json).
- **B4 — SC-007 TTFT tail** ([#12](https://github.com/hermeto/tollgate/issues/12)). p99 +23.5 ms at 1,000
  streams with a co-located load generator; criterion restated (median ≤ +5 ms); maintainer decides.
- **B3 — Code of Conduct contact** ([#2](https://github.com/hermeto/tollgate/issues/2)). Publishing a
  private address is the maintainer's call; CODE_OF_CONDUCT.md points to MAINTAINERS.md.

## Remote setup (playbook §4.4), 2026-11-05

| Item | State | By |
| --- | --- | --- |
| Public repo, description, topics | done | maintainer |
| Issues and Discussions on; wiki and projects off | done | maintainer |
| Squash only, delete branch on merge, auto-merge | done | maintainer |
| Vulnerability alerts, Dependabot security updates, private vulnerability reporting | done | maintainer |
| Secret scanning and push protection | done | maintainer |
| Ruleset `main` from `.github/rulesets/main.json` (id 26592775, all fields accepted) | done | agent |
| 33 playbook labels (GitHub defaults kept) | done | agent |
| Milestones M0–M5 (numbers 1–6) | done | agent |
| `needs-human` issues #1 (B2) and #2 (B3) | done | agent |
| Pinned tracking issue [#3 "v1.0 roadmap"](https://github.com/hermeto/tollgate/issues/3) | done | agent |

## Tool versions (Step 1, 2026-11-04, macOS arm64)

| Tool | Version |
| --- | --- |
| git | 2.54.0 |
| gh | authenticated as `hermeto` (scopes: repo, workflow, read:org, gist) |
| uv | 0.13.7 |
| Python | 3.15.2 |
| Spec Kit (specify-cli) | 1.3.0 (git extension 1.0.3) |
| Go | 1.27.3 |
| Docker Engine / Compose | 29.9.2 / v5.6.1 (10 CPU, 8 GiB) |
| kind / kubectl / helm | 0.33.0 / 1.37.2 / 4.3.1 |
| golangci-lint | 2.15.0 |
| govulncheck | 1.9.0 |
| actionlint | 1.7.12 |
| shellcheck | 0.11.0 |
| yamllint | 1.38.0 |
| markdownlint-cli2 | 0.24.0 |
| pre-commit | 4.7.0 |
| trivy | 0.76.1 |
| k6 | 2.3.0 |
| yq | 4.54.1 |
| promtool (prometheus) | 3.15.1 |

## Decision log (bootstrap)

- D-B1: The local repository already existed with an empty `Initial commit` and no remote, so the
  bootstrap commit is added on top of it instead of running `git init`.
- D-B2: Spec Kit skills are installed as `/speckit-<name>` (skills mode). The skills were installed
  after the agent session started, so the agent follows each `.claude/skills/speckit-*/SKILL.md`
  directly; the outputs are the same.
- D-B3: The playbook was found at the repository root and moved verbatim to `docs/agent/PLAYBOOK.md`.
- D-B4: `branch_numbering: sequential` is the git extension default; auto-commit hooks stay disabled so
  commit messages follow playbook §5.2.
- D-B5: PR descriptions follow playbook §5.3 (project-specific) instead of the maintainer's global
  minimal-description rule; recorded in `CLAUDE.md`.
- D-B6: The Spec Kit skills became visible to the agent's Skill tool partway through feature 001; the
  agent keeps following the same `SKILL.md` instructions directly.
- D-B7: Features larger than the 800-line PR rule are split into slices by user story. Slices are
  chained linearly (`NNN-name` ← `NNN-name-iter1` ← `NNN-name-iter2` …), each PR based on the previous
  slice, and feature N+1 branches from the last slice of N. This keeps one linear, bottom-up merge order.
- D-B8: This file is authoritative on the top branch of the stack. Lower branches keep the copy from
  their last change, so ready PRs are not re-pushed (and re-run in CI) just to refresh progress notes.
- D-B9: Codex support is installed alongside Claude using Spec Kit 1.3.0 multi-install-safe skills;
  Claude remains the project default integration, while Codex receives `.agents/skills/speckit-*`.
- D-003-4-1: OTel providers use the pinned versions in research R5, bounded batch queues and a
  five-second exporter timeout; telemetry is fail-open and access logs are JSON plus otelslog.
- D-003-4-2: Dependency readiness uses TCP checks at five-second intervals until protocol clients are
  introduced; fail-closed dependencies gate readiness and fail-open dependencies degrade it.

## Decision log (features)

Per-feature decisions live in each `specs/NNN-*/spec.md` (`## Clarifications`) and in the PR body drafts.

- 001: see [pr-bodies/001-repo-foundation.md](pr-bodies/001-repo-foundation.md) "Decisions for the reviewer".

## Evidence log

- 001 SC-001: fresh clone `make dev-setup && make lint test` → exit 0 in 7.4 s (warm Go and pre-commit caches, M-series Mac).
- 002 tokens: o200k_base mean absolute error 7.4% (holdout 5.8%), bias +2.2%; 4.2 µs per 4 KB.
- 002 SC-001: `make sdk-compat` 6/6 clients (Python, TypeScript, Go × OpenAI, Anthropic).
- 002 SC-003: k6 10,000 req/s sustained, 0 errors; 30,000 req/s probed (p95 106 µs).
- 002 SC-007: TTFT fixed 100 ms at 1,000 streams: median +3.4 ms, p90 +13.6 ms, p99 +23.5 ms (Go client).
- 002 note: the agent's own `git checkout <file>` during a debug cleanup discarded an uncommitted change
  once (re-applied immediately); keep debug instrumentation in scratch copies, never in tracked files.
- 001 SC-003/SC-006: PR #4 first run — every check green; `ci-gate` reported ~2.5 min after the push (lint 1m56s is the longest job).
- 003 slices 2–3: `go test -race ./...` and `golangci-lint run ./...` green; manual simulator→gateway smoke
  passed for non-streaming JSON, `x-request-id` propagation and SSE streaming. `BenchmarkPassthrough`
  measured 1.97 ms/op on Apple M4 with the simulator in-process (includes upstream wait).
- 003 slice 4: `go test ./...`, `go test -race ./...`, and `golangci-lint run ./...` green; telemetry
  recording against an unavailable collector remained under 250 ms for 2,000 measurements; health
  transition test passed.
