package config

import (
	"testing"
)

func buildPromoteChain() *Chain {
	base := NewLayer("base", map[string]string{
		"APP_HOST": "localhost",
		"APP_PORT": "8080",
		"DB_PASS":  "secret",
	})
	dev := NewLayer("dev", map[string]string{
		"APP_PORT": "9090",
	})
	prod := NewLayer("prod", map[string]string{
		"APP_HOST": "prod.example.com",
	})
	c := &Chain{}
	c.layers = []Layer{base, dev, prod}
	return c
}

func TestPromoteOverwriteCopiesAllKeys(t *testing.T) {
	c := buildPromoteChain()
	res, err := PromoteLayer(c, "dev", "prod", PromoteOverwrite)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Promoted) == 0 {
		t.Fatal("expected promoted keys, got none")
	}
	if len(res.Conflicts) != 0 {
		t.Fatalf("expected no conflicts, got %v", res.Conflicts)
	}
}

func TestPromoteKeepExistingSkipsConflicts(t *testing.T) {
	c := buildPromoteChain()
	res, err := PromoteLayer(c, "dev", "prod", PromoteKeepExisting)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// APP_HOST differs between dev-resolved and prod layer — should be skipped
	for _, k := range res.Skipped {
		if k == "APP_HOST" {
			return
		}
	}
	t.Error("expected APP_HOST to be skipped")
}

func TestPromoteErrorOnConflictReturnsError(t *testing.T) {
	c := buildPromoteChain()
	_, err := PromoteLayer(c, "dev", "prod", PromoteErrorOnConflict)
	if err == nil {
		t.Fatal("expected error on conflict, got nil")
	}
}

func TestPromoteUnknownSrcEnvReturnsError(t *testing.T) {
	c := buildPromoteChain()
	_, err := PromoteLayer(c, "unknown", "prod", PromoteOverwrite)
	if err == nil {
		t.Fatal("expected error for unknown source env")
	}
}

func TestPromoteUnknownDstLayerReturnsError(t *testing.T) {
	c := buildPromoteChain()
	_, err := PromoteLayer(c, "dev", "staging", PromoteOverwrite)
	if err == nil {
		t.Fatal("expected error for unknown destination layer")
	}
}
