package zapadapter

import (
	"context"
	"fmt"
	"strings"

	"github.com/Haydn202/Chatterbox/emit"
	"go.uber.org/zap"
)

// Emitter logs records through a zap.Logger.
type Emitter struct {
	Logger *zap.Logger
	FM     emit.FieldMap
}

// New creates an Emitter. fm zero value uses emit.DefaultFieldMap().
func New(logger *zap.Logger, fm emit.FieldMap) *Emitter {
	if fm == (emit.FieldMap{}) {
		fm = emit.DefaultFieldMap()
	}
	return &Emitter{Logger: logger, FM: fm}
}

// Emit writes one record via the logger.
func (e *Emitter) Emit(ctx context.Context, record map[string]any) error {
	if e.Logger == nil {
		return fmt.Errorf("zapadapter: logger must not be nil")
	}
	_ = ctx
	msg := fmt.Sprint(record[e.FM.MessageKey()])
	fields := mapToFields(e.FM.Remaining(record))
	level := strings.ToLower(fmt.Sprint(record[e.FM.LevelKey()]))

	switch level {
	case "debug":
		e.Logger.Debug(msg, fields...)
	case "warn", "warning":
		e.Logger.Warn(msg, fields...)
	case "error", "err":
		e.Logger.Error(msg, fields...)
	case "dpanic":
		e.Logger.DPanic(msg, fields...)
	case "panic":
		e.Logger.Panic(msg, fields...)
	case "fatal":
		e.Logger.Fatal(msg, fields...)
	default:
		e.Logger.Info(msg, fields...)
	}
	return nil
}

func mapToFields(m map[string]any) []zap.Field {
	fields := make([]zap.Field, 0, len(m))
	for k, v := range m {
		fields = append(fields, zap.Any(k, v))
	}
	return fields
}
