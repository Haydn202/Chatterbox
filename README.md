# Chatterbox

Synthetic, **fuzzy** log lines for testing log aggregation pipelines, indexers, and PII rules. Define a schema in Go, attach fuzzers per field, and stream reproducible JSON Lines.

## Install

```bash
go get github.com/Haydn202/Chatterbox
```

Requires Go 1.22+ (`math/rand/v2`).

## Quickstart

```go
package main

import (
	"os"

	"github.com/Haydn202/Chatterbox"
	"github.com/Haydn202/Chatterbox/fuzz"
)

func main() {
	schema := chatterbox.NewSchema(
		chatterbox.MakeField("timestamp", fuzz.TimestampRFC3339(fuzz.WithJitter(30))),
		chatterbox.MakeField("level", fuzz.LevelWeighted(map[string]float64{
			"info": 0.7, "warn": 0.2, "error": 0.1,
		})),
		chatterbox.MakeField("email", fuzz.Email()),
		chatterbox.MakeField("client_ip", fuzz.IPv4()),
		chatterbox.MakeField("message", fuzz.StringFrom(10, 120)),
	)

	gen := chatterbox.NewGenerator(schema, chatterbox.WithSeed(42))
	_ = gen.WriteN(os.Stdout, 1000)
}
```

## API

| Type | Role |
|------|------|
| `Schema` / `MakeField` | Ordered log fields and fuzzers |
| `Generator` | `Next()`, `NextFormatted()`, `NextJSON()`, `NextN()`, `WriteN()` |
| `Stream` | `Run(ctx, w)` — emit at a rate for a duration or until cancelled |
| `fuzz.Fuzzer` | Pluggable value generation |
| `emit.Formatter` | Encode records (JSONL default, or multiline text) |

Use `chatterbox.WithSeed(uint64)` for reproducible sequences in tests.

## Rate-limited streaming (live servers)

Use `Stream` to emit logs at a fixed rate. Omit duration (or pass zero) to run until the context is cancelled—useful against live aggregators. Pass `context` cancellation via `signal.NotifyContext` for Ctrl+C.

```go
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Haydn202/Chatterbox"
	"github.com/Haydn202/Chatterbox/fuzz"
)

func main() {
	schema := chatterbox.NewSchema(/* fields ... */)
	gen := chatterbox.NewGenerator(schema, chatterbox.WithSeed(42))

	// 25 logs/sec for 10 minutes, then stop.
	stream, err := chatterbox.NewStream(gen, 25,
		chatterbox.WithStreamDuration(10*time.Minute),
	)
	if err != nil {
		panic(err)
	}

	// Or omit WithStreamDuration to run until interrupted:
	// stream, _ := chatterbox.NewStream(gen, 25)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	_ = stream.Run(ctx, os.Stdout)
}
```

| Setting | Behavior |
|---------|----------|
| `Rate` | Logs per second (required, > 0) |
| `WithStreamDuration(0)` | Run until `ctx` is cancelled (default) |
| `WithStreamDuration(d)` | Stop after `d` elapses |

## Multiline / stack traces

For testing filebeat-style **multiline** rules, use `fuzz.StackTrace` with `emit.TextMultilineFormatter`. One log event spans multiple physical lines: a header line, then the stack body.

```go
schema := chatterbox.NewSchema(
    chatterbox.MakeField("timestamp", fuzz.TimestampRFC3339(fuzz.WithJitter(30))),
    chatterbox.MakeField("level", fuzz.LevelWeighted(map[string]float64{
        "error": 1.0,
    })),
    chatterbox.MakeField("message", fuzz.StringFrom(10, 80)),
    chatterbox.MakeField("stacktrace", fuzz.StackTrace()),
)

gen := chatterbox.NewGenerator(schema,
    chatterbox.WithSeed(42),
    chatterbox.WithFormatter(emit.TextMultilineFormatter(emit.TextMultilineConfig{
        HeaderFields: []string{"timestamp", "level", "message"},
        BodyFields:   []string{"stacktrace"},
    })),
)
_ = gen.WriteN(os.Stdout, 100)
```

JSONL mode (default) still works: stack traces are a single JSON string field with escaped `\n` characters.

## Built-in fuzzers

| Fuzzer | Description |
|--------|-------------|
| `fuzz.Email(opts...)` | Varied local parts, domains, typos; optional edge cases |
| `fuzz.TimestampRFC3339(opts...)` | RFC3339 time; `WithJitter`, `WithBaseTime` |
| `fuzz.LevelWeighted(map)` | Weighted log levels |
| `fuzz.StringFrom(min, max)` | Random alphanumeric message |
| `fuzz.UUID()` | UUID v4-style string |
| `fuzz.IPv4(opts...)` | IPv4; optional private ranges |
| `fuzz.URL()` | HTTP/HTTPS URLs |
| `fuzz.Choice(...)` | Uniform choice |
| `fuzz.Weighted(map)` | Weighted string choice |
| `fuzz.Optional(p, inner)` | Sometimes nil |
| `fuzz.StackTrace(opts...)` | Multiline stack trace (`go`, `java`, `python` styles) |

### Stack trace options

- `fuzz.WithStackStyle("go")` — `go`, `java`, or `python`
- `fuzz.WithFrames(min, max)` — frame count (default 3–12)
- `fuzz.WithPanicMessages([]string)` — custom first-line messages

### Email options

- `fuzz.WithTypoRate(0.05)` — adjacent swaps, missing dots, etc.
- `fuzz.WithEdgeCases(true)` — invalid-but-plausible addresses (`@@`, trailing `.`, …)

## Testing

```bash
go test ./...
```

Golden output: `testdata/golden.jsonl`, `testdata/golden-multiline.txt`. Regenerate with:

```bash
# PowerShell
$env:UPDATE_GOLDEN="1"; go test ./...
```

## Adding a fuzzer

1. Implement `fuzz.Fuzzer` (`Generate(*rand.Rand) any`) in `fuzz/`.
2. Add table tests in `fuzz/*_test.go` with a fixed `rand.NewPCG` seed.
3. Document options and defaults in godoc.

## Roadmap

- **Config-driven schemas** (YAML/JSON) parsed into the same `Schema` type via a fuzzer registry.
- Additional emitters (logfmt, syslog, ECS-shaped JSON).

## License

See repository license (add as needed).
