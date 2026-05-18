package chatterbox

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Haydn202/Chatterbox/emit"
	"github.com/Haydn202/Chatterbox/fuzz"
)

func testSchema() *Schema {
	base := time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)
	return NewSchema(
		MakeField("timestamp", fuzz.TimestampRFC3339(
			fuzz.WithBaseTime(base),
			fuzz.WithJitter(0),
		)),
		MakeField("level", fuzz.LevelWeighted(map[string]float64{
			"info": 0.7, "warn": 0.2, "error": 0.1,
		})),
		MakeField("email", fuzz.Email()),
		MakeField("message", fuzz.StringFrom(5, 20)),
	)
}

func TestGenerator_golden(t *testing.T) {
	gen := NewGenerator(testSchema(), WithSeed(42))
	lines, err := gen.NextN(5)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	for _, line := range lines {
		buf.Write(line)
	}

	golden := filepath.Join("testdata", "golden.jsonl")
	if os.Getenv("UPDATE_GOLDEN") != "" {
		if err := os.MkdirAll("testdata", 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(golden, buf.Bytes(), 0644); err != nil {
			t.Fatal(err)
		}
	}

	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("read golden (run with UPDATE_GOLDEN=1 to create): %v", err)
	}
	if !bytes.Equal(buf.Bytes(), want) {
		t.Fatalf("output mismatch:\n%s", buf.String())
	}
}

func TestGenerator_WriteN_jsonValid(t *testing.T) {
	gen := NewGenerator(testSchema(), WithSeed(1))
	var buf bytes.Buffer
	if err := gen.WriteN(&buf, 100); err != nil {
		t.Fatal(err)
	}
	raw := buf.Bytes()
	if bytes.Count(raw, []byte{'\n'}) != 100 {
		t.Fatalf("expected 100 lines, got %d", bytes.Count(raw, []byte{'\n'}))
	}
	start := 0
	for i := 0; i < 100; i++ {
		end := bytes.IndexByte(raw[start:], '\n')
		if end < 0 {
			t.Fatal("missing newline")
		}
		line := raw[start : start+end]
		var m map[string]any
		if err := json.Unmarshal(line, &m); err != nil {
			t.Fatalf("line %d: %v", i, err)
		}
		start += end + 1
	}
}

func multilineSchema() *Schema {
	base := time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)
	return NewSchema(
		MakeField("timestamp", fuzz.TimestampRFC3339(
			fuzz.WithBaseTime(base),
			fuzz.WithJitter(0),
		)),
		MakeField("level", fuzz.Choice("error")),
		MakeField("message", fuzz.StringFrom(12, 12)),
		MakeField("stacktrace", fuzz.StackTrace(
			fuzz.WithStackStyle(fuzz.StackStyleGo),
			fuzz.WithFrames(4, 4),
		)),
	)
}

func TestGenerator_multilineGolden(t *testing.T) {
	fmt := emit.TextMultilineFormatter(emit.TextMultilineConfig{
		HeaderFields: []string{"timestamp", "level", "message"},
		BodyFields:   []string{"stacktrace"},
	})
	gen := NewGenerator(multilineSchema(), WithSeed(42), WithFormatter(fmt))
	lines, err := gen.NextN(2)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	for _, line := range lines {
		buf.Write(line)
	}

	golden := filepath.Join("testdata", "golden-multiline.txt")
	if os.Getenv("UPDATE_GOLDEN") != "" {
		if err := os.MkdirAll("testdata", 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(golden, buf.Bytes(), 0644); err != nil {
			t.Fatal(err)
		}
	}

	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("read golden (run with UPDATE_GOLDEN=1 to create): %v", err)
	}
	if !bytes.Equal(buf.Bytes(), want) {
		t.Fatalf("output mismatch:\n%s", buf.String())
	}

	for i, line := range lines {
		if bytes.Count(line, []byte{'\n'}) < 2 {
			t.Fatalf("event %d: expected multiline output, got %q", i, line)
		}
	}
}

func TestGenerator_Next_reproducible(t *testing.T) {
	a := NewGenerator(testSchema(), WithSeed(99))
	b := NewGenerator(testSchema(), WithSeed(99))
	if string(mustJSON(a)) != string(mustJSON(b)) {
		t.Fatal("same seed should match")
	}
}

func mustJSON(g *Generator) []byte {
	b, err := g.NextJSON()
	if err != nil {
		panic(err)
	}
	return b
}
