package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func buildFreezeChain(t *testing.T) *Chain {
	t.Helper()
	base := writeTemp(t, "base.env", "APP_NAME=envchain\nDB_PASSWORD=secret\nPORT=8080\n")
	prod := writeTemp(t, "prod.env", "PORT=443\n")
	c, err := NewChain("base", base)
	if err != nil {
		t.Fatalf("NewChain: %v", err)
	}
	if err := c.AddLayer("prod", prod); err != nil {
		t.Fatalf("AddLayer prod: %v", err)
	}
	return c
}

func TestFreezeLayerBase(t *testing.T) {
	c := buildFreezeChain(t)
	fl, err := FreezeLayer(c, "base", false)
	if err != nil {
		t.Fatalf("FreezeLayer: %v", err)
	}
	if fl.Env != "base" {
		t.Errorf("env = %q, want base", fl.Env)
	}
	if fl.Values["APP_NAME"] != "envchain" {
		t.Errorf("APP_NAME = %q, want envchain", fl.Values["APP_NAME"])
	}
	if fl.Values["PORT"] != "8080" {
		t.Errorf("PORT = %q, want 8080", fl.Values["PORT"])
	}
}

func TestFreezeLayerProdOverrides(t *testing.T) {
	c := buildFreezeChain(t)
	fl, err := FreezeLayer(c, "prod", false)
	if err != nil {
		t.Fatalf("FreezeLayer: %v", err)
	}
	if fl.Values["PORT"] != "443" {
		t.Errorf("PORT = %q, want 443", fl.Values["PORT"])
	}
}

func TestFreezeLayerRedactsSecrets(t *testing.T) {
	c := buildFreezeChain(t)
	fl, err := FreezeLayer(c, "base", true)
	if err != nil {
		t.Fatalf("FreezeLayer: %v", err)
	}
	if fl.Values["DB_PASSWORD"] != RedactedValue {
		t.Errorf("DB_PASSWORD = %q, want %q", fl.Values["DB_PASSWORD"], RedactedValue)
	}
	if fl.Values["APP_NAME"] != "envchain" {
		t.Errorf("APP_NAME should not be redacted")
	}
}

func TestFreezeLayerUnknownEnvReturnsError(t *testing.T) {
	c := buildFreezeChain(t)
	_, err := FreezeLayer(c, "unknown", false)
	if err == nil {
		t.Fatal("expected error for unknown env")
	}
}

func TestSaveAndLoadFreeze(t *testing.T) {
	c := buildFreezeChain(t)
	fl, err := FreezeLayer(c, "base", false)
	if err != nil {
		t.Fatalf("FreezeLayer: %v", err)
	}
	path := filepath.Join(t.TempDir(), "freeze.json")
	if err := SaveFreeze(fl, path); err != nil {
		t.Fatalf("SaveFreeze: %v", err)
	}
	loaded, err := LoadFreeze(path)
	if err != nil {
		t.Fatalf("LoadFreeze: %v", err)
	}
	if loaded.Env != fl.Env {
		t.Errorf("env mismatch: got %q want %q", loaded.Env, fl.Env)
	}
	if loaded.Values["APP_NAME"] != fl.Values["APP_NAME"] {
		t.Errorf("APP_NAME mismatch after round-trip")
	}
}

func TestFreezeFileIsValidJSON(t *testing.T) {
	c := buildFreezeChain(t)
	fl, _ := FreezeLayer(c, "base", false)
	path := filepath.Join(t.TempDir(), "freeze.json")
	_ = SaveFreeze(fl, path)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Errorf("freeze file is not valid JSON: %v", err)
	}
}

func TestFrozenLayerSortedKeys(t *testing.T) {
	c := buildFreezeChain(t)
	fl, _ := FreezeLayer(c, "base", false)
	keys := fl.SortedKeys()
	for i := 1; i < len(keys); i++ {
		if keys[i] < keys[i-1] {
			t.Errorf("keys not sorted at index %d: %v", i, keys)
		}
	}
}
