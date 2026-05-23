package config

import (
	"testing"
	"time"
)

func buildPinDiffChain(t *testing.T) *Chain {
	t.Helper()
	base := map[string]string{"APP_HOST": "localhost", "APP_PORT": "8080", "APP_DEBUG": "true"}
	ch := NewChain()
	if err := ch.AddLayer("dev", base, nil); err != nil {
		t.Fatal(err)
	}
	return ch
}

func TestDiffAgainstPinNoChanges(t *testing.T) {
	ch := buildPinDiffChain(t)
	entry := PinEntry{
		Env:      "dev",
		PinnedAt: time.Now().UTC(),
		PinnedBy: "ci",
		Values:   map[string]string{"APP_HOST": "localhost", "APP_PORT": "8080", "APP_DEBUG": "true"},
	}
	results, err := DiffAgainstPin(ch, entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected no diffs, got %d", len(results))
	}
}

func TestDiffAgainstPinDetectsChanged(t *testing.T) {
	ch := buildPinDiffChain(t)
	entry := PinEntry{
		Env:    "dev",
		Values: map[string]string{"APP_HOST": "old.host", "APP_PORT": "8080", "APP_DEBUG": "true"},
	}
	results, err := DiffAgainstPin(ch, entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, r := range results {
		if r.Key == "APP_HOST" && r.Status == "changed" {
			found = true
			if r.Pinned != "old.host" {
				t.Errorf("Pinned = %q, want old.host", r.Pinned)
			}
			if r.Live != "localhost" {
				t.Errorf("Live = %q, want localhost", r.Live)
			}
		}
	}
	if !found {
		t.Error("expected changed entry for APP_HOST")
	}
}

func TestDiffAgainstPinDetectsAdded(t *testing.T) {
	ch := buildPinDiffChain(t)
	entry := PinEntry{
		Env:    "dev",
		Values: map[string]string{"APP_HOST": "localhost", "APP_PORT": "8080"},
	}
	results, err := DiffAgainstPin(ch, entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, r := range results {
		if r.Key == "APP_DEBUG" && r.Status == "added" {
			found = true
		}
	}
	if !found {
		t.Error("expected added entry for APP_DEBUG")
	}
}

func TestDiffAgainstPinDetectsRemoved(t *testing.T) {
	ch := buildPinDiffChain(t)
	entry := PinEntry{
		Env: "dev",
		Values: map[string]string{
			"APP_HOST": "localhost", "APP_PORT": "8080",
			"APP_DEBUG": "true", "APP_EXTRA": "gone",
		},
	}
	results, err := DiffAgainstPin(ch, entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, r := range results {
		if r.Key == "APP_EXTRA" && r.Status == "removed" {
			found = true
		}
	}
	if !found {
		t.Error("expected removed entry for APP_EXTRA")
	}
}

func TestDiffAgainstPinUnknownEnvReturnsError(t *testing.T) {
	ch := buildPinDiffChain(t)
	entry := PinEntry{Env: "staging", Values: map[string]string{}}
	_, err := DiffAgainstPin(ch, entry)
	if err == nil {
		t.Fatal("expected error for unknown env")
	}
}
