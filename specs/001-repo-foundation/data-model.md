# Data model: 001 repo-foundation

This feature has no runtime data. The entities below are the versioned configuration objects that
other features and the maintainer rely on.

## Required check — `ci-gate`

| Field | Value / rule |
| --- | --- |
| Name | `ci-gate` (job id and check name; referenced by `.github/rulesets/main.json`) |
| Inputs | `needs.<job>.result` for `lint`, `test`, `security`, `pr-title`, `dco`, `dependency-review` |
| Result values | `success`, `failure`, `cancelled`, `skipped` (GitHub) |
| Pass rule | every input is `success`; on non-`pull_request` events `pr-title`, `dco`, `dependency-review` may be `skipped` |
| Fail output | one line per offending job: `<job>: <result>` |

## Ruleset — `main`

Fields as in `.github/rulesets/main.json`: target `branch`, enforcement `active`, condition
`~DEFAULT_BRANCH`, bypass for repository admins (role id 5) via pull request only; rules: deletion,
non_fast_forward, required_linear_history, pull_request (0 approvals, dismiss stale, thread resolution,
squash only), required_status_checks (`ci-gate`, non-strict).

## Label

| Field | Rule |
| --- | --- |
| name | unique; prefixes `kind/`, `area/` or a fixed name from playbook §4.4 |
| color | 6-hex, no `#` |
| description | one sentence |

## ADR

MADR 4 structure: title `NNNN-kebab-title.md`, status (proposed | accepted | superseded by NNNN),
date, context and problem statement, considered options, decision outcome, consequences.
Numbering is global and follows Appendix A §A.10.
