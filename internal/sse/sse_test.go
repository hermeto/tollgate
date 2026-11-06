package sse

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestReaderSupportsLFAndCRLF(t *testing.T) {
	r := NewReader(strings.NewReader("data: one\n\n:event\r\ndata: two\r\n\r\n"))
	one, err := r.ReadEvent()
	if err != nil {
		t.Fatal(err)
	}
	if string(one) != "data: one\n" {
		t.Fatalf("first event %q", one)
	}
	two, err := r.ReadEvent()
	if err != nil {
		t.Fatal(err)
	}
	if string(two) != ":event\ndata: two\n" {
		t.Fatalf("second event %q", two)
	}
	if _, err := r.ReadEvent(); !errors.Is(err, io.EOF) {
		t.Fatalf("end error = %v", err)
	}
}

func TestReaderBoundsEventAndWriterTerminates(t *testing.T) {
	r := NewReader(strings.NewReader(strings.Repeat("x", MaxEventSize+1) + "\n\n"))
	if _, err := r.ReadEvent(); !errors.Is(err, ErrEventTooLarge) {
		t.Fatalf("oversize error = %v", err)
	}
	var out bytes.Buffer
	if err := NewWriter(&out).WriteEvent([]byte("data: ok")); err != nil {
		t.Fatal(err)
	}
	if out.String() != "data: ok\n\n" {
		t.Fatalf("output %q", out.String())
	}
}

func FuzzReaderNeverPanics(f *testing.F) {
	f.Add([]byte("data: hello\n\n"))
	f.Add([]byte("data: é\r\n\r\n"))
	f.Fuzz(func(t *testing.T, data []byte) {
		r := NewReader(bytes.NewReader(data))
		for i := 0; i < 4; i++ {
			if _, err := r.ReadEvent(); err != nil {
				break
			}
		}
	})
}
