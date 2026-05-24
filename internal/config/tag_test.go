package config

import (
	"testing"
)

func buildTagChain(t *testing.T) *Chain {
	t.Helper()
	base := writeTemp(t, "BASE_URL=https://example.com\nAPI_KEY=secret\n")
	prod := writeTemp(t, "BASE_URL=https://prod.example.com\n")
	c, err := NewChain("base", base)
	if err != nil {
		t.Fatalf("NewChain: %v", err)
	}
	if err := c.AddLayer("prod", prod); err != nil {
		t.Fatalf("AddLayer: %v", err)
	}
	return c
}

func TestTagLayerReturnsAllKeys(t *testing.T) {
	c := buildTagChain(t)
	idx, err := TagLayer(c, "base", []string{"infra", "public"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(idx["base"]) != 2 {
		t.Errorf("expected 2 keys, got %d", len(idx["base"]))
	}
	for _, tags := range idx["base"] {
		if len(tags) != 2 {
			t.Errorf("expected 2 tags per key, got %v", tags)
		}
	}
}

func TestTagLayerInvalidTagReturnsError(t *testing.T) {
	c := buildTagChain(t)
	_, err := TagLayer(c, "base", []string{"UPPER"})
	if err == nil {
		t.Fatal("expected error for uppercase tag")
	}
}

func TestTagLayerUnknownEnvReturnsError(t *testing.T) {
	c := buildTagChain(t)
	_, err := TagLayer(c, "staging", []string{"infra"})
	if err == nil {
		t.Fatal("expected error for unknown env")
	}
}

func TestQueryByTagFindsEntries(t *testing.T) {
	c := buildTagChain(t)
	idx, _ := TagLayer(c, "base", []string{"infra"})
	results := QueryByTag(idx, "infra")
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestQueryByTagMissingTagReturnsEmpty(t *testing.T) {
	c := buildTagChain(t)
	idx, _ := TagLayer(c, "base", []string{"infra"})
	results := QueryByTag(idx, "nonexistent")
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestMergeTagIndexesDeduplicate(t *testing.T) {
	c := buildTagChain(t)
	a, _ := TagLayer(c, "base", []string{"infra", "shared"})
	b, _ := TagLayer(c, "base", []string{"shared", "legacy"})
	merged := MergeTagIndexes(a, b)
	for _, tags := range merged["base"] {
		seen := make(map[string]int)
		for _, tag := range tags {
			seen[tag]++
			if seen[tag] > 1 {
				t.Errorf("duplicate tag %q after merge", tag)
			}
		}
	}
}

func TestMergeTagIndexesCombinesEnvs(t *testing.T) {
	c := buildTagChain(t)
	a, _ := TagLayer(c, "base", []string{"infra"})
	b, _ := TagLayer(c, "prod", []string{"prod-only"})
	merged := MergeTagIndexes(a, b)
	if _, ok := merged["base"]; !ok {
		t.Error("expected base env in merged index")
	}
	if _, ok := merged["prod"]; !ok {
		t.Error("expected prod env in merged index")
	}
}
