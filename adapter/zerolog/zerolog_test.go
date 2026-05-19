package zerologadapter

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/Haydn202/Chatterbox"
	"github.com/Haydn202/Chatterbox/emit"
	"github.com/Haydn202/Chatterbox/fuzz"
	"github.com/rs/zerolog"
)

func TestEmitter_Log(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := zerolog.New(buf)
	em := New(logger, emit.DefaultFieldMap())

	schema := chatterbox.NewSchema(
		chatterbox.MakeField("timestamp", fuzz.TimestampRFC3339()),
		chatterbox.MakeField("level", fuzz.Choice("warn")),
		chatterbox.MakeField("message", fuzz.Choice("alert")),
	)
	gen := chatterbox.NewGenerator(schema, chatterbox.WithSeed(1))

	if err := em.Emit(context.Background(), gen.Next()); err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("not json: %v\n%s", err, buf.Bytes())
	}
	if m["message"] != "alert" {
		t.Fatalf("expected message alert, got %v", m["message"])
	}
}
