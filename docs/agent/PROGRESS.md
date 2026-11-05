# Agent progress

Resume instructions: read this file and [`PLAYBOOK.md`](PLAYBOOK.md), then continue from **Next step**.

## Status

| Field | Value |
| --- | --- |
| Current feature | bootstrap (playbook §4) |
| Spec Kit phase | constitution ratified (1.0.0) |
| Last commit | bootstrap commit on `main` |
| Next step | §4.4 create GitHub repository, then feature 001 |

## Open PRs (merge order)

None yet.

## Blockers

- `.claude/settings.json` (playbook §4.5) was not written: the agent's permission classifier refuses
  self-granted permissions. The intended content is in [`claude-settings.json`](claude-settings.json);
  the maintainer must copy it to `.claude/settings.json`. Tracked in a `needs-human` issue.

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
