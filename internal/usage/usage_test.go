package usage

import "testing"

func TestWriterFlushesAndReplays(t *testing.T) {
	path := t.TempDir() + "/events.jsonl"
	w := NewWriter(NewSpoolSink(path), 2)
	for i := 0; i < 5; i++ {
		if !w.Submit(Event{RequestID: "r"}) {
			t.Fatal("submit rejected")
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	events, err := Replay(path)
	if err != nil || len(events) != 5 {
		t.Fatalf("events=%d err=%v", len(events), err)
	}
}
