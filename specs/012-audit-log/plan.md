# Implementation Plan: Audit Log

Use a mutex-protected append-only event sequence with SHA-256 predecessor hashes. Persistence and
rotation can wrap this core contract later without changing verification semantics.
