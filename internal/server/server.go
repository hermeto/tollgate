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

	"github.com/hermeto/tollgate/internal/api"
	"github.com/hermeto/tollgate/internal/api/anthropic"
	"github.com/hermeto/tollgate/internal/api/openai"
	"github.com/hermeto/tollgate/internal/config"
	"github.com/hermeto/tollgate/internal/health"
	"github.com/hermeto/tollgate/internal/identity"
	"github.com/hermeto/tollgate/internal/pipeline"
	"github.com/hermeto/tollgate/internal/policy"
	"github.com/hermeto/tollgate/internal/sse"
	"github.com/hermeto/tollgate/internal/telemetry"
	"github.com/hermeto/tollgate/internal/upstream"

	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"
)

type Server struct {
	cfg       config.Config
	api, ops  *http.Server
	client    *upstream.Client
	pipeline  *pipeline.Pipeline
	telemetry *telemetry.Telemetry
	health    *health.Manager
	identity  *identity.Store
	policy    *policy.Loader
	draining  atomic.Bool
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
	t, err := telemetry.New(context.Background(), cfg.Telemetry.ServiceName, cfg.Telemetry.OTLPEndpoint)
	if err != nil {
		return nil, err
	}
	s.telemetry = t
	s.health = health.New(cfg.Dependencies, t.Logger)
	if cfg.Identity.KeyStore != "" {
		s.identity, err = identity.Open(cfg.Identity.KeyStore)
		if err != nil {
			return nil, err
		}
	}
	if cfg.Identity.PolicyDir != "" {
		s.policy, err = policy.NewLoader(cfg.Identity.PolicyDir)
		if err != nil {
			return nil, err
		}
	}
	s.api = &http.Server{Addr: cfg.Listen.API, Handler: http.HandlerFunc(s.handle), ReadHeaderTimeout: 10 * time.Second}
	s.ops = &http.Server{Addr: cfg.Listen.Ops, Handler: http.HandlerFunc(s.handleOps), ReadHeaderTimeout: 5 * time.Second}
	return s, nil
}

func (s *Server) Serve(ctx context.Context) error {
	s.health.Start(ctx)
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
	_ = s.telemetry.Shutdown(context.Background())
	return nil
}
func cleanServeError(err error) error {
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *Server) handle(w http.ResponseWriter, r *http.Request) {
	ctx, span := s.telemetry.Tracer.Start(r.Context(), "tollgate.server")
	defer span.End()
	r = r.WithContext(ctx)
	started := time.Now()
	defer func() {
		s.telemetry.Duration.Record(ctx, time.Since(started).Seconds(), otelmetric.WithAttributes(attribute.String("http.request.method", r.Method), attribute.String("url.path", r.URL.Path)))
		s.telemetry.Requests.Add(ctx, 1)
	}()
	validPost := r.Method == http.MethodPost && (r.URL.Path == "/v1/chat/completions" || r.URL.Path == "/v1/embeddings" || r.URL.Path == "/v1/messages")
	validModels := r.Method == http.MethodGet && r.URL.Path == "/v1/models"
	if !validPost && !validModels {
		writeError(w, 404, "not_found", "unknown endpoint")
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, s.cfg.Limits.MaxBodyBytes))
	if err != nil {
		writeError(w, 413, "request_too_large", "request body too large")
		return
	}
	protocol := pipeline.ProtocolOpenAI
	if r.URL.Path == "/v1/messages" {
		protocol = pipeline.ProtocolAnthropic
	}
	requestedModel := modelName(body)
	prepared, target, prepErr := s.prepare(protocol, r.URL.Path, body)
	if prepErr != nil {
		writeProtocolError(w, protocol, prepErr)
		return
	}
	body = prepared
	if r.Method == http.MethodGet && r.URL.Path == "/v1/models" && len(s.cfg.Models) > 0 {
		s.writeLogicalModels(w)
		return
	}
	req := pipeline.NewRequest(protocol, r.Method, r.URL.Path, r.Header.Clone(), body)
	req.UpstreamBaseURL = target.BaseURL
	if s.identity != nil {
		rawKey := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		id, ok := s.identity.Authenticate(rawKey, time.Now().UTC())
		if !ok {
			writeProtocolError(w, protocol, &api.Error{Status: http.StatusUnauthorized, Type: "authentication_error", Message: "invalid or expired key"})
			return
		}
		req.KeyPrefix, req.Team, req.Project = id.Prefix, id.Team, id.Project
		req.ResponseHeader.Set("X-Tollgate-Team", id.Team)
		if s.policy != nil {
			allowed, dryRun := s.policy.Snapshot().Allows(id.Team, id.Project, requestedModel)
			if !allowed && !dryRun {
				writeProtocolError(w, protocol, &api.Error{Status: http.StatusForbidden, Type: "permission_error", Message: "model is not allowed by policy"})
				return
			}
			if !allowed {
				s.telemetry.Logger.Warn("policy dry-run denial", "team", id.Team, "project", id.Project, "model", requestedModel)
			}
		}
	}
	req.Received = time.Now()
	req.Stream = isStream(body)
	passed, decision := s.pipeline.Request(ctx, req)
	span.SetAttributes(attribute.String("tollgate.request_id", req.ID))
	for stage, duration := range pipeline.StageTime(req) {
		s.telemetry.Stages.Record(ctx, duration.Seconds(), otelmetric.WithAttributes(attribute.String("stage", stage)))
	}
	for _, record := range req.Records {
		_, stageSpan := s.telemetry.Tracer.Start(ctx, "tollgate.stage."+record.Stage)
		stageSpan.SetAttributes(attribute.String("tollgate.stage", record.Stage), attribute.String("tollgate.phase", record.Phase.String()), attribute.String("tollgate.decision", record.Kind.String()), attribute.Int64("tollgate.duration_ns", record.Duration.Nanoseconds()))
		stageSpan.End()
	}
	s.telemetry.Logger.InfoContext(ctx, "request received", "request_id", req.ID, "method", req.Method, "path", req.Path, "body_bytes", len(req.Body))
	if decision.Kind == pipeline.Deny {
		writeError(w, decision.Status, decision.Type, decision.Message)
		return
	}
	if decision.Kind == pipeline.Respond {
		s.writeResponse(w, req, decision.Response)
		return
	}
	upstreamCtx, upstreamSpan := s.telemetry.Tracer.Start(ctx, "tollgate.upstream")
	resp, err := s.client.Forward(upstreamCtx, req)
	upstreamSpan.SetAttributes(attribute.String("tollgate.request_id", req.ID), attribute.String("http.request.method", req.Method), attribute.String("url.path", req.Path))
	upstreamSpan.End()
	s.telemetry.Overhead.Record(ctx, time.Since(req.Received).Seconds(), otelmetric.WithAttributes(attribute.String("phase", telemetry.PhaseRequest)))
	if err != nil {
		writeError(w, 502, "upstream_unavailable", "upstream unavailable")
		return
	}
	defer resp.Body.Close()
	responseStarted := time.Now()
	pResp := &pipeline.Response{Status: resp.StatusCode, Header: resp.Header.Clone(), Stream: req.Stream}
	if !req.Stream {
		pResp.Body, err = io.ReadAll(resp.Body)
		if err != nil {
			writeError(w, 502, "upstream_unavailable", "upstream read failed")
			return
		}
		if d := s.pipeline.Response(ctx, req, pResp, passed); d.Stops() {
			writeDecision(w, d)
			return
		}
		s.telemetry.Overhead.Record(ctx, time.Since(responseStarted).Seconds(), otelmetric.WithAttributes(attribute.String("phase", telemetry.PhaseResponse)))
		s.writeResponse(w, req, pResp)
		return
	}
	if d := s.pipeline.Response(ctx, req, pResp, passed); d.Stops() {
		writeDecision(w, d)
		return
	}
	s.telemetry.Overhead.Record(ctx, time.Since(responseStarted).Seconds(), otelmetric.WithAttributes(attribute.String("phase", telemetry.PhaseResponse)))
	copyResponseHeaders(w.Header(), req.ResponseHeader)
	copyResponseHeaders(w.Header(), pResp.Header)
	w.WriteHeader(resp.StatusCode)
	reader, writer := sse.NewReader(resp.Body), sse.NewWriter(w)
	flusher, _ := w.(http.Flusher)
	for {
		event, readErr := reader.ReadEvent()
		if len(event) > 0 {
			chunkStarted := time.Now()
			chunk := &pipeline.Chunk{Data: event}
			if d := s.pipeline.Chunk(ctx, req, chunk, passed); d.Stops() {
				writeStreamError(writer, flusher)
				return
			}
			if err := writer.WriteEvent(chunk.Data); err != nil {
				return
			}
			if flusher != nil {
				flusher.Flush()
			}
			s.telemetry.Overhead.Record(ctx, time.Since(chunkStarted).Seconds(), otelmetric.WithAttributes(attribute.String("phase", telemetry.PhaseChunk)))
		}
		if readErr != nil {
			if errors.Is(readErr, io.EOF) && len(event) > 0 || !errors.Is(readErr, io.EOF) {
				writeStreamError(writer, flusher)
			}
			return
		}
	}
}

func (s *Server) prepare(protocol pipeline.Protocol, path string, body []byte) ([]byte, config.TargetConfig, *api.Error) {
	if path == "/v1/models" {
		return body, config.TargetConfig{}, nil
	}
	var model string
	switch protocol {
	case pipeline.ProtocolAnthropic:
		value, err := anthropic.DecodeMessages(body)
		if err != nil {
			return nil, config.TargetConfig{}, asAPIError(err)
		}
		model = value.Model
	case pipeline.ProtocolOpenAI:
		if path == "/v1/embeddings" {
			value, err := openai.DecodeEmbeddings(body)
			if err != nil {
				return nil, config.TargetConfig{}, asAPIError(err)
			}
			model = value.Model
		} else {
			value, err := openai.DecodeChat(body)
			if err != nil {
				return nil, config.TargetConfig{}, asAPIError(err)
			}
			model = value.Model
		}
	}
	targets := s.cfg.Models[model]
	if len(targets) == 0 {
		return body, config.TargetConfig{}, nil
	}
	target := targets[0]
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(body, &fields); err != nil {
		return nil, config.TargetConfig{}, api.BadRequest("invalid JSON body")
	}
	realModel, _ := json.Marshal(target.Model)
	fields["model"] = realModel
	rewritten, err := json.Marshal(fields)
	if err != nil {
		return nil, config.TargetConfig{}, api.BadRequest("request could not be encoded")
	}
	return rewritten, target, nil
}

func asAPIError(err error) *api.Error {
	var value *api.Error
	if errors.As(err, &value) {
		return value
	}
	return api.BadRequest(err.Error())
}

func modelName(body []byte) string {
	var value struct {
		Model string `json:"model"`
	}
	_ = json.Unmarshal(body, &value)
	return value.Model
}

func (s *Server) writeLogicalModels(w http.ResponseWriter) {
	data := make([]map[string]any, 0, len(s.cfg.Models))
	for name := range s.cfg.Models {
		data = append(data, map[string]any{"id": name, "object": "model", "owned_by": "tollgate"})
	}
	writeJSON(w, http.StatusOK, map[string]any{"object": "list", "data": data})
}

func writeProtocolError(w http.ResponseWriter, protocol pipeline.Protocol, err *api.Error) {
	if protocol == pipeline.ProtocolAnthropic {
		writeJSON(w, err.Status, map[string]any{"type": "error", "error": map[string]string{"type": err.Type, "message": err.Message}})
		return
	}
	writeError(w, err.Status, err.Type, err.Message)
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
		if s.draining.Load() || !s.health.Ready() {
			writeJSON(w, 503, map[string]any{"status": "not_ready", "dependencies": s.health.Report()})
		} else {
			writeJSON(w, 200, map[string]any{"status": "ready", "dependencies": s.health.Report()})
		}
	case "/metrics":
		telemetry.MetricsHandler().ServeHTTP(w, r)
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
