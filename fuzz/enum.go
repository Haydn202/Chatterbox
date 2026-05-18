package fuzz

// LevelWeighted picks a log level using weighted probabilities.
func LevelWeighted(weights map[string]float64) Fuzzer {
	return Weighted(weights)
}
