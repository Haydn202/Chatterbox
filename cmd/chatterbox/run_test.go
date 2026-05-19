package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestExecuteRun_batchJSON(t *testing.T) {
	path := t.TempDir() + "/out.jsonl"
	cfg := runConfig{
		Output:     path,
		Count:      5,
		Format:     "json",
		Seed:       42,
		PresetName: "minimal",
	}
	if err := executeRun(cfg); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 5 {
		t.Fatalf("expected 5 lines, got %d", len(lines))
	}
	for i, line := range lines {
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("line %d: %v", i, err)
		}
	}
}

func TestExecuteRun_correlateBatch(t *testing.T) {
	path := t.TempDir() + "/corr.jsonl"
	cfg := runConfig{
		Output:     path,
		Count:      20,
		Format:     "json",
		Seed:       99,
		PresetName: "default",
		Correlate:  boolPtr(true),
	}
	if err := executeRun(cfg); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	var lastTrace string
	run := 1
	for _, line := range lines {
		var m map[string]any
		_ = json.Unmarshal([]byte(line), &m)
		tid, _ := m["trace_id"].(string)
		if lastTrace == "" {
			lastTrace = tid
			continue
		}
		if tid == lastTrace {
			run++
		} else {
			if run < 3 || run > 8 {
				t.Fatalf("run length %d out of expected 3-8", run)
			}
			run = 1
			lastTrace = tid
		}
	}
}

func TestParseRunFlags_email(t *testing.T) {
	cfg, err := parseRunFlags([]string{"--email", "-n", "1"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Email == nil || !*cfg.Email {
		t.Fatal("expected email true")
	}
}

func TestRunVersion(t *testing.T) {
	if err := runVersion(); err != nil {
		t.Fatal(err)
	}
}
