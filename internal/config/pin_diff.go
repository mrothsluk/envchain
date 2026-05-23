package config

import "fmt"

// PinDiffResult describes how the live environment diverges from a pin.
type PinDiffResult struct {
	Key      string
	Pinned   string
	Live     string
	Status   string // "changed", "added", "removed"
}

// DiffAgainstPin compares the live resolved values for env against a PinEntry.
// It returns a slice of PinDiffResult for any diverging keys.
func DiffAgainstPin(chain *Chain, entry PinEntry) ([]PinDiffResult, error) {
	live, err := chain.Resolve(entry.Env)
	if err != nil {
		return nil, fmt.Errorf("pin diff: resolve %q: %w", entry.Env, err)
	}

	var results []PinDiffResult

	for k, pinnedVal := range entry.Values {
		liveVal, ok := live[k]
		if !ok {
			results = append(results, PinDiffResult{Key: k, Pinned: pinnedVal, Live: "", Status: "removed"})
			continue
		}
		if liveVal != pinnedVal {
			results = append(results, PinDiffResult{Key: k, Pinned: pinnedVal, Live: liveVal, Status: "changed"})
		}
	}

	for k, liveVal := range live {
		if _, exists := entry.Values[k]; !exists {
			results = append(results, PinDiffResult{Key: k, Pinned: "", Live: liveVal, Status: "added"})
		}
	}

	return results, nil
}
