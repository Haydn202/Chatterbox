package chatterbox

import (
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/Haydn202/Chatterbox/fuzz"
)

// CorrelationConfig groups log lines into traces with shared IDs.
type CorrelationConfig struct {
	TraceIDField   string        // default trace_id
	RequestIDField string        // default request_id
	SpanIDField    string        // optional; empty omits span_id
	TimestampField string        // default timestamp; used when TimestampStep > 0
	MinLines       int           // default 2
	MaxLines       int           // default 6
	TimestampStep  time.Duration // add per line within a trace (0 = disabled)
}

type correlationState struct {
	cfg            CorrelationConfig
	linesRemaining int
	traceID        string
	requestID      string
	spanID         string
	traceStart     time.Time
	lineIndex      int
}

// WithCorrelation enables shared trace_id / request_id across consecutive records.
func WithCorrelation(cfg CorrelationConfig) GeneratorOption {
	return func(g *Generator) {
		g.correlation = &correlationState{cfg: normalizeCorrelation(cfg)}
	}
}

func normalizeCorrelation(cfg CorrelationConfig) CorrelationConfig {
	if cfg.TraceIDField == "" {
		cfg.TraceIDField = "trace_id"
	}
	if cfg.RequestIDField == "" {
		cfg.RequestIDField = "request_id"
	}
	if cfg.TimestampField == "" {
		cfg.TimestampField = "timestamp"
	}
	if cfg.MinLines <= 0 {
		cfg.MinLines = 2
	}
	if cfg.MaxLines < cfg.MinLines {
		cfg.MaxLines = cfg.MinLines
	}
	return cfg
}

func (g *Generator) applyCorrelation(record map[string]any) {
	if g.correlation == nil {
		return
	}
	c := g.correlation
	if c.linesRemaining == 0 {
		c.traceID = fuzz.UUID().Generate(g.rng).(string)
		c.requestID = fmt.Sprintf("req-%s", randomHex12(g.rng))
		if c.cfg.SpanIDField != "" {
			c.spanID = fuzz.UUID().Generate(g.rng).(string)
		}
		n := c.cfg.MinLines
		if c.cfg.MaxLines > c.cfg.MinLines {
			n += g.rng.IntN(c.cfg.MaxLines - c.cfg.MinLines + 1)
		}
		c.linesRemaining = n
		c.traceStart = time.Now().UTC()
		c.lineIndex = 0
	}

	record[c.cfg.TraceIDField] = c.traceID
	record[c.cfg.RequestIDField] = c.requestID
	if c.cfg.SpanIDField != "" {
		record[c.cfg.SpanIDField] = c.spanID
	}

	if c.cfg.TimestampStep > 0 {
		if _, ok := record[c.cfg.TimestampField]; ok {
			t := c.traceStart.Add(time.Duration(c.lineIndex) * c.cfg.TimestampStep)
			record[c.cfg.TimestampField] = t.Format(time.RFC3339)
		}
	}

	c.lineIndex++
	c.linesRemaining--
}

func randomHex12(r *rand.Rand) string {
	const hex = "0123456789abcdef"
	b := make([]byte, 12)
	for i := range b {
		b[i] = hex[r.IntN(len(hex))]
	}
	return string(b)
}
