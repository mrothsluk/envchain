package config

import (
	"testing"
)

func buildInterpolateChain(t *testing.T) *Chain {
	t.Helper()
	base := writeTemp(t, "base", map[string]string{
		"APP_HOST": "localhost",
		"APP_URL":  "http://${APP_HOST}:8080",
		"GREETING": "hello $APP_HOST",
	})
	prod := writeTemp(t, "prod", map[string]string{
		"APP_HOST": "prod.example.com",
	})
	c, err := NewChain(base, []string{"dev", "prod"})
	if err != nil {
		t.Fatalf("NewChain: %v", err)
	}
	if err := c.AddLayer("prod", prod); err != nil {
		t.Fatalf("AddLayer: %v", err)
	}
	return c
}

func TestInterpolateLayerResolvesReferences(t *testing.T) {
	input := map[string]string{
		"HOST":    "localhost",
		"URL":     "http://${HOST}:9000",
		"LABEL":   "env=$HOST",
	}
	res := InterpolateLayer(input, false)
	if res.Values["URL"] != "http://localhost:9000" {
		t.Errorf("URL: got %q", res.Values["URL"])
	}
	if res.Values["LABEL"] != "env=localhost" {
		t.Errorf("LABEL: got %q", res.Values["LABEL"])
	}
	if len(res.Missing) != 0 {
		t.Errorf("unexpected missing: %v", res.Missing)
	}
}

func TestInterpolateLayerMissingKeyCollected(t *testing.T) {
	input := map[string]string{
		"URL": "http://${UNKNOWN_HOST}:8080",
	}
	res := InterpolateLayer(input, false)
	if len(res.Missing) != 1 || res.Missing[0] != "UNKNOWN_HOST" {
		t.Errorf("missing: got %v", res.Missing)
	}
	// Original token preserved when unresolved.
	if res.Values["URL"] != "http://${UNKNOWN_HOST}:8080" {
		t.Errorf("URL: got %q", res.Values["URL"])
	}
}

func TestInterpolateLayerFallbackToOS(t *testing.T) {
	t.Setenv("OS_VAR", "from-os")
	input := map[string]string{
		"COMBINED": "${OS_VAR}-suffix",
	}
	res := InterpolateLayer(input, true)
	if res.Values["COMBINED"] != "from-os-suffix" {
		t.Errorf("COMBINED: got %q", res.Values["COMBINED"])
	}
	if len(res.Missing) != 0 {
		t.Errorf("unexpected missing: %v", res.Missing)
	}
}

func TestInterpolateChainProdOverridesHost(t *testing.T) {
	c := buildInterpolateChain(t)
	res, err := InterpolateChain(c, "prod", false)
	if err != nil {
		t.Fatalf("InterpolateChain: %v", err)
	}
	if res.Values["APP_URL"] != "http://prod.example.com:8080" {
		t.Errorf("APP_URL: got %q", res.Values["APP_URL"])
	}
}

func TestInterpolateChainUnknownEnvReturnsError(t *testing.T) {
	c := buildInterpolateChain(t)
	_, err := InterpolateChain(c, "staging", false)
	if err == nil {
		t.Fatal("expected error for unknown env")
	}
}

func TestInterpolateLayerNoReferencesUnchanged(t *testing.T) {
	input := map[string]string{
		"PLAIN": "no-refs-here",
	}
	res := InterpolateLayer(input, false)
	if res.Values["PLAIN"] != "no-refs-here" {
		t.Errorf("PLAIN: got %q", res.Values["PLAIN"])
	}
}
