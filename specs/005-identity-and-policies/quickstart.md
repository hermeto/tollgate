# Quickstart: identity and policies

Create a local key with `tollgate keys create --team payments --project checkout --store /tmp/keys.json`.
Validate examples with `tollgate policy validate policies/examples`. The gateway only enforces identity when a
key store is configured, preserving the feature 003 local profile.
