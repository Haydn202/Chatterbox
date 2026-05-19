package scenario

import (
	"fmt"
	"time"

	"github.com/Haydn202/Chatterbox"
	"github.com/Haydn202/Chatterbox/config"
	"github.com/Haydn202/Chatterbox/emit"
	"github.com/Haydn202/Chatterbox/preset"
)

type serviceRuntime struct {
	name   string
	cfg    ServiceConfig
	gen    *chatterbox.Generator
	patch  patchState
	labels map[string]string
}

// RunnerOptions configures scenario execution.
type RunnerOptions struct {
	Seed      uint64
	Duration  time.Duration
	Format    emit.Format
	Rate      float64
	Correlate *bool // nil = use plan.Correlate
}

// Runner executes a scenario plan.
type Runner struct {
	plan      *Plan
	opts      RunnerOptions
	states    map[string]*serviceRuntime
	shared    *chatterbox.SharedCorrelation
	correlate bool
}

// NewRunner builds generators for each service in the plan.
func NewRunner(plan *Plan, opts RunnerOptions) (*Runner, error) {
	if plan == nil {
		return nil, fmt.Errorf("scenario: plan is nil")
	}
	ro := opts
	if ro.Seed == 0 {
		ro.Seed = plan.Seed
	}
	if ro.Rate <= 0 {
		ro.Rate = plan.Rate
	}
	if ro.Format == "" {
		ro.Format = plan.Format
	}
	if ro.Duration == 0 {
		ro.Duration = plan.Duration
	}

	formatter, err := emit.NewFormatter(ro.Format, emit.Options{})
	if err != nil {
		return nil, err
	}

	shared := chatterbox.NewSharedCorrelation(chatterbox.CorrelationConfig{
		MinLines:      3,
		MaxLines:      8,
		TimestampStep: 2 * time.Millisecond,
	}, ro.Seed)

	states := make(map[string]*serviceRuntime, len(plan.Services))
	var i uint64
	for name, svc := range plan.Services {
		schema, err := buildServiceSchema(svc)
		if err != nil {
			return nil, fmt.Errorf("service %q: %w", name, err)
		}
		seed := ro.Seed + i
		i++
		genOpts := []chatterbox.GeneratorOption{
			chatterbox.WithSeed(seed),
			chatterbox.WithFormatter(formatter),
		}
		states[name] = &serviceRuntime{
			name:   name,
			cfg:    svc,
			gen:    chatterbox.NewGenerator(schema, genOpts...),
			patch:  defaultPatch(),
			labels: svc.Labels,
		}
	}

	correlate := plan.Correlate
	if ro.Correlate != nil {
		correlate = *ro.Correlate
	}

	return &Runner{
		plan:      plan,
		opts:      ro,
		states:    states,
		shared:    shared,
		correlate: correlate,
	}, nil
}

func buildServiceSchema(svc ServiceConfig) (*chatterbox.Schema, error) {
	if svc.SchemaPath != "" {
		return config.LoadSchemaFile(svc.SchemaPath)
	}
	name, err := preset.ParseName(svc.Preset)
	if err != nil {
		return nil, err
	}
	return preset.Build(name, preset.Options{})
}

func (r *Runner) emitFrom(st *serviceRuntime) ([]byte, error) {
	rec := st.gen.GenerateFields()
	rng := st.gen.Rand()
	applyPatchToRecord(rec, st.patch, rng)
	if r.correlate || st.patch.useSharedCorrelation {
		r.shared.ApplyCorrelation(rec)
	}
	rec["service"] = st.name
	for k, v := range st.labels {
		rec[k] = v
	}
	return st.gen.FormatRecord(rec)
}
