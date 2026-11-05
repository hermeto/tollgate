package sim

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestLogRingBuffer(t *testing.T) {
	l := newRequestLog(3)
	for _, p := range []string{"/a", "/b", "/c", "/d", "/e"} {
		l.add(RequestRecord{Path: p})
	}
	got := l.list()
	if len(got) != 3 || got[0].Path != "/c" || got[2].Path != "/e" {
		t.Fatalf("list = %+v, want /c /d /e", got)
	}
	l.clear()
	if len(l.list()) != 0 {
		t.Fatal("clear left records")
	}
	l.add(RequestRecord{Path: "/f"})
	if got := l.list(); len(got) != 1 || got[0].Path != "/f" {
		t.Fatalf("after clear: %+v", got)
	}
	disabled := newRequestLog(0)
	if disabled != nil || len(disabled.list()) != 0 {
		t.Fatal("size 0 must disable the log")
	}
	disabled.clear() // must not panic
}

func readLog(t *testing.T, url string) []RequestRecord {
	t.Helper()
	_, body := doReq(t, http.MethodGet, url+"/_sim/requests", "", nil)
	var out struct {
		Requests []RequestRecord `json:"requests"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatal(err)
	}
	return out.Requests
}

func TestRequestLogEndpoint(t *testing.T) {
	ts := newOpenAI(t, 42)
	headers := map[string]string{"Authorization": "Bearer tg-example-key", "X-Team": "payments"}
	doReq(t, http.MethodPost, ts.URL+"/v1/chat/completions", chatBody, headers)
	doReq(t, http.MethodGet, ts.URL+"/v1/models", "", headers)
	doReq(t, http.MethodGet, ts.URL+"/_sim/healthz", "", nil) // admin: not recorded

	recs := readLog(t, ts.URL)
	if len(recs) != 2 {
		t.Fatalf("records = %d, want 2", len(recs))
	}
	first := recs[0]
	if first.Method != http.MethodPost || first.Path != "/v1/chat/completions" || first.Status != http.StatusOK {
		t.Fatalf("first = %+v", first)
	}
	if first.Body != chatBody {
		t.Fatalf("body = %q", first.Body)
	}
	if first.Headers["X-Team"][0] != "payments" || first.Headers["Authorization"][0] != "Bearer tg-example-key" {
		t.Fatalf("headers = %v", first.Headers)
	}
	if recs[1].Path != "/v1/models" {
		t.Fatalf("second = %+v", recs[1])
	}

	if status, _ := doReq(t, http.MethodDelete, ts.URL+"/_sim/requests", "", nil); status != http.StatusNoContent {
		t.Fatalf("DELETE status = %d", status)
	}
	if n := len(readLog(t, ts.URL)); n != 0 {
		t.Fatalf("after clear: %d records", n)
	}
}

func TestRequestLogRecordsErrorsAndTruncatesBodies(t *testing.T) {
	ts := newOpenAI(t, 42)
	doReq(t, http.MethodPost, ts.URL+"/v1/chat/completions", chatBody, nil) // 401
	big := `{"model":"gpt-4o-mini","messages":[{"role":"user","content":"` + strings.Repeat("x", 2*maxRecordedBody) + `"}]}`
	doReq(t, http.MethodPost, ts.URL+"/v1/chat/completions", big, map[string]string{"Authorization": "Bearer k"})
	recs := readLog(t, ts.URL)
	if len(recs) != 2 || recs[0].Status != http.StatusUnauthorized {
		t.Fatalf("records = %+v", recs)
	}
	if !recs[1].BodyTruncated || len(recs[1].Body) != maxRecordedBody || recs[1].Status != http.StatusOK {
		t.Fatalf("large body: truncated=%v len=%d status=%d", recs[1].BodyTruncated, len(recs[1].Body), recs[1].Status)
	}
}

func TestRequestLogDisabled(t *testing.T) {
	cfg := DefaultConfig(FlavorOpenAI)
	cfg.RequestLog.Size = 0
	s, err := New(cfg, Options{})
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(s.Handler())
	t.Cleanup(ts.Close)
	doReq(t, http.MethodPost, ts.URL+"/v1/chat/completions", chatBody, map[string]string{"Authorization": "Bearer k"})
	if n := len(readLog(t, ts.URL)); n != 0 {
		t.Fatalf("disabled log returned %d records", n)
	}
}
