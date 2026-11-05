#!/usr/bin/env bash
# Verifies a Signed-off-by trailer on every non-merge commit in <base>..<head> (DCO).
# Usage: check-dco.sh <base-ref> <head-ref>
set -euo pipefail

if [[ $# -ne 2 ]]; then
  echo "usage: check-dco.sh <base-ref> <head-ref>" >&2
  exit 2
fi
base="$1"
head="$2"

missing=0
commits="$(git rev-list --no-merges "${base}..${head}")"
for sha in ${commits}; do
  trailers="$(git log -1 --format=%B "${sha}" | git interpret-trailers --parse)"
  if ! grep -Eq '^Signed-off-by: .+ <[^>]+>$' <<<"${trailers}"; then
    echo "$(git log -1 --format='%h %s' "${sha}"): missing Signed-off-by"
    missing=$((missing + 1))
  fi
done

if ((missing > 0)); then
  echo "dco: ${missing} commit(s) without sign-off; fix with: git rebase --signoff ${base}"
  exit 1
fi
echo "dco: all commits signed off"
