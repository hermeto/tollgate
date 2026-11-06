## Summary

Feature 005 adds optional virtual-key identity, atomic local key persistence, declarative
org/team/project policies, policy reload, model allowlists and operator CLI commands.

## Base

Based on `004-protocol-api`; merge after the protocol API PR.

## Requirements covered

FR-ID-1 through FR-ID-6; tasks T001–T012.

## Decisions for the reviewer

- Keys are shown only at creation/rotation time and persisted as SHA-256 digests in a 0600 JSON
  store; authentication uses constant-time comparison and checks expiry/revocation.
- Policy inheritance is deterministic (organization, team, project), with project values replacing
  inherited model lists. Invalid reloads retain the last valid snapshot.
- Identity and policy enforcement are opt-in so existing gateway profiles remain backward compatible.
- `dry-run` records a structured warning and permits the request; `enforce` returns protocol-specific
  401/403 errors.

## Verification

- `go test ./...`
- `go test -race ./internal/identity ./internal/policy ./internal/server`
- `go test -cover ./internal/identity ./internal/policy ./internal/server`
- `golangci-lint run ./...`
- `tollgate policy validate --dir policies/examples`
- `tollgate policy explain --dir policies/examples --team payments --project checkout`
