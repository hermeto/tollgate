// Package anthropic decodes the Anthropic Messages request surface.
package anthropic

import (
	"encoding/json"

	"github.com/hermeto/tollgate/internal/api"
	"github.com/hermeto/tollgate/internal/canonical"
)

type messageRequest struct {
	Model     string              `json:"model"`
	System    json.RawMessage     `json:"system"`
	Messages  []canonical.Message `json:"messages"`
	Tools     []canonical.Tool    `json:"tools"`
	Stream    bool                `json:"stream"`
	MaxTokens *int                `json:"max_tokens"`
}

func DecodeMessages(body []byte) (canonical.Request, error) {
	var value messageRequest
	if err := json.Unmarshal(body, &value); err != nil {
		return canonical.Request{}, api.BadRequest("invalid JSON body")
	}
	if value.Model == "" || value.MaxTokens == nil || len(value.Messages) == 0 {
		return canonical.Request{}, api.BadRequest("model, max_tokens and messages are required")
	}
	return canonical.Request{Protocol: "anthropic", Model: value.Model, System: value.System, Messages: value.Messages, Tools: value.Tools, Stream: value.Stream, MaxTokens: value.MaxTokens, Raw: append([]byte(nil), body...)}, nil
}
