package sim

import (
	"bufio"
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// maxBodyBytes bounds request bodies; larger requests get a provider-shaped 413.
const maxBodyBytes = 32 << 20

// shutdownGrace is how long in-flight requests may run after SIGINT/SIGTERM (spec FR-020).
const shutdownGrace = 5 * time.Second

// Options are runtime settings that are not part of the simulator configuration file.
type Options struct {
	Logger *slog.Logger
	// LogRequests logs method, path, status and duration per request. Never bodies or headers.
	LogRequests bool
}

// Server is one simulator instance speaking a single flavor.
type Server struct {
	cfg    Config
	opts   Options
	log    *slog.Logger
	seeds  *seeds
	models map[string]bool
	mux    *http.ServeMux
	reqlog *requestLog
	faults *faults
}

// New validates the configuration and builds a simulator.
func New(cfg Config, opts Options) (*Server, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	logger := opts.Logger
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	s := &Server{
		cfg:    cfg,
		opts:   opts,
		log:    logger,
		seeds:  newSeeds(cfg.Seed),
		models: make(map[string]bool, len(cfg.Models)),
		mux:    http.NewServeMux(),
		reqlog: newRequestLog(cfg.RequestLog.Size),
		faults: newFaults(cfg.Faults, time.Now()),
	}
	for _, m := range cfg.Models {
		s.models[m] = true
	}
	s.mux.HandleFunc("GET /_sim/healthz", s.handleHealth)
	s.mux.HandleFunc("GET /_sim/requests", s.handleRequestsList)
	s.mux.HandleFunc("DELETE /_sim/requests", s.handleRequestsClear)
	s.mux.HandleFunc("GET /_sim/faults", s.handleFaultsGet)
	s.mux.HandleFunc("PUT /_sim/faults", s.handleFaultsPut)
	s.mux.HandleFunc("PATCH /_sim/faults", s.handleFaultsPatch)
	s.mux.HandleFunc("DELETE /_sim/faults", s.handleFaultsDelete)
	switch cfg.Flavor {
	case FlavorOpenAI:
		s.registerOpenAI()
	case FlavorAnthropic:
		s.registerAnthropic()
	}
	s.mux.HandleFunc("/", s.handleUnknown)
	return s, nil
}

// Config returns the configuration the server runs with.
func (s *Server) Config() Config { return s.cfg }

// Handler returns the HTTP handler: the request log recorder, wrapped with optional request logging.
func (s *Server) Handler() http.Handler {
	h := s.record(s.mux)
	if !s.opts.LogRequests {
		return h
	}
	return s.logRequests(h)
}

// ListenAndServe listens on the configured address until ctx is cancelled.
func (s *Server) ListenAndServe(ctx context.Context) error {
	addr := net.JoinHostPort(s.cfg.Host, strconv.Itoa(s.cfg.Port))
	var lc net.ListenConfig
	ln, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	return s.Serve(ctx, ln)
}

// Serve serves on ln until ctx is cancelled, then drains in-flight requests for up to 5 seconds.
func (s *Server) Serve(ctx context.Context, ln net.Listener) error {
	srv := &http.Server{
		Handler:           s.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	errc := make(chan error, 1)
	go func() { errc <- srv.Serve(ln) }()
	s.log.Info("simulator listening", "flavor", s.cfg.Flavor, "addr", ln.Addr().String(), "seed", s.cfg.Seed)

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownGrace)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		s.log.Warn("grace period ended with requests in flight; closing", "error", err)
		_ = srv.Close()
	}
	if err := <-errc; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	s.log.Info("simulator stopped", "flavor", s.cfg.Flavor)
	return nil
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "flavor": s.cfg.Flavor, "models": s.cfg.Models})
}

func (s *Server) handleUnknown(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/v1/") && !s.authorized(r) {
		s.writeError(w, http.StatusUnauthorized, errAuthentication, s.missingAuthMessage())
		return
	}
	s.writeError(w, http.StatusNotFound, errNotFound, "Invalid URL ("+r.Method+" "+r.URL.Path+")")
}

// authorized reports whether the provider's auth header is present (any value is accepted, FR-016).
func (s *Server) authorized(r *http.Request) bool {
	bearer := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
	hasBearer := strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") && bearer != ""
	if s.cfg.Flavor == FlavorAnthropic {
		return r.Header.Get("X-Api-Key") != "" || hasBearer
	}
	return hasBearer
}

func (s *Server) missingAuthMessage() string {
	if s.cfg.Flavor == FlavorAnthropic {
		return "x-api-key header is required"
	}
	return "You didn't provide an API key. Provide it in the Authorization header (Bearer auth)."
}

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.authorized(r) {
			s.writeError(w, http.StatusUnauthorized, errAuthentication, s.missingAuthMessage())
			return
		}
		next(w, r)
	}
}

// readBody reads a bounded request body, writing a provider-shaped error and returning false on failure.
func (s *Server) readBody(w http.ResponseWriter, r *http.Request) ([]byte, bool) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			s.writeError(w, http.StatusRequestEntityTooLarge, errTooLarge, "request body too large")
			return nil, false
		}
		s.writeError(w, http.StatusBadRequest, errInvalidRequest, "could not read request body")
		return nil, false
	}
	return body, true
}

// statusRecorder captures the status code for request logging and keeps Flush/Hijack reachable.
type statusRecorder struct {
	http.ResponseWriter
	status   int
	hijacked bool
}

// Hijack records that the connection was taken over (an injected reset) and delegates.
func (r *statusRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	r.hijacked = true
	return http.NewResponseController(r.ResponseWriter).Hijack()
}

func (r *statusRecorder) WriteHeader(code int) {
	if r.status == 0 {
		r.status = code
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Unwrap() http.ResponseWriter { return r.ResponseWriter }

func (s *Server) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(rec, r)
		status := rec.status
		if status == 0 {
			status = http.StatusOK
		}
		s.log.Info("request", "method", r.Method, "path", r.URL.Path, "status", status,
			"duration_ms", float64(time.Since(start).Microseconds())/1000)
	})
}
