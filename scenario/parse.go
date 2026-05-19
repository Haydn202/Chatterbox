package scenario

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Haydn202/Chatterbox/emit"
	"gopkg.in/yaml.v3"
)

// Plan is a validated scenario ready to run.
type Plan struct {
	Name                 string
	Seed                 uint64
	Duration             time.Duration
	Format               emit.Format
	Rate                 float64
	DefaultPhaseDuration time.Duration
	Correlate            bool // trace_id / request_id on every line (default true)
	Services             map[string]ServiceConfig
	Phases               []Phase
}

// ServiceConfig binds a service name to schema source and emission weight.
type ServiceConfig struct {
	Name       string
	Preset     string
	SchemaPath string
	RateWeight float64
	Labels     map[string]string
}

// Phase is one timeline segment with an event and target services.
type Phase struct {
	Event     string
	Duration  time.Duration
	Services  []string // empty = all services
}

type scenarioDoc struct {
	Scenario struct {
		Name                 string  `yaml:"name"`
		Seed                 uint64  `yaml:"seed"`
		Duration             string  `yaml:"duration"`
		Format               string  `yaml:"format"`
		Rate                 float64 `yaml:"rate"`
		DefaultPhaseDuration string  `yaml:"default_phase_duration"`
		Correlate            *bool   `yaml:"correlate"`
	} `yaml:"scenario"`
	Services map[string]serviceDoc `yaml:"services"`
	Timeline []phaseDoc            `yaml:"timeline"`
	Events   []string              `yaml:"events"`
}

type serviceDoc struct {
	Preset     string            `yaml:"preset"`
	Schema     string            `yaml:"schema"`
	RateWeight float64           `yaml:"rate_weight"`
	Labels     map[string]string `yaml:"labels"`
}

type phaseDoc struct {
	Event    string   `yaml:"event"`
	Duration string   `yaml:"duration"`
	Services []string `yaml:"services"`
}

// Parse reads and validates a scenario YAML file.
func Parse(r io.Reader, baseDir string) (*Plan, error) {
	var doc scenarioDoc
	dec := yaml.NewDecoder(r)
	dec.KnownFields(false)
	if err := dec.Decode(&doc); err != nil {
		return nil, fmt.Errorf("scenario: parse: %w", err)
	}
	if len(doc.Services) == 0 {
		return nil, fmt.Errorf("scenario: at least one service is required")
	}

	plan := &Plan{
		Name:   strings.TrimSpace(doc.Scenario.Name),
		Seed:   doc.Scenario.Seed,
		Format: emit.Format(doc.Scenario.Format),
		Rate:   doc.Scenario.Rate,
	}
	if plan.Seed == 0 {
		plan.Seed = 1
	}
	if plan.Format == "" {
		plan.Format = emit.FormatJSON
	}
	if plan.Rate <= 0 {
		plan.Rate = 25
	}
	if doc.Scenario.Correlate == nil {
		plan.Correlate = true
	} else {
		plan.Correlate = *doc.Scenario.Correlate
	}
	if doc.Scenario.Duration != "" {
		d, err := time.ParseDuration(doc.Scenario.Duration)
		if err != nil {
			return nil, fmt.Errorf("scenario: duration: %w", err)
		}
		plan.Duration = d
	}
	if doc.Scenario.DefaultPhaseDuration != "" {
		d, err := time.ParseDuration(doc.Scenario.DefaultPhaseDuration)
		if err != nil {
			return nil, fmt.Errorf("scenario: default_phase_duration: %w", err)
		}
		plan.DefaultPhaseDuration = d
	}
	if plan.DefaultPhaseDuration == 0 {
		plan.DefaultPhaseDuration = 3 * time.Minute
	}

	plan.Services = make(map[string]ServiceConfig, len(doc.Services))
	var weightSum float64
	for name, svc := range doc.Services {
		name = strings.TrimSpace(name)
		if name == "" {
			return nil, fmt.Errorf("scenario: service name is required")
		}
		preset := strings.TrimSpace(svc.Preset)
		schemaPath := strings.TrimSpace(svc.Schema)
		if preset == "" && schemaPath == "" {
			return nil, fmt.Errorf("scenario: service %q: preset or schema is required", name)
		}
		if preset != "" && schemaPath != "" {
			return nil, fmt.Errorf("scenario: service %q: use preset or schema, not both", name)
		}
		if schemaPath != "" && !filepath.IsAbs(schemaPath) && baseDir != "" {
			schemaPath = filepath.Join(baseDir, schemaPath)
		}
		w := svc.RateWeight
		if w <= 0 {
			w = 1.0 / float64(len(doc.Services))
		}
		weightSum += w
		labels := make(map[string]string, len(svc.Labels))
		for k, v := range svc.Labels {
			labels[k] = v
		}
		plan.Services[name] = ServiceConfig{
			Name:       name,
			Preset:     preset,
			SchemaPath: schemaPath,
			RateWeight: w,
			Labels:     labels,
		}
	}
	normalizeWeights(plan.Services, weightSum)

	if err := expandTimeline(&doc, plan); err != nil {
		return nil, err
	}
	if len(plan.Phases) == 0 {
		return nil, fmt.Errorf("scenario: timeline is empty")
	}
	if plan.Duration == 0 {
		for _, ph := range plan.Phases {
			plan.Duration += ph.Duration
		}
	}
	return plan, nil
}

func normalizeWeights(services map[string]ServiceConfig, sum float64) {
	if sum <= 0 {
		return
	}
	for name, svc := range services {
		svc.RateWeight /= sum
		services[name] = svc
	}
}

func expandTimeline(doc *scenarioDoc, plan *Plan) error {
	if len(doc.Timeline) > 0 {
		for i, ph := range doc.Timeline {
			event := strings.TrimSpace(ph.Event)
			if event == "" {
				return fmt.Errorf("scenario: timeline[%d]: event is required", i)
			}
			dur := plan.DefaultPhaseDuration
			if ph.Duration != "" {
				parsed, err := time.ParseDuration(ph.Duration)
				if err != nil {
					return fmt.Errorf("scenario: timeline[%d] duration: %w", i, err)
				}
				dur = parsed
			}
			plan.Phases = append(plan.Phases, Phase{
				Event:    event,
				Duration: dur,
				Services: append([]string(nil), ph.Services...),
			})
		}
		return nil
	}
	if len(doc.Events) == 0 {
		return fmt.Errorf("scenario: timeline or events is required")
	}
	plan.Phases = append(plan.Phases, Phase{
		Event:    "baseline",
		Duration: plan.DefaultPhaseDuration,
	})
	for _, ev := range doc.Events {
		ev = strings.TrimSpace(ev)
		if ev == "" {
			continue
		}
		plan.Phases = append(plan.Phases, Phase{
			Event:    ev,
			Duration: plan.DefaultPhaseDuration,
		})
	}
	return nil
}

// ParseFile loads a scenario from a YAML file.
func ParseFile(path string) (*Plan, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	baseDir := filepath.Dir(path)
	plan, err := Parse(f, baseDir)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return plan, nil
}
