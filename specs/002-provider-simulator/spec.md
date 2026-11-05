# Feature Specification: Deterministic OpenAI and Anthropic simulators with fault injection

**Feature Branch**: `002-provider-simulator`

**Created**: 2026-11-04

**Status**: Draft

**Input**: User description: "Provider simulator (`tollgate sim`). Tollgate v1.0 never calls paid
providers. Every test, demo and benchmark needs upstreams that speak the real wire protocols, report
realistic token usage, and can be told to be slow, fail, rate-limit or go down — deterministically."

## Clarifications

### Session 2026-11-04

- Q: How do the simulators treat credentials? → A: They require the provider's auth header (`Authorization: Bearer …` for OpenAI, `x-api-key` for Anthropic) to be present and accept any value; a missing header returns the provider's 401 error. (decided by agent: SDKs always send one, and tests can prove the gateway injects provider credentials without the simulator ever holding real keys.)
- Q: How are tool calls simulated? → A: When the request forces a tool (OpenAI `tool_choice` naming a function or `required`; Anthropic `tool_choice` of type `tool` or `any`), the response is one tool call to the first eligible tool with arguments `{}` and finish reason `tool_calls` / stop reason `tool_use`; otherwise the response is text. (decided by agent: enough for translation tests in 004 without modelling argument generation.)
- Q: Which address does the simulator bind to by default? → A: `127.0.0.1`; `--host` overrides it (the Compose profile uses `0.0.0.0` on a private network). (decided by agent: the admin API is unauthenticated, so the safe default is loopback.)
- Q: How is latency kept deterministic under concurrency? → A: Each request's random stream is seeded from the global seed, a hash of the request body and the occurrence count of that body, so distinct requests get the same samples regardless of interleaving. (decided by agent: arrival-order seeding breaks determinism as soon as requests overlap.)
- Q: Which reference tokenizer defines the "known error"? → A: `o200k_base` via tiktoken (pinned) on a committed corpus; the same approximation is used for the Anthropic flavour, whose tokenizer is not public, and the documentation says so. (decided by agent: o200k_base is the current OpenAI family encoding and freely available offline.)

### Session 2026-11-05

- Q: The original SC-007 (p99 within +5 ms at 1,000 streams) is not met on the available hardware (p99 +23.5 ms with a Go client, +27.3 ms with k6, load generator on the same host); what should the criterion be? → A: Median within +5 ms (measured +3.4 ms), with p90/p99 documented; gateway overhead in later features is measured inside the gateway (per-Stage timing), with client-side differentials only as a cross-check. (decided by agent: GC, CPU saturation, macOS timers and the load generator were ruled out; the tail is OS and runtime scheduling with a co-located load generator. The maintainer decides in #12 whether to keep the p99 target on dedicated hardware.)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - OpenAI-compatible simulator (Priority: P1)

A developer starts an OpenAI-flavoured simulator on a local port and points an unmodified official
OpenAI client at it by changing only the base URL. Chat completions (streaming and non-streaming),
embeddings and the model list behave like the real service, including token usage and streamed usage
when the client asks for it.

**Why this priority**: every gateway feature from 003 onwards forwards to this simulator.

**Independent Test**: start the simulator, run the official OpenAI clients for Python, TypeScript and
Go against it, and check that chat (both modes), embeddings and model listing succeed and report usage.

**Acceptance Scenarios**:

1. **Given** an OpenAI simulator, **When** a client sends a non-streaming chat completion, **Then** it
   receives a `chat.completion` object with one choice, a `stop` or `length` finish reason and usage
   whose total equals prompt plus completion tokens.
2. **Given** a streaming chat completion with usage requested, **When** the stream ends, **Then** the
   client received role, content and finish chunks, a final chunk with empty choices carrying usage,
   and the `[DONE]` terminator.
3. **Given** an embeddings request for N inputs, **When** it is served, **Then** N vectors of the
   documented dimension are returned in the requested encoding (float list or base64), and the same
   input always yields the same vector.
4. **Given** a model list request, **When** it is served, **Then** the configured models are listed.
5. **Given** a request with `max_tokens` (or `max_completion_tokens`) smaller than the planned output,
   **When** it is served, **Then** the output is cut at that limit and the finish reason is `length`.

---

### User Story 2 - Anthropic-compatible simulator (Priority: P1)

A developer starts an Anthropic-flavoured simulator and points an unmodified official Anthropic client
at it. The Messages endpoint returns the official response shape and, when streaming, the official
event sequence with usage fields.

**Why this priority**: Anthropic passthrough (004) and cross-provider fallback (009) need it.

**Independent Test**: run the official Anthropic clients (Python, TypeScript, Go) against the
simulator for streaming and non-streaming messages.

**Acceptance Scenarios**:

1. **Given** an Anthropic simulator, **When** a client creates a message, **Then** it receives a
   `message` object with a text content block, a stop reason and input/output token usage.
2. **Given** a streaming request, **When** the stream is consumed, **Then** the events arrive in the
   documented order: `message_start`, `content_block_start`, `ping`, one or more `content_block_delta`,
   `content_block_stop`, `message_delta` (with stop reason and output usage), `message_stop`.
3. **Given** `max_tokens` smaller than the planned output, **When** served, **Then** the stop reason
   is `max_tokens`.

---

### User Story 3 - Deterministic, configurable latency (Priority: P1)

A benchmarker configures each simulator with a time-to-first-token distribution, an inter-token delay
distribution and an output-length distribution (fixed, normal or lognormal), all seeded, so that runs
are reproducible.

**Why this priority**: every overhead number in the report is measured against this model.

**Independent Test**: run the same request sequence twice with the same seed and compare latencies and
outputs; they are identical. Change the seed and they differ.

**Acceptance Scenarios**:

1. **Given** a fixed TTFT of 200 ms and inter-token delay of 20 ms, **When** a 10-token stream is
   served, **Then** the first content arrives after 200–220 ms and the stream lasts 380–420 ms.
2. **Given** the same seed and the same request sequence, **When** replayed, **Then** response text,
   token counts and sampled delays are identical.
3. **Given** identical request bodies, **When** sent at any time, **Then** the generated text is
   identical (content is a function of the seed and the request, not of arrival order).

---

### User Story 4 - Fault injection, at start or live (Priority: P1)

A benchmarker injects faults through the start-up configuration or changes them live through an admin
endpoint: error rates by status (500, 503, 429 with a retry-after value), connection resets, a timeout
before the first token, a stall mid-stream, and a full outage window.

**Why this priority**: fallback (009), rate limiting (007) and the outage workload depend on it.

**Independent Test**: set each fault through the admin endpoint, send requests, observe the fault, then
clear it and observe normal behaviour.

**Acceptance Scenarios**:

1. **Given** a 100% 429 rate with retry-after 2 s, **When** a request arrives, **Then** it receives 429
   with `retry-after: 2` and a provider-shaped rate-limit error body.
2. **Given** an outage window of 30 s starting now, **When** requests arrive during the window,
   **Then** all fail with the configured status; after the window, requests succeed again.
3. **Given** a 100% connection-reset rate, **When** a request arrives, **Then** the connection is
   closed without an HTTP response.
4. **Given** a timeout-before-first-token fault of 5 s, **When** a streaming request arrives, **Then**
   no byte is sent for 5 s (or until the client gives up).
5. **Given** a mid-stream stall after 3 tokens for 10 s, **When** a stream is served, **Then** 3 content
   chunks arrive, then nothing for 10 s, then the rest.
6. **Given** faults set live, **When** the admin endpoint is read, **Then** it returns the active fault
   configuration; resetting it restores the start-up configuration.

---

### User Story 5 - Documented token counting (Priority: P1)

Token counts approximate real tokenizers with a documented method, and the usage the simulator
reports is consistent with the method within a known, measured error.

**Why this priority**: rate limits, budgets and chargeback reconcile estimates against reported usage.

**Independent Test**: count tokens for a fixed corpus with the simulator's method and with a reference
tokenizer; report the error. The simulator's reported prompt usage equals its own method's count.

**Acceptance Scenarios**:

1. **Given** a request, **When** usage is reported, **Then** prompt tokens equal the documented
   counting method applied to the request, and completion tokens equal the number of generated tokens.
2. **Given** the reference corpus, **When** compared with a reference tokenizer, **Then** the mean
   absolute error is reported in the documentation and is at most 15%.

---

### User Story 6 - Request log for assertions (Priority: P2)

A test asserts what the gateway actually sent upstream (to prove PII was redacted or headers were
stripped) by reading the simulator's request log, and clears it between tests.

**Independent Test**: send two requests, read the log, see both with method, path, headers and body;
clear it; the log is empty.

**Acceptance Scenarios**:

1. **Given** requests were served, **When** the request log is read, **Then** it lists the most recent
   requests (bounded) with headers and bodies, newest last.
2. **Given** the log is cleared, **When** read, **Then** it is empty.

---

### User Story 7 - Deterministic embeddings with an optional real model (Priority: P2)

Embeddings are deterministic by default; an optional mode forwards embeddings to a local embedding
server for semantic tests (feature 013).

**Independent Test**: with the forward mode configured, embeddings come from the local server; without
it, from the deterministic generator.

**Acceptance Scenarios**:

1. **Given** an embeddings forward URL, **When** embeddings are requested, **Then** the response comes
   from that server unchanged.

---

### Edge Cases

- The client disconnects mid-stream: the simulator stops generating and logs the cancellation; no
  goroutine or timer leaks.
- Very long outputs (e.g. 32k tokens): memory stays bounded because output is streamed, never fully
  buffered for streaming requests.
- Concurrent load of 10,000 requests per second on a laptop: the simulator must not be the bottleneck;
  the measured throughput is documented.
- Malformed JSON or missing required fields: a provider-shaped 400 error.
- Unknown model: a provider-shaped 404 error (OpenAI `model_not_found`, Anthropic `not_found_error`).
- Fault rates summing above 100%: the configuration is rejected.
- An outage window that has already ended: ignored.
- `n > 1` choices (OpenAI): rejected with 400 (documented as unsupported).
- Anthropic request without `max_tokens`: 400 `invalid_request_error`, as the real API does.
- Embeddings forward URL unreachable or failing: 502 with an OpenAI-shaped `server_error`; the
  simulator never falls back to synthetic vectors silently.
- `stop` sequences: accepted and ignored (documented); the output never contains them by construction.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The simulator MUST run as `tollgate sim --flavor openai|anthropic --port <p>` with an
  optional configuration file.
- **FR-002**: The OpenAI flavour MUST serve `POST /v1/chat/completions` (streaming and non-streaming),
  `POST /v1/embeddings` and `GET /v1/models` with the official response and server-sent-event chunk
  shapes, including `usage` and, when `stream_options.include_usage` is true, a final usage chunk.
- **FR-003**: The Anthropic flavour MUST serve `POST /v1/messages` with the official response and the
  official streaming event sequence, including usage in `message_start` and `message_delta`.
- **FR-004**: Errors MUST use each provider's error body shape and status codes (400, 401 when an
  auth header is missing, 404, 429, 500, 503, plus Anthropic 529 when configured).
- **FR-005**: Each simulator MUST have a latency model: time to first token and inter-token delay as
  fixed, normal or lognormal distributions, and an output-length distribution, all driven by one seed.
- **FR-006**: Generated text MUST be a deterministic function of the seed and the request body.
- **FR-007**: Faults MUST be configurable at start and live through `/_sim/faults`: error rate by status
  (500, 503, 429 with retry-after, 529 for Anthropic), connection reset, timeout before first token,
  stall mid-stream, and a full outage window.
- **FR-008**: Token counting MUST follow one documented approximation; reported prompt usage MUST equal
  that method's count and completion usage MUST equal the number of generated tokens.
- **FR-009**: Embeddings MUST be deterministic unit vectors of a configurable dimension (default 1536),
  in float or base64 encoding; an optional forward mode MUST proxy embeddings to a configured local URL.
- **FR-010**: `/_sim/requests` MUST return the last N requests (default 1000) with method, path, headers
  and body, and MUST support clearing. Request bodies and auth header values MUST NOT be written to the
  simulator's own logs; they are only available through this endpoint.
- **FR-011**: Configuration MUST be YAML validated by a JSON Schema (`schemas/sim.schema.json`).
- **FR-012**: Wire-format fixtures MUST be stored in `testdata/wire/{openai,anthropic}/` with a README
  citing the official source of each.
- **FR-013**: A compatibility suite MUST run the official OpenAI and Anthropic SDKs for Python,
  TypeScript and Go against both simulators, changing only the base URL, as a CI job and via
  `make sdk-compat`.
- **FR-014**: Health endpoint `/_sim/healthz` MUST return 200 for readiness probes.
- **FR-015**: The simulator MUST never contact a real provider; it has no outbound calls except the
  optional local embeddings forward URL.
- **FR-016**: Requests without the provider's auth header MUST receive the provider's 401 error; any
  header value is accepted.
- **FR-017**: When a request forces a tool call, the response MUST contain one tool call to the first
  eligible tool with arguments `{}` and the provider's tool finish/stop reason.
- **FR-018**: The simulator MUST bind to `127.0.0.1` unless `--host` is given.
- **FR-019**: Per-request randomness MUST be seeded from the global seed, the request-body hash and the
  occurrence count of that body.
- **FR-020**: On SIGINT or SIGTERM the simulator MUST stop accepting connections and let in-flight
  requests finish for up to 5 seconds before exiting.

### Key Entities

- **Simulator config**: flavour, port, seed, models, latency model, output-length model, embedding
  dimension, embeddings forward URL, fault config, request-log size.
- **Distribution**: kind (fixed, normal, lognormal), parameters (value; mean and stddev; mu and sigma),
  optional min and max clamps.
- **Fault config**: per-status error rates, retry-after seconds, reset rate, pre-first-token timeout
  (rate, duration), mid-stream stall (rate, after N tokens, duration), outage window (start, duration,
  status).
- **Request record**: time, method, path, headers, body, response status.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The official OpenAI and Anthropic SDKs in Python, TypeScript and Go (6 clients) pass the
  compatibility suite against the simulators with only the base URL changed.
- **SC-002**: Every fixture in `testdata/wire/` has a cited official source, and the simulator's output
  matches each fixture's structure (same fields and event order).
- **SC-003**: With zero configured latency, one simulator instance sustains at least 10,000
  non-streaming requests per second on a laptop (measured and documented with hardware).
- **SC-004**: Replaying a request sequence with the same seed yields byte-identical response bodies
  (excluding ids and timestamps) and identical sampled delays.
- **SC-005**: Each fault type is observable within one request after being set live and disappears
  within one request after being cleared.
- **SC-006**: Token-count mean absolute error against the `o200k_base` reference tokenizer on the
  committed corpus is documented and ≤ 15%.
- **SC-007**: Under 1,000 concurrent streaming requests with a fixed 100 ms time to first token, the
  observed median time to first token stays within +5 ms of the configured value, and the p90 and p99
  overshoot are measured and documented (restated 2026-11-05, see Clarifications and issue #12).

## Assumptions

- Model quality and real tokenizer parity are out of scope; text is synthetic filler.
- Only one choice (`n = 1`) is supported on OpenAI chat completions.
- The admin endpoints have no authentication; the simulator is a test double, never a production
  component.
- Python, Node.js and Go toolchains are available in CI for the compatibility job.
