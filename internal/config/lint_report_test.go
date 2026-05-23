package config

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestWriteLintReportTextNoFindings(t *testing.T) {
	var buf bytes.Buffer
	err := WriteLintReport(&buf, LintResult{}, ReportText)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "No issues") {
		t.Errorf("expected no-issues message, got: %s", buf.String())
	}
}

func TestWriteLintReportTextWithFindings(t *testing.T) {
	result := LintResult{
		Findings: []LintFinding{
			{Key: "bad_key", Message: "lowercase", Severity: LintError},
			{Key: "API_KEY", Message: "blank value", Severity: LintWarn},
		},
	}
	var buf bytes.Buffer
	if err := WriteLintReport(&buf, result, ReportText); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "bad_key") {
		t.Error("expected bad_key in text output")
	}
	if !strings.Contains(out, "API_KEY") {
		t.Error("expected API_KEY in text output")
	}
}

func TestWriteLintReportJSONStructure(t *testing.T) {
	result := LintResult{
		Findings: []LintFinding{
			{Key: "bad_key", Message: "lowercase", Severity: LintError},
			{Key: "HOST", Message: "trailing space", Severity: LintWarn},
		},
	}
	var buf bytes.Buffer
	if err := WriteLintReport(&buf, result, ReportJSON); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var report struct {
		Total    int `json:"total"`
		Errors   int `json:"errors"`
		Warnings int `json:"warnings"`
	}
	if err := json.Unmarshal(buf.Bytes(), &report); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if report.Total != 2 {
		t.Errorf("expected total 2, got %d", report.Total)
	}
	if report.Errors != 1 {
		t.Errorf("expected 1 error, got %d", report.Errors)
	}
	if report.Warnings != 1 {
		t.Errorf("expected 1 warning, got %d", report.Warnings)
	}
}

func TestWriteLintReportUnknownFormatReturnsError(t *testing.T) {
	var buf bytes.Buffer
	err := WriteLintReport(&buf, LintResult{}, ReportFormat("xml"))
	if err == nil {
		t.Fatal("expected error for unknown format")
	}
}
