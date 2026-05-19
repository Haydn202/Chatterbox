package chatterbox

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/Haydn202/Chatterbox/schedule"
)

// Stream emits logs from a Generator using a rate schedule.
type Stream struct {
	gen      *Generator
	sched    schedule.Schedule
	duration time.Duration // optional total cap (WithStreamDuration)
}

// StreamConfig configures rate-based emission.
type StreamConfig struct {
	// Rate is logs emitted per second (must be > 0).
	Rate float64
	// Duration limits how long to emit. Zero runs until ctx is cancelled.
	Duration time.Duration
}

// StreamOption configures a Stream.
type StreamOption func(*Stream)

// WithStreamDuration sets an overall wall-clock cap. Zero means no cap (phase durations or ctx apply).
func WithStreamDuration(d time.Duration) StreamOption {
	return func(s *Stream) { s.duration = d }
}

// NewStream creates a stream at a constant rate.
func NewStream(gen *Generator, rate float64, opts ...StreamOption) (*Stream, error) {
	sched, err := schedule.FlatRate(rate)
	if err != nil {
		return nil, err
	}
	return NewStreamWithSchedule(gen, sched, opts...)
}

// NewStreamFromConfig is equivalent to NewStream with cfg.Rate and cfg.Duration.
func NewStreamFromConfig(gen *Generator, cfg StreamConfig) (*Stream, error) {
	if cfg.Rate <= 0 {
		return nil, fmt.Errorf("chatterbox: rate must be positive, got %v", cfg.Rate)
	}
	s, err := NewStream(gen, cfg.Rate, WithStreamDuration(cfg.Duration))
	if err != nil {
		return nil, err
	}
	return s, nil
}

// NewStreamWithSchedule creates a stream driven by sched.
func NewStreamWithSchedule(gen *Generator, sched schedule.Schedule, opts ...StreamOption) (*Stream, error) {
	if gen == nil {
		return nil, fmt.Errorf("chatterbox: generator must not be nil")
	}
	if sched == nil {
		return nil, fmt.Errorf("chatterbox: schedule must not be nil")
	}
	s := &Stream{gen: gen, sched: sched}
	for _, o := range opts {
		o(s)
	}
	return s, nil
}

// Run writes formatted logs to w until ctx is cancelled, the schedule ends, or duration cap elapses.
func (s *Stream) Run(ctx context.Context, w io.Writer) error {
	return runScheduled(ctx, s.sched, s.duration, func() error {
		b, err := s.gen.NextFormatted()
		if err != nil {
			return err
		}
		_, err = w.Write(b)
		return err
	})
}
