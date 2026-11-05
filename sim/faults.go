package sim

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sort"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/hermeto/tollgate/schemas"
)

// faults holds the start-up fault configuration and the live one (swapped atomically, so the
// request path never takes a lock).
type faults struct {
	startup *FaultConfig
	live    atomic.Pointer[FaultConfig]
}

func newFaults(cfg FaultConfig, now time.Time) *faults {
	resolved := resolveOutage(cfg, now)
	f := &faults{startup: &resolved}
	f.live.Store(&resolved)
	return f
}

// resolveOutage turns a relative outage window (start_after, duration) into absolute times.
func resolveOutage(cfg FaultConfig, now time.Time) FaultConfig {
	if cfg.Outage == nil {
		return cfg
	}
	o := *cfg.Outage
	if o.Start == nil || o.End == nil {
		start := now.Add(time.Duration(o.StartAfter)).UTC()
		end := start.Add(time.Duration(o.Duration))
		o.Start, o.End = &start, &end
	}
	if o.Status == 0 {
		o.Status = http.StatusServiceUnavailable
	}
	cfg.Outage = &o
	return cfg
}

// faultMessages are the error messages of injected failures.
var faultMessages = map[int]string{
	http.StatusTooManyRequests:     "Rate limit reached (simulated by tollgate sim).",
	http.StatusInternalServerError: "The server had an error while processing your request (simulated by tollgate sim).",
	http.StatusServiceUnavailable:  "The service is temporarily unavailable (simulated by tollgate sim).",
	529:                            "Overloaded (simulated by tollgate sim).",
}

// applyFaults decides this request's fate with its own deterministic random stream. It returns the
// response pace, or false when it already answered (error, outage) or dropped the connection.
// Draws happen in a fixed order so the same request sequence always gets the same faults.
func (s *Server) applyFaults(w http.ResponseWriter, r *http.Request, h uint64) (*pace, bool) {
	cfg := s.faults.live.Load()
	rng := s.seeds.perRequest(h)
	p := &pace{ctx: r.Context(), rng: rng, ttft: s.cfg.Latency.TTFT, itl: s.cfg.Latency.InterToken, stallAt: -1, at: arrival(r)}

	if o := cfg.Outage; o != nil && o.Start != nil && o.End != nil {
		if now := time.Now(); !now.Before(*o.Start) && now.Before(*o.End) {
			s.faultResponse(w, o.Status, cfg.RetryAfterSeconds)
			return nil, false
		}
	}
	u := rng.Float64()
	statuses := make([]int, 0, len(cfg.ErrorRates))
	for st := range cfg.ErrorRates {
		statuses = append(statuses, st)
	}
	sort.Ints(statuses)
	acc := 0.0
	for _, st := range statuses {
		acc += cfg.ErrorRates[st]
		if u < acc {
			s.faultResponse(w, st, cfg.RetryAfterSeconds)
			return nil, false
		}
	}
	if u < acc+cfg.ResetRate {
		resetConnection(w)
		return nil, false
	}
	if t := cfg.TimeoutBeforeFirstToken; t != nil && rng.Float64() < t.Rate {
		p.hold = time.Duration(t.Duration)
	}
	if st := cfg.StallMidStream; st != nil && rng.Float64() < st.Rate {
		p.stallAt, p.stallFor = st.AfterTokens, time.Duration(st.Duration)
	}
	return p, true
}

func (s *Server) faultResponse(w http.ResponseWriter, status, retryAfter int) {
	if retryAfter > 0 && (status == http.StatusTooManyRequests || status == 529 || status == http.StatusServiceUnavailable) {
		w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
	}
	kind := errServer
	switch status {
	case http.StatusTooManyRequests:
		kind = errRateLimit
	case http.StatusServiceUnavailable:
		kind = errUnavailable
	case 529:
		kind = errOverloaded
	}
	s.writeError(w, status, kind, faultMessages[status])
}

// resetConnection closes the client connection with a TCP RST and no HTTP response.
func resetConnection(w http.ResponseWriter) {
	conn, _, err := http.NewResponseController(w).Hijack()
	if err != nil {
		panic(http.ErrAbortHandler) // fallback: net/http aborts the connection
	}
	if tcp, ok := conn.(*net.TCPConn); ok {
		_ = tcp.SetLinger(0)
	}
	_ = conn.Close()
}

// Admin API: /_sim/faults (contracts/admin-api.md).

func (s *Server) handleFaultsGet(w http.ResponseWriter, _ *http.Request) {
	s.writeJSON(w, http.StatusOK, s.faults.live.Load())
}

func (s *Server) handleFaultsPut(w http.ResponseWriter, r *http.Request) {
	s.setFaults(w, r, false)
}

func (s *Server) handleFaultsPatch(w http.ResponseWriter, r *http.Request) {
	s.setFaults(w, r, true)
}

func (s *Server) handleFaultsDelete(w http.ResponseWriter, _ *http.Request) {
	s.faults.live.Store(s.faults.startup)
	s.writeJSON(w, http.StatusOK, s.faults.startup)
}

func (s *Server) setFaults(w http.ResponseWriter, r *http.Request, merge bool) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		s.writeAdminError(w, err)
		return
	}
	var doc any
	if err := json.Unmarshal(body, &doc); err != nil {
		s.writeAdminError(w, fmt.Errorf("invalid JSON: %w", err))
		return
	}
	relative := outageIsRelative(doc)
	if merge {
		cur, _ := json.Marshal(s.faults.live.Load())
		var base any
		_ = json.Unmarshal(cur, &base)
		doc = mergePatch(base, doc)
	}
	if err := schemas.Validate("sim.schema.json", map[string]any{"faults": doc}); err != nil {
		s.writeAdminError(w, err)
		return
	}
	data, _ := json.Marshal(doc)
	var fc FaultConfig
	if err := json.Unmarshal(data, &fc); err != nil {
		s.writeAdminError(w, err)
		return
	}
	cfg := s.cfg
	cfg.Faults = fc
	if err := cfg.Validate(); err != nil {
		s.writeAdminError(w, err)
		return
	}
	if fc.Outage != nil && relative {
		fc.Outage.Start, fc.Outage.End = nil, nil // a new relative window starts now
	}
	resolved := resolveOutage(fc, time.Now())
	s.faults.live.Store(&resolved)
	s.writeJSON(w, http.StatusOK, &resolved)
}

func (s *Server) writeAdminError(w http.ResponseWriter, err error) {
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		s.writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "body too large"})
		return
	}
	s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
}

// outageIsRelative reports whether the request body sets an outage window without absolute times.
func outageIsRelative(doc any) bool {
	m, _ := doc.(map[string]any)
	o, ok := m["outage"].(map[string]any)
	if !ok {
		return false
	}
	_, hasStart := o["start"]
	return !hasStart
}

// mergePatch applies an RFC 7396 JSON merge patch.
func mergePatch(target, patch any) any {
	p, ok := patch.(map[string]any)
	if !ok {
		return patch
	}
	t, ok := target.(map[string]any)
	if !ok {
		t = map[string]any{}
	}
	for k, v := range p {
		if v == nil {
			delete(t, k)
			continue
		}
		t[k] = mergePatch(t[k], v)
	}
	return t
}
