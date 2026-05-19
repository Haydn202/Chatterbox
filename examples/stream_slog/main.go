// Example: emit fuzzy logs through slog.JSONHandler via the slog adapter.
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
	"github.com/Haydn202/Chatterbox/adapter"
	"github.com/Haydn202/Chatterbox/adapter/slog"
	"github.com/Haydn202/Chatterbox/emit"
	"github.com/Haydn202/Chatterbox/fuzz"
)

func main() {
	schema := chatterbox.NewSchema(
		chatterbox.MakeField("timestamp", fuzz.TimestampRFC3339(fuzz.WithJitter(10))),
		chatterbox.MakeField("level", fuzz.LevelWeighted(map[string]float64{
			"info": 0.7, "warn": 0.2, "error": 0.1,
		})),
		chatterbox.MakeField("message", fuzz.StringFrom(10, 80)),
		chatterbox.MakeField("email", fuzz.Email()),
	)
	gen := chatterbox.NewGenerator(schema, chatterbox.WithSeed(42))

	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	em := slogadapter.New(h, emit.DefaultFieldMap())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err := adapter.Stream(ctx, gen, 10, 30*time.Second, em)
	if err != nil && ctx.Err() == nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
