package chatterbox

import (
	"math/rand/v2"
	"sync"
)

// SharedCorrelation coordinates trace_id / request_id across multiple generators.
type SharedCorrelation struct {
	mu  sync.Mutex
	cfg CorrelationConfig
	rng *rand.Rand
	st  *correlationState
}

// NewSharedCorrelation creates a pool for cross-service correlation.
func NewSharedCorrelation(cfg CorrelationConfig, seed uint64) *SharedCorrelation {
	return &SharedCorrelation{
		cfg: normalizeCorrelation(cfg),
		rng: rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15)),
	}
}

// WithSharedCorrelation attaches a generator to a shared correlation pool.
func WithSharedCorrelation(pool *SharedCorrelation) GeneratorOption {
	return func(g *Generator) {
		g.sharedCorrelation = pool
	}
}

// ApplyCorrelation writes shared IDs into record using the pool's state.
func (p *SharedCorrelation) ApplyCorrelation(record map[string]any) {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.st == nil {
		p.st = &correlationState{cfg: p.cfg}
	}
	applyCorrelationState(p.st, record, p.rng)
}
