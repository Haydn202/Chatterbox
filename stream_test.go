package chatterbox

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Haydn202/Chatterbox/fuzz"
)

func TestNewStream_invalidRate(t *testing.T) {
	_, err := NewStream(NewGenerator(testSchema()), 0)
	if err == nil {
		t.Fatal("expected error for zero rate")
	}
}

func TestStream_duration(t *testing.T) {
	gen := NewGenerator(testSchema(), WithSeed(1))
	s, err := NewStream(gen, 50, WithStreamDuration(80*time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	var buf bytes.Buffer
	if err := s.Run(ctx, &buf); err != nil {
		t.Fatal(err)
	}
	// ~4 events at 50/sec for 80ms; allow wide timing slack on CI.
	n := countJSONLEvents(buf.Bytes())
	if n < 1 || n > 12 {
		t.Fatalf("expected roughly 4 events, got %d", n)
	}
}

func TestStream_contextCancel(t *testing.T) {
	gen := NewGenerator(testSchema(), WithSeed(2))
	s, err := NewStream(gen, 1000) // no duration: runs until cancelled
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	var buf bytes.Buffer
	err = s.Run(ctx, &buf)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected at least one event before cancel")
	}
}

func TestStream_foreverUntilCancel(t *testing.T) {
	schema := NewSchema(MakeField("x", fuzz.Choice("a")))
	gen := NewGenerator(schema, WithSeed(3))
	s, err := NewStream(gen, 200, WithStreamDuration(0))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	var buf bytes.Buffer
	go func() { done <- s.Run(ctx, &buf) }()

	time.Sleep(30 * time.Millisecond)
	cancel()

	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancel, got %v", err)
	}
}

func countJSONLEvents(b []byte) int {
	if len(b) == 0 {
		return 0
	}
	n := 0
	for _, c := range b {
		if c == '{' {
			n++
		}
	}
	return n
}
