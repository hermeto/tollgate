## Summary

Slice 3 of 7 of the provider simulator: the documented token-counting approximation (`internal/tokens`,
shared with the gateway's estimator in 006/007) and the deterministic text generator with fixed, normal
and lognormal distributions.
Spec: [`specs/002-provider-simulator/spec.md`](https://github.com/hermeto/tollgate/blob/002-provider-simulator-iter2/specs/002-provider-simulator/spec.md) (US5). Method and measurements: `docs/simulator.md` "Token counting". Tracking: #3.

## Stack

Based on #6; merge after it. Next slice: `002-provider-simulator-iter3` (OpenAI flavour).

## Requirements covered

Feature spec FR-008, SC-006; groundwork for FR-005 (distributions) and FR-006 (deterministic content).

## Decisions for the reviewer

- Reference tokenizer: tiktoken 0.14.0 `o200k_base`, pinned in `scripts/tokens/reference.py.lock` (uv script lock). Reference counts are committed, so CI checks the error offline.
- Approximation: a hand-written scanner that mirrors `o200k_base` pre-tokenization, with 7 constants fitted by grid search. In-sample mean absolute error 7.4%, bias +2.2%; fitting on half the corpus and testing on the other half gives 5.8%. The test enforces ≤ 15% error and ±5% bias.
- The corpus (48 samples, 9 languages, code, JSON, YAML, logs) is synthetic and contains no personal data.
- The Anthropic tokenizer is not public; the Anthropic flavour uses the same method and the docs say its error against Claude models is unknown.
- Generated text uses a 130-word vocabulary where every piece counts as one token, so completion usage equals `tokens.Count(text)` (tested).

## How it was verified

- `make lint test` clean; `TestErrorAgainstReference`: 48 samples, 7.4% error, +2.2% bias; worst samples logged (long compound words, tab-indented Go, emoji, URLs, Korean).
- `BenchmarkCount4KB`: 4.2 µs/op, 968 MB/s, 0 allocs (Apple M-series).
- Generator: determinism, one token per piece, capitalisation; distributions: fixed, clamped normal, lognormal mean within 5% of exp(μ + σ²/2), negative floor; output length clamp.

## Not verified

- Error against Anthropic's tokenizer (not public).

## Performance impact

`tokens.Count` will run on the gateway hot path in 006/007: 4.2 µs per 4 KB prompt, zero allocations.

## Checklist

- [x] Tests added or updated
- [x] Docs updated in this PR (README, docs/, runbooks)
- [x] ADR added or updated for architectural decisions — ADR-0009 ships with the last slice
- [x] No secrets, real keys or real personal data
- [x] New dependencies, images and actions pinned
- [x] No prompt/response content or PII in logs, traces, metrics or events
- [x] Commits signed off (DCO) and Conventional
