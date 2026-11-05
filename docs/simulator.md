# Provider simulator (`tollgate sim`)

Tollgate v1.0 never calls a paid provider. Every test, demo and benchmark talks to `tollgate sim`, a
deterministic test double that speaks the OpenAI or Anthropic wire protocol. The simulator never makes
outbound calls (the optional embeddings forward URL is the only exception) and must not be exposed
outside a test or demo network: its admin endpoints are unauthenticated.

Specification: [`specs/002-provider-simulator/`](../specs/002-provider-simulator/spec.md).

## Running it

```bash
tollgate sim --flavor openai --port 9001
tollgate sim --flavor anthropic --config sim.yaml --seed 7
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--flavor` | from config | `openai` or `anthropic` |
| `--host` | `127.0.0.1` | listen address; use `0.0.0.0` only on a private network |
| `--port` | 9001 / 9002 | listen port |
| `--config` | none | YAML file validated by [`schemas/sim.schema.json`](../schemas/sim.schema.json) |
| `--seed` | 42 | seed for generated content, latency and fault decisions |
| `--log-requests` | off | log method, path, status and duration; never bodies or headers |

On SIGINT or SIGTERM the simulator stops accepting connections and gives in-flight requests 5 seconds
to finish. `GET /_sim/healthz` returns `{"status":"ok","flavor":…,"models":[…]}`.

Every provider endpoint requires the provider's auth header (`Authorization: Bearer …` for OpenAI,
`x-api-key` for Anthropic) and accepts any value, so tests can check that the gateway injects a
credential without the simulator ever holding a real key. A missing header returns the provider's 401
error body.

## Configuration

```yaml
flavor: openai
seed: 42
models: [gpt-4o-mini, text-embedding-3-small]
latency:                       # milliseconds
  ttft: {kind: lognormal, mu: 5.3, sigma: 0.4}
  inter_token: {kind: normal, mean: 20, stddev: 5, min: 1}
output_tokens: {kind: normal, mean: 120, stddev: 40, min: 1, max: 1024}
embeddings: {dimensions: 1536}
request_log: {size: 1000}
```

Distributions are `fixed` (`value`), `normal` (`mean`, `stddev`) or `lognormal` (`mu`, `sigma` of the
underlying normal), each with optional `min` and `max` clamps; negative samples become 0.

## Determinism

- Generated text and output length are a pure function of the seed and the request (path and body):
  the same request always gets the same answer, whatever else is in flight (generator added in the
  next slice of feature 002).
- Latency and fault decisions use a per-request random stream seeded with the seed, the request hash
  and how many times that request has been seen, so overlapping distinct requests do not change each
  other's samples.
- Response ids and timestamps are unique per response and are the only fields that differ on replay.
