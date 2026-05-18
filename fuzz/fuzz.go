package fuzz

import (
	"math/rand/v2"
	"sort"
)

// Fuzzer produces a single field value from the given random source.
type Fuzzer interface {
	Generate(r *rand.Rand) any
}

// Func adapts a function to Fuzzer.
type Func func(r *rand.Rand) any

func (f Func) Generate(r *rand.Rand) any { return f(r) }

// Choice picks uniformly from values.
func Choice(values ...any) Fuzzer {
	return Func(func(r *rand.Rand) any {
		return values[r.IntN(len(values))]
	})
}

// Weighted picks from values using proportional weights (need not sum to 1).
func Weighted(values map[string]float64) Fuzzer {
	keys := make([]string, 0, len(values))
	for k, w := range values {
		if w > 0 {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	weights := make([]float64, len(keys))
	var total float64
	for i, k := range keys {
		weights[i] = values[k]
		total += values[k]
	}
	if len(keys) == 0 {
		return Func(func(*rand.Rand) any { return "" })
	}
	return Func(func(r *rand.Rand) any {
		p := r.Float64() * total
		var cum float64
		for i, w := range weights {
			cum += w
			if p < cum {
				return keys[i]
			}
		}
		return keys[len(keys)-1]
	})
}

// Optional returns value with probability p, otherwise nil.
func Optional(p float64, inner Fuzzer) Fuzzer {
	return Func(func(r *rand.Rand) any {
		if r.Float64() < p {
			return inner.Generate(r)
		}
		return nil
	})
}
