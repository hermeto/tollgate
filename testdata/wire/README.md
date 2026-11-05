# Wire-format fixtures

Recorded request/response shapes of the official provider APIs. They are **never** captured by
calling a paid API: each file is copied or transcribed from official documentation examples or from
the official SDKs' own test suites, at the versions listed below. Contract tests compare the
simulator's output (and, from feature 004, the gateway's) with these files by structure: same fields,
same value types, same event order. Identifiers, timestamps, text and numbers are not compared.

## OpenAI (`openai/`)

Sources, read on 2026-11-05:

- [Chat Completions API reference](https://platform.openai.com/docs/api-reference/chat/create) and
  [streaming reference](https://platform.openai.com/docs/api-reference/chat-streaming) — response and
  chunk examples.
- [openai-python v3.46.0](https://github.com/openai/openai-python/tree/v3.46.0) —
  `src/openai/types/chat/chat_completion.py`, `chat_completion_chunk.py`, `completion_usage.py`,
  `types/create_embedding_response.py` (field names and optionality).
- [openai-node v7.51.0 `tests/lib/ChatCompletionStream.test.ts`](https://github.com/openai/openai-node/blob/v7.51.0/tests/lib/ChatCompletionStream.test.ts)
  — SSE framing (`data: <json>\n\n`, terminated by `data: [DONE]`).
- [Embeddings API reference](https://platform.openai.com/docs/api-reference/embeddings/create) and
  [Error codes guide](https://platform.openai.com/docs/guides/error-codes).

| File | Shape | Source |
| --- | --- | --- |
| `chat_completion.json` | non-streaming text completion | Chat API reference example; usage details from `completion_usage.py` |
| `chat_completion_tool_call.json` | non-streaming tool call (`finish_reason: tool_calls`) | Chat API reference "Functions" example |
| `chat_completion_stream.sse` | stream without usage | Streaming reference example |
| `chat_completion_stream_usage.sse` | stream with `stream_options.include_usage` (usage `null` on every chunk, final chunk with `choices: []`) | Streaming reference, `stream_options` description |
| `chat_completion_tool_call_stream.sse` | streamed tool call deltas | `chat_completion_chunk.py` (`ChoiceDeltaToolCall`) |
| `embeddings_float.json` | `encoding_format: float` | Embeddings API reference example |
| `embeddings_base64.json` | `encoding_format: base64` (little-endian float32; the SDKs request this by default) | `resources/embeddings.py` in openai-python v3.46.0 |
| `models.json` | model list | Models API reference example |
| `error_401.json`, `error_404_model.json`, `error_429.json` | error bodies | Error codes guide |

## Anthropic (`anthropic/`)

Sources, read on 2026-11-05:

- [anthropic-sdk-python v2.5.0](https://github.com/anthropics/anthropic-sdk-python/tree/v2.5.0/tests/lib/streaming/fixtures)
  (Apache-2.0, see `NOTICE`) — streaming fixtures copied verbatim.
- [Messages API reference](https://docs.claude.com/en/api/messages),
  [streaming guide](https://docs.claude.com/en/docs/build-with-claude/streaming),
  [tool use guide](https://docs.claude.com/en/docs/agents-and-tools/tool-use/overview) and
  [errors](https://docs.claude.com/en/api/errors) — response and error examples.

| File | Shape | Source |
| --- | --- | --- |
| `message.json` | non-streaming text message | Messages API reference example |
| `message_tool_use.json` | non-streaming `tool_use` block, `stop_reason: tool_use` | Tool use guide example |
| `message_stream.sse` | `message_start` → `content_block_start` → `ping` → `content_block_delta`… → `content_block_stop` → `message_delta` → `message_stop` | SDK fixture `basic_response.txt` |
| `message_tool_use_stream.sse` | text block followed by a `tool_use` block streamed with `input_json_delta` | SDK fixture `tool_use_response.txt` |
| `error_401.json`, `error_404.json`, `error_429.json`, `error_529.json` | error bodies with `request_id` | Errors reference |

The SDK fixtures were recorded at different API versions: `basic_response.txt` has the older usage object
(`input_tokens`, `output_tokens` only), `tool_use_response.txt` the current one (cache token fields,
`service_tier`, tool `caller`). Contract tests therefore require the simulator to contain **every** field
of a fixture with the same type, and allow newer fields.
