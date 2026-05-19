package chatterbox

import (
	"context"
	"time"

	"github.com/Haydn202/Chatterbox/schedule"
)

// runScheduled emits events using sched until ctx is done, sched finishes, emit returns an error, or totalCap elapses.
// The first event is emitted immediately. totalCap zero means no overall cap.
func runScheduled(ctx context.Context, sched schedule.Schedule, totalCap time.Duration, emit func() error) error {
	if ctx == nil {
		ctx = context.Background()
	}

	var deadline time.Time
	if totalCap > 0 {
		deadline = time.Now().Add(totalCap)
	}

	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if !deadline.IsZero() && !time.Now().Before(deadline) {
			return nil
		}

		if err := emit(); err != nil {
			return err
		}

		wait, ok := sched.NextWait()
		if !ok {
			return nil
		}

		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}
