package adapter

import (
	"context"
	"time"

	"github.com/Haydn202/Chatterbox"
)

// Emitter logs one generated record using a specific logging library.
type Emitter interface {
	Emit(ctx context.Context, record map[string]any) error
}

// GenerateN emits n records through em.
func GenerateN(ctx context.Context, gen *chatterbox.Generator, n int, em Emitter) error {
	return gen.GenerateN(ctx, n, em.Emit)
}

// Stream emits at rate until duration or ctx cancel. duration zero runs until cancelled.
func Stream(ctx context.Context, gen *chatterbox.Generator, rate float64, duration time.Duration, em Emitter) error {
	return gen.StreamRecords(ctx, rate, duration, em.Emit)
}
