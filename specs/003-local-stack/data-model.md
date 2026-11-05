# Data model: 003 local-stack

## Gateway config (`schemas/gateway.schema.json`)

| Field | Type | Default | Rule |
| --- | --- | --- | --- |
| `listen.api` | host:port | `:8080` | |
| `listen.ops` | host:port | `:8081` | must differ from `listen.api` |
| `upstream.base_url` | URL | — | required; http or https |
| `upstream.credential_env` | string | — | name of the environment variable holding the provider key; empty means no credential |
| `upstream.timeouts.connect` | duration | `2s` | |
| `upstream.timeouts.response_header` | duration | `60s` | time to first upstream byte |
| `upstream.timeouts.idle` | duration | `90s` | idle keep-alive connections |
| `limits.max_body_bytes` | int | 8388608 | 1 KiB – 64 MiB |
| `shutdown.drain_timeout` | duration | `30s` | |
| `pipeline` | string[] | `[request_id, timing]` | known Stage names, no duplicates |
| `dependencies[]` | `{name, kind, address, fail_mode}` | `[]` | kind `redis` \| `postgres`; fail_mode `open` \| `closed` |
| `telemetry.otlp_endpoint` | URL | empty | empty disables OTLP export (metrics stay on `/metrics`) |
| `telemetry.service_name` | string | `tollgate` | |
| `telemetry.traces_sample_ratio` | number | 1.0 | 0–1 |

## Pipeline types

- **Request**: `ID`, `Protocol` (`openai`), `Method`, `Path`, `Header` (client headers, mutable), `Body`
  (raw bytes, mutable), `Stream` (bool), `Received` (time), `UpstreamHeader` and `ResponseHeader`
  (headers Stages add), `Decisions` (per Stage: name, kind, duration).
- **Response**: `Status`, `Header`, `Body` (non-streaming only), `Stream` (bool).
- **Chunk**: `Data` (one SSE event, raw bytes without the trailing blank line).
- **Decision**: `Kind` ∈ {Continue, Mutate, Deny, Respond}; Deny carries `Status`, `Type`, `Message`;
  Respond carries a `*Response`.

## Readiness

`Dependency{Name, Kind, Address, FailMode, State}` with `State` ∈ {unknown, up, down}. Readiness =
not draining ∧ every fail-closed dependency up. Report body:
`{"status":"ready|degraded|not_ready","dependencies":{"<name>":"up|down"}}`.
