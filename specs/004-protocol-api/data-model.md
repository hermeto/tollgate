# Data model: 004 protocol-api

```text
LogicalModel { name, targets[] }
Target { provider, model, base_url }
CanonicalRequest { protocol, model, messages[], tools[], stream, raw }
Message { role, content[] }
ContentPart { type, text, data }
Tool { name, description, input_schema }
CanonicalResponse { id, model, choices[], usage, stop_reason }
```

Logical model names are non-empty strings, targets have provider `openai` or `anthropic`, and a model must have
at least one target. Raw request bytes are retained only for same-protocol forwarding.
