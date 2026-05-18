package chatterbox

import "github.com/Haydn202/Chatterbox/fuzz"

// Field binds a log field name to a fuzzer.
type Field struct {
	Name   string
	Fuzzer fuzz.Fuzzer
}

// Schema is an ordered set of fields for one log line.
type Schema struct {
	fields []Field
}

// NewSchema builds a schema from fields.
func NewSchema(fields ...Field) *Schema {
	cp := make([]Field, len(fields))
	copy(cp, fields)
	return &Schema{fields: cp}
}

// Fields returns a copy of the schema fields.
func (s *Schema) Fields() []Field {
	cp := make([]Field, len(s.fields))
	copy(cp, s.fields)
	return cp
}

// MakeField builds a schema field (constructor; distinct from the Field type).
func MakeField(name string, f fuzz.Fuzzer) Field {
	return Field{Name: name, Fuzzer: f}
}
