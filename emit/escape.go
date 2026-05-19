package emit

import (
	"fmt"
	"strings"
)

func valueString(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprint(v)
}

// escapeLogfmt quotes and escapes a value for logfmt (https://brandur.org/logfmt).
func escapeLogfmt(s string) string {
	if s == "" {
		return `""`
	}
	needsQuote := false
	for _, c := range s {
		if c <= ' ' || c == '=' || c == '"' || c == '\\' {
			needsQuote = true
			break
		}
	}
	if !needsQuote {
		return s
	}
	var b strings.Builder
	b.WriteByte('"')
	for _, c := range s {
		switch c {
		case '\\', '"':
			b.WriteByte('\\')
			b.WriteRune(c)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if c <= ' ' {
				b.WriteByte('\\')
				b.WriteRune(c)
			} else {
				b.WriteRune(c)
			}
		}
	}
	b.WriteByte('"')
	return b.String()
}

// escapeCEF escapes extension values for CEF.
func escapeCEF(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, "=", `\=`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	return s
}
