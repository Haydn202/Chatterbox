package emit

import (
	"fmt"
	"strings"
)

// Format names a built-in output encoding.
type Format string

const (
	FormatJSON       Format = "json"
	FormatLogfmt     Format = "logfmt"
	FormatPlain      Format = "plain"
	FormatSyslog     Format = "syslog"
	FormatCEF        Format = "cef"
	FormatMultiline  Format = "multiline"
	FormatSlogJSON   Format = "slog_json"
	FormatSlogText   Format = "slog_text"
	FormatZapJSON    Format = "zap_json"
	FormatZerologJSON Format = "zerolog_json"
)

// Options holds per-format configuration. Only the field matching the chosen format is used.
type Options struct {
	Logfmt    *LogfmtConfig
	Plain     *PlainTextConfig
	Syslog    *SyslogConfig
	CEF       *CEFConfig
	Multiline *TextMultilineConfig
	SlogJSON  *SlogJSONConfig
	SlogText  *SlogTextConfig
	ZapJSON   *ZapJSONConfig
	ZerologJSON *ZerologJSONConfig
}

// NewFormatter returns a Formatter for the given format name.
// Format is case-insensitive. Empty format defaults to JSON.
func NewFormatter(format Format, opts Options) (Formatter, error) {
	switch strings.ToLower(string(format)) {
	case "", string(FormatJSON), "jsonl":
		return JSONLFormatter{}, nil
	case string(FormatLogfmt):
		cfg := LogfmtConfig{}
		if opts.Logfmt != nil {
			cfg = *opts.Logfmt
		}
		return Logfmt(cfg), nil
	case string(FormatPlain), "plaintext", "text":
		cfg := PlainTextConfig{}
		if opts.Plain != nil {
			cfg = *opts.Plain
		}
		return PlainText(cfg), nil
	case string(FormatSyslog):
		cfg := SyslogConfig{}
		if opts.Syslog != nil {
			cfg = *opts.Syslog
		}
		return Syslog(cfg), nil
	case string(FormatCEF):
		cfg := CEFConfig{}
		if opts.CEF != nil {
			cfg = *opts.CEF
		}
		return CEF(cfg), nil
	case string(FormatMultiline):
		if opts.Multiline == nil {
			return nil, fmt.Errorf("emit: multiline format requires Options.Multiline")
		}
		return TextMultilineFormatter(*opts.Multiline), nil
	case string(FormatSlogJSON):
		cfg := SlogJSONConfig{}
		if opts.SlogJSON != nil {
			cfg = *opts.SlogJSON
		}
		return SlogJSON(cfg), nil
	case string(FormatSlogText):
		cfg := SlogTextConfig{}
		if opts.SlogText != nil {
			cfg = *opts.SlogText
		}
		return SlogText(cfg), nil
	case string(FormatZapJSON):
		cfg := ZapJSONConfig{}
		if opts.ZapJSON != nil {
			cfg = *opts.ZapJSON
		}
		return ZapJSON(cfg), nil
	case string(FormatZerologJSON):
		cfg := ZerologJSONConfig{}
		if opts.ZerologJSON != nil {
			cfg = *opts.ZerologJSON
		}
		return ZerologJSON(cfg), nil
	default:
		return nil, fmt.Errorf("emit: unknown format %q", format)
	}
}

// MustFormatter is like NewFormatter but panics on error.
func MustFormatter(format Format, opts Options) Formatter {
	f, err := NewFormatter(format, opts)
	if err != nil {
		panic(err)
	}
	return f
}
