package scenario

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/Haydn202/Chatterbox/schedule"
)

// Run executes the scenario until duration elapses or ctx is cancelled.
func (r *Runner) Run(ctx context.Context, writers WritersConfig) error {
	if ctx == nil {
		ctx = context.Background()
	}
	svcNames := make([]string, 0, len(r.states))
	for name := range r.states {
		svcNames = append(svcNames, name)
	}
	mux, err := newWriterMux(writers, svcNames)
	if err != nil {
		return err
	}
	defer mux.closeAll()

	phaseIdx := 0
	phaseStart := time.Now()
	activeServices := resolvePhaseServices(r.plan.Phases[phaseIdx], r.states)

	if err := r.applyPhase(phaseIdx); err != nil {
		return err
	}

	sched, err := schedule.FlatRate(r.opts.Rate)
	if err != nil {
		return err
	}

	deadline := time.Now().Add(r.opts.Duration)
	rng := rand.New(rand.NewPCG(r.opts.Seed, r.opts.Seed^0xdeadbeef))

	emit := func() error {
		now := time.Now()
		if !now.Before(deadline) {
			return errScenarioDone
		}
		for phaseIdx < len(r.plan.Phases) {
			ph := r.plan.Phases[phaseIdx]
			if ph.Duration > 0 && now.Sub(phaseStart) >= ph.Duration {
				prev := phaseIdx
				phaseIdx++
				if phaseIdx >= len(r.plan.Phases) {
					return errScenarioDone
				}
				if err := r.endPhase(prev); err != nil {
					return err
				}
				activeServices = resolvePhaseServices(r.plan.Phases[phaseIdx], r.states)
				phaseStart = now
				if err := r.applyPhase(phaseIdx); err != nil {
					return err
				}
				continue
			}
			break
		}

		name := pickService(rng, r.states, activeServices)
		st := r.states[name]
		line, err := r.emitFrom(st)
		if err != nil {
			return err
		}
		return mux.write(name, line)
	}

	return runScheduled(ctx, sched, r.opts.Duration, emit)
}

var errScenarioDone = fmt.Errorf("scenario done")

func (r *Runner) applyPhase(idx int) error {
	ph := r.plan.Phases[idx]
	ev, err := lookupEvent(ph.Event)
	if err != nil {
		return err
	}
	targets := resolvePhaseServices(ph, r.states)
	ev.Apply(targets, r.states)
	return nil
}

func (r *Runner) endPhase(idx int) error {
	ph := r.plan.Phases[idx]
	ev, err := lookupEvent(ph.Event)
	if err != nil {
		return err
	}
	targets := resolvePhaseServices(ph, r.states)
	ev.Remove(targets, r.states)
	return nil
}

func allServiceNames(states map[string]*serviceRuntime) []string {
	out := make([]string, 0, len(states))
	for name := range states {
		out = append(out, name)
	}
	return out
}

func resolvePhaseServices(ph Phase, states map[string]*serviceRuntime) []string {
	if len(ph.Services) > 0 {
		return filterExisting(ph.Services, states)
	}
	if ev, err := lookupEvent(ph.Event); err == nil {
		if def := ev.DefaultTargets(); len(def) > 0 {
			return filterExisting(def, states)
		}
	}
	return allServiceNames(states)
}

func filterExisting(names []string, states map[string]*serviceRuntime) []string {
	out := make([]string, 0, len(names))
	for _, n := range names {
		if _, ok := states[n]; ok {
			out = append(out, n)
		}
	}
	return out
}

func pickService(rng *rand.Rand, states map[string]*serviceRuntime, active []string) string {
	if len(active) == 0 {
		active = allServiceNames(states)
	}
	var total float64
	weights := make([]float64, len(active))
	for i, name := range active {
		st := states[name]
		w := st.cfg.RateWeight * st.patch.rateMultiplier
		if w <= 0 {
			w = 0.01
		}
		weights[i] = w
		total += w
	}
	p := rng.Float64() * total
	var cum float64
	for i, w := range weights {
		cum += w
		if p < cum {
			return active[i]
		}
	}
	return active[len(active)-1]
}

// runScheduled mirrors chatterbox.runScheduled but treats errScenarioDone as clean exit.
func runScheduled(ctx context.Context, sched schedule.Schedule, totalCap time.Duration, emit func() error) error {
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
			if err == errScenarioDone {
				return nil
			}
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
