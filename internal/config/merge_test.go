package config

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMergeNoConflicts(t *testing.T) {
	base := map[string]string{"APP_ENV": "dev", "PORT": "8080"}
	overlay := map[string]string{"LOG_LEVEL": "debug"}

	got, err := MergeLayers(base, overlay, MergeStrategyOverwrite)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := map[string]string{"APP_ENV": "dev", "PORT": "8080", "LOG_LEVEL": "debug"}
	if diff := cmp.Diff(want, got.Merged); diff != "" {
		t.Errorf("Merged mismatch (-want +got):\n%s", diff)
	}
	if len(got.Conflicts) != 0 {
		t.Errorf("expected no conflicts, got %v", got.Conflicts)
	}
}

func TestMergeOverwriteStrategy(t *testing.T) {
	base := map[string]string{"PORT": "8080", "APP_ENV": "dev"}
	overlay := map[string]string{"PORT": "9090"}

	got, err := MergeLayers(base, overlay, MergeStrategyOverwrite)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Merged["PORT"] != "9090" {
		t.Errorf("expected PORT=9090, got %q", got.Merged["PORT"])
	}
	if len(got.Conflicts) != 1 || got.Conflicts[0] != "PORT" {
		t.Errorf("expected conflicts=[PORT], got %v", got.Conflicts)
	}
}

func TestMergeKeepExistingStrategy(t *testing.T) {
	base := map[string]string{"PORT": "8080"}
	overlay := map[string]string{"PORT": "9090"}

	got, err := MergeLayers(base, overlay, MergeStrategyKeepExisting)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Merged["PORT"] != "8080" {
		t.Errorf("expected PORT=8080 (kept), got %q", got.Merged["PORT"])
	}
	if len(got.Conflicts) != 1 || got.Conflicts[0] != "PORT" {
		t.Errorf("expected conflicts=[PORT], got %v", got.Conflicts)
	}
}

func TestMergeErrorStrategy(t *testing.T) {
	base := map[string]string{"SECRET_KEY": "abc"}
	overlay := map[string]string{"SECRET_KEY": "xyz"}

	_, err := MergeLayers(base, overlay, MergeStrategyError)
	if err == nil {
		t.Fatal("expected error on conflict, got nil")
	}
}

func TestMergeIdenticalValuesNotConflict(t *testing.T) {
	base := map[string]string{"PORT": "8080"}
	overlay := map[string]string{"PORT": "8080"}

	got, err := MergeLayers(base, overlay, MergeStrategyError)
	if err != nil {
		t.Fatalf("identical values should not conflict, got error: %v", err)
	}
	if len(got.Conflicts) != 0 {
		t.Errorf("expected no conflicts for identical values, got %v", got.Conflicts)
	}
}

func TestMergeConflictsSorted(t *testing.T) {
	base := map[string]string{"Z_KEY": "1", "A_KEY": "1", "M_KEY": "1"}
	overlay := map[string]string{"Z_KEY": "2", "A_KEY": "2", "M_KEY": "2"}

	got, err := MergeLayers(base, overlay, MergeStrategyKeepExisting)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantConflicts := []string{"A_KEY", "M_KEY", "Z_KEY"}
	if diff := cmp.Diff(wantConflicts, got.Conflicts); diff != "" {
		t.Errorf("Conflicts order mismatch (-want +got):\n%s", diff)
	}
}
