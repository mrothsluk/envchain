package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func buildPinChain(t *testing.T) *Chain {
	t.Helper()
	base := map[string]string{"APP_HOST": "localhost", "APP_PORT": "8080"}
	prod := map[string]string{"APP_HOST": "prod.example.com"}
	ch := NewChain()
	if err := ch.AddLayer("base", base, nil); err != nil {
		t.Fatal(err)
	}
	if err := ch.AddLayer("prod", prod, nil); err != nil {
		t.Fatal(err)
	}
	return ch
}

func TestPinLayerRecordsValues(t *testing.T) {
	ch := buildPinChain(t)
	entry, err := PinLayer(ch, "prod", "alice", "release v1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Env != "prod" {
		t.Errorf("env = %q, want prod", entry.Env)
	}
	if entry.Values["APP_HOST"] != "prod.example.com" {
		t.Errorf("APP_HOST = %q, want prod.example.com", entry.Values["APP_HOST"])
	}
	if entry.PinnedBy != "alice" {
		t.Errorf("PinnedBy = %q, want alice", entry.PinnedBy)
	}
	if entry.Comment != "release v1" {
		t.Errorf("Comment = %q, want 'release v1'", entry.Comment)
	}
	if entry.PinnedAt.IsZero() {
		t.Error("PinnedAt should not be zero")
	}
}

func TestPinLayerUnknownEnvReturnsError(t *testing.T) {
	ch := buildPinChain(t)
	_, err := PinLayer(ch, "staging", "bob", "")
	if err == nil {
		t.Fatal("expected error for unknown env")
	}
}

func TestSaveAndLoadPin(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pins.json")
	entry := PinEntry{
		Env:      "prod",
		PinnedAt: time.Now().UTC(),
		PinnedBy: "ci",
		Values:   map[string]string{"APP_HOST": "prod.example.com"},
	}
	if err := SavePin(path, entry); err != nil {
		t.Fatalf("SavePin: %v", err)
	}
	pf, err := LoadPin(path)
	if err != nil {
		t.Fatalf("LoadPin: %v", err)
	}
	if len(pf.Entries) != 1 {
		t.Fatalf("entries count = %d, want 1", len(pf.Entries))
	}
	if pf.Entries[0].Values["APP_HOST"] != "prod.example.com" {
		t.Errorf("APP_HOST = %q", pf.Entries[0].Values["APP_HOST"])
	}
}

func TestLatestPinReturnsNewest(t *testing.T) {
	pf := PinFile{Entries: []PinEntry{
		{Env: "prod", PinnedBy: "first", Values: map[string]string{"APP_HOST": "old"}},
		{Env: "prod", PinnedBy: "second", Values: map[string]string{"APP_HOST": "new"}},
	}}
	e, err := LatestPin(pf, "prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.PinnedBy != "second" {
		t.Errorf("PinnedBy = %q, want second", e.PinnedBy)
	}
}

func TestLatestPinMissingEnvReturnsError(t *testing.T) {
	pf := PinFile{}
	_, err := LatestPin(pf, "prod")
	if err == nil {
		t.Fatal("expected error for missing env")
	}
}

func TestLoadPinMissingFileReturnsError(t *testing.T) {
	_, err := LoadPin(filepath.Join(t.TempDir(), "nope.json"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestSavePinAppendsEntries(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pins.json")
	for i := 0; i < 3; i++ {
		e := PinEntry{Env: "dev", PinnedBy: "user", Values: map[string]string{}}
		if err := SavePin(path, e); err != nil {
			t.Fatalf("SavePin iteration %d: %v", i, err)
		}
	}
	data, _ := os.ReadFile(path)
	pf, err := LoadPin(path)
	if err != nil {
		t.Fatalf("LoadPin: %v — raw: %s", err, data)
	}
	if len(pf.Entries) != 3 {
		t.Errorf("entries = %d, want 3", len(pf.Entries))
	}
}
