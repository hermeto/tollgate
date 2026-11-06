# Contract: logical models

`schemas/models.schema.json` validates a map of logical names to ordered targets. `/v1/models` exposes logical
names only; an unknown logical name returns a protocol-correct error and does not call an upstream.
