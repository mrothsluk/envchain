package config

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
)

// RedactReport describes which keys were redacted in a layer.
type RedactReport struct {
	Env          string   `json:"env"`
	RedactedKeys []string `json:"redacted_keys"`
	TotalKeys    int      `json:"total_keys"`
	RedactedCount int     `json:"redacted_count"`
}

// BuildRedactReport inspects the resolved values for the given env and
// returns a report listing every key whose value is the redacted sentinel.
func BuildRedactReport(c *Chain, env string) (*RedactReport, error) {
	layer, err := c.Resolve(env)
	if err != nil {
		return nil, fmt.Errorf("redact report: %w", err)
	}

	report := &RedactReport{
		Env:       env,
		TotalKeys: len(layer),
	}

	for k, v := range layer {
		if v == redactedSentinel {
			report.RedactedKeys = append(report.RedactedKeys, k)
		}
	}
	sort.Strings(report.RedactedKeys)
	report.RedactedCount = len(report.RedactedKeys)
	return report, nil
}

// WriteRedactReport writes the report to w in the requested format ("text" or "json").
func WriteRedactReport(w io.Writer, r *RedactReport, format string) error {
	switch format {
	case "json":
		return writeRedactJSON(w, r)
	case "text", "":
		return writeRedactText(w, r)
	default:
		return fmt.Errorf("unknown format %q: want text or json", format)
	}
}

func writeRedactText(w io.Writer, r *RedactReport) error {
	if r.RedactedCount == 0 {
		_, err := fmt.Fprintf(w, "[%s] no redacted keys (%d total)\n", r.Env, r.TotalKeys)
		return err
	}
	if _, err := fmt.Fprintf(w, "[%s] %d/%d keys redacted:\n", r.Env, r.RedactedCount, r.TotalKeys); err != nil {
		return err
	}
	for _, k := range r.RedactedKeys {
		if _, err := fmt.Fprintf(w, "  - %s\n", k); err != nil {
			return err
		}
	}
	return nil
}

func writeRedactJSON(w io.Writer, r *RedactReport) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}

// redactedSentinel must match the value used in Redact (env.go).
const redactedSentinel = "***REDACTED***"
