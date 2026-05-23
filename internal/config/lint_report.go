package config

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// ReportFormat controls the output format of a lint report.
type ReportFormat string

const (
	ReportText ReportFormat = "text"
	ReportJSON ReportFormat = "json"
)

// WriteLintReport writes a LintResult to w in the requested format.
func WriteLintReport(w io.Writer, result LintResult, format ReportFormat) error {
	switch format {
	case ReportJSON:
		return writeLintJSON(w, result)
	case ReportText:
		return writeLintText(w, result)
	default:
		return fmt.Errorf("unknown report format %q", format)
	}
}

func writeLintText(w io.Writer, result LintResult) error {
	if len(result.Findings) == 0 {
		_, err := fmt.Fprintln(w, "No issues found.")
		return err
	}
	var sb strings.Builder
	for _, f := range result.Findings {
		sb.WriteString(f.String())
		sb.WriteByte('\n')
	}
	_, err := fmt.Fprint(w, sb.String())
	return err
}

type lintJSONReport struct {
	Total    int           `json:"total"`
	Errors   int           `json:"errors"`
	Warnings int           `json:"warnings"`
	Findings []LintFinding `json:"findings"`
}

func writeLintJSON(w io.Writer, result LintResult) error {
	report := lintJSONReport{
		Total:    len(result.Findings),
		Findings: result.Findings,
	}
	for _, f := range result.Findings {
		switch f.Severity {
		case LintError:
			report.Errors++
		case LintWarn:
			report.Warnings++
		}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}
