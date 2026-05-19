package emit

import (
	"strings"
	"time"
)

// FieldMap maps Chatterbox schema field names to logger-specific names.
type FieldMap struct {
	Time    string // default timestamp
	Level   string // default level
	Message string // default message
}

// DefaultFieldMap returns the conventional Chatterbox schema field names.
func DefaultFieldMap() FieldMap {
	return FieldMap{Time: "timestamp", Level: "level", Message: "message"}
}

// TimeKey returns the schema field name for timestamps.
func (m FieldMap) TimeKey() string {
	return m.timeKey()
}

// LevelKey returns the schema field name for log level.
func (m FieldMap) LevelKey() string {
	return m.levelKey()
}

// MessageKey returns the schema field name for the message.
func (m FieldMap) MessageKey() string {
	return m.messageKey()
}

func (m FieldMap) timeKey() string {
	if m.Time != "" {
		return m.Time
	}
	return "timestamp"
}

func (m FieldMap) levelKey() string {
	if m.Level != "" {
		return m.Level
	}
	return "level"
}

func (m FieldMap) messageKey() string {
	if m.Message != "" {
		return m.Message
	}
	return "message"
}

// Remaining returns record entries excluding the mapped time, level, and message keys.
func (m FieldMap) Remaining(record map[string]any) map[string]any {
	skip := map[string]bool{
		m.timeKey():    true,
		m.levelKey():   true,
		m.messageKey(): true,
	}
	out := make(map[string]any)
	for k, v := range record {
		if !skip[k] && v != nil {
			out[k] = v
		}
	}
	return out
}

// ParseTime reads a timestamp field from a record.
func ParseTime(record map[string]any, key string) time.Time {
	return parseTime(record, key)
}

func parseTime(record map[string]any, key string) time.Time {
	v, ok := record[key]
	if !ok || v == nil {
		return time.Now().UTC()
	}
	switch t := v.(type) {
	case time.Time:
		return t.UTC()
	case string:
		for _, layout := range []string{time.RFC3339, time.RFC3339Nano} {
			if parsed, err := time.Parse(layout, t); err == nil {
				return parsed.UTC()
			}
		}
	}
	return time.Now().UTC()
}

func levelToSlogString(level string) string {
	switch strings.ToLower(level) {
	case "debug":
		return "DEBUG"
	case "warn", "warning":
		return "WARN"
	case "error", "err":
		return "ERROR"
	default:
		return "INFO"
	}
}

func levelToZapString(level string) string {
	switch strings.ToLower(level) {
	case "debug":
		return "debug"
	case "warn", "warning":
		return "warn"
	case "error", "err":
		return "error"
	case "dpanic", "panic", "fatal":
		return strings.ToLower(level)
	default:
		return "info"
	}
}
