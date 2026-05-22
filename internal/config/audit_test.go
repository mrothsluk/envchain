package config

import (
	"strings"
	"testing"
	"time"
)

var fixedTime = time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

func fixedClock() time.Time { return fixedTime }

func TestAuditLogRecord(t *testing.T) {
	log := NewAuditLog(fixedClock)
	log.Record("dev", "DATABASE_URL", "read", false)
	log.Record("prod", "API_SECRET", "read", true)

	entries := log.Entries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	if entries[0].Key != "DATABASE_URL" || entries[0].Action != "read" || entries[0].Redacted {
		t.Errorf("unexpected first entry: %+v", entries[0])
	}
	if entries[1].Key != "API_SECRET" || !entries[1].Redacted {
		t.Errorf("unexpected second entry: %+v", entries[1])
	}
}

func TestAuditLogEntriesIsCopy(t *testing.T) {
	log := NewAuditLog(fixedClock)
	log.Record("dev", "FOO", "set", false)

	e1 := log.Entries()
	e1[0].Key = "MUTATED"

	e2 := log.Entries()
	if e2[0].Key == "MUTATED" {
		t.Error("Entries() should return a copy, not a reference")
	}
}

func TestAuditLogSummaryEmpty(t *testing.T) {
	log := NewAuditLog(fixedClock)
	got := log.Summary()
	if got != "no audit events recorded" {
		t.Errorf("unexpected summary for empty log: %q", got)
	}
}

func TestAuditLogSummaryGroupsByKey(t *testing.T) {
	log := NewAuditLog(fixedClock)
	log.Record("base", "APP_PORT", "read", false)
	log.Record("dev", "APP_PORT", "read", false)
	log.Record("prod", "API_SECRET", "read", true)

	summary := log.Summary()
	lines := strings.Split(summary, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 summary lines, got %d: %v", len(lines), lines)
	}

	if !strings.HasPrefix(lines[0], "API_SECRET") || !strings.Contains(lines[0], "[secret]") {
		t.Errorf("expected API_SECRET line with [secret] tag, got: %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], "APP_PORT") || strings.Contains(lines[1], "[secret]") {
		t.Errorf("expected APP_PORT line without [secret] tag, got: %q", lines[1])
	}
	if !strings.Contains(lines[1], "read(base)") || !strings.Contains(lines[1], "read(dev)") {
		t.Errorf("expected both base and dev read actions for APP_PORT, got: %q", lines[1])
	}
}

func TestAuditLogTimestamp(t *testing.T) {
	log := NewAuditLog(fixedClock)
	log.Record("dev", "KEY", "set", false)

	entries := log.Entries()
	if !entries[0].Timestamp.Equal(fixedTime) {
		t.Errorf("expected timestamp %v, got %v", fixedTime, entries[0].Timestamp)
	}
}
