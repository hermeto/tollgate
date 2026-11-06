package health

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/hermeto/tollgate/internal/config"
)

func TestFailClosedTransitionsToReady(t *testing.T) {
	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m := New([]config.DependencyConfig{{Name: "store", Kind: "redis", Address: listener.Addr().String(), FailMode: "closed"}}, nil)
	if m.Ready() {
		t.Fatal("unprobed fail-closed dependency must not be ready")
	}
	m.Start(ctx)
	deadline := time.After(2 * time.Second)
	for !m.Ready() {
		select {
		case <-deadline:
			t.Fatal("dependency did not become ready")
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}
	if got := len(m.Report()); got != 1 {
		t.Fatalf("report entries = %d, want 1", got)
	}
}
