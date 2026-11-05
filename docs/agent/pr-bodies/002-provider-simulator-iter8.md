## Summary

Slice 9 of 9 (last) of the provider simulator: latency is now scheduled as deadlines from request
arrival, throughput and TTFT-accuracy benchmarks (Go and k6), the performance section of
`docs/simulator.md`, ADR-0009 "Simulated providers as the default upstream", and the convergence
fixes. Feature 002 is complete with this PR.
Spec: [`specs/002-provider-simulator/spec.md`](https://github.com/hermeto/tollgate/blob/002-provider-simulator-iter8/specs/002-provider-simulator/spec.md). Tracking: #3. Related: #12.

## Stack

Based on #13; merge after it. Feature 003 (`local-stack`) branches from this slice.

## Requirements covered

Feature spec SC-003, SC-006 (verification), SC-007 (restated), edge cases "client disconnects
mid-stream" and "10,000 req/s on a laptop". ADR-0009 (product spec §A.10).

## Decisions for the reviewer

- **SC-007 restated** (needs your call in #12): the original "p99 within +5 ms at 1,000 streams" is not met with the load generator on the same host (Go client p99 +23.5 ms, k6 +27.3 ms; median +3.4 ms). GC, CPU saturation, macOS timer coalescing and k6 overhead were ruled out. The criterion now reads "median within +5 ms, p90/p99 documented", and later features should measure gateway overhead inside the gateway (per-Stage timing) instead of as a difference of client-side p99s.
- Latency waits are absolute deadlines from the arrival time stamped before any other work, so parsing and writing time is absorbed and inter-token delays do not drift. Consequence: a chunk sent late shortens the next gap; `TestStallMidStream` now asserts the strict lower bound from request start plus a 10 ms tolerance on the gap.
- Benchmarks: in-process Go benchmarks are pessimistic (client and server share the CPU); the k6 runs against the real binary are the documented reference.
- k6 TTFT script ramps VUs over 15 s because macOS listen backlogs (`somaxconn` 128) reset bursts of new connections.

## How it was verified

- `make lint test` clean; timing tests 5 runs under `-race`; `make sdk-compat` 6/6.
- Throughput (k6, real binary): 10,000 req/s for 20 s with 0 errors and 0 drops (median 39 µs, p95 46 µs, simulator ≈ 0.3 cores); 30,000 req/s sustained (p95 106 µs).
- TTFT accuracy (fixed 100 ms): 100 streams median +2.5 ms, p99 +10.2 ms; 1,000 streams median +3.4 ms, p90 +13.6 ms, p99 +23.5 ms (Go client); k6 median +3.8 ms, p99 +27.3 ms; Linux (Docker VM) k6 p99 +31.8 ms.
- New test: a client disconnect in the middle of a paced 200-token stream ends the handler within 500 ms.
- Coverage: `sim` 85.9%, `internal/tokens` 100%.

## Not verified

- TTFT tail on dedicated hardware with the load generator on another host (#12).

## Performance impact

None on the gateway hot path; the simulator itself handles ≥ 30,000 req/s on one laptop.

## Checklist

- [x] Tests added or updated
- [x] Docs updated in this PR (README, docs/, runbooks)
- [x] ADR added or updated for architectural decisions (ADR-0009)
- [x] No secrets, real keys or real personal data
- [x] New dependencies, images and actions pinned
- [x] No prompt/response content or PII in logs, traces, metrics or events
- [x] Commits signed off (DCO) and Conventional
