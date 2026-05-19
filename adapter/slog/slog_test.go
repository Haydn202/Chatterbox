package slogadapter

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/Haydn202/Chatterbox"
	"github.com/Haydn202/Chatterbox/emit"
	"github.com/Haydn202/Chatterbox/fuzz"
)

func TestEmitter_Handle(t *testing.T) {
	var buf bytes.Buffer
	h := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	em := New(h, emit.DefaultFieldMap())

	schema := chatterbox.NewSchema(
		chatterbox.MakeField("timestamp", fuzz.TimestampRFC3339()),
		chatterbox.MakeField("level", fuzz.Choice("error")),
		chatterbox.MakeField("message", fuzz.Choice("boom")),
	)
	gen := chatterbox.NewGenerator(schema, chatterbox.WithSeed(1))

	if err := em.Emit(context.Background(), gen.Next()); err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("not json: %v\n%s", err, buf.Bytes())
	}
	if m["msg"] != "boom" {
		t.Fatalf("expected msg boom, got %v", m["msg"])
	}
}
