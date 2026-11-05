# Governance

## Roles

- **Contributors**: anyone who opens an issue, discussion or pull request.
- **Maintainers**: listed in [MAINTAINERS.md](MAINTAINERS.md). They review and merge pull requests,
  triage issues, cut releases and enforce the [Code of Conduct](CODE_OF_CONDUCT.md).

## Decision making

- Day-to-day changes are decided in pull requests. A maintainer merges once `ci-gate` is green and
  review threads are resolved.
- Architectural decisions are recorded as ADRs in [`docs/adr/`](docs/adr). An ADR is proposed in a pull
  request and accepted when a maintainer merges it.
- The project [constitution](.specify/memory/constitution.md) supersedes other practices. It changes
  only through a pull request labeled `constitution`, with a semantic version bump.
- Maintainers seek consensus. When it cannot be reached, the lead maintainer decides and records the
  rationale in the pull request or ADR.

## Becoming a maintainer

Contributors with a sustained record of reviewed contributions may be nominated by a maintainer in a
pull request to MAINTAINERS.md. Nomination is accepted when no maintainer objects within 7 days.

## Changes to this document

Through a pull request reviewed by a maintainer.
