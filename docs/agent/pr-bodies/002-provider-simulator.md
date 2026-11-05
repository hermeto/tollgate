## Summary

Slice 1 of 8 of the provider simulator (`tollgate sim`): the Spec Kit artifacts for the whole feature,
the move of the `tollgate` CLI to cobra, and the simulator configuration (`sim.Config`) loaded from YAML
and validated by `schemas/sim.schema.json`, with flavor defaults, flag overrides and the semantic checks
the schema cannot express (fault rates summing above 1, min above max, 529 outside Anthropic).
Spec: [`specs/002-provider-simulator/spec.md`](https://github.com/hermeto/tollgate/blob/002-provider-simulator/specs/002-provider-simulator/spec.md). Tracking: #3.

## Stack

Based on #4; merge after it. Next slice: `002-provider-simulator-iter1` (server skeleton).

## Requirements covered

Feature spec FR-011 (config validated by a JSON Schema) and the configuration part of FR-005, FR-007
and FR-009. The CLI, endpoints and behaviour arrive in later slices.

## Decisions for the reviewer

- CLI framework: cobra v1.10.2 (consistent flags and help for ~10 subcommands; generates the CLI reference in 018). Research R7.
- JSON Schema validator: santhosh-tekuri/jsonschema v6.0.3, draft 2020-12, schemas embedded with `go:embed` (reused by policies in 005). Research R8.
- Latency distributions are in milliseconds; YAML status keys like `429:` are normalised to strings before validation.
- 002 is delivered in 8 stacked slices to respect the 800-line rule (tasks.md "PR slices"); slices chain linearly and 003 branches from the last one.
- Clarifications for the whole feature (auth, tool calls, loopback default, determinism, reference tokenizer): [spec.md](https://github.com/hermeto/tollgate/blob/002-provider-simulator/specs/002-provider-simulator/spec.md#clarifications).

## How it was verified

- `make lint test`: clean; schema tests cover valid, unknown field, bad flavor, missing distribution parameters, bad status key, rate above 1 and bad duration; config tests cover defaults, file plus overrides, rate sum above 1, min above max and missing flavor.
- CLI tests: `version`, `--version`, `-v`, unknown command (exit 2), no arguments (usage, exit 2), bad flag (exit 2).

## Not verified

- Nothing user-visible beyond `tollgate version` yet; the `sim` command arrives in slice 2.

## Performance impact

None on the gateway hot path (no gateway code yet).

## Checklist

- [x] Tests added or updated
- [x] Docs updated in this PR (README, docs/, runbooks) — user docs arrive with the `sim` command in slice 2
- [x] ADR added or updated for architectural decisions — ADR-0009 ships with the last slice
- [x] No secrets, real keys or real personal data
- [x] New dependencies, images and actions pinned
- [x] No prompt/response content or PII in logs, traces, metrics or events
- [x] Commits signed off (DCO) and Conventional
