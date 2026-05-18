package fuzz

import (
	"math/rand/v2"
	"strings"
	"testing"
)

func TestEmail_containsAt(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 2))
	f := Email()
	for i := 0; i < 100; i++ {
		s, ok := f.Generate(r).(string)
		if !ok {
			t.Fatal("expected string")
		}
		if !strings.Contains(s, "@") && !strings.Contains(s, "@@") {
			t.Fatalf("email missing @: %q", s)
		}
	}
}

func TestEmail_reproducible(t *testing.T) {
	a := rand.New(rand.NewPCG(99, 0))
	b := rand.New(rand.NewPCG(99, 0))
	f := Email()
	if f.Generate(a) != f.Generate(b) {
		t.Fatal("same seed should produce same email")
	}
}

func TestEmail_edgeCases(t *testing.T) {
	r := rand.New(rand.NewPCG(7, 7))
	f := Email(WithEdgeCases(true))
	seen := false
	for i := 0; i < 500; i++ {
		s := f.Generate(r).(string)
		if strings.Contains(s, "@@") || s == "user@" || strings.HasPrefix(s, "@") {
			seen = true
			break
		}
	}
	if !seen {
		t.Fatal("expected at least one edge case in 500 draws")
	}
}

func TestWeighted_level(t *testing.T) {
	r := rand.New(rand.NewPCG(3, 3))
	f := LevelWeighted(map[string]float64{"info": 1, "warn": 0, "error": 0})
	for i := 0; i < 20; i++ {
		if f.Generate(r) != "info" {
			t.Fatal("zero weight should never be picked")
		}
	}
}

func TestStringFrom_length(t *testing.T) {
	r := rand.New(rand.NewPCG(4, 4))
	f := StringFrom(5, 5)
	s := f.Generate(r).(string)
	if len(s) != 5 {
		t.Fatalf("got len %d", len(s))
	}
}

func TestUUID_format(t *testing.T) {
	r := rand.New(rand.NewPCG(5, 5))
	s := UUID().Generate(r).(string)
	if len(s) != 36 || s[8] != '-' {
		t.Fatalf("bad uuid: %q", s)
	}
}

func TestIPv4_format(t *testing.T) {
	r := rand.New(rand.NewPCG(6, 6))
	s := IPv4().Generate(r).(string)
	parts := strings.Split(s, ".")
	if len(parts) != 4 {
		t.Fatalf("bad ipv4: %q", s)
	}
}
