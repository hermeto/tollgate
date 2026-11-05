# Contract: Makefile targets

| Target | Behaviour | Network |
| --- | --- | --- |
| `help` (default) | Lists targets with one-line descriptions | none |
| `dev-setup` | Verifies git, Go (exact version from go.mod or newer via toolchain), uv; installs pre-commit with uv; downloads Go tools from `tools/go.mod`; installs git hooks | package registries only |
| `fmt` | `gofmt -s -w` and `golangci-lint fmt` | none |
| `lint` | golangci-lint, govulncheck, `pre-commit run --all-files` (actionlint, shellcheck, yamllint, markdownlint-cli2, editorconfig-checker, whitespace hooks) | vuln.go.dev, hook downloads on first run |
| `test` | `go test -race -count=1 ./...` | none |
| `bench-micro` | `go test -run '^$' -bench . -benchmem ./...` | none |

All targets exit non-zero on the first failure and never call a model provider.
