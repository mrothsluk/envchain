package config

import (
	"fmt"
	"sort"
)

// RollbackResult holds the outcome of a rollback operation.
type RollbackResult struct {
	Env      string
	Restored int
	Skipped  int
}

// RollbackToSnapshot restores a chain layer to the values captured in a
// previously saved snapshot. Only keys present in the snapshot are touched;
// keys that exist in the current layer but are absent from the snapshot are
// left unchanged unless removeExtra is true.
//
// The conflict strategy follows the same semantics as MergeLayers:
//   - "overwrite"      – always restore snapshot value
//   - "keep-existing"  – skip keys that already match current value
//   - "error"          – return an error on any differing value
func RollbackToSnapshot(chain *Chain, env string, snap Snapshot, strategy string, removeExtra bool) (RollbackResult, error) {
	layer, err := chain.Resolve(env)
	if err != nil {
		return RollbackResult{}, fmt.Errorf("rollback: %w", err)
	}

	result := RollbackResult{Env: env}

	// Restore keys from snapshot.
	keys := make([]string, 0, len(snap))
	for k := range snap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		snapVal := snap[key]
		curVal, exists := layer[key]

		switch strategy {
		case "keep-existing":
			if exists && curVal == snapVal {
				result.Skipped++
				continue
			}
			if exists {
				result.Skipped++
				continue
			}
		case "error":
			if exists && curVal != snapVal {
				return RollbackResult{}, fmt.Errorf("rollback: conflict on key %q: current=%q snapshot=%q", key, curVal, snapVal)
			}
		default: // "overwrite"
		}

		layer[key] = snapVal
		result.Restored++
	}

	// Optionally remove keys not present in the snapshot.
	if removeExtra {
		for key := range layer {
			if _, inSnap := snap[key]; !inSnap {
				delete(layer, key)
			}
		}
	}

	if err := chain.SetLayer(env, layer); err != nil {
		return RollbackResult{}, fmt.Errorf("rollback: set layer: %w", err)
	}

	return result, nil
}
