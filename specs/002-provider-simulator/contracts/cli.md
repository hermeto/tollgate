# Contract: `tollgate sim`

```text
tollgate sim --flavor openai|anthropic [--host 127.0.0.1] [--port N] [--config sim.yaml] [--seed N]
```

Flags override the config file. Exit 0 on SIGINT/SIGTERM after in-flight requests finish (5 s grace),
1 on configuration or listen errors (message on stderr). Logs: one JSON line (slog) per start/stop;
no per-request logs by default (`--log-requests` enables them).
