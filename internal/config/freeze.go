package config

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"
)

// FrozenLayer is a read-only snapshot of a resolved environment layer.
type FrozenLayer struct {
	Env       string            `json:"env"`
	FrozenAt  time.Time         `json:"frozen_at"`
	Values    map[string]string `json:"values"`
}

// FreezeResult holds the outcome of a freeze operation.
type FreezeResult struct {
	Frozen  []string
	Skipped []string
}

// FreezeLayer resolves all keys for the given env and returns a FrozenLayer.
// Secret values are redacted when redactSecrets is true.
func FreezeLayer(c *Chain, env string, redactSecrets bool) (*FrozenLayer, error) {
	resolved, err := c.Resolve(env)
	if err != nil {
		return nil, fmt.Errorf("freeze: %w", err)
	}

	values := make(map[string]string, len(resolved))
	for k, v := range resolved {
		if redactSecrets && isSecret(k) {
			values[k] = RedactedValue
		} else {
			values[k] = v
		}
	}

	return &FrozenLayer{
		Env:      env,
		FrozenAt: time.Now().UTC(),
		Values:   values,
	}, nil
}

// SaveFreeze writes a FrozenLayer to a JSON file at path.
func SaveFreeze(fl *FrozenLayer, path string) error {
	data, err := json.MarshalIndent(fl, "", "  ")
	if err != nil {
		return fmt.Errorf("savefreeze marshal: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("savefreeze write: %w", err)
	}
	return nil
}

// LoadFreeze reads a FrozenLayer from a JSON file at path.
func LoadFreeze(path string) (*FrozenLayer, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("loadfreeze read: %w", err)
	}
	var fl FrozenLayer
	if err := json.Unmarshal(data, &fl); err != nil {
		return nil, fmt.Errorf("loadfreeze unmarshal: %w", err)
	}
	return &fl, nil
}

// SortedKeys returns the keys of the FrozenLayer in sorted order.
func (fl *FrozenLayer) SortedKeys() []string {
	keys := make([]string, 0, len(fl.Values))
	for k := range fl.Values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
