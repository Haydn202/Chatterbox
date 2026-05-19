// Example: correlated trace IDs with an incident rate spike.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Haydn202/Chatterbox"
	"github.com/Haydn202/Chatterbox/fuzz"
	"github.com/Haydn202/Chatterbox/schedule"
)

func main() {
	schema := chatterbox.NewSchema(
		chatterbox.MakeField("timestamp", fuzz.TimestampRFC3339(fuzz.WithJitter(5))),
		chatterbox.MakeField("level", fuzz.LevelWeighted(map[string]float64{
			"info": 0.75, "warn": 0.15, "error": 0.1,
		})),
		chatterbox.MakeField("message", fuzz.StringFrom(15, 80)),
		chatterbox.MakeField("stacktrace", fuzz.Optional(0.15, fuzz.StackTrace(
			fuzz.WithFrames(4, 8),
		))),
	)

	gen := chatterbox.NewGenerator(schema,
		chatterbox.WithSeed(42),
		chatterbox.WithCorrelation(chatterbox.CorrelationConfig{
			MinLines:       3,
			MaxLines:       6,
			TimestampStep:  2 * time.Millisecond,
		}),
	)

	sched, err := schedule.PresetIncidentSpike(10, 120, time.Minute, 30*time.Second)
	if err != nil {
		panic(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err = gen.StreamRecordsWithSchedule(ctx, sched, 0, func(ctx context.Context, rec map[string]any) error {
		attrs := []any{
			"trace_id", rec["trace_id"],
			"request_id", rec["request_id"],
		}
		if st, ok := rec["stacktrace"]; ok && st != nil {
			attrs = append(attrs, "stacktrace", st)
		}
		slog.Default().InfoContext(ctx, fmt.Sprint(rec["message"]), attrs...)
		return nil
	})
	if err != nil && ctx.Err() == nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
