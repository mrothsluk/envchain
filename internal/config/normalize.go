package config

import (
	"fmt"
	"strings"
)

// NormalizeRule describes a transformation to apply to env var values.
type NormalizeRule struct {
	TrimSpace    bool
	ToUpper      bool
	ToLower      bool
	ReplaceChars map[string]string
}

// NormalizeResult holds the outcome of normalizing a single key.
type NormalizeResult struct {
	Key      string
	OldValue string
	NewValue string
	Changed  bool
}

// NormalizeLayer applies the given rule to every value in the resolved env,
// returning a slice of results describing what changed.
// Secrets are never mutated — they are left as-is and marked unchanged.
func NormalizeLayer(c *Chain, env string, rule NormalizeRule) ([]NormalizeResult, error) {
	resolved, err := c.Resolve(env)
	if err != nil {
		return nil, fmt.Errorf("normalize: %w", err)
	}

	var results []NormalizeResult

	for key, entry := range resolved {
		if entry.Secret {
			results = append(results, NormalizeResult{
				Key:      key,
				OldValue: entry.Value,
				NewValue: entry.Value,
				Changed:  false,
			})
			continue
		}

		newVal := applyNormalizeRule(entry.Value, rule)
		results = append(results, NormalizeResult{
			Key:      key,
			OldValue: entry.Value,
			NewValue: newVal,
			Changed:  newVal != entry.Value,
		})
	}

	return results, nil
}

func applyNormalizeRule(val string, rule NormalizeRule) string {
	if rule.TrimSpace {
		val = strings.TrimSpace(val)
	}
	if rule.ToUpper {
		val = strings.ToUpper(val)
	}
	if rule.ToLower {
		val = strings.ToLower(val)
	}
	for old, replacement := range rule.ReplaceChars {
		val = strings.ReplaceAll(val, old, replacement)
	}
	return val
}
