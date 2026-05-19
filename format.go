package chatterbox

import "github.com/Haydn202/Chatterbox/emit"

// WithOutputFormat sets the generator formatter from a format name and options.
// Returns an error if the format is unknown or misconfigured.
func WithOutputFormat(format emit.Format, opts emit.Options) (GeneratorOption, error) {
	f, err := emit.NewFormatter(format, opts)
	if err != nil {
		return nil, err
	}
	return WithFormatter(f), nil
}
