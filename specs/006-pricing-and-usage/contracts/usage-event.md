# Usage event contract

Usage events are JSON objects containing only accounting metadata and token counts. `estimated: true`
means provider-reported usage was unavailable. `cost_usd` is computed at ingestion with the effective
price active at the event timestamp.
