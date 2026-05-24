package config

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func buildRedactReportChain(t *testing.T) *Chain {
	t.Helper()
	base := writeTemp(t, "base", map[string]string{
		"APP_HOST":  "localhost",
		"API_TOKEN": "secret-token",
		"DB_PASS":   "hunter2",
	})
	cfg, err := LoadConfig(base, []string{"APP_HOST", "API_TOKEN", "DB_PASS"}, []string{"API_TOKEN", "DB_PASS"})
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	c := NewChain()
	if err := c.AddLayer("base", cfg); err != nil {
		t.Fatalf("AddLayer: %v", err)
	}
	return c
}

func TestBuildRedactReportCountsRedacted(t *testing.T) {
	c := buildRedactReportChain(t)
	r, err := BuildRedactReport(c, "base")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.RedactedCount != 2 {
		t.Errorf("RedactedCount = %d, want 2", r.RedactedCount)
	}
	if r.TotalKeys != 3 {
		t.Errorf("TotalKeys = %d, want 3", r.TotalKeys)
	}
}

func TestBuildRedactReportKeysSorted(t *testing.T) {
	c := buildRedactReportChain(t)
	r, _ := BuildRedactReport(c, "base")
	if len(r.RedactedKeys) != 2 {
		t.Fatalf("want 2 redacted keys, got %d", len(r.RedactedKeys))
	}
	if r.RedactedKeys[0] != "API_TOKEN" || r.RedactedKeys[1] != "DB_PASS" {
		t.Errorf("unexpected order: %v", r.RedactedKeys)
	}
}

func TestBuildRedactReportUnknownEnvReturnsError(t *testing.T) {
	c := buildRedactReportChain(t)
	_, err := BuildRedactReport(c, "prod")
	if err == nil {
		t.Fatal("expected error for unknown env, got nil")
	}
}

func TestWriteRedactReportTextNoRedacted(t *testing.T) {
	r := &RedactReport{Env: "dev", TotalKeys: 2, RedactedCount: 0, RedactedKeys: nil}
	var buf bytes.Buffer
	if err := WriteRedactReport(&buf, r, "text"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "no redacted keys") {
		t.Errorf("expected 'no redacted keys' in output, got: %s", buf.String())
	}
}

func TestWriteRedactReportTextWithRedacted(t *testing.T) {
	r := &RedactReport{Env: "prod", TotalKeys: 3, RedactedCount: 1, RedactedKeys: []string{"SECRET"}}
	var buf bytes.Buffer
	if err := WriteRedactReport(&buf, r, "text"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "SECRET") {
		t.Errorf("expected key SECRET in output, got: %s", out)
	}
}

func TestWriteRedactReportJSONStructure(t *testing.T) {
	r := &RedactReport{Env: "staging", TotalKeys: 4, RedactedCount: 2, RedactedKeys: []string{"A", "B"}}
	var buf bytes.Buffer
	if err := WriteRedactReport(&buf, r, "json"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var out RedactReport
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if out.Env != "staging" || out.RedactedCount != 2 {
		t.Errorf("unexpected decoded report: %+v", out)
	}
}

func TestWriteRedactReportUnknownFormatReturnsError(t *testing.T) {
	r := &RedactReport{}
	var buf bytes.Buffer
	if err := WriteRedactReport(&buf, r, "xml"); err == nil {
		t.Fatal("expected error for unknown format, got nil")
	}
}
