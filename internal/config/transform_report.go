package config

import (
	"encoding/json"
	"fmt"
	"io"
)

// WriteTransformReport writes a human-readable or JSON summary of transform results.
func WriteTransformReport(w io.Writer, results []TransformResult, format string) error {
	switch format {
	case "text", "":
		return writeTransformText(w, results)
	case "json":
		return writeTransformJSON(w, results)
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

func writeTransformText(w io.Writer, results []TransformResult) error {
	changed := 0
	for _, r := range results {
		if r.Changed {
			changed++
			fmt.Fprintf(w, "  ~ %s: %q -> %q\n", r.Key, r.OldValue, r.NewValue)
		} else {
			fmt.Fprintf(w, "    %s: unchanged\n", r.Key)
		}
	}
	fmt.Fprintf(w, "\n%d key(s) transformed.\n", changed)
	return nil
}

func writeTransformJSON(w io.Writer, results []TransformResult) error {
	type entry struct {
		Key      string `json:"key"`
		OldValue string `json:"old_value"`
		NewValue string `json:"new_value"`
		Changed  bool   `json:"changed"`
	}
	var entries []entry
	for _, r := range results {
		entries = append(entries, entry{r.Key, r.OldValue, r.NewValue, r.Changed})
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(entries)
}
