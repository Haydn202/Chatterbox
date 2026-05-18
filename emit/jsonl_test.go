package emit

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestJSONL_roundTrip(t *testing.T) {
	record := map[string]any{"level": "info", "count": float64(3)}
	b, err := JSONL(record)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := WriteJSONL(&buf, record); err != nil {
		t.Fatal(err)
	}
	if buf.Bytes()[len(b)] != '\n' {
		t.Fatal("expected trailing newline")
	}
}
