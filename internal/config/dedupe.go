package config

import "fmt"

// DedupeResult holds the outcome of a deduplication pass on a layer.
type DedupeResult struct {
	Key      string
	Kept     string
	Dropped  []string
	Strategy string
}

// DedupeStrategy controls which value is retained when duplicate keys are found
// across multiple raw entries (e.g. after a merge or import).
type DedupeStrategy string

const (
	DedupeKeepFirst DedupeStrategy = "first"
	DedupeKeepLast  DedupeStrategy = "last"
)

// DedupeLayer scans the resolved values of env inside chain for keys that
// appear more than once in rawPairs (key→[]value) and collapses them
// according to strategy. It writes the winning value back into the layer and
// returns one DedupeResult per collapsed key.
//
// rawPairs simulates a scenario where an import or merge produced duplicate
// entries before they were loaded into the chain (e.g. two .env files both
// defining DATABASE_URL).
func DedupeLayer(chain *Chain, env string, rawPairs map[string][]string, strategy DedupeStrategy) ([]DedupeResult, error) {
	layer, ok := chain.layers[env]
	if !ok {
		return nil, fmt.Errorf("dedupe: unknown env %q", env)
	}

	if strategy != DedupeKeepFirst && strategy != DedupeKeepLast {
		return nil, fmt.Errorf("dedupe: unknown strategy %q", strategy)
	}

	var results []DedupeResult

	for key, values := range rawPairs {
		if len(values) <= 1 {
			continue
		}

		var kept string
		var dropped []string

		if strategy == DedupeKeepFirst {
			kept = values[0]
			dropped = values[1:]
		} else {
			kept = values[len(values)-1]
			dropped = values[:len(values)-1]
		}

		layer[key] = kept

		results = append(results, DedupeResult{
			Key:      key,
			Kept:     kept,
			Dropped:  dropped,
			Strategy: string(strategy),
		})
	}

	return results, nil
}
