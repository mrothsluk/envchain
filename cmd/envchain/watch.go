package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/example/envchain/internal/config"
)

func init() {
	registerCommand("watch", "Poll config files and print change events", runWatch)
}

func runWatch(args []string) error {
	fs := flag.NewFlagSet("watch", flag.ContinueOnError)
	interval := fs.Duration("interval", 2*time.Second, "polling interval (e.g. 500ms, 2s)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	paths := fs.Args()
	if len(paths) == 0 {
		return fmt.Errorf("watch: at least one config file path required")
	}

	w, err := config.NewWatcher(paths, *interval)
	if err != nil {
		return fmt.Errorf("watch: %w", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	fmt.Fprintf(os.Stderr, "watching %d file(s) every %s — press Ctrl+C to stop\n", len(paths), *interval)

	go w.Run(ctx)

	for {
		select {
		case ev, ok := <-w.Events:
			if !ok {
				return nil
			}
			fmt.Printf("%s  changed  %s\n",
				ev.At.Format(time.RFC3339),
				ev.Path,
			)
			fmt.Printf("  old: %.12s...\n", ev.OldHash)
			fmt.Printf("  new: %.12s...\n", ev.NewHash)
		case <-ctx.Done():
			return nil
		}
	}
}
