package zerologadapter

import (
	"context"
	"fmt"
	"strings"

	"github.com/Haydn202/Chatterbox/emit"
	"github.com/rs/zerolog"
)

// Emitter logs records through a zerolog.Logger.
type Emitter struct {
	Logger zerolog.Logger
	FM     emit.FieldMap
}

// New creates an Emitter. fm zero value uses emit.DefaultFieldMap().
func New(logger zerolog.Logger, fm emit.FieldMap) *Emitter {
	if fm == (emit.FieldMap{}) {
		fm = emit.DefaultFieldMap()
	}
	return &Emitter{Logger: logger, FM: fm}
}

// Emit writes one record via the logger.
func (e *Emitter) Emit(ctx context.Context, record map[string]any) error {
	_ = ctx
	msg := fmt.Sprint(record[e.FM.MessageKey()])
	level := strings.ToLower(fmt.Sprint(record[e.FM.LevelKey()]))

	var ev *zerolog.Event
	switch level {
	case "debug":
		ev = e.Logger.Debug()
	case "warn", "warning":
		ev = e.Logger.Warn()
	case "error", "err":
		ev = e.Logger.Error()
	case "fatal":
		ev = e.Logger.Fatal()
	case "panic":
		ev = e.Logger.Panic()
	case "trace":
		ev = e.Logger.Trace()
	default:
		ev = e.Logger.Info()
	}
	for k, v := range e.FM.Remaining(record) {
		ev = ev.Interface(k, v)
	}
	ev.Msg(msg)
	return nil
}
