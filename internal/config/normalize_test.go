package config

import (
	"testing"
)

func buildNormalizeChain(t *testing.T) *Chain {
	t.Helper()
	base := writeTemp(t, "base.env", map[string]string{
		"APP_HOST": "  localhost  ",
		"APP_MODE": "development",
		"DB_PASS":  "s3cr3t",
	})
	prod := writeTemp(t, "prod.env", map[string]string{
		"APP_HOST": "prod.example.com",
		"APP_MODE": "production",
	})

	baseLayer, err := LoadConfig(base)
	if err != nil {
		t.Fatalf("load base: %v", err)
	}
	baseLayer.Secrets = []string{"DB_PASS"}

	prodLayer, err := LoadConfig(prod)
	if err != nil {
		t.Fatalf("load prod: %v", err)
	}

	c := NewChain()
	c.AddLayer("base", baseLayer)
	c.AddLayer("prod", prodLayer)
	return c
}

func TestNormalizeLayerTrimSpace(t *testing.T) {
	c := buildNormalizeChain(t)
	results, err := NormalizeLayer(c, "base", NormalizeRule{TrimSpace: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, r := range results {
		if r.Key == "APP_HOST" {
			if r.NewValue != "localhost" {
				t.Errorf("expected trimmed value, got %q", r.NewValue)
			}
			if !r.Changed {
				t.Error("expected Changed=true for APP_HOST")
			}
		}
	}
}

func TestNormalizeLayerToUpper(t *testing.T) {
	c := buildNormalizeChain(t)
	results, err := NormalizeLayer(c, "base", NormalizeRule{TrimSpace: true, ToUpper: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, r := range results {
		if r.Key == "APP_MODE" {
			if r.NewValue != "DEVELOPMENT" {
				t.Errorf("expected DEVELOPMENT, got %q", r.NewValue)
			}
		}
	}
}

func TestNormalizeLayerSecretNotMutated(t *testing.T) {
	c := buildNormalizeChain(t)
	results, err := NormalizeLayer(c, "base", NormalizeRule{ToUpper: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, r := range results {
		if r.Key == "DB_PASS" {
			if r.Changed {
				t.Error("secret should not be marked as changed")
			}
			if r.NewValue != "s3cr3t" {
				t.Errorf("secret value should be unchanged, got %q", r.NewValue)
			}
		}
	}
}

func TestNormalizeLayerReplaceChars(t *testing.T) {
	c := buildNormalizeChain(t)
	rule := NormalizeRule{ReplaceChars: map[string]string{"-": "_", " ": ""}}
	results, err := NormalizeLayer(c, "base", rule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, r := range results {
		if r.Key == "APP_HOST" {
			if r.NewValue != "localhost" {
				t.Errorf("expected spaces stripped, got %q", r.NewValue)
			}
		}
	}
}

func TestNormalizeLayerUnknownEnvReturnsError(t *testing.T) {
	c := buildNormalizeChain(t)
	_, err := NormalizeLayer(c, "staging", NormalizeRule{})
	if err == nil {
		t.Fatal("expected error for unknown env")
	}
}
