# Data model: 005 identity-and-policies

`VirtualKey`: id, prefix, hash, team, project, created_at, expires_at, revoked_at.
`Policy`: enabled, mode (`enforce|dry-run`), models allowlist and provenance map.
`PolicySnapshot`: version, loaded_at, org policy, team policies and project policies.
