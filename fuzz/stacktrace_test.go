package fuzz

import (
	"math/rand/v2"
	"strings"
	"testing"
)

func TestStackTrace_goStyle(t *testing.T) {
	r := rand.New(rand.NewPCG(11, 11))
	s := StackTrace(WithStackStyle(StackStyleGo), WithFrames(4, 4)).Generate(r).(string)
	if !strings.HasPrefix(s, "panic:") {
		t.Fatalf("expected panic prefix: %q", s[:min(20, len(s))])
	}
	if !strings.Contains(s, "goroutine") {
		t.Fatal("expected goroutine line")
	}
	if strings.Count(s, "\n") < 4 {
		t.Fatalf("expected multiple lines, got %d", strings.Count(s, "\n"))
	}
}

func TestStackTrace_javaStyle(t *testing.T) {
	r := rand.New(rand.NewPCG(12, 12))
	s := StackTrace(WithStackStyle(StackStyleJava), WithFrames(3, 3)).Generate(r).(string)
	if !strings.HasPrefix(s, "Exception in thread") {
		t.Fatal("expected java header")
	}
	if !strings.Contains(s, "\tat ") {
		t.Fatal("expected at frames")
	}
}

func TestStackTrace_pythonStyle(t *testing.T) {
	r := rand.New(rand.NewPCG(13, 13))
	s := StackTrace(WithStackStyle(StackStylePython), WithFrames(2, 2)).Generate(r).(string)
	if !strings.HasPrefix(s, "Traceback") {
		t.Fatal("expected python header")
	}
	if !strings.Contains(s, `File "`) {
		t.Fatal("expected file line")
	}
}

func TestStackTrace_reproducible(t *testing.T) {
	a := rand.New(rand.NewPCG(99, 0))
	b := rand.New(rand.NewPCG(99, 0))
	f := StackTrace(WithFrames(5, 5))
	if f.Generate(a) != f.Generate(b) {
		t.Fatal("same seed should match")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
