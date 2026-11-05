# Contract: Anthropic flavour

| Method | Path | Notes |
| --- | --- | --- |
| POST | `/v1/messages` | `max_tokens` required, `stream`, `system`, `tools`, `tool_choice` |

Auth: `x-api-key: <any>` required, else 401 `authentication_error`. `anthropic-version` is accepted
and not enforced. Response header `request-id` on every response.
Shapes and event order: research R3; fixtures in `testdata/wire/anthropic/`.
