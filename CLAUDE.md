# Tollgate — instructions for Claude Code

**Read [`docs/agent/PROGRESS.md`](docs/agent/PROGRESS.md) first.** It holds the current feature, Spec Kit phase,
open PRs in merge order, blockers and the decision log. The full playbook is
[`docs/agent/PLAYBOOK.md`](docs/agent/PLAYBOOK.md); the product specification is
[`docs/product/spec.md`](docs/product/spec.md); the constitution is
[`.specify/memory/constitution.md`](.specify/memory/constitution.md).

## Operating rules (condensed from playbook §2)

- Work autonomously. Answer Spec Kit clarification questions yourself using the product spec and the
  constitution. Prefer the simplest option that keeps the `local` profile working with simulated
  providers, keeps the hot path fast, and is easiest to measure.
- Record every decision taken on the maintainer's behalf: in the feature `spec.md` under
  `## Clarifications` (`- Q: … → A: … (decided by agent: <rationale>)`), as an ADR in `docs/adr/` when
  architectural, and in the PR under "Decisions for the reviewer".
- Blocked? Open an issue labeled `needs-human`, reference it in the PR, mark tasks `[~]` with the link,
  and continue with the next feature.
- Resolve the latest stable version of every tool, library, image and action, then pin it (go.sum,
  uv.lock, image digests, action SHAs). Never use `latest`.
- Verify upstream contracts (OpenAI, Anthropic, OTel GenAI, FOCUS, Envoy ext_proc) from official docs or
  SDK sources and cite them in `research.md`. Record wire formats as fixtures; never call paid APIs.

### Hard stops

- No calls to real model providers or any paid API; no provider keys read, requested or stored.
- No cloud resources (`aws`, `gcloud`, `az`, `terraform apply`).
- No push to `main` after the bootstrap commit; no force-push to `main`. Force-push feature branches only
  with `--force-with-lease`.
- No merging or approving PRs, no visibility changes, no repo deletion, no edits to other repositories.
- No secrets in git. Example keys are obviously fake (`tg-example-…`).
- Never disable CI checks, linters or tests to get green.
- No real personal data: PII datasets are synthetic and labeled as such.

## Feature loop (playbook §5)

Branches are stacked: `main ← 001-… ← 002-… ← … ← 018-…`. Iterations use `NNN-short-name-iterK`.

1. `git checkout <branch of N-1>`; `/speckit-specify` → `docs(spec): specify NNN <feature>`
2. `/speckit-clarify`, answering every question yourself → `docs(spec): clarify NNN`
3. `/speckit-plan` with the global plan context + feature plan input → `docs(plan): plan NNN`
4. `/speckit-checklist` → `docs(spec): checklist NNN`
5. `/speckit-tasks` → `docs(tasks): tasks NNN`
6. `/speckit-analyze`; fix CRITICAL/HIGH until clean → `docs(spec): resolve analysis findings NNN`
7. Push; open a draft PR (`--base <branch of N-1>`, label `stacked` except for 001)
8. `/speckit-implement` phase by phase; commit per phase (Conventional Commits, `git commit -s`)
9. `make lint test` + the feature's Definition of done
10. `/speckit-converge`; implement appended gap tasks until none
11. Update README, `docs/`, runbooks, ADRs (CHANGELOG is generated — do not edit)
12. Push; `gh pr checks --watch`; fix until green; mark ready; update the roadmap issue and `PROGRESS.md`

## Pull requests

This project overrides the maintainer's global "minimal PR description" rule: PR bodies follow playbook
§5.3 and `.github/PULL_REQUEST_TEMPLATE.md` (summary and spec link, stack note, requirements covered,
**Decisions for the reviewer**, verification, what was not verified, performance impact, checklist).
Still leave out anything that goes stale with a new commit: SHAs, commit counts, line counts.

## Session continuity

Update `docs/agent/PROGRESS.md` after every step on the branch you are on. A fresh session must be able
to resume from it alone.
