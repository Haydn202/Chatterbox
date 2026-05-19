package config_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Haydn202/Chatterbox"
	"github.com/Haydn202/Chatterbox/config"
	"github.com/Haydn202/Chatterbox/fuzz"
)

func TestLoadSchema_minimalYAML(t *testing.T) {
	path := filepath.Join("testdata", "minimal.yaml")
	schema, err := config.LoadSchemaFile(path)
	if err != nil {
		t.Fatal(err)
	}
	base := time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)
	fixed := chatterbox.NewSchema(
		chatterbox.MakeField("timestamp", fuzz.TimestampRFC3339(
			fuzz.WithBaseTime(base),
			fuzz.WithJitter(0),
		)),
		chatterbox.MakeField("level", fuzz.LevelWeighted(map[string]float64{"info": 1})),
		chatterbox.MakeField("message", fuzz.StringFrom(5, 10)),
	)
	_ = fixed

	gen := chatterbox.NewGenerator(schema, chatterbox.WithSeed(99))
	for i := 0; i < 3; i++ {
		rec := gen.Next()
		if rec["level"] != "info" {
			t.Fatalf("line %d: level = %v", i, rec["level"])
		}
		if _, ok := rec["message"].(string); !ok {
			t.Fatalf("line %d: message missing", i)
		}
	}
}

func TestLoadSchema_inline(t *testing.T) {
	const doc = `
fields:
  - name: tag
    type: constant
    value: ok
  - name: level
    type: choice
    values: [a, b]
`
	schema, err := config.LoadSchema(bytes.NewReader([]byte(doc)))
	if err != nil {
		t.Fatal(err)
	}
	gen := chatterbox.NewGenerator(schema, chatterbox.WithSeed(1))
	rec := gen.Next()
	if rec["tag"] != "ok" {
		t.Fatalf("tag = %v", rec["tag"])
	}
}

func TestLoadSchema_golden(t *testing.T) {
	path := filepath.Join("testdata", "minimal.yaml")
	schema, err := config.LoadSchemaFile(path)
	if err != nil {
		t.Fatal(err)
	}
	gen := chatterbox.NewGenerator(schema, chatterbox.WithSeed(42))
	lines, err := gen.NextN(5)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	for _, line := range lines {
		buf.Write(line)
	}
	golden := filepath.Join("testdata", "golden-config.jsonl")
	if os.Getenv("UPDATE_GOLDEN") != "" {
		if err := os.WriteFile(golden, buf.Bytes(), 0644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("read golden (run with UPDATE_GOLDEN=1): %v", err)
	}
	if !bytes.Equal(buf.Bytes(), want) {
		t.Fatalf("output mismatch with %s", golden)
	}
}

func TestLoadSchema_errors(t *testing.T) {
	_, err := config.LoadSchema(bytes.NewReader([]byte("fields: []")))
	if err == nil {
		t.Fatal("expected error for empty fields")
	}
	_, err = config.LoadSchema(bytes.NewReader([]byte(`
fields:
  - name: x
    type: unknown_type
`)))
	if err == nil {
		t.Fatal("expected error for unknown type")
	}
}
