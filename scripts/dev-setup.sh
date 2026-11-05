#!/usr/bin/env bash
# Installs the pinned developer toolchain. Requires git, Go and uv; works without Homebrew.
set -euo pipefail

PRE_COMMIT_VERSION="4.7.0"

need() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "error: '$1' is required. $2" >&2
    exit 1
  fi
}

need git "Install it from https://git-scm.com/downloads."
need go "Install Go from https://go.dev/dl/ (any Go >= 1.21 downloads the toolchain pinned in go.mod)."
need uv "Install uv: curl -LsSf https://astral.sh/uv/install.sh | sh"
need jq "Install jq from your package manager (apt install jq, dnf install jq, brew install jq)."

echo "==> Go toolchain: $(go version)"
echo "==> Downloading pinned Go tools (tools/go.mod)"
go mod download -modfile=tools/go.mod

echo "==> Installing pre-commit ${PRE_COMMIT_VERSION} with uv"
uv tool install --quiet "pre-commit==${PRE_COMMIT_VERSION}"
UV_BIN="$(uv tool dir --bin)"
case ":${PATH}:" in
  *":${UV_BIN}:"*) ;;
  *) echo "note: add ${UV_BIN} to PATH to use pre-commit directly"; export PATH="${UV_BIN}:${PATH}" ;;
esac

echo "==> Installing git hooks and hook environments"
pre-commit install --install-hooks

echo "==> Done. Next: make lint test"
