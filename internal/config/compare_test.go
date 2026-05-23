package config

import (
	"testing"
)

func buildCompareChain(t *testing.T) *Chain {
	t.Helper()
	base := writeTemp(t, "base.env", `
DB_HOST=localhost
DB_PORT=5432
API_KEY=secret123
SHARED=same
`)
	prod := writeTemp(t, "prod.env", `
DB_HOST=prod.db.internal
DB_PASS=supersecret
SHARED=same
`)
	cfgBase, err := LoadConfig(base)
	if err != nil {
		t.Fatal(err)
	}
	cfgProd, err := LoadConfig(prod)
	if err != nil {
		t.Fatal(err)
	}
	chain := NewChain()
	if err := chain.AddLayer("dev", cfgBase); err != nil {
		t.Fatal(err)
	}
	if err := chain.AddLayer("prod", cfgProd); err != nil {
		t.Fatal(err)
	}
	return chain
}

func TestCompareOnlyInA(t *testing.T) {
	chain := buildCompareChain(t)
	res, err := CompareEnvs(chain, "dev", "prod", false)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := res.OnlyInA["DB_PORT"]; !ok {
		t.Error("expected DB_PORT only in dev")
	}
	if _, ok := res.OnlyInA["API_KEY"]; !ok {
		t.Error("expected API_KEY only in dev")
	}
}

func TestCompareOnlyInB(t *testing.T) {
	chain := buildCompareChain(t)
	res, err := CompareEnvs(chain, "dev", "prod", false)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := res.OnlyInB["DB_PASS"]; !ok {
		t.Error("expected DB_PASS only in prod")
	}
}

func TestCompareDifferentValues(t *testing.T) {
	chain := buildCompareChain(t)
	res, err := CompareEnvs(chain, "dev", "prod", false)
	if err != nil {
		t.Fatal(err)
	}
	pair, ok := res.Different["DB_HOST"]
	if !ok {
		t.Fatal("expected DB_HOST in Different")
	}
	if pair[0] != "localhost" || pair[1] != "prod.db.internal" {
		t.Errorf("unexpected pair: %v", pair)
	}
}

func TestCompareIdenticalKeys(t *testing.T) {
	chain := buildCompareChain(t)
	res, err := CompareEnvs(chain, "dev", "prod", false)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, k := range res.Identical {
		if k == "SHARED" {
			found = true
		}
	}
	if !found {
		t.Error("expected SHARED in Identical")
	}
}

func TestCompareUnknownEnvReturnsError(t *testing.T) {
	chain := buildCompareChain(t)
	_, err := CompareEnvs(chain, "dev", "staging", false)
	if err == nil {
		t.Error("expected error for unknown env")
	}
}

func TestCompareSortedHelpers(t *testing.T) {
	chain := buildCompareChain(t)
	res, err := CompareEnvs(chain, "dev", "prod", false)
	if err != nil {
		t.Fatal(err)
	}
	keys := res.SortedOnlyInA()
	for i := 1; i < len(keys); i++ {
		if keys[i] < keys[i-1] {
			t.Error("SortedOnlyInA not sorted")
		}
	}
}
