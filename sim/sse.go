package sim

import (
	"fmt"
	"net/http"
)

// sseWriter writes server-sent events and flushes after each one so streams are never buffered.
type sseWriter struct {
	w  http.ResponseWriter
	rc *http.ResponseController
}

// startSSE sends the event-stream headers and a 200 status.
func startSSE(w http.ResponseWriter) *sseWriter {
	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	s := &sseWriter{w: w, rc: http.NewResponseController(w)}
	_ = s.rc.Flush()
	return s
}

// event writes one event. An empty name writes a data-only frame (OpenAI style).
func (s *sseWriter) event(name string, data []byte) error {
	if name != "" {
		if _, err := fmt.Fprintf(s.w, "event: %s\n", name); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(s.w, "data: %s\n\n", data); err != nil {
		return err
	}
	return s.rc.Flush()
}
