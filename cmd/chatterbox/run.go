package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Haydn202/Chatterbox"
	"github.com/Haydn202/Chatterbox/emit"
	"github.com/Haydn202/Chatterbox/preset"
	"github.com/Haydn202/Chatterbox/schedule"
)

type runConfig struct {
	Output            string
	Count             int
	Rate              float64
	Duration          time.Duration
	Format            string
	Seed              uint64
	PresetName        string
	Email             *bool
	Stacktrace        *bool
	Correlate         *bool
	NoCorrelate       bool
	Burst             bool
	BurstRate         float64
	BurstBaseDuration time.Duration
	BurstPeakDuration time.Duration
}

func parseRunFlags(args []string) (runConfig, error) {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	cfg := runConfig{
		Rate:              10,
		Format:            "json",
		Seed:              1,
		PresetName:        "default",
		BurstRate:         150,
		BurstBaseDuration: time.Minute,
		BurstPeakDuration: 30 * time.Second,
	}
	fs.StringVar(&cfg.Output, "output", "-", "output file path or - for stdout")
	fs.StringVar(&cfg.Output, "o", "-", "output file path or - for stdout")
	fs.IntVar(&cfg.Count, "count", 0, "number of logs to emit (batch mode; 0 = stream)")
	fs.IntVar(&cfg.Count, "n", 0, "number of logs to emit (batch mode; 0 = stream)")
	fs.Float64Var(&cfg.Rate, "rate", cfg.Rate, "logs per second when streaming")
	fs.Float64Var(&cfg.Rate, "r", cfg.Rate, "logs per second when streaming")
	fs.DurationVar(&cfg.Duration, "duration", 0, "max stream duration (0 = until interrupted)")
	fs.DurationVar(&cfg.Duration, "d", 0, "max stream duration (0 = until interrupted)")
	fs.StringVar(&cfg.Format, "format", cfg.Format, "output format: json, logfmt, plain, syslog, cef, multiline, slog_json, zap_json, zerolog_json")
	fs.Uint64Var(&cfg.Seed, "seed", cfg.Seed, "PRNG seed")
	fs.StringVar(&cfg.PresetName, "preset", cfg.PresetName, "schema preset: default, api, minimal, multiline-error")

	var email, noEmail, stacktrace, correlate, noCorrelate bool
	fs.BoolVar(&email, "email", false, "include email field")
	fs.BoolVar(&noEmail, "no-email", false, "omit email field")
	fs.BoolVar(&stacktrace, "stacktrace", false, "include optional stacktrace field")
	fs.BoolVar(&correlate, "correlate", false, "enable trace_id / request_id correlation")
	fs.BoolVar(&noCorrelate, "no-correlate", false, "disable correlation")

	fs.BoolVar(&cfg.Burst, "burst", false, "use incident spike schedule")
	fs.Float64Var(&cfg.BurstRate, "burst-rate", cfg.BurstRate, "peak logs/sec when --burst")
	fs.DurationVar(&cfg.BurstBaseDuration, "burst-base-duration", cfg.BurstBaseDuration, "duration at base rate before spike")
	fs.DurationVar(&cfg.BurstPeakDuration, "burst-peak-duration", cfg.BurstPeakDuration, "duration at peak rate")

	fs.Usage = func() { printRunUsage(fs) }
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return cfg, errHelp
		}
		return cfg, err
	}

	if email && noEmail {
		return cfg, fmt.Errorf("cannot use --email and --no-email together")
	}
	if email {
		cfg.Email = boolPtr(true)
	}
	if noEmail {
		cfg.Email = boolPtr(false)
	}
	if stacktrace {
		cfg.Stacktrace = boolPtr(true)
	}
	if correlate {
		cfg.Correlate = boolPtr(true)
	}
	if noCorrelate {
		cfg.NoCorrelate = true
	}
	return cfg, nil
}

func executeRun(cfg runConfig) error {
	presetName, err := preset.ParseName(cfg.PresetName)
	if err != nil {
		return err
	}
	opt := preset.Options{
		Email:      cfg.Email,
		Stacktrace: cfg.Stacktrace,
		Correlate:  cfg.Correlate,
	}
	if cfg.NoCorrelate {
		f := false
		opt.Correlate = &f
	}

	schema, err := preset.Build(presetName, opt)
	if err != nil {
		return err
	}

	emitOpts, err := formatterOptions(cfg.Format, presetName, opt)
	if err != nil {
		return err
	}
	formatOpt, err := chatterbox.WithOutputFormat(emit.Format(cfg.Format), emitOpts)
	if err != nil {
		return err
	}

	genOpts := []chatterbox.GeneratorOption{chatterbox.WithSeed(cfg.Seed), formatOpt}
	if preset.CorrelateEnabled(presetName, opt) {
		genOpts = append(genOpts, chatterbox.WithCorrelation(chatterbox.CorrelationConfig{
			MinLines:      3,
			MaxLines:      8,
			TimestampStep: 2 * time.Millisecond,
		}))
	}
	gen := chatterbox.NewGenerator(schema, genOpts...)

	out, closeOut, err := openOutput(cfg.Output)
	if err != nil {
		return err
	}
	if closeOut != nil {
		defer closeOut()
	}

	if cfg.Count > 0 {
		return gen.WriteN(out, cfg.Count)
	}

	ctx := context.Background()
	if cfg.Duration == 0 {
		var stop context.CancelFunc
		ctx, stop = signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
		defer stop()
	} else {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.Duration)
		defer cancel()
	}

	var sched schedule.Schedule
	if cfg.Burst {
		sched, err = schedule.PresetIncidentSpike(cfg.Rate, cfg.BurstRate, cfg.BurstBaseDuration, cfg.BurstPeakDuration)
	} else {
		sched, err = schedule.FlatRate(cfg.Rate)
	}
	if err != nil {
		return err
	}

	stream, err := chatterbox.NewStreamWithSchedule(gen, sched)
	if err != nil {
		return err
	}
	return stream.Run(ctx, out)
}

func formatterOptions(format string, name preset.Name, opt preset.Options) (emit.Options, error) {
	f := strings.ToLower(format)
	if f != "multiline" {
		return emit.Options{}, nil
	}
	stack := preset.Defaults(name)
	if opt.Stacktrace != nil {
		stack.Stacktrace = opt.Stacktrace
	}
	if !mergePresetBool(stack.Stacktrace, opt.Stacktrace) {
		return emit.Options{}, fmt.Errorf("format multiline requires stacktrace field (--stacktrace or preset multiline-error)")
	}
	return emit.Options{
		Multiline: &emit.TextMultilineConfig{
			HeaderFields: []string{"timestamp", "level", "message"},
			BodyFields:   []string{"stacktrace"},
		},
	}, nil
}

func mergePresetBool(def, override *bool) bool {
	if override != nil {
		return *override
	}
	if def != nil {
		return *def
	}
	return false
}

func openOutput(path string) (io.Writer, func(), error) {
	if path == "" || path == "-" {
		return os.Stdout, nil, nil
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, nil, err
	}
	return f, func() { _ = f.Close() }, nil
}

func boolPtr(b bool) *bool { return &b }

func printRunUsage(fs *flag.FlagSet) {
	fmt.Fprintf(os.Stderr, "Usage: chatterbox run [flags]\n\n")
	fmt.Fprintf(os.Stderr, "Generate fuzzy logs. Use -n for batch count; omit -n to stream at -r.\n\n")
	fs.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nPresets: default, api, minimal, multiline-error\n\nExamples:\n")
	fmt.Fprintf(os.Stderr, "  chatterbox run -n 1000 --format json\n")
	fmt.Fprintf(os.Stderr, "  chatterbox run -r 25 --duration 5m --preset api\n")
	fmt.Fprintf(os.Stderr, "  chatterbox run --preset multiline-error --format multiline --stacktrace\n")
	fmt.Fprintf(os.Stderr, "  chatterbox run --burst --rate 10 --burst-rate 150\n")
}
