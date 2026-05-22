package config

import (
	"fmt"
	"sort"
	"strings"
)

// DiffEntry represents a single changed key between two environments.
type DiffEntry struct {
	Key      string
	BaseVal  string
	OtherVal string
	Status   string // "added", "removed", "changed", "unchanged"
}

// DiffResult holds all entries comparing two resolved environments.
type DiffResult struct {
	BaseEnv  string
	OtherEnv string
	Entries  []DiffEntry
}

// Changed returns only entries where the value differs between environments.
func (d *DiffResult) Changed() []DiffEntry {
	var out []DiffEntry
	for _, e := range d.Entries {
		if e.Status != "unchanged" {
			out = append(out, e)
		}
	}
	return out
}

// DiffChain compares the resolved values for two named environments within a Chain.
// Secret values are redacted before comparison so sensitive data is never surfaced.
func DiffChain(c *Chain, baseEnv, otherEnv string) (*DiffResult, error) {
	baseVars, err := resolveAll(c, baseEnv)
	if err != nil {
		return nil, fmt.Errorf("diff: resolving base env %q: %w", baseEnv, err)
	}
	otherVars, err := resolveAll(c, otherEnv)
	if err != nil {
		return nil, fmt.Errorf("diff: resolving other env %q: %w", otherEnv, err)
	}

	keys := unionKeys(baseVars, otherVars)
	sort.Strings(keys)

	result := &DiffResult{BaseEnv: baseEnv, OtherEnv: otherEnv}
	for _, k := range keys {
		bv, bOk := baseVars[k]
		ov, oOk := otherVars[k]
		entry := DiffEntry{Key: k, BaseVal: bv, OtherVal: ov}
		switch {
		case bOk && !oOk:
			entry.Status = "removed"
		case !bOk && oOk:
			entry.Status = "added"
		case bv == ov:
			entry.Status = "unchanged"
		default:
			entry.Status = "changed"
		}
		result.Entries = append(result.Entries, entry)
	}
	return result, nil
}

// resolveAll resolves every key known to the chain for the given environment,
// returning a map of key → redacted value.
func resolveAll(c *Chain, env string) (map[string]string, error) {
	out := make(map[string]string)
	for key, cfg := range c.keys {
		val, err := c.Resolve(env, key)
		if err != nil {
			return nil, err
		}
		if cfg.Secret {
			val = strings.Repeat("*", 8)
		}
		out[key] = val
	}
	return out, nil
}

func unionKeys(a, b map[string]string) []string {
	seen := make(map[string]struct{}, len(a)+len(b))
	for k := range a {
		seen[k] = struct{}{}
	}
	for k := range b {
		seen[k] = struct{}{}
	}
	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	return keys
}
