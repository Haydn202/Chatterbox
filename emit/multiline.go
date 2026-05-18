package emit

import (
	"fmt"
	"strings"
)

// TextMultilineConfig controls physical multiline text output.
type TextMultilineConfig struct {
	HeaderFields           []string
	BodyFields             []string
	BlankLineBetweenEvents bool
}

// TextMultiline formats records as a header line plus body fields on following lines.
type TextMultiline struct {
	cfg TextMultilineConfig
}

// TextMultilineFormatter returns a Formatter for multiline text logs.
func TextMultilineFormatter(cfg TextMultilineConfig) Formatter {
	return TextMultiline{cfg: cfg}
}

func (t TextMultiline) Format(record map[string]any) ([]byte, error) {
	var b strings.Builder
	for i, key := range t.cfg.HeaderFields {
		val, ok := record[key]
		if !ok || val == nil {
			continue
		}
		if i > 0 && b.Len() > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(key)
		b.WriteByte('=')
		b.WriteString(formatHeaderValue(val))
	}
	if b.Len() > 0 {
		b.WriteByte('\n')
	}
	for _, key := range t.cfg.BodyFields {
		val, ok := record[key]
		if !ok || val == nil {
			continue
		}
		s, ok := val.(string)
		if !ok {
			s = fmt.Sprint(val)
		}
		if s == "" {
			continue
		}
		b.WriteString(s)
		if !strings.HasSuffix(s, "\n") {
			b.WriteByte('\n')
		}
	}
	if t.cfg.BlankLineBetweenEvents {
		b.WriteByte('\n')
	}
	return []byte(b.String()), nil
}

func formatHeaderValue(v any) string {
	s := fmt.Sprint(v)
	if strings.ContainsAny(s, " \t\"") {
		s = strings.ReplaceAll(s, `\`, `\\`)
		s = strings.ReplaceAll(s, `"`, `\"`)
		return `"` + s + `"`
	}
	return s
}
