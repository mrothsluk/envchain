package config

import (
	"strings"
	"testing"
)

func buildTruncateChain(t *testing.T) *Chain {
	t.Helper()
	base := writeTemp(t, "base.env", map[string]string{
		"APP_NAME":  "myapp",
		"APP_DESC":  "This is a very long description that exceeds the limit",
		"API_TOKEN": "supersecrettoken",
		"SHORT_VAL": "hi",
	})
	c, err := NewChain("dev", base)
	if err != nil {
		t.Fatalf("NewChain: %v", err)
	}
	c.MarkSecret("API_TOKEN")
	return c
}

func TestTruncateLayerShortValues(t *testing.T) {
	c := buildTruncateChain(t)
	out, results, err := TruncateLayer(c, "dev", TruncateOptions{MaxLen: 100})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["APP_NAME"] != "myapp" {
		t.Errorf("expected APP_NAME unchanged, got %q", out["APP_NAME"])
	}
	for _, r := range results {
		if r.Key == "APP_NAME" && r.Truncated {
			t.Errorf("APP_NAME should not be truncated")
		}
	}
}

func TestTruncateLayerLongValueTruncated(t *testing.T) {
	c := buildTruncateChain(t)
	out, results, err := TruncateLayer(c, "dev", TruncateOptions{MaxLen: 10, Suffix: "..."})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(out["APP_DESC"], "...") {
		t.Errorf("expected APP_DESC to end with '...', got %q", out["APP_DESC"])
	}
	if len(out["APP_DESC"]) != 13 {
		t.Errorf("expected length 13, got %d", len(out["APP_DESC"]))
	}
	var found bool
	for _, r := range results {
		if r.Key == "APP_DESC" {
			found = true
			if !r.Truncated {
				t.Errorf("APP_DESC should be marked truncated")
			}
		}
	}
	if !found {
		t.Errorf("APP_DESC not found in results")
	}
}

func TestTruncateLayerSecretIsRedacted(t *testing.T) {
	c := buildTruncateChain(t)
	out, _, err := TruncateLayer(c, "dev", TruncateOptions{MaxLen: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["API_TOKEN"] == "super" {
		t.Errorf("secret should not be truncated to plain value")
	}
	if !strings.Contains(out["API_TOKEN"], "*") && out["API_TOKEN"] != "[REDACTED]" {
		t.Errorf("expected API_TOKEN to be redacted, got %q", out["API_TOKEN"])
	}
}

func TestTruncateLayerInvalidMaxLen(t *testing.T) {
	c := buildTruncateChain(t)
	_, _, err := TruncateLayer(c, "dev", TruncateOptions{MaxLen: 0})
	if err == nil {
		t.Fatal("expected error for MaxLen=0")
	}
}

func TestTruncateLayerUnknownEnvReturnsError(t *testing.T) {
	c := buildTruncateChain(t)
	_, _, err := TruncateLayer(c, "unknown", TruncateOptions{MaxLen: 10})
	if err == nil {
		t.Fatal("expected error for unknown env")
	}
}

func TestTruncateLayerCustomSuffix(t *testing.T) {
	c := buildTruncateChain(t)
	out, _, err := TruncateLayer(c, "dev", TruncateOptions{MaxLen: 4, Suffix: "~~"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(out["APP_DESC"], "~~") {
		t.Errorf("expected custom suffix '~~', got %q", out["APP_DESC"])
	}
}
