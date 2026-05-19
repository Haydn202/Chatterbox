package emit

import (
	"sort"
	"strings"
)

// LogfmtConfig configures logfmt output.
type LogfmtConfig struct {
	// FieldOrder lists keys in emission order. Empty means all keys sorted alphabetically.
	FieldOrder []string
}

// LogfmtFormatter encodes records as logfmt (key=value pairs, one line per event).
type LogfmtFormatter struct {
	cfg LogfmtConfig
}

func Logfmt(cfg LogfmtConfig) Formatter {
	return LogfmtFormatter{cfg: cfg}
}

func (l LogfmtFormatter) Format(record map[string]any) ([]byte, error) {
	keys := l.keys(record)
	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteByte(' ')
		}
		v, ok := record[k]
		if !ok || v == nil {
			continue
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(escapeLogfmt(valueString(v)))
	}
	b.WriteByte('\n')
	return []byte(b.String()), nil
}

func (l LogfmtFormatter) keys(record map[string]any) []string {
	if len(l.cfg.FieldOrder) > 0 {
		out := make([]string, 0, len(l.cfg.FieldOrder))
		for _, k := range l.cfg.FieldOrder {
			if _, ok := record[k]; ok {
				out = append(out, k)
			}
		}
		return out
	}
	keys := make([]string, 0, len(record))
	for k := range record {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
