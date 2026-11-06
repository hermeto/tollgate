# Gateway skeleton and local stack

How `tollgate serve` processes a request and how the local profile (`make up`) is assembled.
Specification: [`specs/003-local-stack/`](../specs/003-local-stack/spec.md). Decisions:
[ADR-0001](adr/0001-standalone-gateway-with-optional-extproc.md) (standalone gateway, optional ext_proc),
[ADR-0002](adr/0002-state-stores.md) (Valkey and PostgreSQL).

## The pipeline

Every policy is a **Stage** (`internal/pipeline`) that implements any of three hooks:

| Hook | Runs | Can |
| --- | --- | --- |
| `OnRequest` | before the upstream call, in pipeline order | continue, mutate the request, deny with a protocol-correct error, or respond (short-circuit, for example a cache hit) |
| `OnResponse` | on the response before anything reaches the client, in reverse order | continue, mutate headers or body, deny |
| `OnChunk` | on each streamed event, in reverse order | continue, mutate the event, deny (ends the stream with an error event) |

A Stage that denies or short-circuits does not see the response it caused; the Stages before it do.
Stages that need to add headers to every outcome (for example the request id) set
`Request.ResponseHeader` in the request phase. Every hook call is timed (`Request.Records`), which feeds
per-Stage spans and the `tollgate_stage_duration_seconds` metric.

Built-in Stages:

- `request_id` — keeps a valid client `x-request-id` (1–128 printable ASCII characters) or generates
  `req_` + 26 random characters, sends it upstream and returns it.
- `timing` — stamps the arrival time if the transport has not, so all timings share one origin.

### Baseline (Apple M4, Go 1.27.3, 2026-11-05)

| Benchmark | Time | Allocations |
| --- | --- | --- |
| `BenchmarkEmptyPipeline` (request + response phase, `request_id` and `timing`) | 540 ns/op | 12 allocs, 1,424 B |
| `BenchmarkChunkNoHooks` (per streamed event, no chunk hooks) | 4 ns/op | 0 |

Run them with `go test ./internal/pipeline/ -bench . -run '^$' -benchmem`.

## Local Compose profile

The reproducible local profile starts the gateway, both deterministic simulators, Valkey, PostgreSQL,
the OpenTelemetry Collector, Prometheus, Tempo, Loki and Grafana. All published ports bind to
`127.0.0.1`; provider calls remain local and use the fake key `tg-example-upstream`.

```bash
make up       # build the pinned gateway image and wait for readiness
make smoke    # non-streaming + streaming request, metrics and Tempo trace
make e2e      # run the stack-gated Go tests
make down     # stop and remove containers, network and volumes
```

URLs:

| Service | URL |
| --- | --- |
| Gateway API | <http://127.0.0.1:8080> |
| Health, readiness and metrics | <http://127.0.0.1:8081> |
| Prometheus | <http://127.0.0.1:9090> |
| Tempo | <http://127.0.0.1:3200> |
| Grafana | <http://127.0.0.1:3000> |

`make up` polls readiness from the host because some observability images do not include a shell or
HTTP client. If it fails, inspect `docker compose --project-name tollgate --file deploy/compose/compose.yaml
logs --no-color`, then rerun `make down` before retrying. A fail-closed Valkey or PostgreSQL check keeps
`/readyz` at 503 until the service is reachable.

The stack uses image index digests recorded in
[`specs/003-local-stack/research.md`](../specs/003-local-stack/research.md); update the digest and its
architecture evidence together when upgrading an image.
