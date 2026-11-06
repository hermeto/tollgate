## Summary

Feature 004 adds the canonical protocol model, bounded OpenAI and Anthropic decoders, logical model
schema/loading, same-protocol route support, OpenAI↔Anthropic translation support and gateway SDK
compatibility.

## Base

Based on `003-local-stack-iter5`; merge after the local-stack PR chain.

## Requirements covered

FR-API-1 through FR-API-7; tasks T001–T015.

## Decisions for the reviewer

- Raw request bytes remain on the canonical request for future policy stages and same-protocol passthrough.
- Logical model targets rewrite only the `model` field and choose the first configured target; routing
  weights/fallback arrive in later features.
- Unsupported translated roles fail explicitly; same-protocol unknown fields are preserved.
- `TARGET=gateway make sdk-compat` runs all six official OpenAI/Anthropic SDK checks through the local
  gateway and routes each provider through its simulator.

## Verification

- `make test`
- `go test -race ./...`
- `golangci-lint run ./...`
- fuzz targets exercised for 60 seconds each
- `TARGET=gateway make sdk-compat` (all 6 clients passed)
