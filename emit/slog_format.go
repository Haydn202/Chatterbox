package emit

import (
	"encoding/json"
	"sort"
	"strings"
)

// SlogJSONConfig configures slog.JSONHandler-shaped output.
type SlogJSONConfig struct {
	FieldMap FieldMap
}

// SlogJSONFormatter encodes records like slog.JSONHandler (time, level, msg + attrs).
type SlogJSONFormatter struct {
	fm FieldMap
}

func SlogJSON(cfg SlogJSONConfig) Formatter {
	fm := cfg.FieldMap
	if fm == (FieldMap{}) {
		fm = DefaultFieldMap()
	}
	return SlogJSONFormatter{fm: fm}
}

func (s SlogJSONFormatter) Format(record map[string]any) ([]byte, error) {
	out := make(map[string]any)
	out["time"] = valueString(record[s.fm.timeKey()])
	out["level"] = levelToSlogString(valueString(record[s.fm.levelKey()]))
	out["msg"] = valueString(record[s.fm.messageKey()])
	for k, v := range s.fm.Remaining(record) {
		out[k] = v
	}
	b, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}
	b = append(b, '\n')
	return b, nil
}

// SlogTextConfig configures slog.TextHandler-shaped output.
type SlogTextConfig struct {
	FieldMap FieldMap
}

// SlogTextFormatter encodes records like slog.TextHandler.
type SlogTextFormatter struct {
	fm FieldMap
}

func SlogText(cfg SlogTextConfig) Formatter {
	fm := cfg.FieldMap
	if fm == (FieldMap{}) {
		fm = DefaultFieldMap()
	}
	return SlogTextFormatter{fm: fm}
}

func (s SlogTextFormatter) Format(record map[string]any) ([]byte, error) {
	var b strings.Builder
	b.WriteString("time=")
	b.WriteString(formatHeaderValue(record[s.fm.timeKey()]))
	b.WriteString(" level=")
	b.WriteString(levelToSlogString(valueString(record[s.fm.levelKey()])))
	b.WriteString(" msg=")
	b.WriteString(formatHeaderValue(record[s.fm.messageKey()]))

	keys := make([]string, 0, len(record))
	for k := range s.fm.Remaining(record) {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		b.WriteByte(' ')
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(formatHeaderValue(s.fm.Remaining(record)[k]))
	}
	b.WriteByte('\n')
	return []byte(b.String()), nil
}
