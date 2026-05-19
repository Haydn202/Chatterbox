package chatterbox

import (
	"io"
	"math/rand/v2"

	"github.com/Haydn202/Chatterbox/emit"
)

// Generator produces log lines from a schema.
type Generator struct {
	schema            *Schema
	rng               *rand.Rand
	formatter         emit.Formatter
	correlation       *correlationState
	sharedCorrelation *SharedCorrelation
}

// GeneratorOption configures a Generator.
type GeneratorOption func(*Generator)

// WithSeed sets the PRNG seed for reproducible output.
func WithSeed(seed uint64) GeneratorOption {
	return func(g *Generator) {
		g.rng = rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))
	}
}

// WithFormatter sets the output formatter (default JSONL).
func WithFormatter(f emit.Formatter) GeneratorOption {
	return func(g *Generator) {
		g.formatter = f
	}
}

// NewGenerator creates a generator for the given schema.
func NewGenerator(schema *Schema, opts ...GeneratorOption) *Generator {
	g := &Generator{
		schema:    schema,
		rng:       rand.New(rand.NewPCG(1, 1)),
		formatter: emit.JSONLFormatter{},
	}
	for _, o := range opts {
		o(g)
	}
	return g
}

// GenerateFields builds a record from fuzzers without correlation IDs.
func (g *Generator) GenerateFields() map[string]any {
	record := make(map[string]any, len(g.schema.fields))
	for _, f := range g.schema.fields {
		record[f.Name] = f.Fuzzer.Generate(g.rng)
	}
	return record
}

// Next generates one log record as a field map.
func (g *Generator) Next() map[string]any {
	record := g.GenerateFields()
	g.applyCorrelation(record)
	return record
}

// NextFormatted returns one encoded event using the configured formatter.
func (g *Generator) NextFormatted() ([]byte, error) {
	return g.FormatRecord(g.Next())
}

// FormatRecord encodes an already-built record.
func (g *Generator) FormatRecord(record map[string]any) ([]byte, error) {
	return g.formatter.Format(record)
}

// Rand returns the generator PRNG (for field overrides in scenario mode).
func (g *Generator) Rand() *rand.Rand {
	return g.rng
}

// NextJSON returns one JSONL-encoded line (with trailing newline).
func (g *Generator) NextJSON() ([]byte, error) {
	b, err := emit.JSONL(g.Next())
	if err != nil {
		return nil, err
	}
	out := make([]byte, len(b)+1)
	copy(out, b)
	out[len(b)] = '\n'
	return out, nil
}

// NextN returns n formatted events (each as returned by the formatter).
func (g *Generator) NextN(n int) ([][]byte, error) {
	lines := make([][]byte, 0, n)
	for i := 0; i < n; i++ {
		line, err := g.NextFormatted()
		if err != nil {
			return nil, err
		}
		lines = append(lines, line)
	}
	return lines, nil
}

// WriteN writes n formatted records to w.
func (g *Generator) WriteN(w io.Writer, n int) error {
	for i := 0; i < n; i++ {
		b, err := g.NextFormatted()
		if err != nil {
			return err
		}
		if _, err := w.Write(b); err != nil {
			return err
		}
	}
	return nil
}
