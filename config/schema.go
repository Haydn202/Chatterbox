package config

// schemaDoc is the top-level YAML document for a schema file.
type schemaDoc struct {
	Schema struct {
		Name string `yaml:"name"`
	} `yaml:"schema"`
	Fields []fieldYAML `yaml:"fields"`
}

// fieldYAML describes one schema field and fuzzer options.
type fieldYAML struct {
	Name          string             `yaml:"name"`
	Type          string             `yaml:"type"`
	JitterSeconds int                `yaml:"jitter_seconds"`
	Weights       map[string]float64 `yaml:"weights"`
	MinLen        int                `yaml:"min_len"`
	MaxLen        int                `yaml:"max_len"`
	Probability   float64            `yaml:"probability"`
	Value         any                `yaml:"value"`
	Values        []any              `yaml:"values"`
	Inner         *fieldYAML         `yaml:"inner"`
	Lang          string             `yaml:"lang"`
	MinFrames     int                `yaml:"min_frames"`
	MaxFrames     int                `yaml:"max_frames"`
	AllowPrivate  *bool              `yaml:"allow_private"`
}
