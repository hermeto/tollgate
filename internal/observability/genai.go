// Package observability centralizes stable GenAI telemetry attribute names.
package observability

import "go.opentelemetry.io/otel/attribute"

func Attributes(provider, model, operation string, input, output int) []attribute.KeyValue {
	return []attribute.KeyValue{attribute.String("gen_ai.system", provider), attribute.String("gen_ai.request.model", model), attribute.String("gen_ai.operation.name", operation), attribute.Int("gen_ai.usage.input_tokens", input), attribute.Int("gen_ai.usage.output_tokens", output)}
}
