// Package usage defines content-free accounting events and a bounded async writer.
package usage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type Event struct {
	RequestID         string            `json:"request_id"`
	Timestamp         time.Time         `json:"timestamp"`
	KeyPrefix         string            `json:"key_prefix,omitempty"`
	Team              string            `json:"team,omitempty"`
	Project           string            `json:"project,omitempty"`
	CostLabels        map[string]string `json:"cost_labels,omitempty"`
	LogicalModel      string            `json:"logical_model"`
	RealModel         string            `json:"real_model"`
	Provider          string            `json:"provider"`
	InputTokens       int               `json:"input_tokens"`
	OutputTokens      int               `json:"output_tokens"`
	CachedInputTokens int               `json:"cached_input_tokens"`
	Estimated         bool              `json:"estimated"`
	CostUSD           float64           `json:"cost_usd"`
	LatencyMS         float64           `json:"latency_ms"`
	Status            int               `json:"status"`
	CacheStatus       string            `json:"cache_status,omitempty"`
	PolicyDecisions   []string          `json:"policy_decisions,omitempty"`
}
type Sink interface{ Write([]Event) error }
type SpoolSink struct {
	mu   sync.Mutex
	path string
}

func NewSpoolSink(path string) *SpoolSink { return &SpoolSink{path: path} }
func (s *SpoolSink) Write(events []Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	f, err := os.OpenFile(s.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, e := range events {
		if err := enc.Encode(e); err != nil {
			return err
		}
	}
	return nil
}

type Writer struct {
	ch    chan Event
	sink  Sink
	done  chan struct{}
	wg    sync.WaitGroup
	errMu sync.Mutex
	err   error
}

func NewWriter(sink Sink, capacity int) *Writer {
	if capacity < 1 {
		capacity = 1
	}
	w := &Writer{ch: make(chan Event, capacity), sink: sink, done: make(chan struct{})}
	w.wg.Add(1)
	go w.run()
	return w
}
func (w *Writer) Submit(e Event) bool {
	select {
	case w.ch <- e:
		return true
	case <-w.done:
		return false
	}
}
func (w *Writer) Close() error {
	close(w.done)
	w.wg.Wait()
	w.errMu.Lock()
	defer w.errMu.Unlock()
	return w.err
}
func (w *Writer) run() {
	defer w.wg.Done()
	batch := make([]Event, 0, 32)
	flush := func() {
		if len(batch) == 0 {
			return
		}
		if err := w.sink.Write(batch); err != nil {
			w.errMu.Lock()
			w.err = fmt.Errorf("write usage batch: %w", err)
			w.errMu.Unlock()
		}
		batch = batch[:0]
	}
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case e := <-w.ch:
			batch = append(batch, e)
			if len(batch) >= 32 {
				flush()
			}
		case <-ticker.C:
			flush()
		case <-w.done:
			for {
				select {
				case e := <-w.ch:
					batch = append(batch, e)
				default:
					flush()
					return
				}
			}
		}
	}
}
func Replay(path string) ([]Event, error) {
	f, err := os.Open(path) //nolint:gosec // replay path is explicitly selected by the operator.
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []Event
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var e Event
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, scanner.Err()
}
