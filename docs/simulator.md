# Provider simulator (`tollgate sim`)

Tollgate v1.0 never calls a paid provider. Every test, demo and benchmark talks to `tollgate sim`, a
deterministic test double that speaks the OpenAI or Anthropic wire protocol. The simulator never makes
outbound calls (the optional embeddings forward URL is the only exception) and must not be exposed
outside a test or demo network: its admin endpoints are unauthenticated.

Design decision: [ADR-0009](adr/0009-simulated-providers-default-upstream.md). Specification:
[`specs/002-provider-simulator/`](../specs/002-provider-simulator/spec.md).

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

## OpenAI flavour

| Method | Path | Behaviour |
| --- | --- | --- |
| POST | `/v1/chat/completions` | JSON or SSE (`stream: true`); `stream_options.include_usage` adds `"usage": null` to every chunk and a final chunk with `choices: []` and usage, before `data: [DONE]`. `max_completion_tokens` (or `max_tokens`) cuts the output with `finish_reason: "length"`. A forced tool (`tool_choice: "required"` or naming a function) returns one call with arguments `{}` and `finish_reason: "tool_calls"`. `n > 1` returns 400. `stop` sequences are accepted and ignored: generated text never contains them. |
| POST | `/v1/embeddings` | Deterministic unit vectors (default 1536 dimensions, `dimensions` overrides) for a string, an array of strings or token arrays; `encoding_format` `float` or `base64` (little-endian float32, the SDKs' default). |
| GET | `/v1/models`, `/v1/models/{model}` | Configured models. |

With `embeddings.forward_url` set (the full URL of an OpenAI-compatible embeddings endpoint, for example
a local text-embeddings-inference container in feature 013), `/v1/embeddings` is forwarded unchanged and
the target's answer is returned; if the target fails or is unreachable the simulator answers 502 and
never falls back to synthetic vectors. This is the simulator's only outbound call.

Errors use the OpenAI body `{"error":{"message","type","param","code"}}`: 400 for malformed requests,
401 `invalid_api_key` without a Bearer token, 404 `model_not_found` for unknown models, 413 for bodies
over 32 MiB. Every response carries `x-request-id`. Shapes are checked against the fixtures in
[`testdata/wire/openai/`](../testdata/wire/README.md).

## Anthropic flavour

| Method | Path | Behaviour |
| --- | --- | --- |
| POST | `/v1/messages` | JSON or SSE (`stream: true`) with the official event order: `message_start`, `content_block_start`, `ping`, `content_block_delta`…, `content_block_stop`, `message_delta` (stop reason and output usage), `message_stop`. `max_tokens` is required, as in the real API; reaching it gives `stop_reason: "max_tokens"`. `stop_sequences` are accepted and ignored. A forced tool (`tool_choice` of type `tool` or `any`) returns one `tool_use` block with input `{}` streamed as `input_json_delta`, and `stop_reason: "tool_use"`. |

`system` and message `content` may be strings or arrays of blocks; text, tool inputs and tool results
count towards `input_tokens`. Usage reports `cache_creation_input_tokens` and `cache_read_input_tokens`
as 0 and `service_tier: "standard"`. Errors use `{"type":"error","error":{"type","message"},
"request_id"}` (`invalid_request_error` 400, `authentication_error` 401, `not_found_error` 404,
`request_too_large` 413). The `anthropic-version` header is accepted and not enforced; every response
carries `request-id`. Shapes are checked against [`testdata/wire/anthropic/`](../testdata/wire/README.md).

## Latency model

`latency.ttft` and `latency.inter_token` are sampled per response from the per-request random
stream. Nothing — not even response headers — is sent before the first token is ready, so time to
first token equals time to first byte, which is what the gateway's fallback logic measures (feature
009). Content chunks are then separated by inter-token delays. Non-streaming responses wait for the
whole generation (first token plus one delay per remaining token). Embeddings wait for one TTFT
sample. Every wait ends as soon as the client disconnects.

## Fault injection

Faults come from the `faults` block of the configuration and can be changed live:

```bash
curl -X PUT   localhost:9001/_sim/faults -d '{"error_rates":{"429":0.2,"500":0.05},"retry_after_seconds":2}'
curl -X PATCH localhost:9001/_sim/faults -d '{"outage":{"start_after":"10s","duration":"30s"}}'
curl          localhost:9001/_sim/faults        # active config, outage with absolute start and end
curl -X DELETE localhost:9001/_sim/faults       # back to the start-up config
```

| Fault | Effect |
| --- | --- |
| `error_rates` | per-status probability (429, 500, 503; 529 on the Anthropic flavour) of a provider-shaped error; `retry_after_seconds` adds `retry-after` to 429, 503 and 529 |
| `reset_rate` | probability of closing the connection with a TCP reset and no response (logged with status 0) |
| `timeout_before_first_token` | with probability `rate`, hold the response `duration` before any byte |
| `stall_mid_stream` | with probability `rate`, pause `duration` after `after_tokens` content tokens (streams only) |
| `outage` | every provider request fails with `status` (default 503) from `start_after` for `duration` |

Each request draws its outcome from its own deterministic random stream in a fixed order (outage,
then one draw shared by error rates and resets, then timeout, then stall), so the same seed and
request sequence always produce the same faults. `error_rates` plus `reset_rate` must not exceed 1.
PUT replaces the config, PATCH applies an RFC 7396 merge patch; a new relative outage window starts
when it is set. Admin endpoints are never affected by faults.

## Admin API

| Method | Path | Result |
| --- | --- | --- |
| GET | `/_sim/healthz` | 200 with flavor and models |
| GET | `/_sim/requests` | `{"requests":[…]}`: the last `request_log.size` provider requests, oldest first, with method, path, headers, body (capped at 64 KiB, `body_truncated` when cut) and status |
| DELETE | `/_sim/requests` | 204, clears the log |
| GET, PUT, PATCH, DELETE | `/_sim/faults` | read, replace, merge or reset the fault config (see above) |

The request log lets end-to-end tests assert what the gateway actually sent upstream, for example that
PII was redacted or that client headers were stripped. It keeps headers verbatim, so it must only ever
see fake keys; the simulator's own logs never contain bodies or header values. Admin calls are not
recorded.

## SDK compatibility

`make sdk-compat` starts both flavours and runs the official SDKs, changing only the base URL:

| Language | OpenAI SDK | Anthropic SDK | Client |
| --- | --- | --- | --- |
| Python (uv, `test/sdk/python/uv.lock`) | `openai` 3.46.0 | `anthropic` 2.5.0 | `test/sdk/python/compat.py` |
| TypeScript (npm, `test/sdk/node/package-lock.json`) | `openai` 7.51.0 | `@anthropic-ai/sdk` 1.5.0 | `test/sdk/node/compat.mjs` |
| Go (`test/sdk/go/go.mod`) | `openai-go/v3` v3.95.0 | `anthropic-sdk-go/v2` v2.6.0 | `test/sdk/go/main.go` |

Each client checks chat (JSON and streaming with usage), a forced tool call (Python, TypeScript),
embeddings and the model list for OpenAI, and messages (JSON and streaming) for Anthropic. The suite
runs in CI as the `sdk-compat` job, aggregated by `ci-gate`. `TARGET=gateway` arrives with feature 004.

## Performance

Measured on 2026-11-05, Apple M4 (10 cores, 24 GiB), macOS, Go 1.27.3, k6 2.3.0, with the load
generator on the same machine. Scripts: [`bench/k6/sim-throughput.js`](../bench/k6/sim-throughput.js) and
[`bench/k6/sim-ttft.js`](../bench/k6/sim-ttft.js) with [`sim-ttft.yaml`](../bench/k6/sim-ttft.yaml).

| Measurement | Result |
| --- | --- |
| Non-streaming chat, zero latency, open-loop 10,000 req/s for 20 s (k6) | 10,000 req/s sustained, 0 errors, 0 dropped; median 39 µs, p95 46 µs; simulator ≈ 0.3 cores |
| Same at 30,000 req/s | sustained (6 of 600,000 iterations dropped), p95 106 µs |
| In-process benchmark `BenchmarkOpenAIChatNonStreaming` (client and server share the CPU) | ≈ 94,000 req/s |
| TTFT accuracy, fixed 100 ms, 100 streams (Go client) | median +2.5 ms, p90 +6.2 ms, p99 +10.2 ms |
| TTFT accuracy, fixed 100 ms, 1,000 streams (Go client, ≈ 7,600 req/s) | median +3.4 ms, p90 +13.6 ms, p99 +23.5 ms |
| Same with k6 | median +3.8 ms, p90 +14.8 ms, p99 +27.3 ms |

The simulator is never the throughput bottleneck. Its time-to-first-token tail under heavy concurrency
is a few milliseconds at the median but tens of milliseconds at p99 with a co-located load generator
(GC, CPU saturation, macOS timers and the load generator were ruled out; see
[#12](https://github.com/hermeto/tollgate/issues/12)). Gateway overhead must therefore be measured
inside the gateway (per-Stage timing) rather than as a difference of client-side p99s.

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
  the same request always gets the same answer, whatever else is in flight.
- Latency and fault decisions use a per-request random stream seeded with the seed, the request hash
  and how many times that request has been seen, so overlapping distinct requests do not change each
  other's samples.
- Response ids and timestamps are unique per response and are the only fields that differ on replay.

## Token counting

`internal/tokens` approximates `o200k_base` without tokenizer files. Text is split the way `o200k_base`
pre-tokenizes it (letter runs with an optional leading space, CJK runs, digit groups of up to three,
symbol runs, whitespace runs) and each piece is converted to tokens with fitted constants:

| Piece | Tokens |
| --- | --- |
| ASCII letter run of n letters | 1 if n ≤ 9, else 1 + ⌈(n − 9) / 4⌉ |
| Letter run with non-ASCII letters | 1 if n ≤ 6, else 1 + ⌈(n − 6) / 5⌉ |
| CJK run of n characters | ⌈n / 1.5⌉ |
| Digit run of n digits | ⌈n / 3⌉ |
| Symbol run of n characters | ⌈n / 3⌉ |
| Newline run | 1 |
| Space run of n | ⌈(n − 1) / 4⌉ (one space is absorbed by the next word) |

Chat prompts add 3 tokens per message and 3 for reply priming (OpenAI cookbook method). The simulator
reports `prompt_tokens` / `input_tokens` with this method and counts one completion token per generated
piece, so the gateway's estimates (features 006 and 007) agree with simulated usage by construction.

**Measured error** against tiktoken 0.14.0 `o200k_base` on the 48-sample synthetic corpus
[`testdata/tokens/corpus.jsonl`](../testdata/tokens/corpus.jsonl) (English, Portuguese, Spanish, French,
German, Russian, Chinese, Japanese, Korean, code, JSON, YAML, logs, numbers):

| Metric | Value |
| --- | --- |
| Mean absolute error | 7.4% (enforced ≤ 15% by `TestErrorAgainstReference`) |
| Aggregate bias | +2.2% (enforced within ±5%) |
| Holdout error (fit on half the corpus, test on the other half) | 5.8% |
| Worst samples | long compound English words (+53%), tab-indented Go (+32%), emoji (−26%), URLs (+26%), Korean (+24%) |
| Speed | 4.2 µs per 4 KB of ASCII text, zero allocations (Apple M-series) |

The Anthropic tokenizer is not public; the Anthropic flavour uses the same approximation, so its error
against Claude models is unknown. Regenerate the reference counts after changing the corpus with
`uv run --locked --script scripts/tokens/reference.py` (downloads the public `o200k_base` file once).
