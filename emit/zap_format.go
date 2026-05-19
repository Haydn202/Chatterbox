package emit

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
)

// ZapJSONConfig configures zap production JSON encoder-shaped output.
type ZapJSONConfig struct {
	FieldMap FieldMap
	// Caller is a synthetic caller string (default generated per line).
	Caller string
}

// ZapJSONFormatter encodes records like zap's production JSON encoder.
type ZapJSONFormatter struct {
	fm     FieldMap
	caller string
	rng    *rand.Rand
}

func ZapJSON(cfg ZapJSONConfig) Formatter {
	fm := cfg.FieldMap
	if fm == (FieldMap{}) {
		fm = DefaultFieldMap()
	}
	return ZapJSONFormatter{
		fm:     fm,
		caller: cfg.Caller,
		rng:    rand.New(rand.NewPCG(1, 1)),
	}
}

func (z ZapJSONFormatter) Format(record map[string]any) ([]byte, error) {
	ts := parseTime(record, z.fm.timeKey())
	secs := float64(ts.UnixNano()) / 1e9

	caller := z.caller
	if caller == "" {
		caller = fmt.Sprintf("main.go:%d", 40+z.rng.IntN(200))
	}

	out := map[string]any{
		"level":  levelToZapString(valueString(record[z.fm.levelKey()])),
		"ts":     secs,
		"caller": caller,
		"msg":    valueString(record[z.fm.messageKey()]),
	}
	for k, v := range z.fm.Remaining(record) {
		out[k] = v
	}
	b, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}
	b = append(b, '\n')
	return b, nil
}
