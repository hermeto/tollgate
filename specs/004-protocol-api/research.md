# Research: 004 protocol-api

- OpenAI Chat Completions and Embeddings wire fixtures already exist under `testdata/wire/openai` and simulator
  handlers are the local contract source.
- Anthropic Messages fixtures and event names are defined by `sim/anthropic.go` and will be copied into synthetic
  fixtures; no paid provider calls are permitted.
- Translation keeps support names in one package because provider conventions change; unsupported fields return a
  protocol error instead of being silently discarded.
- Logical model config is YAML normalized through the existing JSON-schema loader.
