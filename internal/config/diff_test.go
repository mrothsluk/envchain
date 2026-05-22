package config

import (
	"testing"
)

func buildDiffChain(t *testing.T) *Chain {
	t.Helper()

	base := writeTemp(t, "base.yaml", `
env: base
vars:
  APP_NAME:
    value: myapp
  DB_URL:
    value: postgres://localhost/dev
  SECRET_KEY:
    value: base-secret
    secret: true
  BASE_ONLY:
    value: only-in-base
`)
	prod := writeTemp(t, "prod.yaml", `
env: prod
vars:
  DB_URL:
    value: postgres://prod-host/prod
  SECRET_KEY:
    value: prod-secret
    secret: true
  PROD_ONLY:
    value: only-in-prod
`)

	c := NewChain()
	if err := c.AddLayer(base, "base"); err != nil {
		t.Fatalf("AddLayer base: %v", err)
	}
	if err := c.AddLayer(prod, "prod"); err != nil {
		t.Fatalf("AddLayer prod: %v", err)
	}
	return c
}

func TestDiffChangedKeys(t *testing.T) {
	c := buildDiffChain(t)
	result, err := DiffChain(c, "base", "prod")
	if err != nil {
		t.Fatalf("DiffChain: %v", err)
	}

	changed := result.Changed()
	statuses := make(map[string]string, len(changed))
	for _, e := range changed {
		statuses[e.Key] = e.Status
	}

	if statuses["DB_URL"] != "changed" {
		t.Errorf("expected DB_URL to be changed, got %q", statuses["DB_URL"])
	}
	if statuses["BASE_ONLY"] != "removed" {
		t.Errorf("expected BASE_ONLY to be removed, got %q", statuses["BASE_ONLY"])
	}
	if statuses["PROD_ONLY"] != "added" {
		t.Errorf("expected PROD_ONLY to be added, got %q", statuses["PROD_ONLY"])
	}
}

func TestDiffUnchangedKeys(t *testing.T) {
	c := buildDiffChain(t)
	result, err := DiffChain(c, "base", "prod")
	if err != nil {
		t.Fatalf("DiffChain: %v", err)
	}

	for _, e := range result.Entries {
		if e.Key == "APP_NAME" && e.Status != "unchanged" {
			t.Errorf("expected APP_NAME unchanged, got %q", e.Status)
		}
	}
}

func TestDiffRedactsSecrets(t *testing.T) {
	c := buildDiffChain(t)
	result, err := DiffChain(c, "base", "prod")
	if err != nil {
		t.Fatalf("DiffChain: %v", err)
	}

	for _, e := range result.Entries {
		if e.Key == "SECRET_KEY" {
			if e.BaseVal != "********" || e.OtherVal != "********" {
				t.Errorf("secret not redacted: base=%q other=%q", e.BaseVal, e.OtherVal)
			}
			return
		}
	}
	t.Error("SECRET_KEY entry not found in diff")
}

func TestDiffInvalidEnvReturnsError(t *testing.T) {
	c := buildDiffChain(t)
	_, err := DiffChain(c, "base", "nonexistent")
	if err == nil {
		t.Error("expected error for unknown env, got nil")
	}
}
