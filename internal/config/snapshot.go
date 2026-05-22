package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Snapshot captures the resolved state of a Chain at a point in time.
type Snapshot struct {
	Env       string            `json:"env"`
	Timestamp time.Time         `json:"timestamp"`
	Values    map[string]string `json:"values"`
}

// TakeSnapshot resolves all keys in the chain for the given env and returns a Snapshot.
func TakeSnapshot(c *Chain, env string, redactSecrets bool) (*Snapshot, error) {
	if _, ok := c.layers[env]; !ok && env != c.baseEnv {
		return nil, fmt.Errorf("unknown env %q", env)
	}

	keys := c.Keys()
	values := make(map[string]string, len(keys))
	for _, k := range keys {
		v, err := c.Resolve(k, env)
		if err != nil {
			return nil, fmt.Errorf("resolving key %q: %w", k, err)
		}
		if redactSecrets {
			v = Redact(k, v)
		}
		values[k] = v
	}

	return &Snapshot{
		Env:       env,
		Timestamp: time.Now().UTC(),
		Values:    values,
	}, nil
}

// SaveSnapshot writes a Snapshot to a JSON file at the given path.
func SaveSnapshot(s *Snapshot, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating snapshot file: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(s); err != nil {
		return fmt.Errorf("encoding snapshot: %w", err)
	}
	return nil
}

// LoadSnapshot reads a Snapshot from a JSON file at the given path.
func LoadSnapshot(path string) (*Snapshot, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening snapshot file: %w", err)
	}
	defer f.Close()

	var s Snapshot
	if err := json.NewDecoder(f).Decode(&s); err != nil {
		return nil, fmt.Errorf("decoding snapshot: %w", err)
	}
	return &s, nil
}

// DiffSnapshot compares two snapshots and returns added, removed, and changed keys.
func DiffSnapshot(before, after *Snapshot) (added, removed, changed []string) {
	for k, v := range after.Values {
		if bv, ok := before.Values[k]; !ok {
			added = append(added, k)
		} else if bv != v {
			changed = append(changed, k)
		}
	}
	for k := range before.Values {
		if _, ok := after.Values[k]; !ok {
			removed = append(removed, k)
		}
	}
	return
}
