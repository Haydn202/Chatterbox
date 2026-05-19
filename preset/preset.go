package preset

import (
	"fmt"
	"strings"

	"github.com/Haydn202/Chatterbox"
	"github.com/Haydn202/Chatterbox/fuzz"
)

// Name identifies a built-in schema preset.
type Name string

const (
	Default         Name = "default"
	API             Name = "api"
	Minimal         Name = "minimal"
	MultilineError  Name = "multiline-error"
)

// Options toggles optional fields on top of a preset.
type Options struct {
	Email      *bool // nil = use preset default
	Stacktrace *bool
	Correlate  *bool
}

// Defaults returns preset-default toggle values for name.
func Defaults(name Name) Options {
	switch name {
	case Minimal:
		return Options{
			Email:      boolPtr(false),
			Stacktrace: boolPtr(false),
			Correlate:  boolPtr(false),
		}
	case MultilineError:
		return Options{
			Email:      boolPtr(false),
			Stacktrace: boolPtr(true),
			Correlate:  boolPtr(true),
		}
	case API:
		return Options{
			Email:      boolPtr(true),
			Stacktrace: boolPtr(false),
			Correlate:  boolPtr(false),
		}
	default:
		return Options{
			Email:      boolPtr(true),
			Stacktrace: boolPtr(false),
			Correlate:  boolPtr(false),
		}
	}
}

// ParseName parses a preset name (case-insensitive).
func ParseName(s string) (Name, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "default", "":
		return Default, nil
	case "api":
		return API, nil
	case "minimal":
		return Minimal, nil
	case "multiline-error", "multiline_error", "multiline":
		return MultilineError, nil
	default:
		return "", fmt.Errorf("preset: unknown preset %q (use default, api, minimal, multiline-error)", s)
	}
}

// Build constructs a schema for the preset with merged options.
func Build(name Name, opt Options) (*chatterbox.Schema, error) {
	if _, err := ParseName(string(name)); err != nil {
		return nil, err
	}
	def := Defaults(name)
	email := mergeBool(def.Email, opt.Email)
	stacktrace := mergeBool(def.Stacktrace, opt.Stacktrace)
	correlate := mergeBool(def.Correlate, opt.Correlate)

	var fields []chatterbox.Field
	fields = append(fields,
		chatterbox.MakeField("timestamp", fuzz.TimestampRFC3339(fuzz.WithJitter(30))),
		chatterbox.MakeField("level", levelFuzzer(name)),
		chatterbox.MakeField("message", fuzz.StringFrom(10, 120)),
	)
	if email {
		fields = append(fields, chatterbox.MakeField("email", fuzz.Email()))
	}
	if name == API {
		fields = append(fields, chatterbox.MakeField("url", fuzz.URL()))
	}
	if name == Default {
		fields = append(fields, chatterbox.MakeField("client_ip", fuzz.IPv4()))
	}
	if stacktrace {
		fields = append(fields, chatterbox.MakeField("stacktrace", fuzz.Optional(0.15, fuzz.StackTrace(
			fuzz.WithFrames(4, 10),
		))))
	}
	_ = correlate // used by CLI for WithCorrelation, not schema fields
	return chatterbox.NewSchema(fields...), nil
}

// CorrelateEnabled reports whether correlation should be enabled for merged options.
func CorrelateEnabled(name Name, opt Options) bool {
	def := Defaults(name)
	return mergeBool(def.Correlate, opt.Correlate)
}

func levelFuzzer(name Name) fuzz.Fuzzer {
	if name == MultilineError {
		return fuzz.LevelWeighted(map[string]float64{
			"info": 0.2, "warn": 0.2, "error": 0.6,
		})
	}
	return fuzz.LevelWeighted(map[string]float64{
		"info": 0.7, "warn": 0.2, "error": 0.1,
	})
}

func mergeBool(def, override *bool) bool {
	if override != nil {
		return *override
	}
	if def != nil {
		return *def
	}
	return false
}

func boolPtr(b bool) *bool { return &b }
