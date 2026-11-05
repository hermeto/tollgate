# Specification Analysis Report: 002 provider-simulator

Run of `/speckit-analyze` on spec.md, plan.md and tasks.md (2026-11-05).

| ID | Category | Severity | Location(s) | Summary | Recommendation |
| --- | --- | --- | --- | --- | --- |
| A1 | Coverage | MEDIUM | SC-004; T014, T019 | Byte-identical replay had no explicit test task. | Added replay tests to T014 and T019 (applied). |
| A2 | Coverage | MEDIUM | contracts/cli.md; FR-010; T009 | `--log-requests` and the "no bodies or auth in own logs" rule had no task. | T009 now covers both (applied). |
| A3 | Coverage | LOW | FR-015 | "No outbound calls except forward URL" is verified by inspection only. | Accepted: the only HTTP client in `sim/` is created for the forward URL; reviewers check it. |
| A4 | Terminology | LOW | spec, contracts | Prose says "flavour", identifiers say `flavor`. | Accepted: identifiers follow the CLI flag; prose keeps the playbook's spelling. |

## Coverage summary

| Requirement | Tasks |
| --- | --- |
| FR-001 | T009 |
| FR-002 | T015, T016 |
| FR-003 | T020 |
| FR-004 | T017, T020 |
| FR-005 | T021, T022 |
| FR-006 | T006, T012 |
| FR-007 | T023, T024 |
| FR-008 | T010, T011, T012 |
| FR-009 | T016, T027 |
| FR-010 | T009, T025, T026 |
| FR-011 | T003, T004, T005 |
| FR-012 | T013, T018 |
| FR-013 | T028–T031 |
| FR-014, FR-016, FR-018, FR-020 | T008 |
| FR-015 | design (A3) |
| FR-017 | T015, T020 |
| FR-019 | T006 |
| SC-001 | T028–T031 |
| SC-002 | T014, T019 |
| SC-003, SC-007 | T032 |
| SC-004 | T014, T019 |
| SC-005 | T023 |
| SC-006 | T033 |

Constitution alignment: no conflicts. CRITICAL: 0, HIGH: 0, MEDIUM: 2 (fixed), LOW: 2 (accepted).
