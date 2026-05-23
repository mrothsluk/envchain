package config

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"time"
)

// WatchEvent describes a change detected in a watched config file.
type WatchEvent struct {
	Path    string
	OldHash string
	NewHash string
	At      time.Time
}

// Watcher polls a set of config file paths and emits events when their
// content changes.
type Watcher struct {
	paths    []string
	hashes   map[string]string
	interval time.Duration
	Events   chan WatchEvent
}

// NewWatcher creates a Watcher that polls the given paths every interval.
func NewWatcher(paths []string, interval time.Duration) (*Watcher, error) {
	w := &Watcher{
		paths:    paths,
		hashes:   make(map[string]string, len(paths)),
		interval: interval,
		Events:   make(chan WatchEvent, 8),
	}
	for _, p := range paths {
		h, err := hashFile(p)
		if err != nil {
			return nil, fmt.Errorf("watch: initial hash %q: %w", p, err)
		}
		w.hashes[p] = h
	}
	return w, nil
}

// Run starts the polling loop and blocks until ctx is cancelled.
func (w *Watcher) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	defer close(w.Events)
	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			w.poll(t)
		}
	}
}

func (w *Watcher) poll(at time.Time) {
	for _, p := range w.paths {
		newHash, err := hashFile(p)
		if err != nil {
			continue
		}
		if old := w.hashes[p]; old != newHash {
			w.hashes[p] = newHash
			w.Events <- WatchEvent{Path: p, OldHash: old, NewHash: newHash, At: at}
		}
	}
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
