package telemetry

import (
	"context"
	"testing"
	"time"
)

func TestCollectorDownDoesNotBlockRecording(t *testing.T) {
	tel, err := New(context.Background(), "test", "http://127.0.0.1:1")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tel.Shutdown(context.Background()) }()
	ctx, span := tel.Tracer.Start(context.Background(), "test")
	start := time.Now()
	for i := 0; i < 2000; i++ {
		tel.Requests.Add(ctx, 1)
	}
	span.End()
	if elapsed := time.Since(start); elapsed > 250*time.Millisecond {
		t.Fatalf("recording telemetry blocked for %s", elapsed)
	}
}
