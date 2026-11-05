# 0000. Record architecture decisions

- Status: accepted
- Date: 2026-11-04
- Deciders: agent (on the maintainer's behalf, feature 001)

## Context and problem statement

Tollgate is built feature by feature, partly by an autonomous agent that takes decisions on the
maintainer's behalf. Reviewers need to see what was decided, why, and which alternatives were rejected,
long after the pull request that introduced a change has merged.

## Considered options

1. Architecture Decision Records in the repository, MADR format.
2. Decisions only in pull request descriptions.
3. A wiki page.

## Decision outcome

Chosen option: ADRs in `docs/adr/`, using the [MADR](https://adr.github.io/madr/) structure in
[`template.md`](template.md), numbered `NNNN-kebab-title.md` following the plan in the product spec
(Appendix A §A.10).

### Consequences

- Good: decisions are versioned with the code, reviewed in pull requests and searchable.
- Good: decisions taken by the agent are explicit and can be overturned with a superseding ADR.
- Bad: one more file per decision; small decisions stay in the feature spec's Clarifications instead.

## More information

- Michael Nygard, "Documenting Architecture Decisions" (2011).
- MADR: <https://adr.github.io/madr/>
