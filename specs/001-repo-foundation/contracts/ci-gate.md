# Contract: `scripts/ci-gate.sh`

```text
Usage: ci-gate.sh <event-name>
Stdin: JSON object equal to `toJSON(needs)` from GitHub Actions, e.g.
       {"lint":{"result":"success","outputs":{}}, "dco":{"result":"skipped","outputs":{}}}
Exit:  0 when every job passed per the rule below, 1 otherwise, 2 on usage/parse error.
Stdout: "ci-gate: all required jobs passed" or one line per failing job "<job>: <result>".
```

Rule: a job passes when its result is `success`. On events other than `pull_request` and
`pull_request_target`, the PR-only jobs `pr-title`, `dco` and `dependency-review` also pass when
`skipped`. An empty object is a failure (nothing was checked).

# Contract: `scripts/check-dco.sh`

```text
Usage: check-dco.sh <base-ref> <head-ref>
Exit:  0 when every non-merge commit in base..head has a Signed-off-by trailer; 1 otherwise; 2 on usage error.
Stdout: "<short-sha> <subject>: missing Signed-off-by" per offending commit.
```
