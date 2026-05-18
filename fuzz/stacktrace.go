package fuzz

import (
	"fmt"
	"math/rand/v2"
	"strings"
)

const (
	StackStyleGo     = "go"
	StackStyleJava   = "java"
	StackStylePython = "python"
)

// StackTraceOption configures StackTrace fuzzer behavior.
type StackTraceOption func(*stacktraceFuzzer)

// WithStackStyle sets output style: go, java, or python (default go).
func WithStackStyle(style string) StackTraceOption {
	return func(s *stacktraceFuzzer) { s.style = style }
}

// WithFrames sets inclusive min and max frame counts (default 3–12).
func WithFrames(min, max int) StackTraceOption {
	return func(s *stacktraceFuzzer) {
		s.minFrames = min
		s.maxFrames = max
	}
}

// WithPanicMessages overrides the pool of first-line error messages.
func WithPanicMessages(msgs []string) StackTraceOption {
	return func(s *stacktraceFuzzer) {
		if len(msgs) > 0 {
			s.messages = append([]string(nil), msgs...)
		}
	}
}

type stacktraceFuzzer struct {
	style     string
	minFrames int
	maxFrames int
	messages  []string
}

var defaultPanicMessages = []string{
	"connection reset by peer",
	"index out of range [5] with length 3",
	"context deadline exceeded",
	"nil pointer dereference",
	"invalid memory address or nil pointer dereference",
	"runtime error: integer divide by zero",
}

// StackTrace generates a multiline stack trace string (physical newlines between frames).
func StackTrace(opts ...StackTraceOption) Fuzzer {
	s := &stacktraceFuzzer{
		style:     StackStyleGo,
		minFrames: 3,
		maxFrames: 12,
		messages:  defaultPanicMessages,
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

func (s *stacktraceFuzzer) Generate(r *rand.Rand) any {
	frames := s.minFrames
	if s.maxFrames > s.minFrames {
		frames += r.IntN(s.maxFrames - s.minFrames + 1)
	} else if s.maxFrames > 0 {
		frames = s.maxFrames
	}
	switch s.style {
	case StackStyleJava:
		return s.java(r, frames)
	case StackStylePython:
		return s.python(r, frames)
	default:
		return s.goStack(r, frames)
	}
}

func (s *stacktraceFuzzer) pickMessage(r *rand.Rand) string {
	return s.messages[r.IntN(len(s.messages))]
}

func (s *stacktraceFuzzer) goStack(r *rand.Rand, frames int) string {
	var b strings.Builder
	b.WriteString("panic: ")
	b.WriteString(s.pickMessage(r))
	b.WriteByte('\n')
	b.WriteByte('\n')
	gid := r.IntN(64) + 1
	states := []string{"running", "sleep", "chan receive", "select"}
	b.WriteString(fmt.Sprintf("goroutine %d [%s]:\n", gid, states[r.IntN(len(states))]))
	for i := 0; i < frames; i++ {
		pkg := randomIdent(r, 2+r.IntN(2))
		fn := randomAlpha(r, 4, 12)
		if r.IntN(2) == 0 {
			b.WriteString(fmt.Sprintf("%s.%s()\n", pkg, fn))
		} else {
			b.WriteString(fmt.Sprintf("%s.(*%s).%s()\n", pkg, randomAlpha(r, 4, 8), fn))
		}
		path := fmt.Sprintf("/%s/%s/%s.go", randomAlpha(r, 3, 6), randomAlpha(r, 4, 10), randomAlpha(r, 4, 10))
		line := r.IntN(900) + 1
		off := r.IntN(0xffff)
		b.WriteString(fmt.Sprintf("\t%s:%d +0x%x\n", path, line, off))
	}
	return strings.TrimRight(b.String(), "\n")
}

func (s *stacktraceFuzzer) java(r *rand.Rand, frames int) string {
	var b strings.Builder
	exTypes := []string{"java.lang.RuntimeException", "java.lang.NullPointerException", "java.io.IOException"}
	ex := exTypes[r.IntN(len(exTypes))]
	b.WriteString(fmt.Sprintf("Exception in thread \"%s\" %s: %s\n",
		randomAlpha(r, 4, 8), ex, s.pickMessage(r)))
	for i := 0; i < frames; i++ {
		pkg := randomIdent(r, 3+r.IntN(2))
		cls := capitalize(randomAlpha(r, 4, 10))
		method := randomAlpha(r, 4, 12)
		file := cls + ".java"
		line := r.IntN(500) + 1
		b.WriteString(fmt.Sprintf("\tat %s.%s.%s(%s:%d)\n", pkg, cls, method, file, line))
	}
	return strings.TrimRight(b.String(), "\n")
}

func (s *stacktraceFuzzer) python(r *rand.Rand, frames int) string {
	var b strings.Builder
	b.WriteString("Traceback (most recent call last):\n")
	for i := 0; i < frames; i++ {
		file := fmt.Sprintf("/app/%s/%s.py", randomAlpha(r, 3, 8), randomAlpha(r, 4, 10))
		line := r.IntN(400) + 1
		fn := randomAlpha(r, 4, 14)
		b.WriteString(fmt.Sprintf("  File \"%s\", line %d, in %s\n", file, line, fn))
		if i == frames-1 {
			b.WriteString(fmt.Sprintf("%s: %s\n", capitalize(randomAlpha(r, 5, 12))+"Error", s.pickMessage(r)))
		}
	}
	return strings.TrimRight(b.String(), "\n")
}
