# Implementation Plan: Response Cache

Provide an in-memory exact cache with content-addressed keys, TTL and bounded item count. Semantic
cache and distributed invalidation remain outside this first deterministic slice.
