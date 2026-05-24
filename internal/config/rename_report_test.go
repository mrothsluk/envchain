package config

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestWriteRenameReportTextNoResults(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteRenameReport(&buf, nil, "text"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "no keys renamed") {
		t.Errorf("expected empty message, got: %q", buf.String())
	}
}

func TestWriteRenameReportTextWithResults(t *testing.T) {
	results := []RenameResult{
		{Env: "dev", OldKey: "BASE_URL", NewKey: "SERVICE_URL"},
		{Env: "prod", OldKey: "DB_HOST", NewKey: "DATABASE_HOST"},
	}
	var buf bytes.Buffer
	if err := WriteRenameReport(&buf, results, "text"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"dev", "BASE_URL", "SERVICE_URL", "prod", "DB_HOST", "DATABASE_HOST"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output:\n%s", want, out)
		}
	}
}

func TestWriteRenameReportJSONStructure(t *testing.T) {
	results := []RenameResult{
		{Env: "staging", OldKey: "OLD_TOKEN", NewKey: "NEW_TOKEN"},
	}
	var buf bytes.Buffer
	if err := WriteRenameReport(&buf, results, "json"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var parsed []map[string]string
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(parsed) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(parsed))
	}
	if parsed[0]["env"] != "staging" || parsed[0]["old_key"] != "OLD_TOKEN" || parsed[0]["new_key"] != "NEW_TOKEN" {
		t.Errorf("unexpected JSON content: %v", parsed[0])
	}
}

func TestWriteRenameReportUnknownFormatReturnsError(t *testing.T) {
	var buf bytes.Buffer
	err := WriteRenameReport(&buf, nil, "yaml")
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
}
