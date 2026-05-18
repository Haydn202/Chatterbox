package fuzz

import (
	"fmt"
	"math/rand/v2"
)

// IPv4Option configures IPv4 fuzzer.
type IPv4Option func(*ipv4Fuzzer)

// WithPrivateRange allows RFC1918 addresses (default true).
func WithPrivateRange(allow bool) IPv4Option {
	return func(i *ipv4Fuzzer) { i.allowPrivate = allow }
}

type ipv4Fuzzer struct {
	allowPrivate bool
}

// IPv4 generates random IPv4 addresses.
func IPv4(opts ...IPv4Option) Fuzzer {
	i := &ipv4Fuzzer{allowPrivate: true}
	for _, o := range opts {
		o(i)
	}
	return i
}

func (i *ipv4Fuzzer) Generate(r *rand.Rand) any {
	if i.allowPrivate && r.Float64() < 0.4 {
		switch r.IntN(3) {
		case 0:
			return fmt.Sprintf("10.%d.%d.%d", r.IntN(256), r.IntN(256), r.IntN(256))
		case 1:
			return fmt.Sprintf("172.%d.%d.%d", 16+r.IntN(16), r.IntN(256), r.IntN(256))
		case 2:
			return fmt.Sprintf("192.168.%d.%d", r.IntN(256), r.IntN(256))
		}
	}
	return fmt.Sprintf("%d.%d.%d.%d", r.IntN(256), r.IntN(256), r.IntN(256), r.IntN(256))
}

// URL generates http/https URLs with varied paths.
func URL() Fuzzer {
	return Func(func(r *rand.Rand) any {
		scheme := "https"
		if r.IntN(2) == 0 {
			scheme = "http"
		}
		host := randomAlpha(r, 4, 12) + ".example.com"
		path := "/" + randomAlpha(r, 3, 20)
		if r.Float64() < 0.3 {
			path += "/" + randomAlpha(r, 2, 10)
		}
		if r.Float64() < 0.2 {
			return fmt.Sprintf("%s://%s%s?q=%s", scheme, host, path, randomAlpha(r, 2, 8))
		}
		return fmt.Sprintf("%s://%s%s", scheme, host, path)
	})
}
