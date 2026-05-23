package config

import (
	"sort"
	"testing"
)

func buildScopeChain(t *testing.T) *Chain {
	t.Helper()
	base := writeTemp(t, "BASE_URL=https://example.com\nDB_HOST=localhost\nDB_PORT=5432\nSECRET_KEY=topsecret\nAPP_NAME=myapp\n")
	prod := writeTemp(t, "BASE_URL=https://prod.example.com\nDB_HOST=prod-db\n")
	c, err := NewChain("base", base)
	if err != nil {
		t.Fatalf("NewChain: %v", err)
	}
	if err := c.AddLayer("prod", prod, "base"); err != nil {
		t.Fatalf("AddLayer: %v", err)
	}
	return c
}

func TestScopeLayerNoPrefixNoDeny(t *testing.T) {
	c := buildScopeChain(t)
	res, err := ScopeLayer(c, "base", ScopeFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Visible) != 5 {
		t.Errorf("expected 5 visible keys, got %d", len(res.Visible))
	}
	if len(res.Excluded) != 0 {
		t.Errorf("expected 0 excluded keys, got %d", len(res.Excluded))
	}
}

func TestScopeLayerAllowedPrefix(t *testing.T) {
	c := buildScopeChain(t)
	res, err := ScopeLayer(c, "base", ScopeFilter{AllowedPrefixes: []string{"DB_"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := res.Visible["DB_HOST"]; !ok {
		t.Error("expected DB_HOST in visible")
	}
	if _, ok := res.Visible["DB_PORT"]; !ok {
		t.Error("expected DB_PORT in visible")
	}
	if len(res.Visible) != 2 {
		t.Errorf("expected 2 visible keys, got %d", len(res.Visible))
	}
}

func TestScopeLayerDeniedKeys(t *testing.T) {
	c := buildScopeChain(t)
	res, err := ScopeLayer(c, "base", ScopeFilter{DeniedKeys: []string{"SECRET_KEY"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := res.Visible["SECRET_KEY"]; ok {
		t.Error("SECRET_KEY should be excluded")
	}
	sort.Strings(res.Excluded)
	if len(res.Excluded) != 1 || res.Excluded[0] != "SECRET_KEY" {
		t.Errorf("expected [SECRET_KEY] excluded, got %v", res.Excluded)
	}
}

func TestScopeLayerPrefixAndDeny(t *testing.T) {
	c := buildScopeChain(t)
	res, err := ScopeLayer(c, "prod", ScopeFilter{
		AllowedPrefixes: []string{"DB_", "BASE_"},
		DeniedKeys:      []string{"DB_PORT"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := res.Visible["DB_HOST"]; !ok {
		t.Error("expected DB_HOST visible")
	}
	if _, ok := res.Visible["DB_PORT"]; ok {
		t.Error("DB_PORT should be excluded")
	}
	if res.Visible["BASE_URL"] != "https://prod.example.com" {
		t.Errorf("unexpected BASE_URL: %s", res.Visible["BASE_URL"])
	}
}

func TestScopeLayerUnknownEnvReturnsError(t *testing.T) {
	c := buildScopeChain(t)
	_, err := ScopeLayer(c, "staging", ScopeFilter{})
	if err == nil {
		t.Fatal("expected error for unknown env")
	}
}
