package config

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// AuditEntry records a single access or mutation event for an environment variable.
type AuditEntry struct {
	Timestamp time.Time
	Env       string
	Key       string
	Action    string // "read", "set", "delete"
	Redacted  bool
}

// AuditLog holds a list of audit entries for a chain resolution session.
type AuditLog struct {
	entries []AuditEntry
	clock   func() time.Time
}

// NewAuditLog creates an AuditLog. Pass nil to use the real clock.
func NewAuditLog(clock func() time.Time) *AuditLog {
	if clock == nil {
		clock = time.Now
	}
	return &AuditLog{clock: clock}
}

// Record appends a new entry to the log.
func (a *AuditLog) Record(env, key, action string, redacted bool) {
	a.entries = append(a.entries, AuditEntry{
		Timestamp: a.clock(),
		Env:       env,
		Key:       key,
		Action:    action,
		Redacted:  redacted,
	})
}

// Entries returns a copy of all recorded entries.
func (a *AuditLog) Entries() []AuditEntry {
	out := make([]AuditEntry, len(a.entries))
	copy(out, a.entries)
	return out
}

// Summary returns a human-readable summary grouped by key.
func (a *AuditLog) Summary() string {
	if len(a.entries) == 0 {
		return "no audit events recorded"
	}

	type stat struct {
		actions  []string
		redacted bool
	}
	byKey := make(map[string]*stat)
	for _, e := range a.entries {
		s, ok := byKey[e.Key]
		if !ok {
			s = &stat{}
			byKey[e.Key] = s
		}
		s.actions = append(s.actions, fmt.Sprintf("%s(%s)", e.Action, e.Env))
		if e.Redacted {
			s.redacted = true
		}
	}

	keys := make([]string, 0, len(byKey))
	for k := range byKey {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		s := byKey[k]
		redactedTag := ""
		if s.redacted {
			redactedTag = " [secret]"
		}
		fmt.Fprintf(&sb, "%s%s: %s\n", k, redactedTag, strings.Join(s.actions, ", "))
	}
	return strings.TrimRight(sb.String(), "\n")
}
