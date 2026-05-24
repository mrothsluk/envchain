package config

import (
	"testing"
)

func buildRenameChain(t *testing.T) *Chain {
	t.Helper()
	base := writeTemp(t, "BASE_URL=http://localhost\nAPI_KEY=secret\n")
	dev := writeTemp(t, "BASE_URL=http://dev.example.com\n")
	c, err := NewChain("dev", []string{base, dev})
	if err != nil {
		t.Fatalf("NewChain: %v", err)
	}
	return c
}

func TestRenameKeySuccess(t *testing.T) {
	c := buildRenameChain(t)
	res, err := RenameKey(c, "dev", "BASE_URL", "SERVICE_URL", RenameOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.OldKey != "BASE_URL" || res.NewKey != "SERVICE_URL" {
		t.Errorf("unexpected result: %+v", res)
	}
	if _, ok := c.layers["dev"].values["BASE_URL"]; ok {
		t.Error("old key should be removed")
	}
	if _, ok := c.layers["dev"].values["SERVICE_URL"]; !ok {
		t.Error("new key should exist")
	}
}

func TestRenameKeyDryRunDoesNotMutate(t *testing.T) {
	c := buildRenameChain(t)
	_, err := RenameKey(c, "dev", "BASE_URL", "SERVICE_URL", RenameOptions{DryRun: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := c.layers["dev"].values["BASE_URL"]; !ok {
		t.Error("dry run must not remove old key")
	}
	if _, ok := c.layers["dev"].values["SERVICE_URL"]; ok {
		t.Error("dry run must not create new key")
	}
}

func TestRenameKeyFailIfExists(t *testing.T) {
	c := buildRenameChain(t)
	// BASE_URL exists in base layer; inject into dev to create conflict
	c.layers["dev"].values["SERVICE_URL"] = "already"
	_, err := RenameKey(c, "dev", "BASE_URL", "SERVICE_URL", RenameOptions{FailIfExists: true})
	if err == nil {
		t.Fatal("expected error when destination key exists")
	}
}

func TestRenameKeyUnknownEnvReturnsError(t *testing.T) {
	c := buildRenameChain(t)
	_, err := RenameKey(c, "prod", "BASE_URL", "SERVICE_URL", RenameOptions{})
	if err == nil {
		t.Fatal("expected error for unknown env")
	}
}

func TestRenameKeyMissingOldKeyReturnsError(t *testing.T) {
	c := buildRenameChain(t)
	_, err := RenameKey(c, "dev", "NONEXISTENT", "NEW_KEY", RenameOptions{})
	if err == nil {
		t.Fatal("expected error for missing old key")
	}
}

func TestRenameKeyInvalidNewKeyReturnsError(t *testing.T) {
	c := buildRenameChain(t)
	_, err := RenameKey(c, "dev", "BASE_URL", "invalid-key", RenameOptions{})
	if err == nil {
		t.Fatal("expected error for invalid new key name")
	}
}

func TestRenameKeyPreservesSecretFlag(t *testing.T) {
	c := buildRenameChain(t)
	// Mark API_KEY as secret in base layer and rename it
	c.layers["base"].secrets["API_KEY"] = true
	_, err := RenameKey(c, "base", "API_KEY", "TOKEN", RenameOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !c.layers["base"].secrets["TOKEN"] {
		t.Error("secret flag should be transferred to new key")
	}
	if c.layers["base"].secrets["API_KEY"] {
		t.Error("secret flag should be removed from old key")
	}
}
