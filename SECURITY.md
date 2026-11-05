# Security policy

Tollgate is a security control: it holds provider credentials, enforces policies and keeps an audit
trail. We take reports seriously and thank everyone who reports responsibly.

## Reporting a vulnerability

**Do not open a public issue.** Report privately through GitHub's
[private vulnerability reporting](https://github.com/hermeto/tollgate/security/advisories/new).
Include the affected version or commit, a description, reproduction steps and the impact you expect.
Use synthetic data only; never include real prompts, keys or personal data.

## Scope

In scope:

- The gateway request path: authentication, protocol handling, streaming, routing and fallback.
- **Policy bypass**: any way to exceed a rate limit or budget, call a model outside the allowlist, or
  skip a policy that should apply.
- **PII leakage**: PII reaching a provider when a policy redacts or blocks it, or appearing in logs,
  traces, metrics, cache, usage events or audit records.
- **Audit tampering**: modifying, deleting or reordering audit records without `tollgate audit verify`
  detecting it.
- Exposure of provider credentials or virtual keys; admin API access control.
- The release supply chain: signatures, SBOMs, provenance, container images and the Helm chart.

Out of scope:

- Vulnerabilities in model providers themselves, or model output quality.
- Denial of service that requires traffic volumes beyond configured limits, unless a limit is bypassed.
- Findings that require a compromised host or an operator who disables a protection on purpose.
- Missing hardening in the local demo profile (`make up`) when it is documented as dev-only.

## Supported versions

Only the latest released minor version (before 1.0: the latest `0.x` release) and `main` receive
security fixes.

## Response targets

| Step | Target |
| --- | --- |
| Acknowledge the report | within 3 business days |
| Triage and severity assessment | within 7 calendar days |
| Fix or mitigation for HIGH/CRITICAL | within 90 days |
| Coordinated disclosure | at fix release, or 90 days after the report, whichever comes first |

We publish a GitHub Security Advisory with credit to the reporter unless they prefer otherwise.
