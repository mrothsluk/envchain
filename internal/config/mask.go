package config

import (
	"regexp"
	"strings"
)

// MaskRule defines a pattern-based masking rule for env var values.
type MaskRule struct {
	Pattern     *regexp.Regexp
	Replacement string
}

// MaskOptions controls how MaskLayer behaves.
type MaskOptions struct {
	Rules       []MaskRule
	MaskSecrets bool
}

// MaskedEntry holds the original key and its masked value.
type MaskedEntry struct {
	Key      string
	Original string
	Masked   string
	Changed  bool
}

// MaskLayer applies masking rules to all values in the given environment layer.
// If MaskSecrets is true, values for keys marked as secrets are fully redacted.
// Returns a slice of MaskedEntry describing each key's transformation.
func MaskLayer(chain *Chain, env string, opts MaskOptions) ([]MaskedEntry, error) {
	resolved, err := chain.Resolve(env)
	if err != nil {
		return nil, err
	}

	entries := make([]MaskedEntry, 0, len(resolved))
	for k, v := range resolved {
		masked := applyMasks(v, opts.Rules)
		if opts.MaskSecrets && chain.IsSecret(k) {
			masked = "********"
		}
		entries = append(entries, MaskedEntry{
			Key:      k,
			Original: v,
			Masked:   masked,
			Changed:  masked != v,
		})
	}
	return entries, nil
}

// applyMasks runs all rules against a value in order, returning the final result.
func applyMasks(value string, rules []MaskRule) string {
	for _, r := range rules {
		if r.Pattern != nil {
			value = r.Pattern.ReplaceAllString(value, r.Replacement)
		}
	}
	return value
}

// DefaultMaskRules returns a set of built-in masking rules for common sensitive
// patterns such as tokens, passwords embedded in URLs, and credit-card numbers.
func DefaultMaskRules() []MaskRule {
	return []MaskRule{
		{
			Pattern:     regexp.MustCompile(`(?i)(password|passwd|secret|token)=[^&\s]+`),
			Replacement: "${1}=********",
		},
		{
			Pattern:     regexp.MustCompile(`\b\d{4}[- ]?\d{4}[- ]?\d{4}[- ]?\d{4}\b`),
			Replacement: "****-****-****-****",
		},
		{
			Pattern:     regexp.MustCompile(`\b[A-Za-z0-9+/]{40,}={0,2}\b`),
			Replacement: strings.Repeat("*", 8),
		},
	}
}
