package chatterbox

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Haydn202/Chatterbox/schedule"
)

func TestRunScheduled_flatRate(t *testing.T) {
	sched, err := schedule.FlatRate(200)
	if err != nil {
		t.Fatal(err)
	}
	var count atomic.Int32
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()
	err = runScheduled(ctx, sched, 0, func() error {
		count.Add(1)
		return nil
	})
	if err != nil && ctx.Err() == nil {
		t.Fatal(err)
	}
	if count.Load() < 2 {
		t.Fatalf("expected multiple events, got %d", count.Load())
	}
}

func TestRunScheduled_phasesEnd(t *testing.T) {
	sched, err := schedule.NewPhases(schedule.Phase{Rate: 1000, Duration: 1 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	var count atomic.Int32
	err = runScheduled(context.Background(), sched, 0, func() error {
		count.Add(1)
		time.Sleep(2 * time.Millisecond)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if count.Load() == 0 {
		t.Fatal("expected at least one event")
	}
}
