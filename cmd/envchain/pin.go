package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/yourorg/envchain/internal/config"
)

var (
	pinFile    string
	pinBy      string
	pinComment string
	pinList    bool
	pinJSON    bool
)

func init() {
	pinCmd := &cobra.Command{
		Use:   "pin <env> [files...]",
		Short: "Pin resolved environment variables to a versioned record",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runPin,
	}
	pinCmd.Flags().StringVar(&pinFile, "pin-file", ".envchain-pins.json", "path to pin storage file")
	pinCmd.Flags().StringVar(&pinBy, "by", os.Getenv("USER"), "author of the pin")
	pinCmd.Flags().StringVar(&pinComment, "comment", "", "optional comment for this pin")
	pinCmd.Flags().BoolVar(&pinList, "list", false, "list pins instead of creating one")
	pinCmd.Flags().BoolVar(&pinJSON, "json", false, "output as JSON")
	rootCmd.AddCommand(pinCmd)
}

func runPin(cmd *cobra.Command, args []string) error {
	if pinList {
		return runPinList()
	}

	env := args[0]
	files := args[1:]
	if len(files) == 0 {
		return fmt.Errorf("at least one config file required when pinning")
	}

	chain, err := buildChainFromFiles(files)
	if err != nil {
		return err
	}

	entry, err := config.PinLayer(chain, env, pinBy, pinComment)
	if err != nil {
		return err
	}

	if err := config.SavePin(pinFile, entry); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Pinned %q at %s by %s\n", env, entry.PinnedAt.Format(time.RFC3339), entry.PinnedBy)
	return nil
}

func runPinList() error {
	pf, err := config.LoadPin(pinFile)
	if err != nil {
		return fmt.Errorf("load pin file: %w", err)
	}

	if pinJSON {
		return json.NewEncoder(os.Stdout).Encode(pf)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ENV\tPINNED AT\tPINNED BY\tCOMMENT")
	for _, e := range pf.Entries {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", e.Env, e.PinnedAt.Format(time.RFC3339), e.PinnedBy, e.Comment)
	}
	return w.Flush()
}
