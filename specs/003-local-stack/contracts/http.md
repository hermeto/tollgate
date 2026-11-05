# Contract: gateway HTTP surface (003)

API listener (default `:8080`):

| Method | Path | Behaviour |
| --- | --- | --- |
| POST | `/v1/chat/completions` | pipeline → upstream; JSON or SSE relayed unchanged; `x-request-id` on every response |
| any | other | 404 in OpenAI error format |

Errors (OpenAI format `{"error":{"message","type","param","code"}}`): 413 `request_too_large` for bodies
over `limits.max_body_bytes`; 502 `upstream_unavailable` when the upstream cannot be reached; 504
`upstream_timeout` when the first upstream byte exceeds `timeouts.response_header`; Stage denials use the
Stage's status and type. Mid-stream upstream failure: final data event with code `upstream_stream_error`,
no `[DONE]`.

Operations listener (default `:8081`): `GET /healthz` (200 while the process runs), `GET /readyz` (see
data-model "Readiness"; 503 while draining), `GET /metrics` (Prometheus text format).
