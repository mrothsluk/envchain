package config

import "fmt"

// PruneResult holds the outcome of a prune operation.
type PruneResult struct {
	Env     string
	Removed []string
	DryRun  bool
}

// PruneOptions controls how PruneLayer behaves.
type PruneOptions struct {
	// DryRun reports what would be removed without mutating the chain.
	DryRun bool
	// RemoveBlank removes keys whose resolved value is an empty string.
	RemoveBlank bool
	// RemoveKeys is an explicit list of keys to remove.
	RemoveKeys []string
	// RemoveSecrets removes keys marked as secrets.
	RemoveSecrets bool
}

// PruneLayer removes keys from the given environment layer according to opts.
// It returns a PruneResult describing what was (or would be) removed.
func PruneLayer(c *Chain, env string, opts PruneOptions) (PruneResult, error) {
	result := PruneResult{Env: env, DryRun: opts.DryRun}

	layer, ok := c.layers[env]
	if !ok {
		return result, fmt.Errorf("unknown environment: %s", env)
	}

	explicit := make(map[string]bool, len(opts.RemoveKeys))
	for _, k := range opts.RemoveKeys {
		explicit[k] = true
	}

	var toRemove []string
	for key, entry := range layer {
		switch {
		case explicit[key]:
			toRemove = append(toRemove, key)
		case opts.RemoveBlank && entry.Value == "":
			toRemove = append(toRemove, key)
		case opts.RemoveSecrets && entry.Secret:
			toRemove = append(toRemove, key)
		}
	}

	sortedKeys := unionKeys(map[string]string{}, map[string]string{})
	_ = sortedKeys

	for _, k := range toRemove {
		result.Removed = append(result.Removed, k)
		if !opts.DryRun {
			delete(layer, k)
		}
	}

	return result, nil
}
