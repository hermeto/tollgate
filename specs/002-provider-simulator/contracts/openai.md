# Contract: OpenAI flavour

| Method | Path | Notes |
| --- | --- | --- |
| POST | `/v1/chat/completions` | `stream`, `stream_options.include_usage`, `max_tokens` / `max_completion_tokens`, `tools`, `tool_choice`; `n > 1` → 400 |
| POST | `/v1/embeddings` | `input` string or array, `encoding_format` float (default) or base64, `dimensions` |
| GET | `/v1/models` | configured models |

Auth: `Authorization: Bearer <any>` required, else 401 `invalid_api_key`.
Headers on every response: `x-request-id`, `openai-processing-ms`.
Shapes: research R1, R2; fixtures in `testdata/wire/openai/`.
