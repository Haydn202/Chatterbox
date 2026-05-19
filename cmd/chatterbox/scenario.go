package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Haydn202/Chatterbox/emit"
	"github.com/Haydn202/Chatterbox/scenario"
)

type scenarioRunConfig struct {
	File       string
	Seed       uint64
	Duration   time.Duration
	Format     string
	OutputMode string
	Output     string
	OutputDir  string
}

func parseScenarioRunFlags(args []string) (scenarioRunConfig, error) {
	fs := flag.NewFlagSet("scenario run", flag.ContinueOnError)
	cfg := scenarioRunConfig{
		OutputMode: "interleaved",
		Output:     "-",
	}
	fs.StringVar(&cfg.File, "file", "", "scenario YAML file")
	fs.StringVar(&cfg.File, "f", "", "scenario YAML file")
	fs.Uint64Var(&cfg.Seed, "seed", 0, "override scenario seed (0 = use file)")
	fs.DurationVar(&cfg.Duration, "duration", 0, "override total duration (0 = use file)")
	fs.StringVar(&cfg.Format, "format", "", "override output format")
	fs.StringVar(&cfg.OutputMode, "output-mode", cfg.OutputMode, "interleaved, split, or both")
	fs.StringVar(&cfg.Output, "output", cfg.Output, "combined output path or - for stdout")
	fs.StringVar(&cfg.Output, "o", cfg.Output, "combined output path or - for stdout")
	fs.StringVar(&cfg.OutputDir, "output-dir", "", "directory for per-service files (split or both)")
	fs.Usage = func() { printScenarioRunUsage(fs) }
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return cfg, errHelp
		}
		return cfg, err
	}
	if cfg.File == "" {
		return cfg, fmt.Errorf("scenario run: -f/--file is required")
	}
	return cfg, nil
}

func executeScenarioRun(cfg scenarioRunConfig) error {
	plan, err := scenario.ParseFile(cfg.File)
	if err != nil {
		return err
	}
	opts := scenario.RunnerOptions{}
	if cfg.Seed != 0 {
		opts.Seed = cfg.Seed
	}
	if cfg.Duration > 0 {
		opts.Duration = cfg.Duration
	}
	if cfg.Format != "" {
		opts.Format = emit.Format(cfg.Format)
	}

	runner, err := scenario.NewRunner(plan, opts)
	if err != nil {
		return err
	}

	mode, err := parseOutputMode(cfg.OutputMode)
	if err != nil {
		return err
	}
	if mode == scenario.OutputSplit || mode == scenario.OutputBoth {
		if cfg.OutputDir == "" {
			return fmt.Errorf("scenario run: --output-dir is required for output-mode %s", cfg.OutputMode)
		}
	}

	writers := scenario.WritersConfig{Mode: mode, OutputDir: cfg.OutputDir}
	if mode == scenario.OutputInterleaved || mode == scenario.OutputBoth {
		out, closeOut, err := openOutput(cfg.Output)
		if err != nil {
			return err
		}
		if closeOut != nil {
			defer closeOut()
		}
		writers.Combined = out
	}

	ctx := context.Background()
	if cfg.Duration == 0 && plan.Duration == 0 {
		var stop context.CancelFunc
		ctx, stop = signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
		defer stop()
	}

	return runner.Run(ctx, writers)
}

func parseOutputMode(s string) (scenario.OutputMode, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "interleaved", "":
		return scenario.OutputInterleaved, nil
	case "split":
		return scenario.OutputSplit, nil
	case "both":
		return scenario.OutputBoth, nil
	default:
		return "", fmt.Errorf("scenario: unknown output-mode %q", s)
	}
}

func printScenarioRunUsage(fs *flag.FlagSet) {
	fmt.Fprintf(os.Stderr, "Usage: chatterbox scenario run [flags]\n\n")
	fmt.Fprintf(os.Stderr, "Run a multi-service failure scenario from YAML.\n\n")
	fs.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nExamples:\n")
	fmt.Fprintf(os.Stderr, "  chatterbox scenario run -f scenarios/cascading-failure.yaml\n")
	fmt.Fprintf(os.Stderr, "  chatterbox scenario run -f scenarios/cascading-failure.yaml --duration 30s\n")
	fmt.Fprintf(os.Stderr, "  chatterbox scenario run -f scenarios/cascading-failure.yaml --output-mode both -o out/combined.jsonl --output-dir out/\n")
}
