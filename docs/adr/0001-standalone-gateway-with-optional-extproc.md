# 0001. Standalone Go gateway with an optional Envoy ext_proc adapter

- Status: accepted
- Date: 2026-11-05
- Deciders: maintainer (product spec §A.9); documented by the agent in feature 003

## Context and problem statement

Tollgate has to enforce per-team policies (identity, rate limits, budgets, PII, cache, routing) on every
LLM request and measure what each one costs. Many platforms already run Envoy Gateway or Agent Router
(formerly Envoy AI Gateway) and would rather add policies to that data plane than run another proxy.
Where should the policy pipeline live?

## Considered options

1. A standalone Go gateway that owns the request path end to end, plus an adapter that exposes the same
   pipeline as an Envoy External Processing (ext_proc) service.
2. Policies only as an ext_proc service; Envoy (or Agent Router) does all proxying and routing.
3. Extend an existing gateway (for example LiteLLM) with the missing governance features.

## Decision outcome

Chosen option 1, as decided by the maintainer. The gateway (`tollgate serve`) is the primary data plane;
`tollgate extproc` (feature 014) runs the same `internal/pipeline` Stages behind Envoy, and experiment X7
measures the cost of that extra hop.

### Consequences

- Good: one place owns streaming, fallback before the first token and cost attribution, so every policy
  can be measured without an Envoy dependency; the local profile stays small.
- Good: the Stage contract (feature 003) is transport-independent, so the ext_proc adapter maps
  ext_proc messages to the same Request/Chunk types instead of re-implementing policies.
- Bad: two integration surfaces to keep in parity; the parity table in 014 documents which Stages do not
  apply in ext_proc mode (for example cross-provider fallback is Envoy's job there).
- Bad: Tollgate is "another proxy" for teams without Envoy; the report must show its overhead is small.

## More information

- Product spec §A.3 (architecture) and §A.9 (decisions already taken).
- Agent Router: <https://aaif.io/projects/agent-router>.
