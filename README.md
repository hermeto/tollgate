# Tollgate — LLM gateway with platform policies

[![CI](https://github.com/hermeto/tollgate/actions/workflows/ci.yml/badge.svg)](https://github.com/hermeto/tollgate/actions/workflows/ci.yml)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/hermeto/tollgate/badge)](https://scorecard.dev/viewer/?uri=github.com/hermeto/tollgate)
[![License: Apache-2.0](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/hermeto/tollgate)](https://goreportcard.com/report/github.com/hermeto/tollgate)

Tollgate is a Go gateway between engineering teams and model providers (OpenAI, Anthropic, self-hosted)
that enforces per-team rate limits and budgets, attributes every token to a cost owner (chargeback,
FOCUS export), fails over between providers, caches exact and semantically similar requests, redacts
PII, and writes a tamper-evident audit log. A reproducible report measures what each policy costs in
latency and saves in dollars.

> 🚧 **Under construction.** See the [v1.0 roadmap](https://github.com/hermeto/tollgate/issues/3) for progress and merge order, and
> [`docs/product/spec.md`](docs/product/spec.md) for the product specification.

## Documentation

- [Provider simulator](docs/simulator.md) — `tollgate sim`, the deterministic OpenAI and Anthropic test double
  (`make sdk-compat` runs the official SDKs against it).

## Contributing

```bash
make dev-setup && make lint test
```

- [CONTRIBUTING.md](CONTRIBUTING.md) — setup on macOS and Linux, Spec Kit workflow, stacked PRs, DCO,
  Conventional Commits, performance budget, and every tooling and configuration file.
- [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) · [SECURITY.md](SECURITY.md) · [SUPPORT.md](SUPPORT.md) ·
  [GOVERNANCE.md](GOVERNANCE.md) · [MAINTAINERS.md](MAINTAINERS.md)
- [Architecture decisions](docs/adr) · [Constitution](.specify/memory/constitution.md) ·
  [Feature specs](specs)

## License

Apache-2.0. See [LICENSE](LICENSE) and [NOTICE](NOTICE).
