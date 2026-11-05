# Contract: stack commands

| Command | Behaviour |
| --- | --- |
| `make up [PROFILE=local]` | builds the gateway image, starts the Compose project `tollgate`, waits until every service is ready (≤ 5 min), prints URLs and example keys |
| `make smoke [PROFILE=local]` | one non-streaming and one streaming request through the gateway; asserts 200, `[DONE]`, a trace in Tempo with the request's `tollgate.request_id`, and `tollgate_requests_total` > 0 in Prometheus |
| `make down [PROFILE=local]` | `docker compose down --volumes --remove-orphans`; afterwards no `tollgate` container, network or volume exists |
| `make e2e` | `TOLLGATE_E2E=1 go test ./test/e2e/... -count=1` against a running stack |

Other profiles (`envoy`, `k8s`) arrive in 014 and 015; unknown profiles exit 2.
