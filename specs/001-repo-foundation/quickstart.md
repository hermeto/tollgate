# Quickstart: 001 repo-foundation

## Prerequisites

git ≥ 2.40, Go 1.27.3 (or any Go ≥ 1.21 that can download the toolchain named in go.mod), uv, jq.
On macOS `brew install go uv jq` is the shortest path; on Linux use the official Go tarball and
`curl -LsSf https://astral.sh/uv/install.sh | sh`.

## Validate locally

```bash
make dev-setup          # installs pre-commit (uv) and pinned Go tools
make lint test          # all linters + race tests; exit 0 on a fresh clone
go run ./cmd/tollgate version
```

Expected: both make targets exit 0; the binary prints `tollgate dev (commit none, built unknown)`.

## Validate the gate logic

```bash
echo '{"lint":{"result":"success"},"test":{"result":"skipped"}}' | scripts/ci-gate.sh pull_request; echo $?   # 1
go test ./internal/repotest/...   # ci-gate, DCO and workflow-pinning invariants
```

## Validate remotely (after the repository exists)

Open a PR; `gh pr checks --watch` shows `ci-gate` green. Push a commit without `-s` and the `dco` job
fails, turning `ci-gate` red.
