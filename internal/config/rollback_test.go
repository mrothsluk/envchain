package config

import (
	"testing"
)

func buildRollbackChain(t *testing.T) *Chain {
	t.Helper()
	base := writeTemp(t, "BASE_URL=https://base.example.com\nAPI_KEY=old-key\nDEBUG=false\n")
	dev := writeTemp(t, "API_KEY=dev-key\n")

	cfgBase, err := LoadConfig(base)
	if err != nil {
		t.Fatal(err)
	}
	cfgDev, err := LoadConfig(dev)
	if err != nil {
		t.Fatal(err)
	}

	chain := NewChain()
	if err := chain.AddLayer("base", cfgBase); err != nil {
		t.Fatal(err)
	}
	if err := chain.AddLayer("dev", cfgDev); err != nil {
		t.Fatal(err)
	}
	return chain
}

func TestRollbackRestoresSnapshot(t *testing.T) {
	chain := buildRollbackChain(t)
	snap := Snapshot{"API_KEY": "old-key", "BASE_URL": "https://base.example.com", "DEBUG": "false"}

	res, err := RollbackToSnapshot(chain, "base", snap, "overwrite", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Restored == 0 {
		t.Error("expected at least one key to be restored")
	}
}

func TestRollbackKeepExistingSkipsDifferentKeys(t *testing.T) {
	chain := buildRollbackChain(t)
	snap := Snapshot{"API_KEY": "snapshot-key"}

	res, err := RollbackToSnapshot(chain, "dev", snap, "keep-existing", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// dev layer already has API_KEY set to a different value — should be skipped
	if res.Skipped == 0 {
		t.Error("expected key to be skipped under keep-existing strategy")
	}
}

func TestRollbackErrorStrategyOnConflict(t *testing.T) {
	chain := buildRollbackChain(t)
	snap := Snapshot{"API_KEY": "different-value"}

	_, err := RollbackToSnapshot(chain, "dev", snap, "error", false)
	if err == nil {
		t.Fatal("expected error on conflicting key with error strategy")
	}
}

func TestRollbackUnknownEnvReturnsError(t *testing.T) {
	chain := buildRollbackChain(t)
	snap := Snapshot{"API_KEY": "v"}

	_, err := RollbackToSnapshot(chain, "unknown", snap, "overwrite", false)
	if err == nil {
		t.Fatal("expected error for unknown environment")
	}
}

func TestRollbackRemoveExtraDeletesUnsnapshotted(t *testing.T) {
	chain := buildRollbackChain(t)
	// Snapshot only contains BASE_URL — DEBUG should be removed with removeExtra=true
	snap := Snapshot{"BASE_URL": "https://base.example.com"}

	_, err := RollbackToSnapshot(chain, "base", snap, "overwrite", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	layer, _ := chain.Resolve("base")
	if _, ok := layer["DEBUG"]; ok {
		t.Error("expected DEBUG to be removed after rollback with removeExtra=true")
	}
	if _, ok := layer["BASE_URL"]; !ok {
		t.Error("expected BASE_URL to be present after rollback")
	}
}
