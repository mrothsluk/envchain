package config

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// ExportFormat defines the output format for exported environment variables.
type ExportFormat string

const (
	FormatShell  ExportFormat = "shell"
	FormatDotenv ExportFormat = "dotenv"
	FormatJSON   ExportFormat = "json"
)

// ExportOptions controls how variables are exported.
type ExportOptions struct {
	Format  ExportFormat
	Redact  bool
	SortKeys bool
}

// Export writes the resolved environment variables from the chain to w
// using the specified format. Secret values are redacted if opts.Redact is true.
func Export(c *Chain, env string, opts ExportOptions, w io.Writer) error {
	resolved, err := c.Resolve(env)
	if err != nil {
		return fmt.Errorf("export: resolve %q: %w", env, err)
	}

	if opts.Redact {
		resolved = Redact(resolved)
	}

	keys := make([]string, 0, len(resolved))
	for k := range resolved {
		keys = append(keys, k)
	}
	if opts.SortKeys {
		sort.Strings(keys)
	}

	switch opts.Format {
	case FormatShell:
		return exportShell(keys, resolved, w)
	case FormatDotenv:
		return exportDotenv(keys, resolved, w)
	case FormatJSON:
		return exportJSON(keys, resolved, w)
	default:
		return fmt.Errorf("export: unknown format %q", opts.Format)
	}
}

func exportShell(keys []string, vars map[string]string, w io.Writer) error {
	for _, k := range keys {
		if _, err := fmt.Fprintf(w, "export %s=%q\n", k, vars[k]); err != nil {
			return err
		}
	}
	return nil
}

func exportDotenv(keys []string, vars map[string]string, w io.Writer) error {
	for _, k := range keys {
		v := vars[k]
		// Quote value if it contains spaces or special chars.
		if strings.ContainsAny(v, " \t\n#") {
			v = fmt.Sprintf("%q", v)
		}
		if _, err := fmt.Fprintf(w, "%s=%s\n", k, v); err != nil {
			return err
		}
	}
	return nil
}

func exportJSON(keys []string, vars map[string]string, w io.Writer) error {
	if _, err := fmt.Fprint(w, "{\n"); err != nil {
		return err
	}
	for i, k := range keys {
		comma := ","
		if i == len(keys)-1 {
			comma = ""
		}
		if _, err := fmt.Fprintf(w, "  %q: %q%s\n", k, vars[k], comma); err != nil {
			return err
		}
	}
	_, err := fmt.Fprint(w, "}\n")
	return err
}
