## Summary

Feature 018 closes the v1.0 backlog with a changelog, maintainer release checklist and manually
triggered simulator-safe CI verification workflow.

## Base

Based on `017-report-generator`; merge after PR #34.

## Verification

- `make lint test`
- `go test -race ./...`
- `actionlint .github/workflows/release.yml`
