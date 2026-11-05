#!/usr/bin/env bash
# Aggregates GitHub Actions job results into the single required `ci-gate` status.
# Usage: ci-gate.sh <event-name> < needs.json   (stdin is `toJSON(needs)`)
# Contract: specs/001-repo-foundation/contracts/ci-gate.md
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: ci-gate.sh <event-name> < needs.json" >&2
  exit 2
fi
event="$1"
needs="$(cat)"

if ! jq -e 'type == "object"' >/dev/null 2>&1 <<<"${needs}"; then
  echo "ci-gate: stdin is not a JSON object" >&2
  exit 2
fi
if [[ "$(jq 'length' <<<"${needs}")" -eq 0 ]]; then
  echo "ci-gate: no jobs to check"
  exit 1
fi

is_pr=false
case "${event}" in
  pull_request | pull_request_target) is_pr=true ;;
esac

# Jobs that only run on pull requests; they may be skipped on other events.
pr_only='["pr-title","dco","dependency-review"]'

failures="$(jq -r --argjson pronly "${pr_only}" --argjson ispr "${is_pr}" '
  to_entries[]
  | .key as $k
  | (.value.result // "missing") as $r
  | select($r != "success")
  | select((($r == "skipped") and ($ispr | not) and ($pronly | index($k) != null)) | not)
  | "\($k): \($r)"
' <<<"${needs}")"

if [[ -n "${failures}" ]]; then
  echo "${failures}"
  exit 1
fi
echo "ci-gate: all required jobs passed"
