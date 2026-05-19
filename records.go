package chatterbox

import (
	"context"
	"fmt"
	"time"

	"github.com/Haydn202/Chatterbox/schedule"
)

// RecordHandler receives one generated log record. Return an error to stop emission.
type RecordHandler func(ctx context.Context, record map[string]any) error

func copyRecord(src map[string]any) map[string]any {
	dst := make(map[string]any, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// GenerateN calls handler n times with a copy of each generated record.
func (g *Generator) GenerateN(ctx context.Context, n int, h RecordHandler) error {
	if h == nil {
		return fmt.Errorf("chatterbox: handler must not be nil")
	}
	if n < 0 {
		return fmt.Errorf("chatterbox: n must be non-negative, got %d", n)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	for i := 0; i < n; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := h(ctx, copyRecord(g.Next())); err != nil {
			return err
		}
	}
	return nil
}

// StreamRecords emits at a constant rate until duration elapses or ctx is cancelled.
func (g *Generator) StreamRecords(ctx context.Context, rate float64, duration time.Duration, h RecordHandler) error {
	sched, err := schedule.FlatRate(rate)
	if err != nil {
		return err
	}
	return g.StreamRecordsWithSchedule(ctx, sched, duration, h)
}

// StreamRecordsWithSchedule emits using sched. totalCap zero means no overall cap.
func (g *Generator) StreamRecordsWithSchedule(ctx context.Context, sched schedule.Schedule, totalCap time.Duration, h RecordHandler) error {
	if h == nil {
		return fmt.Errorf("chatterbox: handler must not be nil")
	}
	if sched == nil {
		return fmt.Errorf("chatterbox: schedule must not be nil")
	}
	return runScheduled(ctx, sched, totalCap, func() error {
		return h(ctx, copyRecord(g.Next()))
	})
}
