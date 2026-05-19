// Example: receive fuzzy records and log them with slog (any logger works the same way).
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Haydn202/Chatterbox"
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err := gen.StreamRecords(ctx, 10, 30*time.Second, func(ctx context.Context, rec map[string]any) error {
		level := parseLevel(fmt.Sprint(rec["level"]))
		msg := fmt.Sprint(rec["message"])
		attrs := attrsExcept(rec, "timestamp", "level", "message")
		slog.Default().LogAttrs(ctx, level, msg, attrs...)
		return nil
	})
	if err != nil && ctx.Err() == nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
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

func attrsExcept(rec map[string]any, skip ...string) []slog.Attr {
	skipSet := make(map[string]bool, len(skip))
	for _, k := range skip {
		skipSet[k] = true
	}
	var attrs []slog.Attr
	for k, v := range rec {
		if !skipSet[k] && v != nil {
			attrs = append(attrs, slog.Any(k, v))
		}
	}
	return attrs
}
