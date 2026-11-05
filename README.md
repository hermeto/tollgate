# Tollgate — LLM gateway with platform policies

[![License: Apache-2.0](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](LICENSE)

Tollgate is a Go gateway between engineering teams and model providers (OpenAI, Anthropic, self-hosted)
that enforces per-team rate limits and budgets, attributes every token to a cost owner (chargeback,
FOCUS export), fails over between providers, caches exact and semantically similar requests, redacts
PII, and writes a tamper-evident audit log. A reproducible report measures what each policy costs in
latency and saves in dollars.

> 🚧 **Under construction.** See the "v1.0 roadmap" issue for progress and merge order, and
> [`docs/product/spec.md`](docs/product/spec.md) for the product specification.

## License

Apache-2.0. See [LICENSE](LICENSE) and [NOTICE](NOTICE).
