## Summary

Feature 006 adds effective-dated USD price tables, token cost calculation, content-free usage events,
and a bounded asynchronous JSONL spool writer with graceful shutdown flush.

## Base

Based on `005-identity-and-policies`; merge after PR #22.

## Requirements covered

FR-FIN-1, FR-FIN-2 and FR-FIN-6; tasks T001–T008.

## Decisions for the reviewer

- Prices are USD per million tokens and selected by effective timestamp.
- Missing provider usage is represented explicitly by `estimated`; no prompt or response content enters
  an event.
- The sink interface keeps the hot path independent of PostgreSQL while the bounded JSONL spool gives
  local development deterministic durability.

## Verification

- `go test ./internal/pricing/... ./internal/usage/... -race -cover`
- `go test ./...`
- `golangci-lint run ./...`
