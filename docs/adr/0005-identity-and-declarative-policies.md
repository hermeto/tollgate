# ADR 0005: Virtual identity keys and declarative policy inheritance

## Status

Accepted.

## Decision

The gateway supports optional, locally persisted virtual keys. Only a SHA-256
digest is stored; the bearer secret is displayed once by the CLI. Keys carry
team and project context, and revocation, rotation, and expiry are enforced at
the gateway boundary.

Authorization is expressed as YAML policy documents inherited in deterministic
order: organization, team, then project. A project policy replaces the model
allowlist inherited from its parent. Policy reloads are atomic: an invalid
update leaves the last valid snapshot active. `enforce` denies unauthorized
models, while `dry-run` records the decision and permits the request.

## Consequences

This keeps the default gateway deployment backward compatible when identity is
not configured, while providing an auditable control plane that is easy to
review in source control. The local JSON store and filesystem policy loader
are intentionally replaceable seams for a future database/control-plane
integration.
