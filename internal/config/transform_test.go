package config

import (
	"testing"
)

func buildTransformChain(t *testing.T) *Chain {
	t.Helper()
	base := writeTemp(t, "BASE", map[string]string{
		"APP_NAME":  "  myapp  ",
		"APP_ENV":   "development",
		"API_TOKEN": "secret-value",
	})
	cfg, err := LoadConfig(base)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	cfg.Secrets = []string{"API_TOKEN"}
	c := NewChain()
	if err := c.AddLayer("base", cfg); err != nil {
		t.Fatalf("AddLayer: %v", err)
	}
	return c
}

func TestTransformLayerTrim(t *testing.T) {
	c := buildTransformChain(t)
	results, err := TransformLayer(c, "base", []string{"trim"}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, r := range results {
		if r.Key == "APP_NAME" {
			if r.NewValue != "myapp" {
				t.Errorf("expected 'myapp', got %q", r.NewValue)
			}
			if !r.Changed {
				t.Error("expected Changed=true")
			}
		}
	}
}

func TestTransformLayerSkipsSecrets(t *testing.T) {
	c := buildTransformChain(t)
	results, err := TransformLayer(c, "base", []string{"upper"}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, r := range results {
		if r.Key == "API_TOKEN" {
			t.Error("secret key should have been skipped")
		}
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestTransformLayerDryRunDoesNotMutate(t *testing.T) {
	c := buildTransformChain(t)
	_, err := TransformLayer(c, "base", []string{"upper"}, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, _ := c.Resolve("base", "APP_ENV")
	if val != "development" {
		t.Errorf("dry run mutated value: got %q", val)
	}
}

func TestTransformLayerUnknownRuleReturnsError(t *testing.T) {
	c := buildTransformChain(t)
	_, err := TransformLayer(c, "base", []string{"nonexistent"}, false)
	if err == nil {
		t.Fatal("expected error for unknown rule")
	}
}

func TestTransformLayerUnknownEnvReturnsError(t *testing.T) {
	c := buildTransformChain(t)
	_, err := TransformLayer(c, "ghost", []string{"trim"}, false)
	if err == nil {
		t.Fatal("expected error for unknown env")
	}
}

func TestTransformLayerMultipleRules(t *testing.T) {
	c := buildTransformChain(t)
	results, err := TransformLayer(c, "base", []string{"trim", "upper"}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, r := range results {
		if r.Key == "APP_NAME" && r.NewValue != "MYAPP" {
			t.Errorf("expected 'MYAPP', got %q", r.NewValue)
		}
	}
}
