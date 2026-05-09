package config

import (
	"os"
	"path/filepath"
	"testing"
)

const sampleConfig = `
version: 1
layers:
  base:
    - key: APP_NAME
      value: envchain
    - key: LOG_LEVEL
      value: info
    - key: DB_PASSWORD
      value: base_secret
      secret: true
  dev:
    - key: LOG_LEVEL
      value: debug
    - key: DB_PASSWORD
      value: dev_secret
      secret: true
  prod:
    - key: LOG_LEVEL
      value: warn
    - key: DB_PASSWORD
      value: prod_secret
      secret: true
`

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "envchain.yaml")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}
	return p
}

func TestLoadConfig(t *testing.T) {
	p := writeTemp(t, sampleConfig)
	cfg, err := LoadConfig(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Version != 1 {
		t.Errorf("expected version 1, got %d", cfg.Version)
	}
	if len(cfg.Layers["base"]) != 3 {
		t.Errorf("expected 3 base vars, got %d", len(cfg.Layers["base"]))
	}
}

func TestResolveDevOverridesBase(t *testing.T) {
	p := writeTemp(t, sampleConfig)
	cfg, _ := LoadConfig(p)

	vars, err := cfg.Resolve(EnvDev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	byKey := make(map[string]EnvVar)
	for _, v := range vars {
		byKey[v.Key] = v
	}

	if byKey["LOG_LEVEL"].Value != "debug" {
		t.Errorf("expected LOG_LEVEL=debug, got %s", byKey["LOG_LEVEL"].Value)
	}
	if byKey["APP_NAME"].Value != "envchain" {
		t.Errorf("expected APP_NAME=envchain, got %s", byKey["APP_NAME"].Value)
	}
}

func TestResolveUnknownEnvReturnsError(t *testing.T) {
	p := writeTemp(t, sampleConfig)
	cfg, _ := LoadConfig(p)

	_, err := cfg.Resolve("unknown")
	if err == nil {
		t.Error("expected error for unknown environment, got nil")
	}
}

func TestRedact(t *testing.T) {
	secret := EnvVar{Key: "DB_PASSWORD", Value: "s3cr3t", Secret: true}
	plain := EnvVar{Key: "APP_NAME", Value: "envchain", Secret: false}

	if Redact(secret) != "********" {
		t.Errorf("expected redacted secret, got %q", Redact(secret))
	}
	if Redact(plain) != "envchain" {
		t.Errorf("expected plain value, got %q", Redact(plain))
	}
}
