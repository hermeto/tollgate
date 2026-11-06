# Release checklist

- [ ] Review stacked PRs bottom-up.
- [ ] Run `make lint test`.
- [ ] Run `go test -race ./...`.
- [ ] Run simulator-only SDK compatibility checks.
- [ ] Verify no provider credentials or real personal data are present.
- [ ] Tag the release only after required CI checks pass.
