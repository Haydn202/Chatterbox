package fuzz

import (
	"fmt"
	"math/rand/v2"
)

// StringFrom generates a string with length between min and max (inclusive).
func StringFrom(min, max int) Fuzzer {
	if min < 0 {
		min = 0
	}
	if max < min {
		max = min
	}
	return Func(func(r *rand.Rand) any {
		n := min
		if max > min {
			n += r.IntN(max - min + 1)
		}
		const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 "
		b := make([]byte, n)
		for i := range b {
			b[i] = chars[r.IntN(len(chars))]
		}
		return string(b)
	})
}

// UUID generates a random UUID v4-style string (stdlib only).
func UUID() Fuzzer {
	return Func(func(r *rand.Rand) any {
		b := make([]byte, 16)
		for i := range b {
			b[i] = byte(r.IntN(256))
		}
		b[6] = (b[6] & 0x0f) | 0x40
		b[8] = (b[8] & 0x3f) | 0x80
		return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
			uint32(b[0])<<24|uint32(b[1])<<16|uint32(b[2])<<8|uint32(b[3]),
			uint16(b[4])<<8|uint16(b[5]),
			uint16(b[6])<<8|uint16(b[7]),
			uint16(b[8])<<8|uint16(b[9]),
			uint64(b[10])<<40|uint64(b[11])<<32|uint64(b[12])<<24|
				uint64(b[13])<<16|uint64(b[14])<<8|uint64(b[15]),
		)
	})
}
