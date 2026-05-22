package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func buildSnapshotChain(t *testing.T) *Chain {
	t.Helper()
	base := writeTemp(t, "base.env", "APP_HOST=localhost\nAPP_SECRET=topsecret\nAPP_PORT=8080\n")
	prod := writeTemp(t, "prod.env", "APP_HOST=prod.example.com\nAPP_SECRET=prodsecret\n")
	c, err := NewChain(base, "dev")
	if err != nil {
		t.Fatalf("NewChain: %v", err)
	}
	if err := c.AddLayer(prod, "prod"); err != nil {
		t.Fatalf("AddLayer: %v", err)
	}
	return c
}

func TestTakeSnapshotBase(t *testing.T) {
	c := buildSnapshotChain(t)
	s, err := TakeSnapshot(c, "dev", false)
	if err != nil {
		t.Fatalf("TakeSnapshot: %v", err)
	}
	if s.Env != "dev" {
		t.Errorf("expected env=dev, got %q", s.Env)
	}
	if s.Values["APP_HOST"] != "localhost" {
		t.Errorf("expected APP_HOST=localhost, got %q", s.Values["APP_HOST"])
	}
	if s.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestTakeSnapshotProdOverrides(t *testing.T) {
	c := buildSnapshotChain(t)
	s, err := TakeSnapshot(c, "prod", false)
	if err != nil {
		t.Fatalf("TakeSnapshot: %v", err)
	}
	if s.Values["APP_HOST"] != "prod.example.com" {
		t.Errorf("expected prod override, got %q", s.Values["APP_HOST"])
	}
	if s.Values["APP_PORT"] != "8080" {
		t.Errorf("expected base APP_PORT, got %q", s.Values["APP_PORT"])
	}
}

func TestTakeSnapshotRedactsSecrets(t *testing.T) {
	c := buildSnapshotChain(t)
	s, err := TakeSnapshot(c, "dev", true)
	if err != nil {
		t.Fatalf("TakeSnapshot: %v", err)
	}
	if s.Values["APP_SECRET"] != "***" {
		t.Errorf("expected redacted secret, got %q", s.Values["APP_SECRET"])
	}
	if s.Values["APP_HOST"] != "localhost" {
		t.Errorf("expected unredacted host, got %q", s.Values["APP_HOST"])
	}
}

func TestTakeSnapshotUnknownEnvReturnsError(t *testing.T) {
	c := buildSnapshotChain(t)
	_, err := TakeSnapshot(c, "staging", false)
	if err == nil {
		t.Error("expected error for unknown env")
	}
}

func TestSaveAndLoadSnapshot(t *testing.T) {
	s := &Snapshot{
		Env:       "prod",
		Timestamp: time.Now().UTC().Truncate(time.Second),
		Values:    map[string]string{"APP_HOST": "prod.example.com", "APP_PORT": "443"},
	}
	path := filepath.Join(t.TempDir(), "snap.json")
	if err := SaveSnapshot(s, path); err != nil {
		t.Fatalf("SaveSnapshot: %v", err)
	}
	loaded, err := LoadSnapshot(path)
	if err != nil {
		t.Fatalf("LoadSnapshot: %v", err)
	}
	if loaded.Env != s.Env {
		t.Errorf("env mismatch: got %q", loaded.Env)
	}
	if loaded.Values["APP_HOST"] != s.Values["APP_HOST"] {
		t.Errorf("value mismatch: got %q", loaded.Values["APP_HOST"])
	}
}

func TestLoadSnapshotMissingFile(t *testing.T) {
	_, err := LoadSnapshot(filepath.Join(t.TempDir(), "missing.json"))
	if !os.IsNotExist(err) && err == nil {
		t.Error("expected error for missing file")
	}
}

func TestDiffSnapshot(t *testing.T) {
	before := &Snapshot{Values: map[string]string{"A": "1", "B": "2", "C": "3"}}
	after := &Snapshot{Values: map[string]string{"A": "1", "B": "changed", "D": "4"}}
	added, removed, changed := DiffSnapshot(before, after)
	if len(added) != 1 || added[0] != "D" {
		t.Errorf("expected added=[D], got %v", added)
	}
	if len(removed) != 1 || removed[0] != "C" {
		t.Errorf("expected removed=[C], got %v", removed)
	}
	if len(changed) != 1 || changed[0] != "B" {
		t.Errorf("expected changed=[B], got %v", changed)
	}
}
