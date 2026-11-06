// Package canonical contains protocol-neutral request and response values.
package canonical

import "encoding/json"

type ContentPart struct {
	Type string          `json:"type"`
	Text string          `json:"text,omitempty"`
	Data json.RawMessage `json:"data,omitempty"`
}
type Message struct {
	Role       string          `json:"role"`
	Content    json.RawMessage `json:"content"`
	Name       string          `json:"name,omitempty"`
	ToolCallID string          `json:"tool_call_id,omitempty"`
}
type Tool struct {
	Type        string          `json:"type,omitempty"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"input_schema,omitempty"`
	Parameters  json.RawMessage `json:"parameters,omitempty"`
}
type Request struct {
	Protocol  string
	Model     string
	Messages  []Message
	System    json.RawMessage
	Tools     []Tool
	Stream    bool
	MaxTokens *int
	Raw       json.RawMessage
}
type Usage struct {
	InputTokens       int
	OutputTokens      int
	CachedInputTokens int
}
type Response struct {
	ID         string
	Model      string
	Content    json.RawMessage
	StopReason string
	Usage      Usage
}
