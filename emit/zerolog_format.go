package emit

import "encoding/json"

// ZerologJSONConfig configures zerolog JSON output shape.
type ZerologJSONConfig struct {
	FieldMap FieldMap
}

// ZerologJSONFormatter encodes records like zerolog's default JSON writer.
type ZerologJSONFormatter struct {
	fm FieldMap
}

func ZerologJSON(cfg ZerologJSONConfig) Formatter {
	fm := cfg.FieldMap
	if fm == (FieldMap{}) {
		fm = DefaultFieldMap()
	}
	return ZerologJSONFormatter{fm: fm}
}

func (z ZerologJSONFormatter) Format(record map[string]any) ([]byte, error) {
	ts := parseTime(record, z.fm.timeKey())
	out := map[string]any{
		"level":   levelToZapString(valueString(record[z.fm.levelKey()])),
		"time":    ts.Format("2006-01-02T15:04:05Z07:00"),
		"message": valueString(record[z.fm.messageKey()]),
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
