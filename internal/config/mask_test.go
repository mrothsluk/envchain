package config

import (
	"regexp"
	"testing"
)

func buildMaskChain(t *testing.T) *Chain {
	t.Helper()
	base := writeTemp(t, "base", map[string]string{
		"API_URL":      "https://host?token=supersecret",
		"PLAIN":        "hello world",
		"CARD":         "4111 1111 1111 1111",
		"SECRET_TOKEN": "abc123",
	})
	cfg, err := LoadConfig(base)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	cfg.Secrets = map[string]bool{"SECRET_TOKEN": true}
	chain, err := NewChain(cfg, "base")
	if err != nil {
		t.Fatalf("NewChain: %v", err)
	}
	return chain
}

func TestMaskLayerNoRulesNoChange(t *testing.T) {
	chain := buildMaskChain(t)
	entries, err := MaskLayer(chain, "base", MaskOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, e := range entries {
		if e.Changed {
			t.Errorf("key %s unexpectedly changed: %q -> %q", e.Key, e.Original, e.Masked)
		}
	}
}

func TestMaskLayerCustomRule(t *testing.T) {
	chain := buildMaskChain(t)
	opts := MaskOptions{
		Rules: []MaskRule{
			{Pattern: regexp.MustCompile(`world`), Replacement: "***"},
		},
	}
	entries, err := MaskLayer(chain, "base", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, e := range entries {
		if e.Key == "PLAIN" {
			if e.Masked != "hello ***" {
				t.Errorf("expected 'hello ***', got %q", e.Masked)
			}
			if !e.Changed {
				t.Error("expected Changed=true for PLAIN")
			}
		}
	}
}

func TestMaskLayerMaskSecretsFlag(t *testing.T) {
	chain := buildMaskChain(t)
	opts := MaskOptions{MaskSecrets: true}
	entries, err := MaskLayer(chain, "base", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, e := range entries {
		if e.Key == "SECRET_TOKEN" {
			if e.Masked != "********" {
				t.Errorf("expected redacted secret, got %q", e.Masked)
			}
			return
		}
	}
	t.Error("SECRET_TOKEN entry not found")
}

func TestMaskLayerDefaultRulesRedactToken(t *testing.T) {
	chain := buildMaskChain(t)
	opts := MaskOptions{Rules: DefaultMaskRules()}
	entries, err := MaskLayer(chain, "base", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, e := range entries {
		if e.Key == "API_URL" && !e.Changed {
			t.Errorf("expected API_URL to be masked, got %q", e.Masked)
		}
		if e.Key == "CARD" && !e.Changed {
			t.Errorf("expected CARD to be masked, got %q", e.Masked)
		}
	}
}

func TestMaskLayerUnknownEnvReturnsError(t *testing.T) {
	chain := buildMaskChain(t)
	_, err := MaskLayer(chain, "nonexistent", MaskOptions{})
	if err == nil {
		t.Fatal("expected error for unknown env, got nil")
	}
}
