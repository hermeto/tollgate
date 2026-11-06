// Package server contains the gateway HTTP API and its graceful-drain lifecycle.
package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/hermeto/tollgate/internal/config"
	"github.com/hermeto/tollgate/internal/pipeline"
	"github.com/hermeto/tollgate/internal/sse"
	"github.com/hermeto/tollgate/internal/upstream"
)

type Server struct {
	cfg      config.Config
	api, ops *http.Server
	client   *upstream.Client
	pipeline *pipeline.Pipeline
	draining atomic.Bool
}

func New(cfg config.Config) (*Server, error) {
	client, err := upstream.New(cfg.Upstream)
	if err != nil {
		return nil, err
	}
	stages := make([]pipeline.Stage, 0, len(cfg.Pipeline))
	for _, name := range cfg.Pipeline {
		switch name {
		case "request_id":
			stages = append(stages, pipeline.RequestID{})
		case "timing":
			stages = append(stages, pipeline.Timing{})
		}
	}
	p, err := pipeline.New(stages...)
	if err != nil {
		return nil, err
	}
	s := &Server{cfg: cfg, client: client, pipeline: p}
	s.api = &http.Server{Addr: cfg.Listen.API, Handler: http.HandlerFunc(s.handle), ReadHeaderTimeout: 10 * time.Second}
	s.ops = &http.Server{Addr: cfg.Listen.Ops, Handler: http.HandlerFunc(s.handleOps), ReadHeaderTimeout: 5 * time.Second}
	return s, nil
}

func (s *Server) Serve(ctx context.Context) error {
	apiLn, err := (&net.ListenConfig{}).Listen(ctx, "tcp", s.api.Addr)
	if err != nil {
		return err
	}
	opsLn, err := (&net.ListenConfig{}).Listen(ctx, "tcp", s.ops.Addr)
	if err != nil {
		_ = apiLn.Close()
		return err
	}
	apiErr, opsErr := make(chan error, 1), make(chan error, 1)
	go func() { apiErr <- s.api.Serve(apiLn) }()
	go func() { opsErr <- s.ops.Serve(opsLn) }()
	select {
	case err := <-apiErr:
		_ = opsLn.Close()
		return cleanServeError(err)
	case err := <-opsErr:
		_ = apiLn.Close()
		return cleanServeError(err)
	case <-ctx.Done():
	}
	s.draining.Store(true)
	drain, cancel := context.WithTimeout(context.Background(), time.Duration(s.cfg.Shutdown.DrainTimeout))
	defer cancel()
	_ = s.ops.Shutdown(context.Background())
	if err := s.api.Shutdown(drain); err != nil {
		_ = s.api.Close()
		return err
	}
	return nil
}
func cleanServeError(err error) error {
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *Server) handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || r.URL.Path != "/v1/chat/completions" {
		writeError(w, 404, "not_found", "unknown endpoint")
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, s.cfg.Limits.MaxBodyBytes))
	if err != nil {
		writeError(w, 413, "request_too_large", "request body too large")
		return
	}
	req := pipeline.NewRequest(pipeline.ProtocolOpenAI, r.Method, r.URL.Path, r.Header.Clone(), body)
	req.Received = time.Now()
	req.Stream = isStream(body)
	passed, decision := s.pipeline.Request(r.Context(), req)
	if decision.Kind == pipeline.Deny {
		writeError(w, decision.Status, decision.Type, decision.Message)
		return
	}
	if decision.Kind == pipeline.Respond {
		s.writeResponse(w, req, decision.Response)
		return
	}
	resp, err := s.client.Forward(r.Context(), req)
	if err != nil {
		writeError(w, 502, "upstream_unavailable", "upstream unavailable")
		return
	}
	defer resp.Body.Close()
	pResp := &pipeline.Response{Status: resp.StatusCode, Header: resp.Header.Clone(), Stream: req.Stream}
	if !req.Stream {
		pResp.Body, err = io.ReadAll(resp.Body)
		if err != nil {
			writeError(w, 502, "upstream_unavailable", "upstream read failed")
			return
		}
		if d := s.pipeline.Response(r.Context(), req, pResp, passed); d.Stops() {
			writeDecision(w, d)
			return
		}
		s.writeResponse(w, req, pResp)
		return
	}
	if d := s.pipeline.Response(r.Context(), req, pResp, passed); d.Stops() {
		writeDecision(w, d)
		return
	}
	copyResponseHeaders(w.Header(), req.ResponseHeader)
	copyResponseHeaders(w.Header(), pResp.Header)
	w.WriteHeader(resp.StatusCode)
	reader, writer := sse.NewReader(resp.Body), sse.NewWriter(w)
	flusher, _ := w.(http.Flusher)
	for {
		event, readErr := reader.ReadEvent()
		if len(event) > 0 {
			chunk := &pipeline.Chunk{Data: event}
			if d := s.pipeline.Chunk(r.Context(), req, chunk, passed); d.Stops() {
				writeStreamError(writer, flusher)
				return
			}
			if err := writer.WriteEvent(chunk.Data); err != nil {
				return
			}
			if flusher != nil {
				flusher.Flush()
			}
		}
		if readErr != nil {
			if errors.Is(readErr, io.EOF) && len(event) > 0 || !errors.Is(readErr, io.EOF) {
				writeStreamError(writer, flusher)
			}
			return
		}
	}
}

func (s *Server) writeResponse(w http.ResponseWriter, req *pipeline.Request, resp *pipeline.Response) {
	copyResponseHeaders(w.Header(), req.ResponseHeader)
	copyResponseHeaders(w.Header(), resp.Header)
	w.WriteHeader(resp.Status)
	_, _ = w.Write(resp.Body)
}
func (s *Server) handleOps(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/healthz":
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	case "/readyz":
		if s.draining.Load() {
			writeJSON(w, 503, map[string]string{"status": "not_ready"})
		} else {
			writeJSON(w, 200, map[string]string{"status": "ready"})
		}
	case "/metrics":
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		w.WriteHeader(200)
	default:
		http.NotFound(w, r)
	}
}
func isStream(body []byte) bool {
	var v struct {
		Stream bool `json:"stream"`
	}
	return json.Unmarshal(body, &v) == nil && v.Stream
}
func writeDecision(w http.ResponseWriter, d pipeline.Decision) {
	writeError(w, d.Status, d.Type, d.Message)
}
func writeError(w http.ResponseWriter, status int, typ, message string) {
	if status == 0 {
		status = 500
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]string{"message": message, "type": typ, "code": typ}})
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func copyResponseHeaders(dst, src http.Header) {
	for k, values := range src {
		if strings.EqualFold(k, "Content-Length") || strings.EqualFold(k, "Transfer-Encoding") {
			continue
		}
		for _, v := range values {
			dst.Add(k, v)
		}
	}
}
func writeStreamError(w *sse.Writer, f http.Flusher) {
	_ = w.WriteEvent([]byte(`data: {"error":{"message":"upstream closed the stream","type":"server_error","code":"upstream_stream_error"}}`))
	if f != nil {
		f.Flush()
	}
}
