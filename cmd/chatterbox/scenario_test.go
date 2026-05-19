package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestExecuteScenarioRun_short(t *testing.T) {
	root := filepath.Join("..", "..")
	scenarioFile := filepath.Join(root, "scenarios", "cascading-failure.yaml")
	if _, err := os.Stat(scenarioFile); err != nil {
		t.Skip(scenarioFile)
	}
	outDir := t.TempDir()
	cfg := scenarioRunConfig{
		File:       scenarioFile,
		Duration:   300 * time.Millisecond,
		OutputMode: "both",
		Output:     filepath.Join(outDir, "combined.jsonl"),
		OutputDir:  filepath.Join(outDir, "split"),
	}
	if err := executeScenarioRun(cfg); err != nil {
		t.Fatal(err)
	}
	combined, err := os.ReadFile(cfg.Output)
	if err != nil {
		t.Fatal(err)
	}
	if len(combined) == 0 {
		t.Fatal("combined output empty")
	}
	for _, svc := range []string{"api", "auth", "redis", "postgres"} {
		p := filepath.Join(cfg.OutputDir, svc+".jsonl")
		if _, err := os.Stat(p); err != nil {
			t.Errorf("missing split file %s: %v", p, err)
		}
	}
	seen := make(map[string]bool)
	for _, line := range strings.Split(strings.TrimSpace(string(combined)), "\n") {
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
	if len(seen) == 0 {
		t.Fatal("no service field in combined output")
	}
}
