// Package translate converts the supported text/tool chat subset between protocols.
package translate

import (
	"encoding/json"
	"fmt"

	"github.com/hermeto/tollgate/internal/api"
	"github.com/hermeto/tollgate/internal/canonical"
)

type Feature string

const (
	FeatureMultimodal        Feature = "multimodal"
	FeatureProviderExtension Feature = "provider_extension"
)

func OpenAIToAnthropic(req canonical.Request) ([]byte, error) {
	if len(req.Messages) == 0 {
		return nil, api.BadRequest("messages must not be empty")
	}
	if req.MaxTokens == nil {
		return nil, api.BadRequest("max_tokens is required for Anthropic translation")
	}
	var system json.RawMessage
	messages := make([]canonical.Message, 0, len(req.Messages))
	for _, msg := range req.Messages {
		switch msg.Role {
		case "system":
			if len(system) != 0 {
				return nil, api.BadRequest("multiple system messages are unsupported")
			}
			system = msg.Content
		case "user", "assistant":
			messages = append(messages, msg)
		default:
			return nil, api.BadRequest(fmt.Sprintf("role %q is unsupported by translation", msg.Role))
		}
	}
	type request struct {
		Model     string              `json:"model"`
		System    json.RawMessage     `json:"system,omitempty"`
		Messages  []canonical.Message `json:"messages"`
		Tools     []canonical.Tool    `json:"tools,omitempty"`
		MaxTokens int                 `json:"max_tokens"`
		Stream    bool                `json:"stream,omitempty"`
	}
	return json.Marshal(request{Model: req.Model, System: system, Messages: messages, Tools: req.Tools, MaxTokens: *req.MaxTokens, Stream: req.Stream})
}

func AnthropicToOpenAI(req canonical.Request) ([]byte, error) {
	if len(req.Messages) == 0 {
		return nil, api.BadRequest("messages must not be empty")
	}
	messages := make([]canonical.Message, 0, len(req.Messages)+1)
	if len(req.System) > 0 {
		messages = append(messages, canonical.Message{Role: "system", Content: req.System})
	}
	for _, msg := range req.Messages {
		if msg.Role != "user" && msg.Role != "assistant" {
			return nil, api.BadRequest(fmt.Sprintf("role %q is unsupported by translation", msg.Role))
		}
		messages = append(messages, msg)
	}
	type request struct {
		Model     string              `json:"model"`
		Messages  []canonical.Message `json:"messages"`
		Tools     []canonical.Tool    `json:"tools,omitempty"`
		MaxTokens *int                `json:"max_tokens,omitempty"`
		Stream    bool                `json:"stream,omitempty"`
	}
	return json.Marshal(request{Model: req.Model, Messages: messages, Tools: req.Tools, MaxTokens: req.MaxTokens, Stream: req.Stream})
}
