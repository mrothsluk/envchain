package config

import (
	"fmt"
	"strings"
)

// TransformRule defines a named transformation to apply to env var values.
type TransformRule struct {
	Name string
	Fn   func(string) string
}

// TransformResult holds the outcome of transforming a single key.
type TransformResult struct {
	Key      string
	OldValue string
	NewValue string
	Changed  bool
}

// BuiltinTransforms provides a set of ready-made transform rules.
var BuiltinTransforms = map[string]TransformRule{
	"upper":      {Name: "upper", Fn: strings.ToUpper},
	"lower":      {Name: "lower", Fn: strings.ToLower},
	"trim":       {Name: "trim", Fn: strings.TrimSpace},
	"trim_quotes": {Name: "trim_quotes", Fn: trimQuotes},
}

func trimQuotes(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// TransformLayer applies one or more named transform rules to all non-secret
// values in the given environment within the chain. Secret keys are skipped.
// Returns a slice of TransformResult describing each key's outcome.
func TransformLayer(c *Chain, env string, ruleNames []string, dryRun bool) ([]TransformResult, error) {
	layer, ok := c.layers[env]
	if !ok {
		return nil, fmt.Errorf("unknown environment: %s", env)
	}

	var rules []TransformRule
	for _, name := range ruleNames {
		r, found := BuiltinTransforms[name]
		if !found {
			return nil, fmt.Errorf("unknown transform rule: %s", name)
		}
		rules = append(rules, r)
	}

	var results []TransformResult
	for _, k := range sortedKeys(layer.data) {
		if layer.secrets[k] {
			continue
		}
		old := layer.data[k]
		newVal := old
		for _, r := range rules {
			newVal = r.Fn(newVal)
		}
		changed := newVal != old
		if changed && !dryRun {
			layer.data[k] = newVal
		}
		results = append(results, TransformResult{Key: k, OldValue: old, NewValue: newVal, Changed: changed})
	}
	return results, nil
}
