package emit

import (
	"bytes"
	"strings"
	"testing"
)

func TestTextMultiline_physicalLines(t *testing.T) {
	record := map[string]any{
		"timestamp":  "2026-05-18T12:00:00Z",
		"level":      "error",
		"message":    "handler failed",
		"stacktrace": "panic: boom\n\ngoroutine 1 [running]:\nmain.main()\n\t/main.go:1 +0x1",
	}
	f := TextMultilineFormatter(TextMultilineConfig{
		HeaderFields: []string{"timestamp", "level", "message"},
		BodyFields:   []string{"stacktrace"},
	})
	b, err := f.Format(record)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Count(b, []byte{'\n'}) < 2 {
		t.Fatalf("expected multiple physical lines, got:\n%s", b)
	}
	lines := strings.Split(strings.TrimSuffix(string(b), "\n"), "\n")
	if !strings.Contains(lines[0], "level=error") {
		t.Fatalf("bad header: %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], "panic:") {
		t.Fatalf("bad stack start: %q", lines[1])
	}
}

func TestTextMultiline_quoteSpaces(t *testing.T) {
	record := map[string]any{
		"message": "hello world",
	}
	f := TextMultilineFormatter(TextMultilineConfig{
		HeaderFields: []string{"message"},
	})
	b, err := f.Format(record)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `message="hello world"`) {
		t.Fatalf("expected quoted value: %q", b)
	}
}
