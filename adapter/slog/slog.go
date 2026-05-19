package slogadapter

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Haydn202/Chatterbox/emit"
)

// Emitter logs records through a slog.Handler.
type Emitter struct {
	Handler slog.Handler
	FM      emit.FieldMap
}

// New creates an Emitter. fm zero value uses emit.DefaultFieldMap().
func New(h slog.Handler, fm emit.FieldMap) *Emitter {
	if fm == (emit.FieldMap{}) {
		fm = emit.DefaultFieldMap()
	}
	return &Emitter{Handler: h, FM: fm}
}

// Emit writes one record via the handler.
func (e *Emitter) Emit(ctx context.Context, record map[string]any) error {
	if e.Handler == nil {
		return fmt.Errorf("slogadapter: handler must not be nil")
	}
	t := emit.ParseTime(record, e.FM.TimeKey())
	level := parseLevel(fmt.Sprint(record[e.FM.LevelKey()]))
	msg := fmt.Sprint(record[e.FM.MessageKey()])

	var pcs [1]uintptr
	r := slog.NewRecord(t, level, msg, pcs[0])
	r.AddAttrs(mapToAttrs(e.FM.Remaining(record))...)
	return e.Handler.Handle(ctx, r)
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error", "err":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func mapToAttrs(m map[string]any) []slog.Attr {
	attrs := make([]slog.Attr, 0, len(m))
	for k, v := range m {
		attrs = append(attrs, slog.Any(k, v))
	}
	return attrs
}
