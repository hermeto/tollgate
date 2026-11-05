# Research: 002 provider-simulator

Sources were read on 2026-11-04 at the pinned versions. No provider API was called.

## R1. OpenAI Chat Completions wire format

- **Non-streaming**: object `chat.completion` with `id` (`chatcmpl-…`), `created` (Unix seconds),
  `model`, `choices[{index, message{role, content, refusal, annotations}, logprobs, finish_reason}]`,
  `usage{prompt_tokens, completion_tokens, total_tokens, prompt_tokens_details{cached_tokens,
  audio_tokens}, completion_tokens_details{reasoning_tokens, audio_tokens, accepted_prediction_tokens,
  rejected_prediction_tokens}}`, `service_tier`, `system_fingerprint`.
- **Streaming**: `text/event-stream`, frames `data: <json>\n\n`, terminated by `data: [DONE]\n\n`.
  Chunks are `chat.completion.chunk` objects with `choices[{index, delta{role?, content?, tool_calls?},
  logprobs, finish_reason}]`. First delta carries `role: "assistant"` and `content: ""`; the finish chunk
  has an empty delta and `finish_reason`. With `stream_options.include_usage: true`, every chunk has
  `"usage": null` and one extra chunk with `"choices": []` and the usage object precedes `[DONE]`.
- **Tool calls**: message `tool_calls[{id: "call_…", type: "function", function{name, arguments}}]`,
  `content: null`, `finish_reason: "tool_calls"`; streamed as `delta.tool_calls[{index, id, type,
  function{name, arguments}}]`.
- **Errors**: `{"error": {"message", "type", "param", "code"}}`; 401 `invalid_request_error`/
  `invalid_api_key`, 404 `invalid_request_error`/`model_not_found`, 429 `requests`/`rate_limit_exceeded`
  with `retry-after`, 500 `server_error`, 503 `service_unavailable`.
- **Sources**: SDK types `src/openai/types/chat/chat_completion.py`, `chat_completion_chunk.py`,
  `completion_usage.py` in [openai-python v3.46.0](https://github.com/openai/openai-python/tree/v3.46.0);
  SSE framing and `[DONE]` in
  [openai-node v7.51.0 `tests/lib/ChatCompletionStream.test.ts`](https://github.com/openai/openai-node/blob/v7.51.0/tests/lib/ChatCompletionStream.test.ts);
  [API reference: chat](https://platform.openai.com/docs/api-reference/chat/create),
  [streaming](https://platform.openai.com/docs/api-reference/chat-streaming),
  [errors](https://platform.openai.com/docs/guides/error-codes).

## R2. OpenAI Embeddings and Models

- **Embeddings**: request `{model, input: string | string[], encoding_format?: "float"|"base64",
  dimensions?}`; response `{"object":"list","data":[{"object":"embedding","index","embedding"}],
  "model","usage":{"prompt_tokens","total_tokens"}}`. With `base64`, `embedding` is the base64 of
  little-endian float32 values.
- **SDK behaviour that matters**: openai-python and openai-node request `encoding_format: "base64"` by
  default when the caller does not set it and decode client-side, so base64 support is mandatory for
  SDK compatibility.
- **Models**: `{"object":"list","data":[{"id","object":"model","created","owned_by"}]}`.
- **Sources**: `src/openai/resources/embeddings.py` and `types/create_embedding_response.py`
  (openai-python v3.46.0); [API reference: embeddings](https://platform.openai.com/docs/api-reference/embeddings/create).

## R3. Anthropic Messages wire format

- **Request headers**: `x-api-key`, `anthropic-version: 2023-06-01` (sent by every SDK).
- **Non-streaming**: `{"id":"msg_…","type":"message","role":"assistant","model","content":[{"type":
  "text","text"}],"stop_reason":"end_turn"|"max_tokens"|"stop_sequence"|"tool_use","stop_sequence":null,
  "usage":{"input_tokens","output_tokens","cache_creation_input_tokens","cache_read_input_tokens"}}`.
- **Streaming**: `event: <name>\ndata: <json>\n\n` with the sequence `message_start` (message with empty
  content and usage `{input_tokens, output_tokens: 1}`), `content_block_start` (index 0, empty text),
  `ping`, `content_block_delta` (`text_delta`) × N, `content_block_stop`, `message_delta`
  (`delta{stop_reason, stop_sequence}`, `usage{output_tokens}`), `message_stop`. Tool use streams a
  `tool_use` block with `input_json_delta` deltas.
- **Errors**: `{"type":"error","error":{"type","message"},"request_id"}`; types
  `invalid_request_error` 400, `authentication_error` 401, `not_found_error` 404,
  `rate_limit_error` 429 (with `retry-after`), `api_error` 500, `overloaded_error` 529. Mid-stream
  errors are sent as `event: error`.
- **Sources**: fixtures `tests/lib/streaming/fixtures/basic_response.txt` and `tool_use_response.txt` in
  [anthropic-sdk-python v2.5.0](https://github.com/anthropics/anthropic-sdk-python/tree/v2.5.0/tests/lib/streaming/fixtures)
  (MIT); [Messages API](https://docs.claude.com/en/api/messages),
  [streaming](https://docs.claude.com/en/docs/build-with-claude/streaming),
  [errors](https://docs.claude.com/en/api/errors).

## R4. Token counting approximation

- **Decision**: `internal/tokens` pre-tokenizes text with an RE2 pattern close to the `o200k_base`
  split (letter runs with an optional leading space, digit groups of 1–3, punctuation runs, whitespace
  runs) and counts `ceil(runes / k)` tokens per piece, with `k` tuned on the committed corpus. Chat
  messages add 3 tokens each plus 3 for reply priming (OpenAI cookbook method). Completion tokens are
  the number of generated pieces (each generated piece is one dictionary word, one token).
- **Rationale**: no tokenizer files or network needed at runtime; the gateway estimator (006) and the
  simulator agree by construction, and the error vs a real tokenizer is measured, not assumed.
- **Alternatives considered**: embedding tiktoken BPE ranks (≈ 4 MB, exact for OpenAI but not for
  Anthropic, slower); `len/4` heuristic (error > 20% on code and non-English text).
- **Measurement**: `test/sdk/python/token_error.py` compares the simulator's reported `prompt_tokens`
  with `tiktoken` 0.14.0 `o200k_base` on `testdata/tokens/corpus.jsonl`; result in `docs/simulator.md`.

## R5. Determinism

- **Decision**: content RNG = PCG(seed, fnv64(body)); latency RNG = PCG(seed ^ fnv64(body),
  occurrence(body)). Output length is drawn from the content RNG, so identical bodies get identical text.
- **Rationale**: spec Clarifications (FR-019). `math/rand/v2` PCG is fast and stable across Go versions.

## R6. Fault injection mechanics

- Error statuses: decided per request with the latency RNG; checked before any byte is written.
- Connection reset: `http.Hijacker`, `SetLinger(0)` on the TCP connection, then `Close()` → RST.
- Timeout before first token: wait for the duration or the request context, whichever first, before
  writing headers.
- Stall mid-stream: after N content tokens, wait for the duration (or context cancel).
- Outage window: `{start_after, duration, status}` resolved to absolute times when set; requests in
  the window get the status; expired windows are ignored.
- Live changes: `PUT /_sim/faults` replaces, `PATCH` merges, `DELETE` resets to start-up config;
  state is swapped with `atomic.Pointer` so the request path never locks.

## R7. CLI framework

- **Decision**: `spf13/cobra` v1.10.2.
- **Rationale**: the binary grows to ~10 subcommands (serve, sim, load, policy, keys, chargeback, audit,
  bench, extproc, report); cobra gives consistent flags, help and generated reference docs (018).
- **Alternatives considered**: standard `flag` with a hand-written dispatcher (no doc generation).

## R8. JSON Schema validation

- **Decision**: `santhosh-tekuri/jsonschema/v6` v6.0.3, draft 2020-12; schemas embedded with
  `go:embed` from `schemas/`.
- **Rationale**: pure Go, full draft 2020-12 support, good error paths; reused for policies (005).

## R9. SDK compatibility harness

- **Decision**: `scripts/sdk-compat.sh` builds `tollgate`, starts both flavours on free ports, waits for
  `/_sim/healthz`, then runs `test/sdk/python` (uv, `uv.lock`), `test/sdk/node` (`npm ci`,
  `package-lock.json`) and `test/sdk/go` (own `go.mod`), each exercising chat (both modes), embeddings
  and models (OpenAI) and messages (both modes, Anthropic). A `sdk-compat` CI job runs it and is
  aggregated by `ci-gate`. `TARGET=gateway` is reserved for 004.
- **Rationale**: separate modules keep SDK dependencies out of the main module and out of govulncheck
  noise for the binary.

## R10. Throughput measurement

- **Decision**: `BenchmarkOpenAIChatNonStreaming` (in-process, parallel) plus a documented `k6` run
  against `tollgate sim` (`bench/k6/sim-throughput.js` is created in 016; this feature records the
  numbers from a one-off run in `docs/simulator.md`).
