# Quickstart: protocol API

Start the local profile with `make up`, then use `make sdk-compat TARGET=gateway`. The OpenAI endpoint is
`http://127.0.0.1:8080/v1`; the Anthropic endpoint is `http://127.0.0.1:8080`.

Translation and parser tests use `go test ./internal/api/... ./internal/translate/... -race` and never require
provider credentials.
