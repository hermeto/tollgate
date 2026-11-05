# Feature Specification: `tollgate serve` skeleton, pipeline interface and the local profile

**Feature Branch**: `003-local-stack`

**Created**: 2026-11-05

**Status**: Draft

**Input**: User description: "Gateway skeleton and one-command local environment. Every later policy
plugs into a pipeline; the pipeline and the local stack must exist first, with observability from day
one, so each feature lands with measurements."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Faithful streaming passthrough (Priority: P1)

A developer starts the gateway with a configuration file and sends OpenAI chat completions to it; the
gateway forwards them to one configured upstream (the OpenAI simulator) and relays responses, including
server-sent-event streams, without buffering. Stopping the gateway lets in-flight streams finish.

**Why this priority**: every policy feature (005–013) runs inside this request path.

**Independent Test**: start the simulator and the gateway, send a streaming and a non-streaming request
through the gateway, and compare them with the same requests sent directly to the simulator.

**Acceptance Scenarios**:

1. **Given** the gateway forwarding to the OpenAI simulator, **When** a client sends a non-streaming
   chat completion, **Then** it receives the simulator's response body unchanged and the upstream status.
2. **Given** a streaming request, **When** the upstream emits chunks, **Then** each chunk reaches the
   client as soon as the upstream sends it (no buffering until the end).
3. **Given** a stream in flight, **When** the gateway receives SIGTERM, **Then** it stops accepting new
   connections and the stream completes before the process exits (within the configured drain timeout).
4. **Given** a client that disconnects mid-stream, **When** the gateway notices, **Then** it cancels the
   upstream request.
5. **Given** a request body above the configured limit, **When** it arrives, **Then** the gateway answers
   413 in the OpenAI error format without calling the upstream.

---

### User Story 2 - Policy pipeline contract (Priority: P1)

A developer adding a policy implements one Stage interface. Stages run in a declared order on every
request (request phase), on the response (response phase) and on each streamed chunk, and each Stage
can continue, mutate, deny with a protocol-correct error, or short-circuit with its own response. This
feature ships two Stages: request ID and a no-op timing Stage.

**Why this priority**: the interface is the extension point for every later feature.

**Independent Test**: unit tests drive a pipeline of test Stages through each decision type and assert
the order of calls and the resulting client response.

**Acceptance Scenarios**:

1. **Given** Stages A, B, C, **When** a request flows through, **Then** request hooks run A→B→C and
   response and chunk hooks run C→B→A.
2. **Given** Stage B denies, **When** the request arrives, **Then** C and the upstream are not called, the
   client gets the error in the caller's protocol format, and A's response hook still runs.
3. **Given** Stage B short-circuits with a response, **When** the request arrives, **Then** the upstream
   is not called and the client gets that response.
4. **Given** the request-ID Stage, **When** a request arrives without `x-request-id`, **Then** one is
   generated and returned; a valid incoming `x-request-id` is propagated upstream and returned.

---

### User Story 3 - One-command local environment (Priority: P1)

A developer runs one command to start the gateway, both simulators, Redis-compatible store,
PostgreSQL, OpenTelemetry Collector, Prometheus, Tempo, Loki and Grafana; the command prints URLs and
example keys. Another command removes everything, including volumes. A smoke command sends one streaming
and one non-streaming request and checks that a trace and a metric exist.

**Why this priority**: every later feature is demonstrated and tested on this environment.

**Independent Test**: `make up && make smoke && make down` on a clean machine; after `make down` no
container, network or volume of the project remains.

**Acceptance Scenarios**:

1. **Given** Docker is running, **When** the developer runs `make up`, **Then** all services become
   healthy and the command prints the gateway, Grafana, Prometheus and simulator URLs and example keys.
2. **Given** the stack is up, **When** `make smoke` runs, **Then** both requests succeed, a trace for the
   request is found in the trace store and a gateway request metric is found in the metrics store.
3. **Given** the stack is up, **When** `make down` runs, **Then** no project container, network or volume
   remains.

---

### User Story 4 - The local stack is tested on every PR (Priority: P1)

The maintainer sees a required CI job that brings the local stack up, runs the smoke checks and the
end-to-end tests, and tears it down on every pull request.

**Independent Test**: the `e2e-local` job is aggregated by `ci-gate`; breaking the passthrough makes it
fail.

**Acceptance Scenarios**:

1. **Given** a pull request, **When** CI runs, **Then** `e2e-local` runs up → smoke → e2e tests → down
   and `ci-gate` requires it.

---

### User Story 5 - Health and metrics endpoints (Priority: P2)

An operator probes `/healthz` (process alive), `/readyz` (ready to serve) and scrapes `/metrics`.
Readiness reflects the health of each configured dependency according to its declared fail mode.

**Acceptance Scenarios**:

1. **Given** a fail-closed dependency is unreachable, **When** `/readyz` is probed, **Then** it returns
   503 and names the dependency.
2. **Given** a fail-open dependency is unreachable, **When** `/readyz` is probed, **Then** it returns 200,
   reports the dependency as degraded, and the event is logged.
3. **Given** traffic has flowed, **When** `/metrics` is scraped, **Then** request counts, request
   duration and gateway overhead are exposed.

---

### Edge Cases

- The upstream closes the connection mid-stream: the client stream ends with a protocol-correct error
  event and the request is recorded as an upstream error.
- The client cancels: the upstream request is cancelled within one read.
- Request bodies over the configured limit: 413, no upstream call.
- The upstream is down at start-up: the gateway starts and answers 502 in the caller's protocol format.
- The telemetry backend is down: requests are never blocked by telemetry export (constitution IV).
- Hop-by-hop headers and the client's credentials are not forwarded upstream.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `tollgate serve --config <file>` MUST start an HTTP server that forwards
  `POST /v1/chat/completions` to the configured upstream (FR-API-1 partial).
- **FR-002**: Streaming responses MUST be relayed chunk by chunk with a flush after each upstream read.
- **FR-003**: On SIGINT/SIGTERM the gateway MUST stop accepting connections, mark itself not ready, and
  drain in-flight requests up to a configurable timeout (default 30 s).
- **FR-004**: A client disconnect MUST cancel the upstream request.
- **FR-005**: Request bodies above a configurable limit (default 8 MiB) MUST be rejected with 413 in the
  OpenAI error format before any upstream call.
- **FR-006**: A pipeline MUST run Stages in configured order with request, response and per-chunk hooks;
  a Stage decision MUST be one of continue, mutate, deny (protocol-correct error) or respond
  (short-circuit).
- **FR-007**: Response and chunk hooks MUST run in reverse order of the request hooks, and only for
  Stages whose request hook ran.
- **FR-008**: The request-ID Stage MUST propagate a valid incoming `x-request-id` (or generate one), send
  it upstream and return it to the client.
- **FR-009**: The timing Stage MUST record per-Stage and total gateway overhead (time not spent waiting
  for the upstream) as a metric.
- **FR-010**: Every request MUST produce a server span with one child span per Stage and a client span for
  the upstream call, exported over OTLP; export failures MUST NOT block requests (FR-OBS-1 skeleton).
- **FR-011**: The gateway MUST expose `/healthz`, `/readyz` and `/metrics` on a separate port from the API.
- **FR-012**: Dependencies (Redis-compatible store, PostgreSQL) MUST be declared in configuration with a
  fail mode (`open` or `closed`); readiness MUST fail for an unreachable fail-closed dependency and
  report fail-open ones as degraded; every transition MUST be logged (NFR-2).
- **FR-013**: `make up`, `make down` and `make smoke` MUST manage a Docker Compose stack with the gateway,
  both simulators, a Redis-compatible store, PostgreSQL, OpenTelemetry Collector, Prometheus, Tempo,
  Loki and Grafana, all images pinned by digest and available for arm64 and amd64.
- **FR-014**: `make up` MUST print the service URLs and example keys; `make down` MUST remove all project
  containers, networks and volumes.
- **FR-015**: A CI job `e2e-local` MUST run up → smoke → end-to-end tests → down on every PR and be
  aggregated by `ci-gate` (NFR-8).
- **FR-016**: Gateway configuration MUST be YAML validated by `schemas/gateway.schema.json`; the default
  upstream in every shipped config MUST be the simulator.
- **FR-017**: The gateway MUST NOT forward hop-by-hop headers or the client's `Authorization` header; the
  upstream credential comes from gateway configuration (environment variable reference), never from the
  client.

### Key Entities

- **Stage**: a named policy step with request, response and chunk hooks.
- **Decision**: continue, mutate, deny (status, error type, message) or respond (status, headers, body).
- **Request context**: request ID, protocol, raw body, headers, timings, per-Stage decisions.
- **Dependency**: name, kind, address, fail mode, last health state.
- **Gateway config**: listen addresses, upstream (base URL, credential reference, timeouts), pipeline
  order, body limit, drain timeout, dependencies, telemetry endpoint.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `e2e-local` is green on every PR and `make up` reaches healthy in under 5 minutes on a
  laptop with images cached.
- **SC-002**: The empty pipeline (request-ID and timing Stages) has a committed micro-benchmark with
  baseline numbers.
- **SC-003**: Passthrough gateway overhead (time spent in the gateway, excluding upstream wait) is
  ≤ 1 ms at p99 against the simulator on the CI runner.
- **SC-004**: After `make down`, zero project containers, networks or volumes remain.
- **SC-005**: With the telemetry collector stopped, request success rate and latency are unchanged.

## Assumptions

- Only the OpenAI chat completions endpoint is proxied in this feature; Anthropic, embeddings, models and
  translation arrive in 004; authentication and policies in 005.
- Redis-compatible store and PostgreSQL are present in the stack and checked for readiness but not yet
  used by the request path (006, 007).
- Docker (Desktop or Colima) with Compose v2 is available locally and on CI runners.
- Gateway overhead is measured inside the gateway (issue #12): the simulator's own timing tail under
  load makes client-side differences unreliable at millisecond resolution.
