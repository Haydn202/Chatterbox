package fuzz

import (
	"math/rand/v2"
	"strings"
)

// EmailOption configures Email fuzzer behavior.
type EmailOption func(*emailFuzzer)

// WithTypoRate sets the probability of introducing a typo in the address (default 0.05).
func WithTypoRate(rate float64) EmailOption {
	return func(e *emailFuzzer) { e.typoRate = rate }
}

// WithEdgeCases enables occasional invalid-but-plausible addresses (default false).
func WithEdgeCases(enable bool) EmailOption {
	return func(e *emailFuzzer) { e.edgeCases = enable }
}

type emailFuzzer struct {
	typoRate  float64
	edgeCases bool
}

// Email generates varied, mostly-valid email addresses.
func Email(opts ...EmailOption) Fuzzer {
	e := &emailFuzzer{typoRate: 0.05}
	for _, o := range opts {
		o(e)
	}
	return e
}

func (e *emailFuzzer) Generate(r *rand.Rand) any {
	if e.edgeCases && r.Float64() < 0.08 {
		return e.edgeCase(r)
	}
	local := e.localPart(r)
	domain := e.domain(r)
	addr := local + "@" + domain
	if r.Float64() < e.typoRate {
		addr = e.applyTypo(addr, r)
	}
	return addr
}

var (
	localBases = []string{"user", "jane.doe", "admin", "support", "test", "alice", "bob.smith", "noreply"}
	domains    = []string{"example.com", "mail.io", "corp.net", "service.org", "app.dev", "logs.test"}
	tlds       = []string{"com", "net", "org", "io", "dev"}
)

func (e *emailFuzzer) localPart(r *rand.Rand) string {
	base := localBases[r.IntN(len(localBases))]
	switch r.IntN(4) {
	case 0:
		return strings.ToLower(base)
	case 1:
		return mixedCase(base, r)
	case 2:
		return base + "+" + randomAlpha(r, 4, 8)
	case 3:
		return base + string(rune('0'+r.IntN(10))) + randomAlpha(r, 0, 3)
	default:
		return base
	}
}

func (e *emailFuzzer) domain(r *rand.Rand) string {
	d := domains[r.IntN(len(domains))]
	if r.Float64() < 0.3 {
		sub := randomAlpha(r, 3, 8)
		parts := strings.SplitN(d, ".", 2)
		if len(parts) == 2 {
			return sub + "." + parts[1]
		}
		return sub + "." + d
	}
	if r.Float64() < 0.15 {
		return randomAlpha(r, 5, 10) + "." + tlds[r.IntN(len(tlds))]
	}
	return d
}

func (e *emailFuzzer) applyTypo(s string, r *rand.Rand) string {
	if len(s) < 2 {
		return s
	}
	switch r.IntN(3) {
	case 0:
		i := r.IntN(len(s) - 1)
		return s[:i] + s[i+1:i+2] + s[i:i+1] + s[i+2:]
	case 1:
		return strings.Replace(s, ".", "", 1)
	case 2:
		i := r.IntN(len(s))
		return s[:i] + s[i:]
	default:
		return s
	}
}

func (e *emailFuzzer) edgeCase(r *rand.Rand) string {
	cases := []string{
		"user@@example.com",
		"user@example.com.",
		".user@example.com",
		"user@",
		"@example.com",
		"user name@example.com",
	}
	return cases[r.IntN(len(cases))]
}

func mixedCase(s string, r *rand.Rand) string {
	var b strings.Builder
	for _, c := range s {
		if c >= 'a' && c <= 'z' && r.IntN(2) == 0 {
			b.WriteRune(c - 32)
		} else {
			b.WriteRune(c)
		}
	}
	return b.String()
}
