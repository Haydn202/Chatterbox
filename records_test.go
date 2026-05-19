package chatterbox

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Haydn202/Chatterbox/fuzz"
)

func TestGenerateN(t *testing.T) {
	gen := NewGenerator(testSchema(), WithSeed(1))
	var count atomic.Int32
	err := gen.GenerateN(context.Background(), 10, func(ctx context.Context, rec map[string]any) error {
		count.Add(1)
		if _, ok := rec["email"]; !ok {
			t.Fatal("expected email field")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if count.Load() != 10 {
		t.Fatalf("got %d records", count.Load())
	}
}

func TestGenerateN_handlerStops(t *testing.T) {
	gen := NewGenerator(testSchema(), WithSeed(1))
	err := gen.GenerateN(context.Background(), 100, func(ctx context.Context, rec map[string]any) error {
		return errors.New("stop")
	})
	if err == nil || err.Error() != "stop" {
		t.Fatalf("expected stop error, got %v", err)
	}
}

func TestStreamRecords_cancel(t *testing.T) {
	gen := NewGenerator(testSchema(), WithSeed(2))
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	var count atomic.Int32
	err := gen.StreamRecords(ctx, 500, 0, func(ctx context.Context, rec map[string]any) error {
		count.Add(1)
		return nil
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline, got %v", err)
	}
	if count.Load() == 0 {
		t.Fatal("expected at least one record")
	}
}

func TestStreamRecords_duration(t *testing.T) {
	schema := NewSchema(MakeField("x", fuzz.Choice("a")))
	gen := NewGenerator(schema, WithSeed(3))
	var count atomic.Int32
	err := gen.StreamRecords(context.Background(), 50, 80*time.Millisecond, func(ctx context.Context, rec map[string]any) error {
		count.Add(1)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if count.Load() < 1 || count.Load() > 12 {
		t.Fatalf("expected ~4 records, got %d", count.Load())
	}
}
