package config

import (
	"fmt"
	"sort"
)

// MergeStrategy controls how conflicting keys are handled during a merge.
type MergeStrategy int

const (
	// MergeStrategyOverwrite replaces existing values with incoming values.
	MergeStrategyOverwrite MergeStrategy = iota
	// MergeStrategyKeepExisting retains the existing value when a conflict occurs.
	MergeStrategyKeepExisting
	// MergeStrategyError returns an error on any key conflict.
	MergeStrategyError
)

// MergeResult holds the outcome of a merge operation.
type MergeResult struct {
	// Merged is the combined key/value map.
	Merged map[string]string
	// Conflicts lists keys that had conflicting values (only populated for
	// MergeStrategyOverwrite and MergeStrategyKeepExisting).
	Conflicts []string
}

// MergeLayers combines two resolved environment maps according to the given
// strategy. base and overlay must not be nil.
func MergeLayers(base, overlay map[string]string, strategy MergeStrategy) (MergeResult, error) {
	result := MergeResult{
		Merged: make(map[string]string, len(base)),
	}

	// Copy base into merged.
	for k, v := range base {
		result.Merged[k] = v
	}

	conflictSet := map[string]struct{}{}

	for k, overlayVal := range overlay {
		existingVal, exists := result.Merged[k]
		if !exists {
			result.Merged[k] = overlayVal
			continue
		}

		if existingVal == overlayVal {
			// No real conflict — values are identical.
			continue
		}

		switch strategy {
		case MergeStrategyOverwrite:
			result.Merged[k] = overlayVal
			conflictSet[k] = struct{}{}
		case MergeStrategyKeepExisting:
			// Keep base value; just record the conflict.
			conflictSet[k] = struct{}{}
		case MergeStrategyError:
			return MergeResult{}, fmt.Errorf("merge conflict on key %q: base=%q overlay=%q", k, existingVal, overlayVal)
		default:
			return MergeResult{}, fmt.Errorf("unknown merge strategy: %d", strategy)
		}
	}

	for k := range conflictSet {
		result.Conflicts = append(result.Conflicts, k)
	}
	sort.Strings(result.Conflicts)

	return result, nil
}
