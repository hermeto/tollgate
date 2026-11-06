# Research

## Decisions

- Use YAML for operator-maintained price tables because the repository already validates YAML config
  and policy files.
- Use USD per million tokens and integer token counters to avoid floating point token ambiguity.
- Use an interface-backed asynchronous writer with JSONL spool for deterministic local operation; a
  PostgreSQL adapter can be added without changing event producers.

## Alternatives considered

Embedding prices in gateway config was rejected because price revisions should be independently
reviewable. Synchronous database writes were rejected because they violate the request latency budget.
