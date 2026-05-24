package config

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func buildPruneChain(t *testing.T) *Chain {
	t.Helper()
	base := mustLoad(t, "base", map[string]string{
		"APP_NAME": "envchain",
		"LOG_LEVEL": "",
		"SECRET_KEY": "hunter2",
		"DB_PASS": "s3cr3t",
	})
	base["SECRET_KEY"] = Entry{Value: "hunter2", Secret: true}
	base["DB_PASS"] = Entry{Value: "s3cr3t", Secret: true}

	c := NewChain()
	if err := c.AddLayer("base", base); err != nil {
		t.Fatal(err)
	}
	return c
}

func mustLoad(t *testing.T, env string, kv map[string]string) map[string]Entry {
	t.Helper()
	l := make(map[string]Entry, len(kv))
	for k, v := range kv {
		l[k] = Entry{Value: v}
	}
	return l
}

func TestPruneExplicitKeys(t *testing.T) {
	c := buildPruneChain(t)
	res, err := PruneLayer(c, "base", PruneOptions{RemoveKeys: []string{"APP_NAME"}})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]string{"APP_NAME"}, res.Removed); diff != "" {
		t.Fatalf("removed mismatch (-want +got):\n%s", diff)
	}
	if _, exists := c.layers["base"]["APP_NAME"]; exists {
		t.Error("expected APP_NAME to be deleted")
	}
}

func TestPruneBlankValues(t *testing.T) {
	c := buildPruneChain(t)
	res, err := PruneLayer(c, "base", PruneOptions{RemoveBlank: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Removed) != 1 || res.Removed[0] != "LOG_LEVEL" {
		t.Fatalf("expected only LOG_LEVEL removed, got %v", res.Removed)
	}
}

func TestPruneSecrets(t *testing.T) {
	c := buildPruneChain(t)
	res, err := PruneLayer(c, "base", PruneOptions{RemoveSecrets: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Removed) != 2 {
		t.Fatalf("expected 2 secrets removed, got %v", res.Removed)
	}
}

func TestPruneDryRunDoesNotMutate(t *testing.T) {
	c := buildPruneChain(t)
	before := len(c.layers["base"])
	res, err := PruneLayer(c, "base", PruneOptions{RemoveBlank: true, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if !res.DryRun {
		t.Error("expected DryRun=true in result")
	}
	if len(c.layers["base"]) != before {
		t.Error("dry run must not mutate the layer")
	}
}

func TestPruneUnknownEnvReturnsError(t *testing.T) {
	c := buildPruneChain(t)
	_, err := PruneLayer(c, "ghost", PruneOptions{RemoveBlank: true})
	if err == nil {
		t.Fatal("expected error for unknown env")
	}
}
