package fuzz

import (
	"math/rand/v2"
	"strings"
)

func randomAlpha(r *rand.Rand, min, maxLen int) string {
	n := min
	if maxLen > min {
		n += r.IntN(maxLen - min + 1)
	}
	const letters = "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[r.IntN(len(letters))]
	}
	return string(b)
}

func randomIdent(r *rand.Rand, parts int) string {
	if parts < 1 {
		parts = 1
	}
	var s string
	for i := 0; i < parts; i++ {
		if i > 0 {
			s += "."
		}
		s += randomAlpha(r, 3, 10)
	}
	return s
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
