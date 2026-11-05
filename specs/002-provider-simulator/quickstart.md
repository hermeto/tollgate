# Quickstart: 002 provider-simulator

```bash
go run ./cmd/tollgate sim --flavor openai --port 9001 &
go run ./cmd/tollgate sim --flavor anthropic --port 9002 &

curl -s localhost:9001/v1/chat/completions -H 'authorization: Bearer tg-example-key' \
  -d '{"model":"gpt-4o-mini","messages":[{"role":"user","content":"hi"}]}' | jq .usage
curl -sN localhost:9001/v1/chat/completions -H 'authorization: Bearer tg-example-key' \
  -d '{"model":"gpt-4o-mini","stream":true,"stream_options":{"include_usage":true},"messages":[{"role":"user","content":"hi"}]}'
curl -sN localhost:9002/v1/messages -H 'x-api-key: tg-example-key' -H 'anthropic-version: 2023-06-01' \
  -d '{"model":"claude-sonnet-4-5","max_tokens":64,"stream":true,"messages":[{"role":"user","content":"hi"}]}'

# Faults, live
curl -s -X PUT localhost:9001/_sim/faults -d '{"error_rates":{"429":1},"retry_after_seconds":2}'
curl -si localhost:9001/v1/chat/completions -H 'authorization: Bearer x' -d '{"model":"gpt-4o-mini","messages":[]}' | head -1   # 429
curl -s -X DELETE localhost:9001/_sim/faults

# Request log
curl -s localhost:9001/_sim/requests | jq '.requests | length'
```

Validation suites:

```bash
go test ./sim/... ./internal/tokens/... -race -cover
make sdk-compat                  # official SDKs (Python, TypeScript, Go) against both flavours
go test ./sim/... -bench . -run '^$'
```

Expected: tests green; `make sdk-compat` prints `PASS` for 6 clients; the benchmark reports
≥ 10,000 req/s equivalent for the non-streaming path.
