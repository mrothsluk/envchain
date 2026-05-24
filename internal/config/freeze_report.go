package config

import (
	"encoding/json"
	"fmt"
	"io"
)

// FreezeSummary is the JSON-serialisable form of a freeze report.
type FreezeSummary struct {
	Env      string            `json:"env"`
	FrozenAt string            `json:"frozen_at"`
	Count    int               `json:"count"`
	Values   map[string]string `json:"values"`
}

// WriteFreezeReport writes a human-readable or JSON freeze report to w.
func WriteFreezeReport(w io.Writer, fl *FrozenLayer, format string) error {
	switch format {
	case "text", "":
		return writeFreezeText(w, fl)
	case "json":
		return writeFreezeJSON(w, fl)
	default:
		return fmt.Errorf("freeze report: unknown format %q", format)
	}
}

func writeFreezeText(w io.Writer, fl *FrozenLayer) error {
	fmt.Fprintf(w, "Frozen env : %s\n", fl.Env)
	fmt.Fprintf(w, "Frozen at  : %s\n", fl.FrozenAt.Format("2006-01-02T15:04:05Z"))
	fmt.Fprintf(w, "Keys       : %d\n", len(fl.Values))
	if len(fl.Values) == 0 {
		return nil
	}
	fmt.Fprintln(w)
	for _, k := range fl.SortedKeys() {
		fmt.Fprintf(w, "  %s=%s\n", k, fl.Values[k])
	}
	return nil
}

func writeFreezeJSON(w io.Writer, fl *FrozenLayer) error {
	summary := FreezeSummary{
		Env:      fl.Env,
		FrozenAt: fl.FrozenAt.Format("2006-01-02T15:04:05Z"),
		Count:    len(fl.Values),
		Values:   fl.Values,
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(summary)
}
