package config

import (
	"context"
	"os"
	"testing"
	"time"
)

func writeTempWatch(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "envchain-watch-*.env")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestWatcherNoChangeEmitsNoEvent(t *testing.T) {
	path := writeTempWatch(t, "APP_ENV=dev\n")
	w, err := NewWatcher([]string{path}, 20*time.Millisecond)
	if err != nil {
		t.Fatalf("NewWatcher: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()
	go w.Run(ctx)
	<-ctx.Done()
	if len(w.Events) != 0 {
		t.Errorf("expected no events, got %d", len(w.Events))
	}
}

func TestWatcherDetectsChange(t *testing.T) {
	path := writeTempWatch(t, "APP_ENV=dev\n")
	w, err := NewWatcher([]string{path}, 20*time.Millisecond)
	if err != nil {
		t.Fatalf("NewWatcher: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go w.Run(ctx)
	time.Sleep(30 * time.Millisecond)
	if err := os.WriteFile(path, []byte("APP_ENV=prod\n"), 0o644); err != nil {
		t.Fatalf("update file: %v", err)
	}
	select {
	case ev := <-w.Events:
		if ev.Path != path {
			t.Errorf("path = %q, want %q", ev.Path, path)
		}
		if ev.OldHash == ev.NewHash {
			t.Error("old and new hash should differ")
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for change event")
	}
}

func TestNewWatcherMissingFileReturnsError(t *testing.T) {
	_, err := NewWatcher([]string{"/nonexistent/path.env"}, time.Second)
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestWatcherMultipleFiles(t *testing.T) {
	p1 := writeTempWatch(t, "A=1\n")
	p2 := writeTempWatch(t, "B=2\n")
	w, err := NewWatcher([]string{p1, p2}, 20*time.Millisecond)
	if err != nil {
		t.Fatalf("NewWatcher: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go w.Run(ctx)
	time.Sleep(30 * time.Millisecond)
	if err := os.WriteFile(p2, []byte("B=99\n"), 0o644); err != nil {
		t.Fatalf("update p2: %v", err)
	}
	select {
	case ev := <-w.Events:
		if ev.Path != p2 {
			t.Errorf("expected change on p2, got %q", ev.Path)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for event on p2")
	}
}
