# ADR 0006: Cost export format

## Status

Accepted.

## Decision

Usage events use stable provider-neutral accounting fields and USD amounts. Future chargeback output
will map these fields to the current FOCUS specification; event ingestion itself remains independent of
the export format.

## Consequences

Historical events can be recomputed and exported without storing request content or coupling the hot
path to a reporting schema.
