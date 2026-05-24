package config

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func sampleResults() []TransformResult {
	return []TransformResult{
		{Key: "APP_NAME", OldValue: "  myapp  ", NewValue: "myapp", Changed: true},
		{Key: "APP_ENV", OldValue: "development", NewValue: "development", Changed: false},
	}
}

func TestWriteTransformReportTextChanged(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteTransformReport(&buf, sampleResults(), "text"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "APP_NAME") {
		t.Error("expected APP_NAME in output")
	}
	if !strings.Contains(out, "1 key(s) transformed") {
		t.Error("expected summary line")
	}
}

func TestWriteTransformReportTextUnchanged(t *testing.T) {
	var buf bytes.Buffer
	results := []TransformResult{
		{Key: "APP_ENV", OldValue: "prod", NewValue: "prod", Changed: false},
	}
	if err := WriteTransformReport(&buf, results, ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "0 key(s) transformed") {
		t.Error("expected zero count")
	}
}

func TestWriteTransformReportJSONStructure(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteTransformReport(&buf, sampleResults(), "json"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var out []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(out) != 2 {
		t.Errorf("expected 2 entries, got %d", len(out))
	}
	if _, ok := out[0]["changed"]; !ok {
		t.Error("expected 'changed' field in JSON")
	}
}

func TestWriteTransformReportUnknownFormatReturnsError(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteTransformReport(&buf, sampleResults(), "xml"); err == nil {
		t.Fatal("expected error for unknown format")
	}
}
