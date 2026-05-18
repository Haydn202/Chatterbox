package chatterbox

import (
	"context"
	"fmt"
	"io"
	"time"
)

// Stream emits logs from a Generator at a fixed rate for a configurable duration.
type Stream struct {
	gen      *Generator
	rate     float64       // logs per second
	duration time.Duration // zero = until context is cancelled
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

// WithStreamDuration sets how long to emit. Zero means until context cancellation.
func WithStreamDuration(d time.Duration) StreamOption {
	return func(s *Stream) { s.duration = d }
}

// NewStream creates a rate-limited emitter for gen.
func NewStream(gen *Generator, rate float64, opts ...StreamOption) (*Stream, error) {
	if rate <= 0 {
		return nil, fmt.Errorf("chatterbox: rate must be positive, got %v", rate)
	}
	if gen == nil {
		return nil, fmt.Errorf("chatterbox: generator must not be nil")
	}
	s := &Stream{gen: gen, rate: rate}
	for _, o := range opts {
		o(s)
	}
	return s, nil
}

// NewStreamFromConfig is equivalent to NewStream with cfg.Rate and cfg.Duration.
func NewStreamFromConfig(gen *Generator, cfg StreamConfig) (*Stream, error) {
	if cfg.Rate <= 0 {
		return nil, fmt.Errorf("chatterbox: rate must be positive, got %v", cfg.Rate)
	}
	return NewStream(gen, cfg.Rate, WithStreamDuration(cfg.Duration))
}

// Run writes formatted logs to w at the configured rate until duration elapses,
// ctx is cancelled, or an I/O error occurs. Returns nil when duration completes
// normally; returns ctx.Err() when cancelled.
func (s *Stream) Run(ctx context.Context, w io.Writer) error {
	if ctx == nil {
		ctx = context.Background()
	}

	interval := time.Duration(float64(time.Second) / s.rate)
	if interval < time.Microsecond {
		interval = time.Microsecond
	}

	var deadline time.Time
	if s.duration > 0 {
		deadline = time.Now().Add(s.duration)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if !deadline.IsZero() && !time.Now().Before(deadline) {
			return nil
		}

		b, err := s.gen.NextFormatted()
		if err != nil {
			return err
		}
		if _, err := w.Write(b); err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
