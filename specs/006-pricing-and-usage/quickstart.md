# Quickstart

```sh
go test ./internal/pricing/... ./internal/usage/... -race -cover
go test ./internal/usage/... -bench . -run '^$'
```

The tests load `pricing/examples`, compute a simulator fixture cost, write events to a temporary JSONL
spool, and verify shutdown flush and replay behavior.
