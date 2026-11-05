package sim

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// shape reduces a JSON value to its structure: object keys, value types and, for arrays, the shape
// of the first element. Fixtures and simulator output must have the same shape (spec SC-002).
func shape(v any) any {
	switch t := v.(type) {
	case map[string]any:
		m := make(map[string]any, len(t))
		for k, val := range t {
			m[k] = shape(val)
		}
		return m
	case []any:
		if len(t) == 0 {
			return []any{}
		}
		return []any{shape(t[0])}
	case string:
		return "string"
	case float64, json.Number:
		return "number"
	case bool:
		return "bool"
	case nil:
		return "null"
	default:
		return "unknown"
	}
}

func jsonShape(t *testing.T, data []byte) any {
	t.Helper()
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		t.Fatalf("invalid JSON %q: %v", data, err)
	}
	return shape(v)
}

// fixture reads testdata/wire/<path>, for example fixture(t, "openai/models.json").
func fixture(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "testdata", "wire", filepath.FromSlash(path)))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// sseFrame is one server-sent event.
type sseFrame struct {
	event string
	data  string
}

func parseSSE(t *testing.T, body []byte) []sseFrame {
	t.Helper()
	var frames []sseFrame
	for _, block := range strings.Split(strings.TrimSpace(string(bytes.ReplaceAll(body, []byte("\r\n"), []byte("\n")))), "\n\n") {
		var f sseFrame
		for _, line := range strings.Split(block, "\n") {
			switch {
			case strings.HasPrefix(line, "event: "):
				f.event = strings.TrimPrefix(line, "event: ")
			case strings.HasPrefix(line, "data: "):
				f.data = strings.TrimPrefix(line, "data: ")
			}
		}
		if f.data != "" {
			frames = append(frames, f)
		}
	}
	return frames
}

// streamShape returns the sequence of frame shapes with consecutive duplicates collapsed, so a
// fixture with two content chunks matches a stream with any number of them.
func streamShape(t *testing.T, frames []sseFrame) []any {
	t.Helper()
	var out []any
	for _, f := range frames {
		var s any = f.data
		if f.data != "[DONE]" {
			s = map[string]any{"event": f.event, "data": jsonShape(t, []byte(f.data))}
		}
		if len(out) > 0 && reflect.DeepEqual(out[len(out)-1], s) {
			continue
		}
		out = append(out, s)
	}
	return out
}

func assertSameShape(t *testing.T, name string, want, got any) {
	t.Helper()
	if !reflect.DeepEqual(want, got) {
		w, _ := json.MarshalIndent(want, "", "  ")
		g, _ := json.MarshalIndent(got, "", "  ")
		t.Fatalf("%s: shape mismatch\nfixture:\n%s\nsimulator:\n%s", name, w, g)
	}
}
