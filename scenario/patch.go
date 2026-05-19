package scenario

import (
	"math/rand/v2"

	"github.com/Haydn202/Chatterbox/fuzz"
)

func applyPatchToRecord(rec map[string]any, p patchState, rng *rand.Rand) {
	if p.levelWeights != nil {
		rec["level"] = fuzz.Weighted(p.levelWeights).Generate(rng)
	}
	if p.messageFuzzer != nil {
		rec["message"] = p.messageFuzzer.Generate(rng)
	}
}
