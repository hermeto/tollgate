# Data model: 002 provider-simulator

## Config (`schemas/sim.schema.json`)

| Field | Type | Default | Rule |
| --- | --- | --- | --- |
| `flavor` | `openai` \| `anthropic` | — | required (flag `--flavor` overrides) |
| `host` | string | `127.0.0.1` | |
| `port` | int | 9001 (openai), 9002 (anthropic) | 1–65535 |
| `seed` | uint64 | 42 | |
| `models` | string[] | flavour defaults | ≥ 1 entry |
| `latency.ttft` | Distribution | fixed 0 ms | |
| `latency.inter_token` | Distribution | fixed 0 ms | |
| `output_tokens` | Distribution | normal(mean 120, stddev 40, min 1, max 1024) | clamped by request max tokens |
| `embeddings.dimensions` | int | 1536 | 1–8192; request `dimensions` overrides |
| `embeddings.forward_url` | URL | empty | when set, `/v1/embeddings` is proxied |
| `request_log.size` | int | 1000 | 0 disables |
| `faults` | FaultConfig | none | |

## Distribution

| Field | Type | Rule |
| --- | --- | --- |
| `kind` | `fixed` \| `normal` \| `lognormal` | required |
| `value` | number | `fixed` only (ms for latency, tokens for output) |
| `mean`, `stddev` | number | `normal` only; stddev ≥ 0 |
| `mu`, `sigma` | number | `lognormal` only (of the underlying normal, in ln ms / ln tokens) |
| `min`, `max` | number | optional clamps; min ≤ max |

## FaultConfig

| Field | Type | Rule |
| --- | --- | --- |
| `error_rates` | map status → rate | status ∈ {429, 500, 503, 529}; each rate 0–1 |
| `retry_after_seconds` | int | ≥ 0, used with 429 and 529 |
| `reset_rate` | number | 0–1 |
| `timeout_before_first_token` | `{rate, duration}` | duration Go syntax (`5s`) |
| `stall_mid_stream` | `{rate, after_tokens, duration}` | after_tokens ≥ 0 |
| `outage` | `{start_after, duration, status}` | status default 503; resolved to absolute `start`, `end` when applied |

Validation: the sum of `error_rates` and `reset_rate` MUST be ≤ 1.

## RequestRecord

`{time, method, path, headers (map, values verbatim), body (string), status}` — ring buffer of
`request_log.size` entries, newest last.

## State transitions

Fault state: `startup` → (PUT/PATCH) → `live` → (DELETE) → `startup`. An outage window is
`pending` → `active` (start ≤ now < end) → `expired`.
