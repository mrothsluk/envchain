package config

import (
	"testing"
)

func TestLintLayerCleanReturnsNoFindings(t *testing.T) {
	layer := map[string]string{
		"APP_ENV": "production",
		"PORT":    "8080",
	}
	result := LintLayer(layer)
	if len(result.Findings) != 0 {
		t.Fatalf("expected no findings, got %d: %v", len(result.Findings), result.Findings)
	}
}

func TestLintLayerLowercaseKeyIsError(t *testing.T) {
	layer := map[string]string{"app_env": "prod"}
	result := LintLayer(layer)
	if !result.HasErrors() {
		t.Fatal("expected error finding for lowercase key")
	}
	if result.Findings[0].Key != "app_env" {
		t.Errorf("unexpected key: %s", result.Findings[0].Key)
	}
}

func TestLintLayerBlankValueIsWarn(t *testing.T) {
	layer := map[string]string{"API_KEY": "   "}
	result := LintLayer(layer)
	if len(result.Findings) == 0 {
		t.Fatal("expected warning for blank value")
	}
	if result.Findings[0].Severity != LintWarn {
		t.Errorf("expected warn, got %s", result.Findings[0].Severity)
	}
}

func TestLintLayerLeadingSpaceIsWarn(t *testing.T) {
	layer := map[string]string{"DB_HOST": " localhost"}
	result := LintLayer(layer)
	var found bool
	for _, f := range result.Findings {
		if f.Severity == LintWarn && f.Key == "DB_HOST" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected warn for leading space in value")
	}
}

func TestLintLayerNewlineInValueIsWarn(t *testing.T) {
	layer := map[string]string{"CERT": "line1\nline2"}
	result := LintLayer(layer)
	var found bool
	for _, f := range result.Findings {
		if f.Key == "CERT" && f.Severity == LintWarn {
			found = true
		}
	}
	if !found {
		t.Fatal("expected warn for newline in value")
	}
}

func TestLintChainDetectsShadowedKey(t *testing.T) {
	base := map[string]string{"PORT": "8080"}
	override := map[string]string{"PORT": "9090"}
	chain := &Chain{}
	chain.layers = []layer{
		{name: "base", data: base},
		{name: "dev", data: override},
	}
	result := LintChain(chain)
	var found bool
	for _, f := range result.Findings {
		if f.Key == "PORT" && f.Severity == LintWarn {
			found = true
		}
	}
	if !found {
		t.Fatal("expected warn for shadowed key PORT")
	}
}

func TestLintFindingString(t *testing.T) {
	f := LintFinding{Key: "FOO", Message: "some issue", Severity: LintError}
	got := f.String()
	expected := "[error] FOO: some issue"
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}
