package preset

import (
	"testing"
)

func TestBuild_default(t *testing.T) {
	s, err := Build(Default, Options{})
	if err != nil {
		t.Fatal(err)
	}
	fields := s.Fields()
	if len(fields) < 5 {
		t.Fatalf("expected at least 5 fields, got %d", len(fields))
	}
}

func TestBuild_minimalNoEmail(t *testing.T) {
	s, err := Build(Minimal, Options{})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range s.Fields() {
		if f.Name == "email" {
			t.Fatal("minimal preset should not include email")
		}
	}
}

func TestBuild_emailToggle(t *testing.T) {
	s, err := Build(Minimal, Options{Email: boolPtr(true)})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, f := range s.Fields() {
		if f.Name == "email" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected email when toggled on")
	}
}

func TestCorrelateEnabled_multilineDefault(t *testing.T) {
	if !CorrelateEnabled(MultilineError, Options{}) {
		t.Fatal("multiline-error should correlate by default")
	}
}

func TestParseName(t *testing.T) {
	n, err := ParseName("API")
	if err != nil || n != API {
		t.Fatalf("got %q %v", n, err)
	}
}
