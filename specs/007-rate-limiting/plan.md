# Implementation Plan: Rate Limiting

Use a token bucket behind a small limiter interface. The first slice provides deterministic local
state and decision metadata; a Redis atomic implementation can satisfy the same contract later.
