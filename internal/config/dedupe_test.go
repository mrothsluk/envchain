package config

import (
	"testing"
)

func buildDedupeChain(t *testing.T) *Chain {
	t.Helper()
	base := map[string]string{
		"DATABASE_URL": "postgres://localhost/dev",
		"API_KEY":      "dev-key",
		"DEBUG":        "true",
	}
	chain, err := NewChain(base)
	if err != nil {
		t.Fatalf("NewChain: %v", err)
	}
	if err := chain.AddLayer("dev", base); err != nil {
		t.Fatalf("AddLayer dev: %v", err)
	}
	return chain
}

func TestDedupeKeepFirstStrategy(t *testing.T) {
	chain := buildDedupeChain(t)
	rawPairs := map[string][]string{
		"DATABASE_URL": {"postgres://first/db", "postgres://second/db", "postgres://third/db"},
	}
	results, err := DedupeLayer(chain, "dev", rawPairs, DedupeKeepFirst)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Kept != "postgres://first/db" {
		t.Errorf("expected first value kept, got %q", results[0].Kept)
	}
	if len(results[0].Dropped) != 2 {
		t.Errorf("expected 2 dropped, got %d", len(results[0].Dropped))
	}
}

func TestDedupeKeepLastStrategy(t *testing.T) {
	chain := buildDedupeChain(t)
	rawPairs := map[string][]string{
		"API_KEY": {"key-a", "key-b", "key-c"},
	}
	results, err := DedupeLayer(chain, "dev", rawPairs, DedupeKeepLast)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results[0].Kept != "key-c" {
		t.Errorf("expected last value kept, got %q", results[0].Kept)
	}
}

func TestDedupeNoDuplicatesReturnsEmpty(t *testing.T) {
	chain := buildDedupeChain(t)
	rawPairs := map[string][]string{
		"DEBUG": {"true"},
	}
	results, err := DedupeLayer(chain, "dev", rawPairs, DedupeKeepFirst)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected no results for single-value keys, got %d", len(results))
	}
}

func TestDedupeUnknownEnvReturnsError(t *testing.T) {
	chain := buildDedupeChain(t)
	_, err := DedupeLayer(chain, "staging", map[string][]string{}, DedupeKeepFirst)
	if err == nil {
		t.Fatal("expected error for unknown env, got nil")
	}
}

func TestDedupeUnknownStrategyReturnsError(t *testing.T) {
	chain := buildDedupeChain(t)
	_, err := DedupeLayer(chain, "dev", map[string][]string{}, DedupeStrategy("random"))
	if err == nil {
		t.Fatal("expected error for unknown strategy, got nil")
	}
}

func TestDedupeResultStrategyField(t *testing.T) {
	chain := buildDedupeChain(t)
	rawPairs := map[string][]string{
		"DATABASE_URL": {"url-1", "url-2"},
	}
	results, err := DedupeLayer(chain, "dev", rawPairs, DedupeKeepLast)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results[0].Strategy != string(DedupeKeepLast) {
		t.Errorf("expected strategy %q, got %q", DedupeKeepLast, results[0].Strategy)
	}
}
