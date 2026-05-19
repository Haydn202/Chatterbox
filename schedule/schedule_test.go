package schedule

import (
	"testing"
	"time"
)

func TestFlatRate(t *testing.T) {
	s, err := FlatRate(100)
	if err != nil {
		t.Fatal(err)
	}
	w, ok := s.NextWait()
	if !ok || w != 10*time.Millisecond {
		t.Fatalf("got %v ok=%v", w, ok)
	}
}

func TestNewPhases_validation(t *testing.T) {
	_, err := NewPhases()
	if err == nil {
		t.Fatal("expected error for empty phases")
	}
	_, err = NewPhases(Phase{Rate: 0, Duration: time.Second})
	if err == nil {
		t.Fatal("expected error for zero rate")
	}
	_, err = NewPhases(
		Phase{Rate: 10, Duration: 0},
		Phase{Rate: 20, Duration: time.Second},
	)
	if err == nil {
		t.Fatal("expected error for zero duration on non-last phase")
	}
}

func TestPhases_exhausts(t *testing.T) {
	s, err := NewPhases(
		Phase{Rate: 1000, Duration: 5 * time.Millisecond},
	)
	if err != nil {
		t.Fatal(err)
	}
	for {
		_, ok := s.NextWait()
		if !ok {
			return
		}
		time.Sleep(2 * time.Millisecond)
	}
}

func TestPresetIncidentSpike(t *testing.T) {
	s, err := PresetIncidentSpike(10, 100, time.Millisecond, time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	_, ok := s.NextWait()
	if !ok {
		t.Fatal("expected schedule to continue")
	}
}
