package config

import "fmt"

// ScopeFilter defines which keys are visible in a given scope.
type ScopeFilter struct {
	AllowedPrefixes []string
	DeniedKeys      []string
}

// ScopeResult holds the filtered key-value pairs and any keys that were excluded.
type ScopeResult struct {
	Visible  map[string]string
	Excluded []string
}

// ScopeLayer returns only the keys from the resolved environment that match
// the given ScopeFilter. Keys matching a denied key are always excluded.
// If AllowedPrefixes is non-empty, only keys with a matching prefix are included.
func ScopeLayer(chain *Chain, env string, filter ScopeFilter) (ScopeResult, error) {
	resolved, err := chain.Resolve(env)
	if err != nil {
		return ScopeResult{}, fmt.Errorf("scope: resolve %q: %w", env, err)
	}

	denied := make(map[string]bool, len(filter.DeniedKeys))
	for _, k := range filter.DeniedKeys {
		denied[k] = true
	}

	result := ScopeResult{
		Visible: make(map[string]string),
	}

	for k, v := range resolved {
		if denied[k] {
			result.Excluded = append(result.Excluded, k)
			continue
		}
		if len(filter.AllowedPrefixes) == 0 {
			result.Visible[k] = v
			continue
		}
		matched := false
		for _, prefix := range filter.AllowedPrefixes {
			if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
				matched = true
				break
			}
		}
		if matched {
			result.Visible[k] = v
		} else {
			result.Excluded = append(result.Excluded, k)
		}
	}

	return result, nil
}
