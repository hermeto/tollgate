package sim

import (
	"bytes"
	"encoding/json"
	"math/rand/v2"
	"net/http"
	"strings"

	"github.com/hermeto/tollgate/internal/tokens"
)

// Anthropic Messages wire types (research R3; fixtures in testdata/wire/anthropic).

type antRequest struct {
	Model      string          `json:"model"`
	MaxTokens  *int            `json:"max_tokens"`
	Messages   []antMessage    `json:"messages"`
	System     json.RawMessage `json:"system"`
	Stream     bool            `json:"stream"`
	Tools      []antTool       `json:"tools"`
	ToolChoice *antToolChoice  `json:"tool_choice"`
}

type antMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

type antTool struct {
	Name string `json:"name"`
}

type antToolChoice struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type antUsage struct {
	InputTokens              int    `json:"input_tokens"`
	CacheCreationInputTokens int    `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int    `json:"cache_read_input_tokens"`
	OutputTokens             int    `json:"output_tokens"`
	ServiceTier              string `json:"service_tier"`
}

type antMessageResponse struct {
	ID           string           `json:"id"`
	Type         string           `json:"type"`
	Role         string           `json:"role"`
	Model        string           `json:"model"`
	Content      []map[string]any `json:"content"`
	StopReason   *string          `json:"stop_reason"`
	StopSequence *string          `json:"stop_sequence"`
	Usage        antUsage         `json:"usage"`
}

func (s *Server) registerAnthropic() {
	s.mux.HandleFunc("POST /v1/messages", s.requireAuth(s.handleAnthropicMessages))
}

func (s *Server) handleAnthropicMessages(w http.ResponseWriter, r *http.Request) {
	body, ok := s.readBody(w, r)
	if !ok {
		return
	}
	h := requestHash(r.URL.Path, body)
	p, ok := s.applyFaults(w, r, h)
	if !ok {
		return
	}
	var req antRequest
	if err := json.Unmarshal(body, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, errInvalidRequest, "invalid JSON body: "+err.Error())
		return
	}
	switch {
	case req.Model == "":
		s.writeError(w, http.StatusBadRequest, errInvalidRequest, "model: Field required")
		return
	case !s.models[req.Model]:
		s.writeError(w, http.StatusNotFound, errModelNotFound, "model: "+req.Model)
		return
	case req.MaxTokens == nil:
		s.writeError(w, http.StatusBadRequest, errInvalidRequest, "max_tokens: Field required")
		return
	case *req.MaxTokens < 1:
		s.writeError(w, http.StatusBadRequest, errInvalidRequest, "max_tokens: must be greater than or equal to 1")
		return
	case len(req.Messages) == 0:
		s.writeError(w, http.StatusBadRequest, errInvalidRequest, "messages: at least one message is required")
		return
	}

	rep := s.anthropicReply(s.seeds.content(h), req)
	usage := antUsage{InputTokens: anthropicInputTokens(req), OutputTokens: rep.completionTokens(), ServiceTier: "standard"}
	stop := "end_turn"
	switch {
	case rep.toolName != "":
		stop = "tool_use"
	case rep.truncated:
		stop = "max_tokens"
	}
	id := newID("msg_", 24)
	w.Header().Set("Request-Id", newID("req_", 24))
	if req.Stream {
		s.streamAnthropic(w, r, p, req, rep, usage, id, stop)
		return
	}
	if !p.whole(usage.OutputTokens) {
		return
	}
	s.writeJSON(w, http.StatusOK, antMessageResponse{
		ID: id, Type: "message", Role: "assistant", Model: req.Model,
		Content:    []map[string]any{anthropicBlock(rep, false)},
		StopReason: &stop, Usage: usage,
	})
}

// anthropicReply decides between a forced tool use (spec FR-017) and generated text.
func (s *Server) anthropicReply(rng *rand.Rand, req antRequest) reply {
	if name := forcedAnthropicTool(req); name != "" {
		return reply{toolName: name, toolArgs: "{}"}
	}
	n, cut := s.outputLength(rng, *req.MaxTokens)
	return reply{pieces: genText(rng, n), truncated: cut}
}

// forcedAnthropicTool returns the tool a request forces: {"type":"tool","name":X} picks X and
// {"type":"any"} picks the first tool.
func forcedAnthropicTool(req antRequest) string {
	if req.ToolChoice == nil || len(req.Tools) == 0 {
		return ""
	}
	switch req.ToolChoice.Type {
	case "tool":
		return req.ToolChoice.Name
	case "any":
		return req.Tools[0].Name
	}
	return ""
}

// anthropicInputTokens applies the documented counting method to the system prompt, the messages
// and the tool definitions (the Anthropic tokenizer is not public; see docs/simulator.md).
func anthropicInputTokens(req antRequest) int {
	msgs := make([]tokens.Message, 0, len(req.Messages)+1)
	if sys := blocksText(req.System); sys != "" {
		msgs = append(msgs, tokens.Message{Role: "system", Content: sys})
	}
	for _, m := range req.Messages {
		msgs = append(msgs, tokens.Message{Role: m.Role, Content: blocksText(m.Content)})
	}
	n := tokens.CountChat(msgs)
	if len(req.Tools) > 0 {
		if b, err := json.Marshal(req.Tools); err == nil {
			n += tokens.Count(string(b))
		}
	}
	return n
}

// blocksText extracts the text of a string or an array of content blocks (text, tool_use input,
// tool_result content).
func blocksText(raw json.RawMessage) string {
	if len(raw) == 0 || bytes.Equal(raw, jsonNull) {
		return ""
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	var blocks []struct {
		Type    string          `json:"type"`
		Text    string          `json:"text"`
		Input   json.RawMessage `json:"input"`
		Content json.RawMessage `json:"content"`
	}
	if json.Unmarshal(raw, &blocks) != nil {
		return string(raw)
	}
	var b strings.Builder
	for _, bl := range blocks {
		b.WriteString(bl.Text)
		if len(bl.Input) > 0 {
			b.Write(bl.Input)
		}
		if len(bl.Content) > 0 {
			b.WriteString(blocksText(bl.Content))
		}
	}
	return b.String()
}

// anthropicBlock renders the reply as one content block. For streaming starts, text is empty and
// tool input is {} (deltas carry the content).
func anthropicBlock(rep reply, streamStart bool) map[string]any {
	if rep.toolName != "" {
		return map[string]any{
			"type": "tool_use", "id": newID("toolu_", 24), "name": rep.toolName,
			"caller": map[string]string{"type": "direct"}, "input": map[string]any{},
		}
	}
	text := ""
	if !streamStart {
		text = strings.Join(rep.pieces, "")
	}
	return map[string]any{"type": "text", "text": text}
}

func (s *Server) streamAnthropic(w http.ResponseWriter, r *http.Request, p *pace, req antRequest, rep reply, usage antUsage, id, stop string) {
	if !p.first() {
		return
	}
	sse := startSSE(w)
	send := func(event string, v any) bool {
		data, err := json.Marshal(v)
		if err != nil || r.Context().Err() != nil {
			return false
		}
		return sse.event(event, data) == nil
	}
	startUsage := usage
	startUsage.OutputTokens = 1
	start := map[string]any{"type": "message_start", "message": antMessageResponse{
		ID: id, Type: "message", Role: "assistant", Model: req.Model,
		Content: []map[string]any{}, Usage: startUsage,
	}}
	if !send("message_start", start) ||
		!send("content_block_start", map[string]any{"type": "content_block_start", "index": 0, "content_block": anthropicBlock(rep, true)}) ||
		!send("ping", map[string]string{"type": "ping"}) {
		return
	}
	if rep.toolName != "" {
		for i, partial := range []string{"", rep.toolArgs} {
			delta := map[string]any{"type": "content_block_delta", "index": 0,
				"delta": map[string]string{"type": "input_json_delta", "partial_json": partial}}
			if (i > 0 && !p.next(i)) || !send("content_block_delta", delta) {
				return
			}
		}
	} else {
		for i, piece := range rep.pieces {
			if i > 0 && !p.next(i) {
				return // client went away: stop generating
			}
			delta := map[string]any{"type": "content_block_delta", "index": 0,
				"delta": map[string]string{"type": "text_delta", "text": piece}}
			if !send("content_block_delta", delta) {
				return
			}
		}
	}
	if !send("content_block_stop", map[string]any{"type": "content_block_stop", "index": 0}) {
		return
	}
	msgDelta := map[string]any{"type": "message_delta",
		"delta": map[string]any{"stop_reason": stop, "stop_sequence": nil},
		"usage": map[string]int{"output_tokens": usage.OutputTokens}}
	if !send("message_delta", msgDelta) {
		return
	}
	send("message_stop", map[string]string{"type": "message_stop"})
}
