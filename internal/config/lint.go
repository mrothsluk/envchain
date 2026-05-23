package config

import (
	"fmt"
	"strings"
)

// LintSeverity represents the severity level of a lint finding.
type LintSeverity string

const (
	LintWarn  LintSeverity = "warn"
	LintError LintSeverity = "error"
)

// LintFinding describes a single lint issue found in a layer or chain.
type LintFinding struct {
	Key      string
	Message  string
	Severity LintSeverity
}

func (f LintFinding) String() string {
	return fmt.Sprintf("[%s] %s: %s", f.Severity, f.Key, f.Message)
}

// LintResult holds all findings for a lint run.
type LintResult struct {
	Findings []LintFinding
}

func (r *LintResult) HasErrors() bool {
	for _, f := range r.Findings {
		if f.Severity == LintError {
			return true
		}
	}
	return false
}

func (r *LintResult) add(key, msg string, sev LintSeverity) {
	r.Findings = append(r.Findings, LintFinding{Key: key, Message: msg, Severity: sev})
}

// LintLayer checks a single layer's entries for common issues.
func LintLayer(layer map[string]string) LintResult {
	var result LintResult
	for k, v := range layer {
		if strings.TrimSpace(v) == "" {
			result.add(k, "value is blank or whitespace-only", LintWarn)
		}
		if strings.HasPrefix(v, " ") || strings.HasSuffix(v, " ") {
			result.add(k, "value has leading or trailing spaces", LintWarn)
		}
		if strings.ToUpper(k) != k {
			result.add(k, "key contains lowercase characters", LintError)
		}
		if strings.Contains(v, "\n") {
			result.add(k, "value contains newline characters", LintWarn)
		}
	}
	return result
}

// LintChain runs lint checks across all layers in a chain and aggregates findings.
func LintChain(chain *Chain) LintResult {
	var result LintResult
	seen := make(map[string]string) // key -> first layer name that defined it
	for _, layer := range chain.layers {
		lr := LintLayer(layer.data)
		result.Findings = append(result.Findings, lr.Findings...)
		for k := range layer.data {
			if prev, ok := seen[k]; ok {
				result.add(k, fmt.Sprintf("key also defined in layer %q (shadowed)", prev), LintWarn)
			} else {
				seen[k] = layer.name
			}
		}
	}
	return result
}
