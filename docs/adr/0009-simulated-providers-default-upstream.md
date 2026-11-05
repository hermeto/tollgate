# 0009. Simulated providers as the default upstream

- Status: accepted
- Date: 2026-11-05
- Deciders: maintainer (v1.0 uses simulated providers only, product spec §A.9); mechanism by agent (feature 002)

## Context and problem statement

Tollgate's report claims what each policy costs in latency and saves in dollars. Those numbers must be
reproducible by anyone, at zero cost, without provider keys, and every test, demo and benchmark needs
upstreams that can be told to be slow, fail, rate-limit or go down on demand. Real providers are
non-deterministic, cost money, rate-limit the tester and cannot be made to fail on cue.

## Considered options

1. A built-in simulator (`tollgate sim`) that speaks the OpenAI and Anthropic wire protocols, with a
   seeded latency model and live fault injection, as the default upstream in every profile.
2. Recorded HTTP cassettes replayed by a generic mock server (for example VCR-style fixtures).
3. Real providers with small budgets in CI, simulators only for failure tests.

## Decision outcome

Chosen option 1. Every profile (`local`, `envoy`, `k8s`) points at the simulators by default; real
providers are configured only by explicit operator configuration (post-v1.0 `live` profile), and code
never falls back to a real endpoint when keys are absent.

### Consequences

- Good: zero cost and no secrets anywhere in CI; deterministic content and fault sequences per seed;
  latency and failures controllable per experiment; the same binary ships the simulator.
- Good: wire-format fidelity is checked twice: structural contract tests against fixtures from official
  documentation and SDK test suites, and the official SDKs (Python, TypeScript, Go) run against the
  simulators in CI.
- Bad: results say nothing about real model quality or real provider latency distributions; the report
  lists this under threats to validity, and validation with real providers is post-v1.0.
- Bad: token counts are an approximation (7.4% mean error against `o200k_base`; unknown for Claude).
- Bad: under 1,000 concurrent streams with a co-located load generator the simulator's TTFT tail reaches
  +23 ms at p99, so gateway overhead is measured inside the gateway, not as a client-side p99 difference
  (issue #12).
- Option 2 was rejected because cassettes cannot inject faults or vary latency, and replaying recorded
  traffic would need captures from paid APIs. Option 3 was rejected by the constitution (Principle VI).

## More information

- [docs/simulator.md](../simulator.md), [specs/002-provider-simulator/](../../specs/002-provider-simulator/spec.md)
- Fixtures and sources: [testdata/wire/README.md](../../testdata/wire/README.md)
