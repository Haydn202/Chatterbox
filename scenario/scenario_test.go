package scenario_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Haydn202/Chatterbox/scenario"
)

func TestParse_shorthandEvents(t *testing.T) {
	const doc = `
scenario:
  name: test
  default_phase_duration: 1s
services:
  api: { preset: api }
  redis: { preset: minimal }
events:
  - redis_latency_spike
`
	plan, err := scenario.Parse(strings.NewReader(doc), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Phases) < 2 {
		t.Fatalf("phases = %d, want >= 2", len(plan.Phases))
	}
	if plan.Phases[0].Event != "baseline" {
		t.Fatalf("first phase = %q", plan.Phases[0].Event)
	}
}

func TestRunner_allServices(t *testing.T) {
	path := filepath.Join("..", "scenarios", "cascading-failure.yaml")
	if _, err := os.Stat(path); err != nil {
		t.Skip("scenarios/cascading-failure.yaml not found yet")
	}
	plan, err := scenario.ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	plan.Duration = 500 * time.Millisecond
	for i := range plan.Phases {
		plan.Phases[i].Duration = 100 * time.Millisecond
	}
	runner, err := scenario.NewRunner(plan, scenario.RunnerOptions{
		Rate: 50,
	})
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err = runner.Run(ctx, scenario.WritersConfig{
		Mode:     scenario.OutputInterleaved,
		Combined: &buf,
	})
	if err != nil {
		t.Fatal(err)
	}
	seen := make(map[string]bool)
	for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		if line == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("json: %v", err)
		}
		if s, _ := m["service"].(string); s != "" {
			seen[s] = true
		}
	}
	for _, svc := range []string{"api", "auth", "redis", "postgres"} {
		if !seen[svc] {
			t.Errorf("output missing service %q", svc)
		}
	}
}

func TestRunner_correlationIDs(t *testing.T) {
	const doc = `
scenario:
  name: corr-test
  correlate: true
  default_phase_duration: 1s
services:
  api: { preset: minimal }
timeline:
  - event: baseline
    duration: 1s
`
	plan, err := scenario.Parse(strings.NewReader(doc), "")
	if err != nil {
		t.Fatal(err)
	}
	plan.Duration = 200 * time.Millisecond
	runner, err := scenario.NewRunner(plan, scenario.RunnerOptions{Rate: 40})
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := runner.Run(ctx, scenario.WritersConfig{
		Mode:     scenario.OutputInterleaved,
		Combined: &buf,
	}); err != nil {
		t.Fatal(err)
	}
	foundTrace := false
	for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		if line == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatal(err)
		}
		if tid, _ := m["trace_id"].(string); tid != "" {
			foundTrace = true
			if _, ok := m["request_id"].(string); !ok {
				t.Fatal("missing request_id")
			}
			break
		}
	}
	if !foundTrace {
		t.Fatal("expected trace_id on correlated scenario output")
	}
}
