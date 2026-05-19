package emit

import (
	"fmt"
	"strings"
)

// SyslogConfig configures RFC5424-style syslog lines (simplified for testing).
type SyslogConfig struct {
	Hostname string // default chatterbox
	AppName  string // default app
	// Facility is the syslog facility code 0–23 (default 1, user-level).
	Facility int
	// LevelField names the record field used for severity (default "level").
	LevelField string
	// MessageField is the human-readable MSG part (default "message").
	MessageField string
}

// SyslogFormatter encodes one syslog line per event.
type SyslogFormatter struct {
	cfg SyslogConfig
}

func Syslog(cfg SyslogConfig) Formatter {
	return SyslogFormatter{cfg: cfg}
}

func (s SyslogFormatter) Format(record map[string]any) ([]byte, error) {
	host := s.cfg.Hostname
	if host == "" {
		host = "chatterbox"
	}
	app := s.cfg.AppName
	if app == "" {
		app = "app"
	}
	facility := s.cfg.Facility
	if facility < 0 || facility > 23 {
		facility = 1
	}
	levelField := s.cfg.LevelField
	if levelField == "" {
		levelField = "level"
	}
	msgField := s.cfg.MessageField
	if msgField == "" {
		msgField = "message"
	}

	severity := levelToSeverity(valueString(record[levelField]))
	pri := facility*8 + severity

	ts := valueString(record["timestamp"])
	if ts == "" {
		ts = "-"
	}
	msg := valueString(record[msgField])
	if msg == "" {
		msg = "-"
	}

	// Structured data: remaining fields as [chatterbox key="val" ...]
	var sd strings.Builder
	sd.WriteString("[chatterbox")
	for k, v := range record {
		if k == "timestamp" || k == levelField || k == msgField || v == nil {
			continue
		}
		sd.WriteByte(' ')
		sd.WriteString(k)
		sd.WriteString(`="`)
		sd.WriteString(strings.ReplaceAll(valueString(v), `"`, `\"`))
		sd.WriteByte('"')
	}
	sd.WriteByte(']')

	line := fmt.Sprintf("<%d>1 %s %s %s - - - %s %s\n", pri, ts, host, app, sd.String(), msg)
	return []byte(line), nil
}

func levelToSeverity(level string) int {
	switch strings.ToLower(level) {
	case "emerg", "emergency":
		return 0
	case "alert":
		return 1
	case "crit", "critical":
		return 2
	case "err", "error":
		return 3
	case "warning", "warn":
		return 4
	case "notice":
		return 5
	case "info":
		return 6
	case "debug":
		return 7
	default:
		return 6
	}
}
