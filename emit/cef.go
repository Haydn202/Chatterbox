package emit

import (
	"fmt"
	"sort"
	"strings"
)

// CEFConfig configures ArcSight CEF lines.
type CEFConfig struct {
	Vendor  string // default Chatterbox
	Product string // default Generator
	Version string // default 1.0
	// NameField maps to CEF Name (default "message").
	NameField string
	// SeverityField maps to CEF Severity 0–10 (default "level").
	SeverityField string
}

// CEFFormatter encodes one CEF line per event.
type CEFFormatter struct {
	cfg CEFConfig
}

func CEF(cfg CEFConfig) Formatter {
	return CEFFormatter{cfg: cfg}
}

func (c CEFFormatter) Format(record map[string]any) ([]byte, error) {
	vendor := c.cfg.Vendor
	if vendor == "" {
		vendor = "Chatterbox"
	}
	product := c.cfg.Product
	if product == "" {
		product = "Generator"
	}
	version := c.cfg.Version
	if version == "" {
		version = "1.0"
	}
	nameField := c.cfg.NameField
	if nameField == "" {
		nameField = "message"
	}
	sevField := c.cfg.SeverityField
	if sevField == "" {
		sevField = "level"
	}

	name := valueString(record[nameField])
	if name == "" {
		name = "event"
	}
	severity := levelToCEFSeverity(valueString(record[sevField]))
	sigID := "0"

	ext := c.extension(record, nameField, sevField)
	line := fmt.Sprintf("CEF:0|%s|%s|%s|%s|%s|%d|%s\n",
		escapeCEFPipe(vendor),
		escapeCEFPipe(product),
		escapeCEFPipe(version),
		sigID,
		escapeCEFPipe(name),
		severity,
		ext,
	)
	return []byte(line), nil
}

func (c CEFFormatter) extension(record map[string]any, skip ...string) string {
	skipSet := make(map[string]bool)
	for _, k := range skip {
		skipSet[k] = true
	}
	keys := make([]string, 0, len(record))
	for k := range record {
		if !skipSet[k] {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		v := record[k]
		if v == nil {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s=%s", k, escapeCEF(valueString(v))))
	}
	return strings.Join(parts, " ")
}

func escapeCEFPipe(s string) string {
	return strings.ReplaceAll(s, "|", `\|`)
}

func levelToCEFSeverity(level string) int {
	switch strings.ToLower(level) {
	case "emerg", "emergency", "alert":
		return 10
	case "crit", "critical":
		return 9
	case "err", "error":
		return 8
	case "warning", "warn":
		return 6
	case "notice", "info":
		return 3
	case "debug":
		return 1
	default:
		return 3
	}
}
