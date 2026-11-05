# Contract: `tollgate serve`

```text
tollgate serve --config <file>     # exit 0 after a clean drain, 1 on config/listen errors
tollgate probe <url>               # exit 0 when GET <url> returns 2xx within 2 s (container healthchecks)
```

Config: data-model "Gateway config". Unknown fields are rejected by the schema.
