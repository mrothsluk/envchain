package config

import (
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"
)

// WriteRenameReport writes rename results to w in the requested format.
// Supported formats: "text", "json".
func WriteRenameReport(w io.Writer, results []RenameResult, format string) error {
	switch format {
	case "text":
		return writeRenameText(w, results)
	case "json":
		return writeRenameJSON(w, results)
	default:
		return fmt.Errorf("rename report: unsupported format %q", format)
	}
}

func writeRenameText(w io.Writer, results []RenameResult) error {
	if len(results) == 0 {
		_, err := fmt.Fprintln(w, "no keys renamed")
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ENV\tOLD KEY\tNEW KEY")
	for _, r := range results {
		fmt.Fprintf(tw, "%s\t%s\t%s\n", r.Env, r.OldKey, r.NewKey)
	}
	return tw.Flush()
}

func writeRenameJSON(w io.Writer, results []RenameResult) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	type jsonResult struct {
		Env    string `json:"env"`
		OldKey string `json:"old_key"`
		NewKey string `json:"new_key"`
	}
	out := make([]jsonResult, len(results))
	for i, r := range results {
		out[i] = jsonResult{Env: r.Env, OldKey: r.OldKey, NewKey: r.NewKey}
	}
	return enc.Encode(out)
}
