# Quickstart: 003 local-stack

## Host run (no Docker)

```bash
go run ./cmd/tollgate sim --flavor openai --port 9001 &
TOLLGATE_UPSTREAM_OPENAI_KEY=tg-example-upstream go run ./cmd/tollgate serve --config config/local.yaml &
curl -s localhost:8080/v1/chat/completions -d '{"model":"gpt-4o-mini","messages":[{"role":"user","content":"hi"}]}' | jq .usage
curl -sN localhost:8080/v1/chat/completions -d '{"model":"gpt-4o-mini","stream":true,"messages":[{"role":"user","content":"hi"}]}'
curl -s localhost:8081/readyz; curl -s localhost:8081/metrics | grep tollgate_
```

## Local stack

```bash
make up        # prints URLs: gateway :8080, ops :8081, Grafana :3000, Prometheus :9090, simulators :9001/:9002
make smoke     # streaming + non-streaming request, trace in Tempo, metric in Prometheus
make e2e       # end-to-end tests against the stack
make down      # removes containers, networks and volumes
```

## Benchmarks

```bash
go test ./internal/pipeline/... -bench . -run '^$'
go test ./internal/server/... -bench Passthrough -run '^$'
```
