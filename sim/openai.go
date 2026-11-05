package sim

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"strings"
	"time"

	"github.com/hermeto/tollgate/internal/tokens"
)

// OpenAI wire types (research R1, R2; fixtures in testdata/wire/openai).

type oaiChatRequest struct {
	Model               string           `json:"model"`
	Messages            []oaiMessage     `json:"messages"`
	Stream              bool             `json:"stream"`
	StreamOptions       *oaiStreamOption `json:"stream_options"`
	MaxTokens           *int             `json:"max_tokens"`
	MaxCompletionTokens *int             `json:"max_completion_tokens"`
	N                   *int             `json:"n"`
	Tools               []oaiTool        `json:"tools"`
	ToolChoice          json.RawMessage  `json:"tool_choice"`
}

type oaiStreamOption struct {
	IncludeUsage bool `json:"include_usage"`
}

type oaiMessage struct {
	Role      string          `json:"role"`
	Content   json.RawMessage `json:"content"`
	ToolCalls json.RawMessage `json:"tool_calls"`
}

type oaiTool struct {
	Type     string `json:"type"`
	Function struct {
		Name string `json:"name"`
	} `json:"function"`
}

type oaiToolCall struct {
	Index    *int            `json:"index,omitempty"`
	ID       string          `json:"id,omitempty"`
	Type     string          `json:"type,omitempty"`
	Function oaiFunctionCall `json:"function"`
}

type oaiFunctionCall struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments"`
}

type oaiUsage struct {
	PromptTokens        int `json:"prompt_tokens"`
	CompletionTokens    int `json:"completion_tokens"`
	TotalTokens         int `json:"total_tokens"`
	PromptTokensDetails struct {
		CachedTokens int `json:"cached_tokens"`
		AudioTokens  int `json:"audio_tokens"`
	} `json:"prompt_tokens_details"`
	CompletionTokensDetails struct {
		ReasoningTokens          int `json:"reasoning_tokens"`
		AudioTokens              int `json:"audio_tokens"`
		AcceptedPredictionTokens int `json:"accepted_prediction_tokens"`
		RejectedPredictionTokens int `json:"rejected_prediction_tokens"`
	} `json:"completion_tokens_details"`
}

type oaiCompletion struct {
	ID                string      `json:"id"`
	Object            string      `json:"object"`
	Created           int64       `json:"created"`
	Model             string      `json:"model"`
	Choices           []oaiChoice `json:"choices"`
	Usage             oaiUsage    `json:"usage"`
	ServiceTier       string      `json:"service_tier"`
	SystemFingerprint string      `json:"system_fingerprint"`
}

type oaiChoice struct {
	Index        int             `json:"index"`
	Message      oaiReplyMessage `json:"message"`
	Logprobs     *struct{}       `json:"logprobs"`
	FinishReason string          `json:"finish_reason"`
}

type oaiReplyMessage struct {
	Role        string        `json:"role"`
	Content     *string       `json:"content"`
	Refusal     *string       `json:"refusal"`
	Annotations []struct{}    `json:"annotations"`
	ToolCalls   []oaiToolCall `json:"tool_calls,omitempty"`
}

type oaiChunk struct {
	ID                string           `json:"id"`
	Object            string           `json:"object"`
	Created           int64            `json:"created"`
	Model             string           `json:"model"`
	ServiceTier       string           `json:"service_tier"`
	SystemFingerprint string           `json:"system_fingerprint"`
	Choices           []oaiChunkChoice `json:"choices"`
	// Usage is omitted unless stream_options.include_usage is set; then it is null on every chunk
	// except the final one.
	Usage json.RawMessage `json:"usage,omitempty"`
}

type oaiChunkChoice struct {
	Index        int       `json:"index"`
	Delta        oaiDelta  `json:"delta"`
	Logprobs     *struct{} `json:"logprobs"`
	FinishReason *string   `json:"finish_reason"`
}

type oaiDelta struct {
	Role      string          `json:"role,omitempty"`
	Content   json.RawMessage `json:"content,omitempty"`
	ToolCalls []oaiToolCall   `json:"tool_calls,omitempty"`
	Refusal   json.RawMessage `json:"refusal,omitempty"`
}

const systemFingerprint = "fp_tollgate_sim"

var jsonNull = json.RawMessage("null")

func (s *Server) registerOpenAI() {
	s.mux.HandleFunc("POST /v1/chat/completions", s.requireAuth(s.handleOpenAIChat))
}

// reply is the generated answer to a chat request.
type reply struct {
	pieces    []string // text pieces, one token each
	toolName  string   // set when the request forced a tool call
	toolArgs  string
	truncated bool
}

func (r reply) completionTokens() int {
	if r.toolName != "" {
		return tokens.Count(r.toolName) + tokens.Count(r.toolArgs)
	}
	return len(r.pieces)
}

func (s *Server) handleOpenAIChat(w http.ResponseWriter, r *http.Request) {
	body, ok := s.readBody(w, r)
	if !ok {
		return
	}
	var req oaiChatRequest
	if err := json.Unmarshal(body, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, errInvalidRequest, "We could not parse the JSON body of your request: "+err.Error())
		return
	}
	switch {
	case req.Model == "":
		s.writeError(w, http.StatusBadRequest, errInvalidRequest, "you must provide a model parameter")
		return
	case !s.models[req.Model]:
		s.writeError(w, http.StatusNotFound, errModelNotFound, fmt.Sprintf("The model `%s` does not exist or you do not have access to it.", req.Model))
		return
	case len(req.Messages) == 0:
		s.writeError(w, http.StatusBadRequest, errInvalidRequest, "[] is too short - 'messages'")
		return
	case req.N != nil && *req.N != 1:
		s.writeError(w, http.StatusBadRequest, errInvalidRequest, "the simulator supports only n=1")
		return
	}

	h := requestHash(r.URL.Path, body)
	rep := s.openAIReply(s.seeds.content(h), req)
	prompt := openAIPromptTokens(req)
	usage := oaiUsage{PromptTokens: prompt, CompletionTokens: rep.completionTokens()}
	usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens

	id := newID("chatcmpl-", 29)
	created := time.Now().Unix()
	w.Header().Set("X-Request-Id", newID("req_", 32))
	if req.Stream {
		s.streamOpenAIChat(w, r, req, rep, usage, id, created)
		return
	}
	msg := oaiReplyMessage{Role: "assistant", Annotations: []struct{}{}}
	finish := "stop"
	switch {
	case rep.toolName != "":
		msg.ToolCalls = []oaiToolCall{{ID: newID("call_", 24), Type: "function", Function: oaiFunctionCall{Name: rep.toolName, Arguments: rep.toolArgs}}}
		finish = "tool_calls"
	default:
		text := strings.Join(rep.pieces, "")
		msg.Content = &text
		if rep.truncated {
			finish = "length"
		}
	}
	s.writeJSON(w, http.StatusOK, oaiCompletion{
		ID: id, Object: "chat.completion", Created: created, Model: req.Model,
		Choices:     []oaiChoice{{Index: 0, Message: msg, FinishReason: finish}},
		Usage:       usage,
		ServiceTier: "default", SystemFingerprint: systemFingerprint,
	})
}

// openAIReply decides between a forced tool call (spec FR-017) and generated text.
func (s *Server) openAIReply(rng *rand.Rand, req oaiChatRequest) reply {
	if name := forcedOpenAITool(req); name != "" {
		return reply{toolName: name, toolArgs: "{}"}
	}
	limit := 0
	switch {
	case req.MaxCompletionTokens != nil:
		limit = *req.MaxCompletionTokens
	case req.MaxTokens != nil:
		limit = *req.MaxTokens
	}
	n, cut := s.outputLength(rng, limit)
	return reply{pieces: genText(rng, n), truncated: cut}
}

// forcedOpenAITool returns the function a request forces: tool_choice "required" picks the first
// function tool; {"type":"function","function":{"name":X}} picks X.
func forcedOpenAITool(req oaiChatRequest) string {
	if len(req.Tools) == 0 || len(req.ToolChoice) == 0 {
		return ""
	}
	var choice string
	if json.Unmarshal(req.ToolChoice, &choice) == nil {
		if choice == "required" {
			return req.Tools[0].Function.Name
		}
		return ""
	}
	var named struct {
		Type     string `json:"type"`
		Function struct {
			Name string `json:"name"`
		} `json:"function"`
	}
	if json.Unmarshal(req.ToolChoice, &named) == nil && named.Type == "function" {
		return named.Function.Name
	}
	return ""
}

// openAIPromptTokens applies the documented counting method (internal/tokens) to the messages and,
// when present, the tool definitions.
func openAIPromptTokens(req oaiChatRequest) int {
	msgs := make([]tokens.Message, 0, len(req.Messages))
	for _, m := range req.Messages {
		msgs = append(msgs, tokens.Message{Role: m.Role, Content: messageText(m.Content) + string(m.ToolCalls)})
	}
	n := tokens.CountChat(msgs)
	if len(req.Tools) > 0 {
		if b, err := json.Marshal(req.Tools); err == nil {
			n += tokens.Count(string(b))
		}
	}
	return n
}

// messageText extracts text from a content field that is either a string or an array of parts.
func messageText(raw json.RawMessage) string {
	if len(raw) == 0 || bytes.Equal(raw, jsonNull) {
		return ""
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	var parts []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if json.Unmarshal(raw, &parts) != nil {
		return string(raw)
	}
	var b strings.Builder
	for _, p := range parts {
		b.WriteString(p.Text)
	}
	return b.String()
}

func (s *Server) streamOpenAIChat(w http.ResponseWriter, r *http.Request, req oaiChatRequest, rep reply, usage oaiUsage, id string, created int64) {
	includeUsage := req.StreamOptions != nil && req.StreamOptions.IncludeUsage
	chunk := func(delta oaiDelta, finish *string) oaiChunk {
		c := oaiChunk{
			ID: id, Object: "chat.completion.chunk", Created: created, Model: req.Model,
			ServiceTier: "default", SystemFingerprint: systemFingerprint,
			Choices: []oaiChunkChoice{{Index: 0, Delta: delta, FinishReason: finish}},
		}
		if includeUsage {
			c.Usage = jsonNull
		}
		return c
	}
	sse := startSSE(w)
	send := func(v any) bool {
		data, err := json.Marshal(v)
		if err != nil || r.Context().Err() != nil {
			return false
		}
		return sse.event("", data) == nil
	}

	finish := "stop"
	if rep.toolName != "" {
		zero := 0
		first := oaiDelta{Role: "assistant", Content: jsonNull, Refusal: jsonNull, ToolCalls: []oaiToolCall{{
			Index: &zero, ID: newID("call_", 24), Type: "function", Function: oaiFunctionCall{Name: rep.toolName},
		}}}
		args := oaiDelta{ToolCalls: []oaiToolCall{{Index: &zero, Function: oaiFunctionCall{Arguments: rep.toolArgs}}}}
		if !send(chunk(first, nil)) || !send(chunk(args, nil)) {
			return
		}
		finish = "tool_calls"
	} else {
		if !send(chunk(oaiDelta{Role: "assistant", Content: json.RawMessage(`""`), Refusal: jsonNull}, nil)) {
			return
		}
		for _, p := range rep.pieces {
			content, _ := json.Marshal(p)
			if !send(chunk(oaiDelta{Content: content}, nil)) {
				return // client went away: stop generating
			}
		}
		if rep.truncated {
			finish = "length"
		}
	}
	if !send(chunk(oaiDelta{}, &finish)) {
		return
	}
	if includeUsage {
		u, _ := json.Marshal(usage)
		final := chunk(oaiDelta{}, nil)
		final.Choices = []oaiChunkChoice{}
		final.Usage = u
		if !send(final) {
			return
		}
	}
	_ = sse.event("", []byte("[DONE]"))
}
