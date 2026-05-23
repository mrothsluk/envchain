package config

import "fmt"

// TruncateOptions controls how values are truncated.
type TruncateOptions struct {
	// MaxLen is the maximum allowed value length before truncation.
	MaxLen int
	// Suffix is appended to truncated values (default: "...").
	Suffix string
	// SkipSecrets prevents truncation of secret values (they remain redacted).
	SkipSecrets bool
}

// TruncateResult holds a single truncation finding.
type TruncateResult struct {
	Key       string
	OrigLen   int
	Truncated bool
}

// TruncateLayer truncates long values in the given environment layer.
// Secret keys are redacted rather than truncated when SkipSecrets is false.
// Returns a slice of TruncateResult describing what was changed.
func TruncateLayer(c *Chain, env string, opts TruncateOptions) (map[string]string, []TruncateResult, error) {
	if opts.MaxLen <= 0 {
		return nil, nil, fmt.Errorf("truncate: MaxLen must be positive")
	}
	suffix := opts.Suffix
	if suffix == "" {
		suffix = "..."
	}

	resolved, err := c.Resolve(env)
	if err != nil {
		return nil, nil, fmt.Errorf("truncate: %w", err)
	}

	output := make(map[string]string, len(resolved))
	var results []TruncateResult

	for k, v := range resolved {
		isSecret := c.IsSecret(k)
		if isSecret && opts.SkipSecrets {
			output[k] = Redact(k, v)
			continue
		}
		if isSecret {
			output[k] = Redact(k, v)
			continue
		}
		if len(v) > opts.MaxLen {
			truncVal := v[:opts.MaxLen] + suffix
			output[k] = truncVal
			results = append(results, TruncateResult{Key: k, OrigLen: len(v), Truncated: true})
		} else {
			output[k] = v
			results = append(results, TruncateResult{Key: k, OrigLen: len(v), Truncated: false})
		}
	}
	return output, results, nil
}
