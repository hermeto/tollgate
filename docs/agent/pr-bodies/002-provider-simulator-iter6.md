## Summary

Slice 7 of 9 of the provider simulator: the seeded latency model (time to first token, inter-token
delay) applied to both flavours, and live fault injection through `/_sim/faults`: error rates by
status with `retry-after`, TCP resets, timeouts before the first token, mid-stream stalls and outage
windows.
Spec: [`specs/002-provider-simulator/spec.md`](https://github.com/hermeto/tollgate/blob/002-provider-simulator-iter6/specs/002-provider-simulator/spec.md) (US3, US4). Tracking: #3.

## Stack

Based on #10; merge after it. Next slice: `002-provider-simulator-iter7` (SDK compatibility, benchmarks, ADR-0009).

## Requirements covered

Feature spec FR-005, FR-007, FR-019 (fault and latency draws), SC-004 (latency samples), SC-005.

## Decisions for the reviewer

- Nothing is sent before the first token, including headers: TTFT = time to first byte. Fallback (009) holds the client until the first upstream byte, so this is the behaviour it needs to observe.
- Non-streaming responses wait TTFT plus one inter-token delay per remaining token; embeddings wait one TTFT sample; `/v1/models` has no latency but is subject to faults.
- Fault draw order per request: outage → one uniform draw shared by error rates (sorted by status) and `reset_rate` → timeout → stall. The rates must sum to at most 1 (validated).
- Reset = hijack + `SetLinger(0)` + close (TCP RST); the request log records status 0.
- `retry_after_seconds` applies to 429, 503 and 529; outage status defaults to 503.
- PUT replaces, PATCH is an RFC 7396 merge patch, DELETE restores the start-up config. A body that sets a relative outage window (no `start`) restarts it from now; a config read back from GET (with absolute `start`/`end`) keeps its window.
- Live config is swapped with `atomic.Pointer`: the request path never locks.

## How it was verified

- `make lint test` clean; timing tests also run 3 times with `-race`.
- US3 AS1: TTFT 200 ms + 20 ms × 9 → first content 200–220 ms, last content 380–420 ms (up to 3 attempts for noisy runners; lower bounds never relaxed). Non-streaming waits ≥ 100 ms for 50 + 10 × 5. Latency samples are identical for the same seed and request.
- Faults: 429 with `retry-after: 2` and the fixture body; Anthropic 529 and 429 bodies; outage before/inside/after the window; reset (client error, log status 0); TTFT hold ≥ 300 ms and a 100 ms client timeout fails; 300 ms stall exactly after 3 tokens; PUT/PATCH/GET/DELETE; invalid configs (rate sum, unknown field, 529 on OpenAI, bad JSON) → 400; admin endpoints unaffected while `/v1/models` is; a 50% error rate replays the same status sequence.

## Not verified

- TTFT accuracy under 1,000 concurrent streams (SC-007) and throughput (SC-003): slice 8.

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
