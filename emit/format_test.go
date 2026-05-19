package emit

import (
	"encoding/json"
	"strings"
	"testing"
)

func sampleRecord() map[string]any {
	return map[string]any{
		"timestamp": "2026-05-18T12:00:00Z",
		"level":     "error",
		"message":   "handler failed",
		"email":     "user@example.com",
	}
}

func TestNewFormatter_json(t *testing.T) {
	f, err := NewFormatter(FormatJSON, Options{})
	if err != nil {
		t.Fatal(err)
	}
	b, err := f.Format(sampleRecord())
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(bytesTrimNL(b), &m); err != nil {
		t.Fatal(err)
	}
}

func TestNewFormatter_logfmt(t *testing.T) {
	f, err := NewFormatter(FormatLogfmt, Options{})
	if err != nil {
		t.Fatal(err)
	}
	b, err := f.Format(sampleRecord())
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.Contains(s, `level=error`) || !strings.Contains(s, `message=`) {
		t.Fatalf("unexpected logfmt: %q", s)
	}
}

func TestNewFormatter_plain(t *testing.T) {
	f, err := NewFormatter(FormatPlain, Options{})
	if err != nil {
		t.Fatal(err)
	}
	b, err := f.Format(sampleRecord())
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.HasPrefix(s, "2026-05-18T12:00:00Z ERROR handler failed") {
		t.Fatalf("unexpected plain: %q", s)
	}
}

func TestNewFormatter_syslog(t *testing.T) {
	f, err := NewFormatter(FormatSyslog, Options{Syslog: &SyslogConfig{Hostname: "host1"}})
	if err != nil {
		t.Fatal(err)
	}
	b, err := f.Format(sampleRecord())
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.HasPrefix(s, "<") || !strings.Contains(s, "host1") {
		t.Fatalf("unexpected syslog: %q", s)
	}
}

func TestNewFormatter_cef(t *testing.T) {
	f, err := NewFormatter(FormatCEF, Options{})
	if err != nil {
		t.Fatal(err)
	}
	b, err := f.Format(sampleRecord())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(b), "CEF:0|") {
		t.Fatalf("unexpected cef: %q", b)
	}
}

func TestNewFormatter_multilineRequiresConfig(t *testing.T) {
	_, err := NewFormatter(FormatMultiline, Options{})
	if err == nil {
		t.Fatal("expected error without multiline config")
	}
}

func TestNewFormatter_slogJSON(t *testing.T) {
	f, err := NewFormatter(FormatSlogJSON, Options{})
	if err != nil {
		t.Fatal(err)
	}
	b, err := f.Format(sampleRecord())
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.Contains(s, `"time":`) || !strings.Contains(s, `"msg":`) || !strings.Contains(s, `"level":"ERROR"`) {
		t.Fatalf("unexpected slog json: %q", s)
	}
}

func TestNewFormatter_zapJSON(t *testing.T) {
	f, err := NewFormatter(FormatZapJSON, Options{})
	if err != nil {
		t.Fatal(err)
	}
	b, err := f.Format(sampleRecord())
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.Contains(s, `"ts":`) || !strings.Contains(s, `"caller":`) {
		t.Fatalf("unexpected zap json: %q", s)
	}
}

func TestNewFormatter_zerologJSON(t *testing.T) {
	f, err := NewFormatter(FormatZerologJSON, Options{})
	if err != nil {
		t.Fatal(err)
	}
	b, err := f.Format(sampleRecord())
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.Contains(s, `"message":`) || !strings.Contains(s, `"time":`) {
		t.Fatalf("unexpected zerolog json: %q", s)
	}
}

func TestNewFormatter_unknown(t *testing.T) {
	_, err := NewFormatter(Format("xml"), Options{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func bytesTrimNL(b []byte) []byte {
	return []byte(strings.TrimSuffix(string(b), "\n"))
}
