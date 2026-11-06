# Rate-limit decision contract

`Decision` returns `allowed`, `limit`, `remaining` and `reset`. An adapter maps a denied decision to
the caller's protocol-specific 429 response and `Retry-After` header.
