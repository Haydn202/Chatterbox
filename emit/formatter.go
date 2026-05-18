package emit

// Formatter encodes one log record. Output may contain newlines (multiline events).
type Formatter interface {
	Format(record map[string]any) ([]byte, error)
}

// JSONLFormatter encodes each record as a single JSON line with a trailing newline.
type JSONLFormatter struct{}

func (JSONLFormatter) Format(record map[string]any) ([]byte, error) {
	b, err := JSONL(record)
	if err != nil {
		return nil, err
	}
	out := make([]byte, len(b)+1)
	copy(out, b)
	out[len(b)] = '\n'
	return out, nil
}
