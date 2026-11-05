package sim

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// maxRecordedBody caps each recorded body so the log stays bounded (size × 64 KiB at most).
const maxRecordedBody = 64 << 10

// RequestRecord is one entry of the request log (contracts/admin-api.md).
type RequestRecord struct {
	Time          time.Time           `json:"time"`
	Method        string              `json:"method"`
	Path          string              `json:"path"`
	Headers       map[string][]string `json:"headers"`
	Body          string              `json:"body"`
	BodyTruncated bool                `json:"body_truncated,omitempty"`
	Status        int                 `json:"status"`
}

// requestLog is a fixed-size ring buffer of the most recent provider requests.
type requestLog struct {
	mu   sync.Mutex
	buf  []RequestRecord
	next int
	full bool
}

func newRequestLog(size int) *requestLog {
	if size <= 0 {
		return nil
	}
	return &requestLog{buf: make([]RequestRecord, size)}
}

func (l *requestLog) add(r RequestRecord) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.buf[l.next] = r
	l.next = (l.next + 1) % len(l.buf)
	if l.next == 0 {
		l.full = true
	}
}

// list returns the records oldest first, newest last.
func (l *requestLog) list() []RequestRecord {
	if l == nil {
		return []RequestRecord{}
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.full {
		return append([]RequestRecord(nil), l.buf[:l.next]...)
	}
	out := make([]RequestRecord, 0, len(l.buf))
	out = append(out, l.buf[l.next:]...)
	return append(out, l.buf[:l.next]...)
}

func (l *requestLog) clear() {
	if l == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	clear(l.buf)
	l.next, l.full = 0, false
}

// record wraps provider endpoints so each request is added to the log after it completes.
// Admin endpoints (/_sim/*) are not recorded.
func (s *Server) record(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.reqlog == nil || strings.HasPrefix(r.URL.Path, "/_sim/") {
			next.ServeHTTP(w, r)
			return
		}
		head, err := io.ReadAll(io.LimitReader(r.Body, maxRecordedBody+1))
		if err != nil {
			head = nil
		}
		// The handler still reads the full body (and enforces its own size limit).
		r.Body = struct {
			io.Reader
			io.Closer
		}{io.MultiReader(bytes.NewReader(head), r.Body), r.Body}
		rec := RequestRecord{Time: time.Now().UTC(), Method: r.Method, Path: r.URL.Path, Headers: r.Header.Clone()}
		if len(head) > maxRecordedBody {
			head, rec.BodyTruncated = head[:maxRecordedBody], true
		}
		rec.Body = string(head)
		sr := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(sr, r)
		rec.Status = sr.status
		if rec.Status == 0 {
			rec.Status = http.StatusOK
		}
		s.reqlog.add(rec)
	})
}

func (s *Server) handleRequestsList(w http.ResponseWriter, _ *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]any{"requests": s.reqlog.list()})
}

func (s *Server) handleRequestsClear(w http.ResponseWriter, _ *http.Request) {
	s.reqlog.clear()
	w.WriteHeader(http.StatusNoContent)
}
