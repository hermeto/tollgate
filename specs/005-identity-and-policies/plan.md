# Implementation Plan: Identity and declarative policies

## Summary

Add a local-first identity store, policy loader with inheritance/provenance, CLI administration and gateway
authentication/model allowlist stages. Keep persistence and authorization behind small interfaces for shared stores.

## Technical Context

Go 1.27.3; SHA-256 hashes for high-entropy virtual keys; atomic JSON replacement for local persistence; YAML v4
and JSON Schema validation; `atomic.Pointer` snapshots for reloads. No provider calls or secrets in tests.

## Structure

- `internal/identity`: key generation, hashing, authentication and file persistence.
- `internal/policy`: schema, deterministic merge, provenance, loader and watcher.
- `cmd/tollgate`: `keys` and `policy` administration commands.
- `internal/server`: optional identity and policy enforcement.
- `policies/examples`: synthetic example policies.

## Constitution Check

Pass: secrets are shown once, raw keys never enter logs, policy decisions are measurable, reload is atomic,
local profile remains functional without identity configuration and tests use synthetic values.
