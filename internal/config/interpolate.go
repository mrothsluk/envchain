package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// interpolatePattern matches ${VAR} and $VAR style references.
var interpolatePattern = regexp.MustCompile(`\$\{([A-Z_][A-Z0-9_]*)\}|\$([A-Z_][A-Z0-9_]*)`)

// InterpolateResult holds the interpolated values and any warnings.
type InterpolateResult struct {
	Values   map[string]string
	Missing  []string
}

// InterpolateLayer resolves variable references within a layer's values.
// References can point to other keys in the resolved map or to OS environment
// variables. Missing references are collected rather than causing an error.
func InterpolateLayer(resolved map[string]string, fallbackToOS bool) InterpolateResult {
	result := make(map[string]string, len(resolved))
	var missing []string
	seen := map[string]bool{}

	for k, v := range resolved {
		interpolated, refs := interpolateValue(v, resolved, fallbackToOS)
		result[k] = interpolated
		for _, ref := range refs {
			if !seen[ref] {
				seen[ref] = true
				missing = append(missing, ref)
			}
		}
	}

	return InterpolateResult{Values: result, Missing: missing}
}

// interpolateValue replaces variable references in a single string value.
// Returns the substituted string and a list of keys that could not be resolved.
func interpolateValue(value string, lookup map[string]string, fallbackToOS bool) (string, []string) {
	var missing []string

	replaced := interpolatePattern.ReplaceAllStringFunc(value, func(match string) string {
		key := extractKey(match)
		if v, ok := lookup[key]; ok {
			return v
		}
		if fallbackToOS {
			if v, ok := os.LookupEnv(key); ok {
				return v
			}
		}
		missing = append(missing, key)
		return match
	})

	return replaced, missing
}

// extractKey strips the sigil and braces from a matched token.
func extractKey(match string) string {
	match = strings.TrimPrefix(match, "$")
	match = strings.Trim(match, "{}")
	return match
}

// InterpolateChain resolves a full chain for a given environment and then
// interpolates all variable references, returning the final flat map.
func InterpolateChain(c *Chain, env string, fallbackToOS bool) (InterpolateResult, error) {
	resolved, err := c.Resolve(env)
	if err != nil {
		return InterpolateResult{}, fmt.Errorf("interpolate: resolve %q: %w", env, err)
	}
	return InterpolateLayer(resolved, fallbackToOS), nil
}
