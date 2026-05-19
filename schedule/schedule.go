package schedule

import (
	"fmt"
	"time"
)

// Schedule controls wait time between emitted events.
type Schedule interface {
	// NextWait returns how long to wait before the next event after the current one.
	// ok false means the schedule is finished (no more phases).
	NextWait() (wait time.Duration, ok bool)
}

// Phase is one rate segment.
type Phase struct {
	Rate     float64       // events per second, must be > 0
	Duration time.Duration // 0 = run until context cancel (only valid for the last phase)
}

type flatRate struct {
	interval time.Duration
}

func (f *flatRate) NextWait() (time.Duration, bool) {
	return f.interval, true
}

// FlatRate emits at a constant events-per-second forever.
func FlatRate(eventsPerSec float64) (Schedule, error) {
	if eventsPerSec <= 0 {
		return nil, fmt.Errorf("schedule: rate must be positive, got %v", eventsPerSec)
	}
	return &flatRate{interval: intervalForRate(eventsPerSec)}, nil
}

type phaseSchedule struct {
	segments    []Phase
	phaseIndex  int
	phaseStart  time.Time
	phaseEvents int
}

// NewPhases runs phases in order. Only the last phase may have Duration 0 (runs until the schedule is stopped externally).
func NewPhases(segments ...Phase) (Schedule, error) {
	if len(segments) == 0 {
		return nil, fmt.Errorf("schedule: at least one phase required")
	}
	for i, p := range segments {
		if p.Rate <= 0 {
			return nil, fmt.Errorf("schedule: phase %d rate must be positive", i)
		}
		if p.Duration == 0 && i != len(segments)-1 {
			return nil, fmt.Errorf("schedule: only the last phase may have zero duration")
		}
	}
	cp := make([]Phase, len(segments))
	copy(cp, segments)
	return &phaseSchedule{segments: cp, phaseStart: time.Now()}, nil
}

func (p *phaseSchedule) NextWait() (time.Duration, bool) {
	if p.phaseIndex >= len(p.segments) {
		return 0, false
	}
	ph := p.segments[p.phaseIndex]
	if ph.Duration > 0 && time.Since(p.phaseStart) >= ph.Duration {
		p.phaseIndex++
		p.phaseStart = time.Now()
		p.phaseEvents = 0
		if p.phaseIndex >= len(p.segments) {
			return 0, false
		}
		ph = p.segments[p.phaseIndex]
	}
	p.phaseEvents++
	return intervalForRate(ph.Rate), true
}

// PresetIncidentSpike returns base rate, then peak rate, then base rate until stopped.
func PresetIncidentSpike(baseRate, peakRate float64, baseDur, peakDur time.Duration) (Schedule, error) {
	return NewPhases(
		Phase{Rate: baseRate, Duration: baseDur},
		Phase{Rate: peakRate, Duration: peakDur},
		Phase{Rate: baseRate, Duration: 0},
	)
}

func intervalForRate(eventsPerSec float64) time.Duration {
	d := time.Duration(float64(time.Second) / eventsPerSec)
	if d < time.Microsecond {
		return time.Microsecond
	}
	return d
}
