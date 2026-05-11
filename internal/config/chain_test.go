package config

import (
	"testing"
)

func TestChainResolveBase(t *testing.T) {
	c := NewChain()
	_ = c.AddLayer(EnvBase, map[string]string{"HOST": "localhost", "PORT": "5432"})

	got, err := c.Resolve(EnvBase)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["HOST"] != "localhost" || got["PORT"] != "5432" {
		t.Errorf("unexpected base result: %v", got)
	}
}

func TestChainResolveDevOverridesBase(t *testing.T) {
	c := NewChain()
	_ = c.AddLayer(EnvBase, map[string]string{"HOST": "localhost", "DEBUG": "false"})
	_ = c.AddLayer(EnvDev, map[string]string{"DEBUG": "true"})

	got, err := c.Resolve(EnvDev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["DEBUG"] != "true" {
		t.Errorf("expected dev to override base DEBUG, got %q", got["DEBUG"])
	}
	if got["HOST"] != "localhost" {
		t.Errorf("expected HOST to be inherited from base, got %q", got["HOST"])
	}
}

func TestChainResolveProdSkipsDevLayer(t *testing.T) {
	c := NewChain()
	_ = c.AddLayer(EnvBase, map[string]string{"HOST": "localhost"})
	_ = c.AddLayer(EnvDev, map[string]string{"HOST": "dev-host"})
	_ = c.AddLayer(EnvProd, map[string]string{"HOST": "prod-host"})

	got, err := c.Resolve(EnvProd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["HOST"] != "prod-host" {
		t.Errorf("expected prod HOST, got %q", got["HOST"])
	}
}

func TestChainResolveUnknownEnvReturnsError(t *testing.T) {
	c := NewChain()
	_, err := c.Resolve("unknown")
	if err == nil {
		t.Fatal("expected error for unknown env, got nil")
	}
}

func TestAddLayerUnknownEnvReturnsError(t *testing.T) {
	c := NewChain()
	err := c.AddLayer("canary", map[string]string{"X": "1"})
	if err == nil {
		t.Fatal("expected error for unknown layer env, got nil")
	}
}
