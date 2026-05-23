package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// PinEntry records a resolved environment snapshot pinned to a specific version.
type PinEntry struct {
	Env       string            `json:"env"`
	PinnedAt  time.Time         `json:"pinned_at"`
	PinnedBy  string            `json:"pinned_by"`
	Values    map[string]string `json:"values"`
	Comment   string            `json:"comment,omitempty"`
}

// PinFile holds all pin entries persisted to disk.
type PinFile struct {
	Entries []PinEntry `json:"entries"`
}

// PinLayer resolves the given env in chain and records a PinEntry.
func PinLayer(chain *Chain, env, pinnedBy, comment string) (PinEntry, error) {
	resolved, err := chain.Resolve(env)
	if err != nil {
		return PinEntry{}, fmt.Errorf("pin: resolve %q: %w", env, err)
	}
	return PinEntry{
		Env:      env,
		PinnedAt: time.Now().UTC(),
		PinnedBy: pinnedBy,
		Values:   resolved,
		Comment:  comment,
	}, nil
}

// SavePin appends a PinEntry to the given file path (creates if absent).
func SavePin(path string, entry PinEntry) error {
	pf, _ := LoadPin(path) // ignore error; treat missing file as empty
	pf.Entries = append(pf.Entries, entry)
	data, err := json.MarshalIndent(pf, "", "  ")
	if err != nil {
		return fmt.Errorf("pin: marshal: %w", err)
	}
	return os.WriteFile(path, data, 0o600)
}

// LoadPin reads a PinFile from disk.
func LoadPin(path string) (PinFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return PinFile{}, err
	}
	var pf PinFile
	if err := json.Unmarshal(data, &pf); err != nil {
		return PinFile{}, fmt.Errorf("pin: unmarshal: %w", err)
	}
	return pf, nil
}

// LatestPin returns the most recently saved entry for env, or an error if none.
func LatestPin(pf PinFile, env string) (PinEntry, error) {
	for i := len(pf.Entries) - 1; i >= 0; i-- {
		if pf.Entries[i].Env == env {
			return pf.Entries[i], nil
		}
	}
	return PinEntry{}, fmt.Errorf("pin: no entry found for env %q", env)
}
