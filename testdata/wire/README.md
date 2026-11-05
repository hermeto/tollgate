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
| `error_401.json`, `error_404_model.json`, `error_429.json` | error bodies | Error codes guide |
