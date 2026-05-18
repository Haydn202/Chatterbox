package fuzz

import (
	"math/rand/v2"
	"time"
)

// TimestampOption configures TimestampRFC3339.
type TimestampOption func(*timestampFuzzer)

// WithJitter sets max seconds to add/subtract from base time (default 0).
func WithJitter(seconds int) TimestampOption {
	return func(t *timestampFuzzer) { t.jitterSec = seconds }
}

// WithBaseTime sets the reference time (default time.Now() at generation).
func WithBaseTime(base time.Time) TimestampOption {
	return func(t *timestampFuzzer) {
		t.base = base
		t.useBase = true
	}
}

type timestampFuzzer struct {
	jitterSec int
	base      time.Time
	useBase   bool
}

// TimestampRFC3339 generates RFC3339 timestamps with optional jitter.
func TimestampRFC3339(opts ...TimestampOption) Fuzzer {
	t := &timestampFuzzer{}
	for _, o := range opts {
		o(t)
	}
	return t
}

func (t *timestampFuzzer) Generate(r *rand.Rand) any {
	base := t.base
	if !t.useBase {
		base = time.Now().UTC()
	} else {
		base = base.UTC()
	}
	if t.jitterSec > 0 {
		delta := r.IntN(2*t.jitterSec+1) - t.jitterSec
		base = base.Add(time.Duration(delta) * time.Second)
	}
	return base.Format(time.RFC3339)
}
