package config

import (
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/Haydn202/Chatterbox/fuzz"
)

func buildFuzzer(f fieldYAML) (fuzz.Fuzzer, error) {
	typ := strings.ToLower(strings.TrimSpace(f.Type))
	if typ == "" {
		return nil, fmt.Errorf("config: field %q: type is required", f.Name)
	}
	switch typ {
	case "timestamp_rfc3339":
		var opts []fuzz.TimestampOption
		if f.JitterSeconds > 0 {
			opts = append(opts, fuzz.WithJitter(f.JitterSeconds))
		}
		return fuzz.TimestampRFC3339(opts...), nil
	case "level_weighted":
		if len(f.Weights) == 0 {
			return nil, fmt.Errorf("config: field %q: level_weighted requires weights", f.Name)
		}
		if err := validateWeights(f.Name, f.Weights); err != nil {
			return nil, err
		}
		return fuzz.LevelWeighted(f.Weights), nil
	case "weighted":
		if len(f.Weights) == 0 {
			return nil, fmt.Errorf("config: field %q: weighted requires weights", f.Name)
		}
		return fuzz.Weighted(f.Weights), nil
	case "string":
		min, max := f.MinLen, f.MaxLen
		if min == 0 && max == 0 {
			min, max = 10, 120
		}
		if max < min {
			return nil, fmt.Errorf("config: field %q: max_len must be >= min_len", f.Name)
		}
		return fuzz.StringFrom(min, max), nil
	case "email":
		return fuzz.Email(), nil
	case "uuid":
		return fuzz.UUID(), nil
	case "ipv4":
		var opts []fuzz.IPv4Option
		if f.AllowPrivate != nil {
			opts = append(opts, fuzz.WithPrivateRange(*f.AllowPrivate))
		}
		return fuzz.IPv4(opts...), nil
	case "url":
		return fuzz.URL(), nil
	case "choice":
		if len(f.Values) == 0 {
			return nil, fmt.Errorf("config: field %q: choice requires values", f.Name)
		}
		return fuzz.Choice(f.Values...), nil
	case "constant":
		v := f.Value
		if v == nil {
			return nil, fmt.Errorf("config: field %q: constant requires value", f.Name)
		}
		return fuzz.Func(func(*rand.Rand) any { return v }), nil
	case "optional":
		if f.Inner == nil {
			return nil, fmt.Errorf("config: field %q: optional requires inner", f.Name)
		}
		p := f.Probability
		if p <= 0 {
			p = 0.15
		}
		inner, err := buildFuzzer(*f.Inner)
		if err != nil {
			return nil, err
		}
		return fuzz.Optional(p, inner), nil
	case "stacktrace":
		var opts []fuzz.StackTraceOption
		if lang := strings.TrimSpace(f.Lang); lang != "" {
			opts = append(opts, fuzz.WithStackStyle(lang))
		}
		minF, maxF := f.MinFrames, f.MaxFrames
		if minF > 0 || maxF > 0 {
			if minF <= 0 {
				minF = 3
			}
			if maxF <= 0 {
				maxF = 12
			}
			opts = append(opts, fuzz.WithFrames(minF, maxF))
		}
		return fuzz.StackTrace(opts...), nil
	default:
		return nil, fmt.Errorf("config: field %q: unknown type %q", f.Name, f.Type)
	}
}

func validateWeights(field string, weights map[string]float64) error {
	var sum float64
	for _, w := range weights {
		if w < 0 {
			return fmt.Errorf("config: field %q: negative weight", field)
		}
		sum += w
	}
	if sum <= 0 {
		return fmt.Errorf("config: field %q: weights must sum to a positive value", field)
	}
	return nil
}
