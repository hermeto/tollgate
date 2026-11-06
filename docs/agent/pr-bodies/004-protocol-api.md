## Summary

Initial slice of feature 004 adds the canonical protocol model, bounded OpenAI and Anthropic decoders,
logical model schema/loading, same-protocol route support and the first OpenAI↔Anthropic translation
support matrix.

## Base

Based on `003-local-stack-iter5`; merge after the local-stack PR chain.

## Requirements covered

FR-API-1 through FR-API-7 (foundational portions); tasks T001–T014.

## Decisions for the reviewer

- Raw request bytes remain on the canonical request for future policy stages and same-protocol passthrough.
- Logical model targets rewrite only the `model` field and choose the first configured target; routing
  weights/fallback arrive in later features.
- Unsupported translated roles fail explicitly; same-protocol unknown fields are preserved.
- T015 remains open for the follow-up slice because gateway SDK execution requires the Compose daemon and
  the full 60-second fuzz/SDK validation belongs with that compatibility pass.

## Verification

- `make test`
- `go test -race ./...`
- `golangci-lint run ./...`
- fuzz targets exercised for 30 seconds each
