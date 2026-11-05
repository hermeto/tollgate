# Agent progress

Resume instructions: read this file and [`PLAYBOOK.md`](PLAYBOOK.md), then continue from **Next step**.

## Status

| Field | Value |
| --- | --- |
| Current feature | 001 repo-foundation — complete locally (steps 1–11); step 12 blocked by B1 |
| Spec Kit phase | converge clean |
| Last commit | `docs(agent): record 001 progress` on `001-repo-foundation` |
| Next step | feature 002 provider-simulator from `001-repo-foundation` |

## Features (merge order)

| # | Branch | Local status | PR | PR body draft |
| --- | --- | --- | --- | --- |
| 001 | `001-repo-foundation` | done (lint, test, DoD local parts) | not opened (B1) | [pr-bodies/001-repo-foundation.md](pr-bodies/001-repo-foundation.md) |

PR metadata to apply when opening: title `<type>(<area>): <summary> (#NNN)`, base = previous feature
branch (`main` for 001), labels `kind/*`, `area/*`, `spec`, `stacked` (not 001), milestone per playbook
§6, assignee `hermeto`.

## Blockers

- **B1 — GitHub repository not created.** `gh repo create tollgate --public …` (playbook §4.4) was
  refused by the agent's permission classifier ("Create Public Surface"). The agent did not work around
  it (no private repo, no other host). Everything in §4.4 and every push/PR step is queued below. The
  maintainer creates the repository (command in "Remote steps queued"), then a session resumes.
- **B2 — `.claude/settings.json` not written.** Playbook §4.5 asks the agent to write its own
  permission file; the classifier refuses self-granted permissions. The intended content is in
  [`claude-settings.json`](claude-settings.json); the maintainer copies it to `.claude/settings.json`.
  Becomes a `needs-human` issue once the repository exists.
- **B3 — Code of Conduct contact.** Publishing a private contact address is the maintainer's call;
  CODE_OF_CONDUCT.md points to MAINTAINERS.md, which says the address is pending. Becomes a
  `needs-human` issue once the repository exists.

## Remote steps queued (blocked by B1)

Run by the maintainer, from the repository root:

```bash
gh repo create tollgate --public --source . --remote origin --push \
  --description "LLM gateway with platform policies: per-team rate limits and budgets, cost chargeback (FOCUS), provider fallback, semantic cache, PII redaction and tamper-evident audit log. Go, OpenTelemetry, Envoy ext_proc."
```

Then an agent session resumes with: repository settings, security settings, ruleset, labels,
milestones, roadmap issue (playbook §4.4); `needs-human` issue for B2; push every local feature branch
and open the stacked PRs in order; watch CI and fix until green.

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

## Decision log (features)

Per-feature decisions live in each `specs/NNN-*/spec.md` (`## Clarifications`) and in the PR body drafts.

- 001: see [pr-bodies/001-repo-foundation.md](pr-bodies/001-repo-foundation.md) "Decisions for the reviewer".

## Evidence log

- 001 SC-001: fresh clone `make dev-setup && make lint test` → exit 0 in 7.4 s (warm Go and pre-commit caches, M-series Mac).
