# Research: 003 local-stack

Versions resolved on 2026-11-05 (`go list -m -versions`, `gh release view`, `docker buildx imagetools
inspect`). All images are multi-arch (linux/amd64 and linux/arm64) and pinned by index digest.

## R1. Container images

| Service | Image | Digest |
| --- | --- | --- |
| Valkey | `valkey/valkey:9.1.3-alpine` | `sha256:bf09d58e96d6b4524a15396b65b242bbe18beebbcc15a04d8ee75ddac62fae0b` |
| PostgreSQL | `postgres:18.6-alpine` | `sha256:10a5c69a83035ab5239f5620d708520e8b28f5aa1f2b20143e6024532edadfb5` |
| OpenTelemetry Collector (contrib) | `otel/opentelemetry-collector-contrib:0.164.0` | `sha256:340c9ca6a36617084c6661bdc4f0fe92b7cda5a805d61673ec8db375fddc522b` |
| Prometheus | `prom/prometheus:v3.16.0` | `sha256:e04987641b57704f408e31f6373ba4138ecf27cb4a165d24aabb6c38fddd0f06` |
| Tempo | `grafana/tempo:3.1.1` | `sha256:64e2fbd56d9c2b3412a4b430a58ea59e91a670b1e75f920a0c702213ad145044` |
| Loki | `grafana/loki:3.7.9` | `sha256:0db216b630ae1f348a1584c67f3cead8b4c1fc839169a39501db664744158e28` |
| Grafana | `grafana/grafana:13.3.0` | `sha256:488984b716c0862639f91a698adbe2862677a5a07bad42da5f03f3030883b205` |
| Gateway build stage | `golang:1.27.3-alpine` | `sha256:d200d2dd76b03a99b52651fd1325aff7dcd35422c35c01bac56571b0507f3deb` |
| Gateway runtime | `gcr.io/distroless/static-debian13:nonroot` | `sha256:25ac8f0c3b2334a9eb05452266bd4697864a8c8ef0ef71e5e6480d7f92868329` |

- **Decision**: Valkey instead of Redis (spec Clarifications: licence). PostgreSQL 18.6 is the latest
  release with an Alpine image (no 19.x image exists yet).

## R2. Pipeline design

- **Decision**: three optional hook interfaces (`RequestHook`, `ResponseHook`, `ChunkHook`) on a `Stage`
  with a `Name()`; the pipeline resolves them once at construction, so the hot path iterates plain slices.
  Decisions are a small value type (`Continue`, `Mutate`, `Deny`, `Respond`). Response and chunk hooks run
  in reverse order over the Stages whose request hook ran (spec FR-007), mirroring middleware semantics.
- **Rationale**: no allocation per decision, no reflection, and later policies implement only the hooks
  they need. Per-Stage spans and durations are recorded by the runner, not by each Stage.
- **Alternatives considered**: `net/http` middleware chains (no chunk hook, hard to short-circuit with a
  protocol-correct error); a single `Handle(ctx, *Exchange)` method (forces every Stage to handle every
  phase).

## R3. Streaming relay

- **Decision**: `internal/sse.Reader` splits upstream bytes into events at blank lines (`\n\n`, `\r\n\r\n`)
  with a bounded event size (1 MiB); each event passes the chunk hooks and is written and flushed
  immediately with `http.ResponseController`. Non-SSE bodies are copied with a 32 KiB buffer.
- **Rationale**: chunk hooks (PII restore in 011, cache capture in 013) need whole events; flushing per
  event preserves the upstream's pacing (FR-002).
- **Sources**: WHATWG HTML Living Standard, "Server-sent events" (event stream interpretation);
  <https://html.spec.whatwg.org/multipage/server-sent-events.html>.

## R4. Upstream client

- **Decision**: one `http.Transport` per upstream with `MaxIdleConnsPerHost` 256, `IdleConnTimeout` 90 s,
  connect timeout 2 s, response-header timeout 60 s, `ForceAttemptHTTP2` (HTTP/2 over TLS for real
  providers; plain HTTP/1.1 keep-alive to the simulators). The request context carries cancellation, so a
  client disconnect cancels the upstream call (FR-004).
- **Hop-by-hop headers** (RFC 9110 §7.6.1): `Connection`, `Keep-Alive`, `Proxy-Connection`,
  `Transfer-Encoding`, `Upgrade`, `TE`, `Trailer`, plus any listed in `Connection`, are dropped; so are the
  client's `Authorization` and `X-Api-Key` (FR-017).

## R5. Telemetry

- **Decision**: OpenTelemetry Go v1.47.0. Traces: batch span processor → OTLP/HTTP (`otlptracehttp`) to
  the Collector → Tempo. Logs: `slog` fan-out to JSON on stdout and to the `otelslog` bridge → OTLP/HTTP
  (`otlploghttp` v0.23.0) → Collector → Loki's native OTLP endpoint. Metrics: OTel meter provider with the
  Prometheus exporter on `/metrics` (port 8081), scraped by Prometheus. W3C trace context is propagated
  upstream. Instrument and attribute names live in `internal/telemetry` (ADR-0012 comes with 010).
- **Fail-open**: batch processors use bounded queues and drop on overflow; exporter timeouts are 5 s and
  run off the request path, so a stopped Collector never blocks requests (spec SC-005).
- **Semantic conventions**: `go.opentelemetry.io/otel/semconv/v1.44.0` (latest in otel-go v1.47.0) for HTTP
  attributes; GenAI attributes arrive in 010.
- **Sources**: <https://opentelemetry.io/docs/languages/go/>, <https://grafana.com/docs/loki/latest/send-data/otel/>,
  <https://grafana.com/docs/tempo/latest/>.

## R6. Readiness and dependency checks

- **Decision**: each declared dependency is checked every 5 s with a TCP connect (1 s timeout). A
  fail-closed dependency that is unreachable makes `/readyz` return 503; fail-open ones are reported as
  `degraded` with 200. Transitions are logged at WARN/INFO. Protocol-level checks replace TCP checks when
  the Valkey and PostgreSQL clients arrive (006, 007).
- **Rationale**: the request path does not use the stores yet; reachability is the honest signal now.

## R7. End-to-end tests: Compose stack, not testcontainers

- **Decision**: `test/e2e` runs against the stack started by `make up` (`TOLLGATE_E2E=1`, base URLs from
  environment with local defaults). Tests skip otherwise.
- **Rationale**: one definition of the environment for development, smoke checks and CI; the stack starts
  once per job instead of per test; testcontainers-go would duplicate eleven service definitions in Go.
- **Alternatives considered**: testcontainers-go (better isolation per test, slower, duplicated config).

## R8. Migrations tool (ADR-0002)

- **Decision**: goose v3.29.0 for PostgreSQL migrations (used from 006): embeddable with `embed.FS`, plain
  SQL migrations, small dependency surface, CLI and library modes.
- **Alternatives considered**: golang-migrate (also embeddable; more drivers than needed, separate CLI
  conventions).

## R9. Gateway overhead measurement

- **Decision**: the server records `tollgate_gateway_overhead_seconds{phase}` with phases `request`
  (arrival → upstream request written), `response` (upstream headers → client headers written) and
  `chunk` (event read → event flushed), plus `tollgate_stage_duration_seconds{stage}`. SC-003 is checked on
  the CI runner by querying p99 of the request+response overhead from Prometheus after a load step.
- **Rationale**: issue #12 showed client-side differences cannot resolve millisecond overheads.
