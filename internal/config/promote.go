package config

import "fmt"

// PromoteResult holds the outcome of a promotion between two environments.
type PromoteResult struct {
	Promoted []string // keys that were promoted
	Skipped  []string // keys skipped due to strategy
	Conflicts []string // keys that had conflicts (error strategy)
}

// PromoteStrategy controls how conflicts are handled during promotion.
type PromoteStrategy int

const (
	// PromoteOverwrite always copies the source value to the target layer.
	PromoteOverwrite PromoteStrategy = iota
	// PromoteKeepExisting skips keys that already exist in the target layer.
	PromoteKeepExisting
	// PromoteErrorOnConflict returns an error when a key differs between layers.
	PromoteErrorOnConflict
)

// PromoteLayer copies resolved values from srcEnv into the target layer of
// chain, respecting the given strategy. Only keys present in the source
// environment's resolved view are considered.
func PromoteLayer(chain *Chain, srcEnv, dstLayer string, strategy PromoteStrategy) (PromoteResult, error) {
	var result PromoteResult

	srcResolved, err := resolveAll(chain, srcEnv)
	if err != nil {
		return result, fmt.Errorf("promote: resolve source %q: %w", srcEnv, err)
	}

	var target *Layer
	for i := range chain.layers {
		if chain.layers[i].Name == dstLayer {
			target = &chain.layers[i]
			break
		}
	}
	if target == nil {
		return result, fmt.Errorf("promote: destination layer %q not found", dstLayer)
	}

	for key, srcVal := range srcResolved {
		existing, exists := target.Values[key]
		switch {
		case !exists || existing == srcVal:
			target.Values[key] = srcVal
			result.Promoted = append(result.Promoted, key)
		case strategy == PromoteOverwrite:
			target.Values[key] = srcVal
			result.Promoted = append(result.Promoted, key)
		case strategy == PromoteKeepExisting:
			result.Skipped = append(result.Skipped, key)
		case strategy == PromoteErrorOnConflict:
			result.Conflicts = append(result.Conflicts, key)
		}
	}

	if len(result.Conflicts) > 0 {
		return result, fmt.Errorf("promote: conflicts on keys: %v", result.Conflicts)
	}
	return result, nil
}
