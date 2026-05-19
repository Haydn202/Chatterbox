package zapadapter

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/Haydn202/Chatterbox"
	"github.com/Haydn202/Chatterbox/emit"
	"github.com/Haydn202/Chatterbox/fuzz"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestEmitter_Log(t *testing.T) {
	var buf bytes.Buffer
	enc := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(enc, zapcore.AddSync(&buf), zapcore.DebugLevel)
	logger := zap.New(core)
	em := New(logger, emit.DefaultFieldMap())

	schema := chatterbox.NewSchema(
		chatterbox.MakeField("timestamp", fuzz.TimestampRFC3339()),
		chatterbox.MakeField("level", fuzz.Choice("info")),
		chatterbox.MakeField("message", fuzz.Choice("hello")),
		chatterbox.MakeField("email", fuzz.Email()),
	)
	gen := chatterbox.NewGenerator(schema, chatterbox.WithSeed(1))

	if err := em.Emit(context.Background(), gen.Next()); err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("not json: %v\n%s", err, buf.Bytes())
	}
	if m["msg"] != "hello" {
		t.Fatalf("expected msg hello, got %v", m["msg"])
	}
}
