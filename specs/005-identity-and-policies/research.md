# Research: 005 identity-and-policies

- A 32-byte random secret encoded base62 provides sufficient entropy; SHA-256 is appropriate for lookup of
  high-entropy secrets and avoids password-style slow hashing on the hot path.
- Atomic replacement uses a temporary file in the same directory followed by rename and restrictive permissions.
- Policy snapshots are immutable values swapped atomically; malformed updates never replace the current snapshot.
