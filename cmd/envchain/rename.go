package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourorg/envchain/internal/config"
)

var renameCmd = &cobra.Command{
	Use:   "rename <env> <old-key> <new-key> [files...]",
	Short: "Rename an environment variable key within a layer",
	Args:  cobra.MinimumNArgs(3),
	RunE:  runRename,
}

var (
	renameFailIfExists bool
	renameDryRun       bool
)

func init() {
	renameCmd.Flags().BoolVar(&renameFailIfExists, "fail-if-exists", false, "error if the new key already exists")
	renameCmd.Flags().BoolVar(&renameDryRun, "dry-run", false, "print what would change without modifying")
	rootCmd.AddCommand(renameCmd)
}

func runRename(cmd *cobra.Command, args []string) error {
	env := args[0]
	oldKey := args[1]
	newKey := args[2]
	files := args[3:]

	if len(files) == 0 {
		return fmt.Errorf("at least one config file is required")
	}

	c, err := config.NewChain(env, files)
	if err != nil {
		return fmt.Errorf("loading chain: %w", err)
	}

	opts := config.RenameOptions{
		FailIfExists: renameFailIfExists,
		DryRun:       renameDryRun,
	}

	result, err := config.RenameKey(c, env, oldKey, newKey, opts)
	if err != nil {
		return err
	}

	if renameDryRun {
		fmt.Fprintf(os.Stdout, "[dry-run] would rename %s -> %s in env %q\n",
			result.OldKey, result.NewKey, result.Env)
	} else {
		fmt.Fprintf(os.Stdout, "renamed %s -> %s in env %q\n",
			result.OldKey, result.NewKey, result.Env)
	}
	return nil
}
