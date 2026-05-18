package emit

import (
	"encoding/json"
	"io"
)

// JSONL encodes a single log record as one JSON line (no trailing newline).
func JSONL(record map[string]any) ([]byte, error) {
	return json.Marshal(record)
}

// WriteJSONL marshals record and writes it followed by a newline.
func WriteJSONL(w io.Writer, record map[string]any) error {
	b, err := JSONL(record)
	if err != nil {
		return err
	}
	if _, err := w.Write(b); err != nil {
		return err
	}
	_, err = w.Write([]byte{'\n'})
	return err
}
