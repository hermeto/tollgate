package sim

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"strconv"
)

// errKind is a provider-neutral error category mapped to each provider's error body.
type errKind int

const (
	errInvalidRequest errKind = iota
	errAuthentication
	errNotFound
	errModelNotFound
	errTooLarge
	errRateLimit
	errServer
	errUnavailable
	errOverloaded
)

type openAIError struct {
	Error openAIErrorBody `json:"error"`
}

type openAIErrorBody struct {
	Message string  `json:"message"`
	Type    string  `json:"type"`
	Param   *string `json:"param"`
	Code    *string `json:"code"`
}

type anthropicError struct {
	Type      string             `json:"type"`
	Error     anthropicErrorBody `json:"error"`
	RequestID string             `json:"request_id"`
}

type anthropicErrorBody struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func strp(s string) *string { return &s }

// writeError writes an error in the simulator flavor's wire format (spec FR-004).
func (s *Server) writeError(w http.ResponseWriter, status int, kind errKind, msg string) {
	if s.cfg.Flavor == FlavorAnthropic {
		s.writeJSON(w, status, s.anthropicErrorBody(kind, msg))
		return
	}
	s.writeJSON(w, status, openAIErrorFor(kind, msg))
}

func openAIErrorFor(kind errKind, msg string) openAIError {
	body := openAIErrorBody{Message: msg, Type: "invalid_request_error"}
	switch kind {
	case errAuthentication:
		body.Code = strp("invalid_api_key")
	case errModelNotFound:
		body.Code = strp("model_not_found")
		body.Param = strp("model")
	case errRateLimit:
		body.Type, body.Code = "requests", strp("rate_limit_exceeded")
	case errServer:
		body.Type = "server_error"
	case errUnavailable, errOverloaded:
		body.Type, body.Code = "server_error", strp("service_unavailable")
	}
	return openAIError{Error: body}
}

func (s *Server) anthropicErrorBody(kind errKind, msg string) anthropicError {
	typ := "invalid_request_error"
	switch kind {
	case errAuthentication:
		typ = "authentication_error"
	case errNotFound, errModelNotFound:
		typ = "not_found_error"
	case errTooLarge:
		typ = "request_too_large"
	case errRateLimit:
		typ = "rate_limit_error"
	case errServer, errUnavailable:
		typ = "api_error"
	case errOverloaded:
		typ = "overloaded_error"
	}
	return anthropicError{Type: "error", Error: anthropicErrorBody{Type: typ, Message: msg}, RequestID: newID("req_", 24)}
}

// writeJSON writes v as a JSON response with the given status.
func (s *Server) writeJSON(w http.ResponseWriter, status int, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		http.Error(w, "encode response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(status)
	_, _ = w.Write(data)
}

const idAlphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// newID returns a provider-style identifier. Ids are unique, not deterministic (SC-004 excludes them).
func newID(prefix string, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = idAlphabet[rand.IntN(len(idAlphabet))] // #nosec G404 -- identifiers, not secrets
	}
	return fmt.Sprintf("%s%s", prefix, b)
}
