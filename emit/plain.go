package emit

import (
	"fmt"
	"sort"
	"strings"
)

// PlainTextConfig configures single-line human-readable logs.
type PlainTextConfig struct {
	// Fields is emission order. Empty uses timestamp, level, message when present, then other keys sorted.
	Fields []string
	// Separator between parts (default single space).
	Separator string
}

// PlainTextFormatter encodes one line per event: "timestamp LEVEL message key=value ..."
type PlainTextFormatter struct {
	cfg PlainTextConfig
}

func PlainText(cfg PlainTextConfig) Formatter {
	return PlainTextFormatter{cfg: cfg}
}

func (p PlainTextFormatter) Format(record map[string]any) ([]byte, error) {
	sep := p.cfg.Separator
	if sep == "" {
		sep = " "
	}
	keys := p.resolveKeys(record)
	var parts []string
	for _, k := range keys {
		v, ok := record[k]
		if !ok || v == nil {
			continue
		}
		s := valueString(v)
		if s == "" {
			continue
		}
		switch k {
		case "timestamp":
			parts = append(parts, s)
		case "level":
			parts = append(parts, strings.ToUpper(s))
		case "message":
			parts = append(parts, s)
		default:
			parts = append(parts, fmt.Sprintf("%s=%s", k, s))
		}
	}
	line := strings.Join(parts, sep) + "\n"
	return []byte(line), nil
}

func (p PlainTextFormatter) resolveKeys(record map[string]any) []string {
	if len(p.cfg.Fields) > 0 {
		out := make([]string, 0, len(p.cfg.Fields))
		for _, k := range p.cfg.Fields {
			if _, ok := record[k]; ok {
				out = append(out, k)
			}
		}
		return out
	}
	defaults := []string{"timestamp", "level", "message"}
	seen := make(map[string]bool)
	var out []string
	for _, k := range defaults {
		if _, ok := record[k]; ok {
			out = append(out, k)
			seen[k] = true
		}
	}
	rest := make([]string, 0)
	for k := range record {
		if !seen[k] {
			rest = append(rest, k)
		}
	}
	sort.Strings(rest)
	return append(out, rest...)
}
