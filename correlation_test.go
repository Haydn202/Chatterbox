package chatterbox

import (
	"testing"
	"time"

	"github.com/Haydn202/Chatterbox/fuzz"
)

func TestCorrelation_sharedIDs(t *testing.T) {
	schema := NewSchema(
		MakeField("timestamp", fuzz.TimestampRFC3339(
			fuzz.WithBaseTime(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
			fuzz.WithJitter(0),
		)),
		MakeField("level", fuzz.Choice("info")),
		MakeField("message", fuzz.StringFrom(5, 5)),
	)
	gen := NewGenerator(schema, WithSeed(99), WithCorrelation(CorrelationConfig{
		MinLines: 2,
		MaxLines: 4,
		TimestampStep: time.Millisecond,
	}))

	var records []map[string]any
	for i := 0; i < 20; i++ {
		records = append(records, gen.Next())
	}

	runLen := 1
	for i := 1; i < len(records); i++ {
		prev, cur := records[i-1], records[i]
		if prev["trace_id"] == cur["trace_id"] && prev["request_id"] == cur["request_id"] {
			runLen++
			continue
		}
		if runLen < 2 || runLen > 4 {
			t.Fatalf("run length %d out of range at index %d", runLen, i)
		}
		runLen = 1
	}
	if runLen < 2 || runLen > 4 {
		t.Fatalf("final run length %d out of range", runLen)
	}
}

func TestCorrelation_noCorrelationIndependent(t *testing.T) {
	schema := NewSchema(
		MakeField("trace_id", fuzz.UUID()),
		MakeField("message", fuzz.Choice("a", "b")),
	)
	gen := NewGenerator(schema, WithSeed(1))
	a := gen.Next()["trace_id"]
	b := gen.Next()["trace_id"]
	if a == b {
		// unlikely but possible; try a few more
		seen := false
		for i := 0; i < 10; i++ {
			if gen.Next()["trace_id"] != a {
				seen = true
				break
			}
		}
		if !seen {
			t.Fatal("expected varying trace_id without correlation")
		}
	}
}
