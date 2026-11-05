package sim

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"
)

func benchServer(b *testing.B, mutate func(*Config)) *httptest.Server {
	b.Helper()
	cfg := DefaultConfig(FlavorOpenAI)
	mutate(&cfg)
	s, err := New(cfg, Options{})
	if err != nil {
		b.Fatal(err)
	}
	ts := httptest.NewServer(s.Handler())
	b.Cleanup(ts.Close)
	return ts
}

// BenchmarkOpenAIChatNonStreaming measures simulator throughput with zero latency (spec SC-003).
// Client and server share the process, so the req/s figure is a lower bound.
func BenchmarkOpenAIChatNonStreaming(b *testing.B) {
	ts := benchServer(b, func(c *Config) { c.OutputTokens = Distribution{Kind: "fixed", Value: 20} })
	client := &http.Client{Transport: &http.Transport{MaxIdleConnsPerHost: 1024}}
	body := []byte(chatBody)
	url := ts.URL + "/v1/chat/completions"
	b.ReportAllocs()
	b.ResetTimer()
	start := time.Now()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequestWithContext(b.Context(), http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set("Authorization", "Bearer tg-example-bench")
			resp, err := client.Do(req)
			if err != nil {
				b.Error(err)
				return
			}
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				b.Errorf("status %d", resp.StatusCode)
				return
			}
		}
	})
	b.ReportMetric(float64(b.N)/time.Since(start).Seconds(), "req/s")
}

// BenchmarkTTFTUnder1000Streams measures how far observed time to first token drifts above the
// configured 100 ms when 1,000 streams start at once (spec SC-007, target: p99 ≤ +5 ms). macOS
// listen backlogs are small (somaxconn 128), so a warm-up round holds every request open with the
// simulator's own timeout fault while connections are dialled slowly; the measured rounds then reuse
// those 1,000 pooled connections and time the simulator, not TCP setup. Client and server share the
// CPU here, so the figures are pessimistic; the external measurement with a separate client process
// (bench/k6/sim-ttft.js, docs/simulator.md) is the reference.
func BenchmarkTTFTUnder1000Streams(b *testing.B) {
	const streams = 1000
	ts := benchServer(b, func(c *Config) {
		c.Latency.TTFT = Distribution{Kind: "fixed", Value: 100}
		c.OutputTokens = Distribution{Kind: "fixed", Value: 5}
		c.RequestLog.Size = 0
	})
	dials := make(chan struct{}, 8)
	dialer := &net.Dialer{}
	client := &http.Client{Transport: &http.Transport{
		MaxIdleConnsPerHost: streams,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			dials <- struct{}{}
			defer func() { time.Sleep(time.Millisecond); <-dials }()
			return dialer.DialContext(ctx, network, addr)
		},
	}}
	setFaults := func(method, body string) {
		req, _ := http.NewRequestWithContext(b.Context(), method, ts.URL+"/_sim/faults", strings.NewReader(body))
		resp, err := client.Do(req)
		if err != nil {
			b.Fatal(err)
		}
		_ = resp.Body.Close()
	}
	body := strings.Replace(chatBody, `{"model"`, `{"stream":true,"model"`, 1)
	round := func() []time.Duration {
		var mu sync.Mutex
		var wg sync.WaitGroup
		out := make([]time.Duration, 0, streams)
		for range streams {
			wg.Go(func() {
				req, _ := http.NewRequestWithContext(b.Context(), http.MethodPost, ts.URL+"/v1/chat/completions", strings.NewReader(body))
				req.Header.Set("Authorization", "Bearer tg-example-bench")
				start := time.Now()
				resp, err := client.Do(req)
				if err != nil {
					b.Error(err)
					return
				}
				defer resp.Body.Close()
				sc := bufio.NewScanner(resp.Body)
				first := time.Duration(0)
				for sc.Scan() {
					if first == 0 && strings.HasPrefix(sc.Text(), "data: ") {
						first = time.Since(start)
					}
				}
				mu.Lock()
				out = append(out, first-100*time.Millisecond)
				mu.Unlock()
			})
		}
		wg.Wait()
		return out
	}
	setFaults(http.MethodPut, `{"timeout_before_first_token":{"rate":1,"duration":"3s"}}`)
	round() // warm-up: all requests held open while 1,000 connections are dialled
	setFaults(http.MethodDelete, "")
	var overshoots []time.Duration
	b.ResetTimer()
	for range b.N {
		overshoots = append(overshoots, round()...)
	}
	b.StopTimer()
	if len(overshoots) == 0 {
		b.Fatal("no samples")
	}
	slices.Sort(overshoots)
	pct := func(p float64) float64 {
		return float64(overshoots[int(p*float64(len(overshoots)-1))]) / float64(time.Millisecond)
	}
	b.ReportMetric(pct(0.50), "p50-overshoot-ms")
	b.ReportMetric(pct(0.99), "p99-overshoot-ms")
}
