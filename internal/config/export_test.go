package config

import (
	"strings"
	"testing"
)

func buildTestChain(t *testing.T) *Chain {
	t.Helper()

	basePath := writeTemp(t, "base.yaml", `
env: base
vars:
  APP_NAME: myapp
  DB_HOST: localhost
  DB_PASS: supersecret
  SECRET_KEY: topsecret
`)
	devPath := writeTemp(t, "dev.yaml", `
env: dev
vars:
  DB_HOST: dev-db.internal
`)

	c := NewChain()
	if err := c.AddLayer("base", basePath); err != nil {
		t.Fatalf("AddLayer base: %v", err)
	}
	if err := c.AddLayer("dev", devPath); err != nil {
		t.Fatalf("AddLayer dev: %v", err)
	}
	return c
}

func TestExportShellFormat(t *testing.T) {
	c := buildTestChain(t)
	var buf strings.Builder
	err := Export(c, "dev", ExportOptions{Format: FormatShell, SortKeys: true}, &buf)
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "export APP_NAME=\"") {
		t.Errorf("expected shell export line, got:\n%s", out)
	}
	if !strings.Contains(out, "export DB_HOST=\"dev-db.internal\"") {
		t.Errorf("expected dev DB_HOST override, got:\n%s", out)
	}
}

func TestExportDotenvFormat(t *testing.T) {
	c := buildTestChain(t)
	var buf strings.Builder
	err := Export(c, "dev", ExportOptions{Format: FormatDotenv, SortKeys: true}, &buf)
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "APP_NAME=myapp") {
		t.Errorf("expected APP_NAME in dotenv output, got:\n%s", out)
	}
}

func TestExportJSONFormat(t *testing.T) {
	c := buildTestChain(t)
	var buf strings.Builder
	err := Export(c, "base", ExportOptions{Format: FormatJSON, SortKeys: true}, &buf)
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	out := buf.String()
	if !strings.HasPrefix(out, "{") || !strings.HasSuffix(strings.TrimSpace(out), "}") {
		t.Errorf("expected JSON object, got:\n%s", out)
	}
	if !strings.Contains(out, `"APP_NAME": "myapp"`) {
		t.Errorf("expected APP_NAME in JSON output, got:\n%s", out)
	}
}

func TestExportRedactsSecrets(t *testing.T) {
	c := buildTestChain(t)
	var buf strings.Builder
	err := Export(c, "base", ExportOptions{Format: FormatDotenv, Redact: true, SortKeys: true}, &buf)
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "supersecret") || strings.Contains(out, "topsecret") {
		t.Errorf("expected secrets to be redacted, got:\n%s", out)
	}
}

func TestExportUnknownEnvReturnsError(t *testing.T) {
	c := buildTestChain(t)
	var buf strings.Builder
	err := Export(c, "staging", ExportOptions{Format: FormatShell}, &buf)
	if err == nil {
		t.Fatal("expected error for unknown env, got nil")
	}
}

func TestExportUnknownFormatReturnsError(t *testing.T) {
	c := buildTestChain(t)
	var buf strings.Builder
	err := Export(c, "base", ExportOptions{Format: "xml"}, &buf)
	if err == nil {
		t.Fatal("expected error for unknown format, got nil")
	}
}

// TestExportDevInheritsBaseVars verifies that env-specific exports include
// variables defined only in the base layer (i.e. inheritance works correctly).
func TestExportDevInheritsBaseVars(t *testing.T) {
	c := buildTestChain(t)
	var buf strings.Builder
	err := Export(c, "dev", ExportOptions{Format: FormatDotenv, SortKeys: true}, &buf)
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	out := buf.String()
	// APP_NAME is only defined in base, but should appear when exporting dev.
	if !strings.Contains(out, "APP_NAME=myapp") {
		t.Errorf("expected inherited APP_NAME from base layer, got:\n%s", out)
	}
	// DB_HOST should reflect the dev override, not the base value.
	if strings.Contains(out, "DB_HOST=localhost") {
		t.Errorf("expected dev override of DB_HOST, but got base value:\n%s", out)
	}
}
