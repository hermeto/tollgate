# 0002. State stores: Valkey for counters, PostgreSQL for events, keys and audit

- Status: accepted
- Date: 2026-11-05
- Deciders: agent (on the maintainer's behalf, feature 003)

## Context and problem statement

Rate limits and budgets need counters shared by all replicas with sub-millisecond atomic updates.
Usage events, virtual keys and the audit trail need durable, queryable, append-heavy storage with
transactions and SQL for chargeback. Which stores, and how is the schema evolved?

## Considered options

1. Redis-compatible store for counters and PostgreSQL for everything durable.
2. PostgreSQL only (counters as rows with `UPDATE … RETURNING`).
3. Redis-compatible store only (events and audit in streams).

## Decision outcome

Chosen option 1, with **Valkey** (BSD-3-Clause, Redis-protocol compatible) as the default
Redis-compatible store and **PostgreSQL 18** for usage events, keys and audit. The Go driver for
PostgreSQL is `jackc/pgx/v5`; the Redis client is chosen in feature 007 (go-redis or rueidis, by
benchmark). Schema migrations use **goose** (v3) with SQL files embedded in the binary.

### Consequences

- Good: counters stay off the database's write path (Lua scripts give atomic GCRA/sliding-window
  updates); PostgreSQL gives partitioning, append-only grants for audit and SQL for chargeback.
- Good: Valkey avoids pulling a copyleft or source-available server into an Apache-2.0 reference stack,
  and any Redis-compatible service (including managed Redis) still works.
- Bad: two stateful dependencies to operate; each needs an explicit fail mode per policy (constitution IV).
- goose was chosen over golang-migrate for its small API, embedded SQL migrations and library mode; either
  would work.

## More information

- Research: [specs/003-local-stack/research.md](../../specs/003-local-stack/research.md) R1, R8.
- Valkey licence: <https://github.com/valkey-io/valkey/blob/unstable/COPYING>.
