## Summary

Slice 2 of 7 of the provider simulator: the HTTP server skeleton and the `tollgate sim` command.
The server answers `/_sim/healthz`, requires the provider's auth header on `/v1/*` (any value), returns
provider-shaped 401/404 bodies, binds to `127.0.0.1` by default and drains in-flight requests for 5 s
on SIGINT/SIGTERM. Adds the deterministic per-request RNG and the SSE writer used by the flavours.
Spec: [`specs/002-provider-simulator/spec.md`](https://github.com/hermeto/tollgate/blob/002-provider-simulator-iter1/specs/002-provider-simulator/spec.md). Docs: `docs/simulator.md`. Tracking: #3.

## Stack

Based on #5; merge after it. Next slice: `002-provider-simulator-iter2` (token counting).

## Requirements covered

Feature spec FR-001, FR-010 (no bodies or auth headers in the simulator's own logs), FR-014, FR-016,
FR-018, FR-019, FR-020.

## Decisions for the reviewer

- Anthropic flavour also accepts `Authorization: Bearer` (the Anthropic SDKs support auth tokens), OpenAI only Bearer.
- Content randomness = PCG(seed, hash(path, body)); latency/fault randomness = PCG(seed ^ hash, occurrence of that request). The occurrence table is bounded at 65,536 entries and cleared when full; sequential replays clear it at the same point, so determinism holds. Research R5.
- Shutdown: `http.Server.Shutdown` with a 5 s grace, then `Close`.
- `--log-requests` wraps the response writer with `Unwrap()` so flushing and hijacking (used by faults later) keep working.

## How it was verified

- `make lint test` clean; tests for health, 401 per flavour, any key accepted, empty Bearer rejected, `requireAuth`, invalid config, loopback default, graceful shutdown on context cancel, RNG determinism under interleaving and bounded occurrence table, SSE framing and flushing.
- Manual: `tollgate sim --flavor anthropic --port 19002` → healthz 200, `POST /v1/messages` without key → 401, SIGTERM → "simulator stopped" within 0.3 s.

## Not verified

- Provider endpoints do not exist yet (slices 4 and 5).

## Performance impact

None on the gateway hot path.

## Checklist

- [x] Tests added or updated
- [x] Docs updated in this PR (README, docs/, runbooks)
- [x] ADR added or updated for architectural decisions — ADR-0009 ships with the last slice
- [x] No secrets, real keys or real personal data
- [x] New dependencies, images and actions pinned
- [x] No prompt/response content or PII in logs, traces, metrics or events
- [x] Commits signed off (DCO) and Conventional
