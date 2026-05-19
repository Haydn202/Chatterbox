package config

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Haydn202/Chatterbox"
	"gopkg.in/yaml.v3"
)

// LoadSchema parses a schema YAML document and builds a chatterbox.Schema.
func LoadSchema(r io.Reader) (*chatterbox.Schema, error) {
	var doc schemaDoc
	dec := yaml.NewDecoder(r)
	dec.KnownFields(false)
	if err := dec.Decode(&doc); err != nil {
		return nil, fmt.Errorf("config: parse schema: %w", err)
	}
	if len(doc.Fields) == 0 {
		return nil, fmt.Errorf("config: schema has no fields")
	}
	var fields []chatterbox.Field
	seen := make(map[string]struct{})
	for _, f := range doc.Fields {
		name := strings.TrimSpace(f.Name)
		if name == "" {
			return nil, fmt.Errorf("config: field name is required")
		}
		if _, dup := seen[name]; dup {
			return nil, fmt.Errorf("config: duplicate field %q", name)
		}
		seen[name] = struct{}{}
		fuzzer, err := buildFuzzer(f)
		if err != nil {
			return nil, err
		}
		fields = append(fields, chatterbox.MakeField(name, fuzzer))
	}
	return chatterbox.NewSchema(fields...), nil
}

// LoadSchemaFile loads a schema from a YAML file path.
func LoadSchemaFile(path string) (*chatterbox.Schema, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	schema, err := LoadSchema(f)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return schema, nil
}
