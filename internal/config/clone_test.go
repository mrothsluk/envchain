package config

import (
	"testing"
)

func buildCloneChain(t *testing.T) *Chain {
	t.Helper()

	base := map[string]Entry{
		"APP_NAME":  {Value: "envchain", Secret: false},
		"DB_PASS":   {Value: "secret123", Secret: true},
		"LOG_LEVEL": {Value: "info", Secret: false},
	}
	dev := map[string]Entry{
		"APP_NAME": {Value: "envchain-dev", Secret: false},
	}

	c := NewChain()
	c.AddLayerEntries("base", base)
	c.AddLayerEntries("dev", dev)
	c.AddLayerEntries("staging", map[string]Entry{})
	return c
}

func TestCloneAllKeys(t *testing.T) {
	c := buildCloneChain(t)
	n, err := CloneLayer(c, "base", "staging", CloneOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 3 {
		t.Fatalf("expected 3 keys copied, got %d", n)
	}
	val, _ := c.layers["staging"].entries["APP_NAME"]
	if val.Value != "envchain" {
		t.Errorf("expected APP_NAME=envchain, got %q", val.Value)
	}
}

func TestCloneSkipSecrets(t *testing.T) {
	c := buildCloneChain(t)
	n, err := CloneLayer(c, "base", "staging", CloneOptions{SkipSecrets: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected 2 keys copied, got %d", n)
	}
	if _, ok := c.layers["staging"].entries["DB_PASS"]; ok {
		t.Error("DB_PASS should not have been cloned")
	}
}

func TestCloneOnlyKeys(t *testing.T) {
	c := buildCloneChain(t)
	n, err := CloneLayer(c, "base", "staging", CloneOptions{OnlyKeys: []string{"LOG_LEVEL"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 key copied, got %d", n)
	}
	if _, ok := c.layers["staging"].entries["APP_NAME"]; ok {
		t.Error("APP_NAME should not have been cloned")
	}
}

func TestCloneUnknownSrcReturnsError(t *testing.T) {
	c := buildCloneChain(t)
	_, err := CloneLayer(c, "nope", "staging", CloneOptions{})
	if err == nil {
		t.Fatal("expected error for unknown source env")
	}
}

func TestCloneUnknownDstReturnsError(t *testing.T) {
	c := buildCloneChain(t)
	_, err := CloneLayer(c, "base", "prod", CloneOptions{})
	if err == nil {
		t.Fatal("expected error for unknown destination env")
	}
}
