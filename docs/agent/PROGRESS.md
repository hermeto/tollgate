# Agent progress

Resume instructions: read this file and [`PLAYBOOK.md`](PLAYBOOK.md), then continue from **Next step**.

## Status

| Field | Value |
| --- | --- |
| Current feature | 002 provider-simulator — slice 1 of 7 (`002-provider-simulator`) |
| Spec Kit phase | implement |
| Last commit | `feat(simulator): add simulator configuration and JSON Schema` on `002-provider-simulator` |
| Next step | CI on PR #5; then slice 2 (`002-provider-simulator-iter1`) from the local backup branch `tmp-002-impl` |

## Features (merge order)

| # | Branch | Local status | PR | PR body draft |
| --- | --- | --- | --- | --- |
| 001 | `001-repo-foundation` | done; CI green | [#4](https://github.com/hermeto/tollgate/pull/4) ready | [pr-bodies/001-repo-foundation.md](pr-bodies/001-repo-foundation.md) |
| 002 (1/7) | `002-provider-simulator` | CLI, schema, config | [#5](https://github.com/hermeto/tollgate/pull/5) | [pr-bodies/002-provider-simulator.md](pr-bodies/002-provider-simulator.md) |
| 002 (2/7) | `002-provider-simulator-iter1` | server skeleton (code on `tmp-002-impl`) | — | — |
| 002 (3/7) | `002-provider-simulator-iter2` | tokens, generator (code on `tmp-002-impl`) | — | — |
| 002 (4/7) | `002-provider-simulator-iter3` | pending: OpenAI flavour | — | — |
| 002 (5/7) | `002-provider-simulator-iter4` | pending: Anthropic, request log | — | — |
| 002 (6/7) | `002-provider-simulator-iter5` | pending: latency, faults | — | — |
| 002 (7/7) | `002-provider-simulator-iter6` | pending: SDK compat, polish | — | — |

PR metadata: title `<type>(<area>): <summary> (#NNN)`, base = previous feature
branch (`main` for 001), labels `kind/*`, `area/*`, `spec`, `stacked` (not 001), milestone per playbook
§6, assignee `hermeto`.

## Blockers

- **B1 — resolved 2026-11-05.** The classifier refused `gh repo create --public`; the maintainer
  created `hermeto/tollgate` and pushed `main`, `001-repo-foundation` and `002-provider-simulator`.
- **B2 — `.claude/settings.json` not written** ([#1](https://github.com/hermeto/tollgate/issues/1)).
  The classifier refuses self-granted permissions; the intended content is in
  [`claude-settings.json`](claude-settings.json).
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

## Decision log (features)

Per-feature decisions live in each `specs/NNN-*/spec.md` (`## Clarifications`) and in the PR body drafts.

- 001: see [pr-bodies/001-repo-foundation.md](pr-bodies/001-repo-foundation.md) "Decisions for the reviewer".

## Evidence log

- 001 SC-001: fresh clone `make dev-setup && make lint test` → exit 0 in 7.4 s (warm Go and pre-commit caches, M-series Mac).
- 002 tokens: o200k_base mean absolute error 7.4% (holdout 5.8%), bias +2.2%; 4.2 µs per 4 KB.
- 001 SC-003/SC-006: PR #4 first run — every check green; `ci-gate` reported ~2.5 min after the push (lint 1m56s is the longest job).
