// Package openai decodes the OpenAI request surface used by the canonical pipeline.
package openai

import (
	"bytes"
	"encoding/json"

	"github.com/hermeto/tollgate/internal/api"
	"github.com/hermeto/tollgate/internal/canonical"
)

type chatRequest struct {
	Model               string              `json:"model"`
	Messages            []canonical.Message `json:"messages"`
	Stream              bool                `json:"stream"`
	Tools               []canonical.Tool    `json:"tools"`
	MaxTokens           *int                `json:"max_tokens"`
	MaxCompletionTokens *int                `json:"max_completion_tokens"`
}
type EmbeddingsRequest struct {
	Model string          `json:"model"`
	Input json.RawMessage `json:"input"`
}

func DecodeChat(body []byte) (canonical.Request, error) {
	var value chatRequest
	dec := json.NewDecoder(bytes.NewReader(body))
	if err := dec.Decode(&value); err != nil {
		return canonical.Request{}, api.BadRequest("invalid JSON body")
	}
	if value.Model == "" {
		return canonical.Request{}, api.BadRequest("model is required")
	}
	if len(value.Messages) == 0 {
		return canonical.Request{}, api.BadRequest("messages must not be empty")
	}
	max := value.MaxTokens
	if max == nil {
		max = value.MaxCompletionTokens
	}
	return canonical.Request{Protocol: "openai", Model: value.Model, Messages: value.Messages, Tools: value.Tools, Stream: value.Stream, MaxTokens: max, Raw: append([]byte(nil), body...)}, nil
}

func DecodeEmbeddings(body []byte) (EmbeddingsRequest, error) {
	var value EmbeddingsRequest
	if err := json.Unmarshal(body, &value); err != nil || value.Model == "" {
		return EmbeddingsRequest{}, api.BadRequest("model and input are required")
	}
	return value, nil
}
